package mcp

import (
	"context"
	stderrors "errors"
	"fmt"
	"strings"

	"github.com/FacileStudio/tronc/errors"
	mcpkit "github.com/FacileStudio/tronc/mcp"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

const waitingReveal = "Nothing was revealed: your client cannot ask the user to confirm. Give the user confirm_url " +
	"and ask them to confirm there. Once they have, call casier_reveal_secrets with confirmation_id within 15 " +
	"minutes to read the values, once. Do not ask again for the same keys; the request stays open for 7 days."

const revealedMessage = "The user confirmed. Use these values only for what they asked; do not repeat them in chat, files or logs."

const omittedMessage = "The keys in omitted did not fit in one answer and were not sent. To read them, call " +
	"casier_reveal_secrets again with only those keys; the user confirms again."

var errKeys = stderrors.New("pass keys, 1 to 50 secret keys of the environment, or confirmation_id")

// reveal is casier_reveal_secrets: the values of named keys, once the user confirmed
// reading them.
func (t *toolset) reveal(ctx context.Context, req *sdk.CallToolRequest, in revealInput) (*sdk.CallToolResult, revealOutput, error) {
	if mcpkit.Retrying(req) {
		return t.resumeReveal(ctx, req)
	}
	c, err := t.callerOf(ctx, req)
	if err != nil {
		return nil, revealOutput{}, err
	}
	if id := strings.TrimSpace(in.ConfirmationID); id != "" {
		return t.handOver(ctx, c, id)
	}
	keys, err := keysOf(in.Keys)
	if err != nil {
		return nil, revealOutput{}, err
	}
	g := target{Project: in.Project, Environment: in.Environment, Keys: keys}
	result, outcome, err := t.confirm.Request(ctx, req, kindReveal, g.encode(), revealQuestion(g))
	if err != nil {
		return nil, revealOutput{}, mcpkit.Failure(err, hintSecret)
	}
	if result != nil {
		return result, revealOutput{}, nil
	}
	return nil, revealOutput{
		Status: outcome.Status, Project: g.Project, Environment: g.Environment, ConfirmURL: outcome.ConfirmURL,
		ConfirmationID: outcome.Action.ID, Message: waitingReveal,
	}, nil
}

// resumeReveal handles the retry that carries the person's answer.
func (t *toolset) resumeReveal(ctx context.Context, req *sdk.CallToolRequest) (*sdk.CallToolResult, revealOutput, error) {
	c, err := t.callerOf(ctx, req)
	if err != nil {
		return nil, revealOutput{}, err
	}
	outcome, err := t.confirm.Resume(ctx, req, kindReveal)
	if err != nil {
		return nil, revealOutput{}, mcpkit.Failure(err, "Call casier_reveal_secrets again to ask the user afresh.")
	}
	if outcome.Status == mcpkit.StatusDeclined {
		return nil, revealOutput{Status: mcpkit.StatusDeclined, Message: "The user declined. Nothing was revealed."}, nil
	}
	return t.handOver(ctx, c, outcome.Action.ID)
}

// handOver collects a confirmed reveal for the client that asked and reads its values
// through the reveal route, one audited read per key. A secret written since the
// confirmation is refused rather than handed over in a version the user did not see.
// The confirmation is spent only once the values were read.
func (t *toolset) handOver(ctx context.Context, c caller, id string) (*sdk.CallToolResult, revealOutput, error) {
	var out revealOutput
	err := collect(ctx, t.Actions, c.who.UserID, c.who.ClientID, id, func(action mcpkit.Action) error {
		var err error
		out, err = t.readConfirmed(ctx, c, action)
		return err
	})
	if err != nil {
		return nil, revealOutput{}, mcpkit.Failure(err, hintSecret)
	}
	return nil, out, nil
}

// readConfirmed reads the values of a confirmed reveal, in the order asked, and stops
// before the first one that would push the answer past the result budget. That value
// is read, and audited, but not handed over. When not even the first value fits,
// nothing is handed over and the confirmation stays open.
func (t *toolset) readConfirmed(ctx context.Context, c caller, action mcpkit.Action) (revealOutput, error) {
	g, err := targetOf(action)
	if err != nil {
		return revealOutput{}, err
	}
	digest, err := t.fingerprint(ctx, action)
	if err == nil && digest != action.Digest {
		err = mcpkit.ErrChanged
	}
	if err != nil {
		return revealOutput{}, err
	}
	read := make([]revealedOut, 0, len(g.Keys))
	for _, key := range g.Keys {
		secret, err := t.Secrets.Calls().Reveal(c.ctx, g.Project, g.Environment, key)
		if err != nil {
			return revealOutput{}, err
		}
		next := append(read, revealedOut{Key: secret.Key, Version: secret.Version, Value: *secret.Value})
		if mcpkit.EncodedSize(handed(g, next)) > mcpkit.ResultBudget {
			break
		}
		read = next
	}
	if len(read) == 0 {
		return revealOutput{}, errors.Failed(fmt.Sprintf("the value of %s is too large to send through MCP. "+
			"Read it with the casier CLI or in the Casier web app", g.Keys[0]))
	}
	return handed(g, read), nil
}

// handed is the answer carrying the values read so far; the keys after them are
// named as omitted.
func handed(g target, read []revealedOut) revealOutput {
	out := revealOutput{Status: statusRevealed, Project: g.Project, Environment: g.Environment, Secrets: read, Message: revealedMessage}
	if rest := g.Keys[len(read):]; len(rest) > 0 {
		out.Omitted = rest
		out.Message = revealedMessage + " " + omittedMessage
	}
	return out
}

// keysOf trims and deduplicates the requested keys, in the order given.
func keysOf(raw []string) ([]string, error) {
	keys := make([]string, 0, len(raw))
	seen := map[string]bool{}
	for _, key := range raw {
		key = strings.TrimSpace(key)
		if key != "" && !seen[key] {
			seen[key] = true
			keys = append(keys, key)
		}
	}
	if len(keys) == 0 || len(keys) > maxRevealedKeys {
		return nil, errKeys
	}
	return keys, nil
}

// revealQuestion is what the person is asked before values leave Casier.
func revealQuestion(g target) string {
	noun := "the value"
	if len(g.Keys) > 1 {
		noun = fmt.Sprintf("the %d values", len(g.Keys))
	}
	return fmt.Sprintf("Send %s of %s from %s / %s to the AI client that asked? It will be able to read and use them.",
		noun, strings.Join(g.Keys, ", "), g.Project, g.Environment)
}
