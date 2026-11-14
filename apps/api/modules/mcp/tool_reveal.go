package mcp

import (
	"context"
	stderrors "errors"
	"fmt"
	"strings"
	"time"

	"github.com/FacileStudio/tronc/errors"
	mcpkit "github.com/FacileStudio/tronc/mcp"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

const collectAfterConfirm = " Once they have confirmed, call casier_reveal_secrets with only confirmation_id, within " +
	"15 minutes, to read the values once."

const revealedMessage = "The user confirmed. Use these values only for what they asked; do not repeat them in chat, files or logs."

const omittedMessage = "The keys in omitted did not fit in one answer and were not sent. To read them, call " +
	"casier_reveal_secrets again with only those keys; the user confirms again."

var errKeys = stderrors.New("pass keys, 1 to 50 secret keys of the environment, or confirmation_id")

var errNoClient = errors.Forbidden("this token names no OAuth client, so Casier cannot bind a reveal to it. " +
	"Sign in again with an MCP client that registers itself")

var errInClient = errors.Failed("Casier sends secret values only after the user confirms on the confirm page " +
	"in their browser, never through an answer in the client. Call casier_reveal_secrets again with keys to get confirm_url")

// reveal is casier_reveal_secrets: the values of named keys, once the user confirmed
// reading them on the confirm page in their own browser. A client that can elicit is
// still sent to the page, and an in-client answer is refused: a client could forge it.
func (t *toolset) reveal(ctx context.Context, req *sdk.CallToolRequest, in revealInput) (*sdk.CallToolResult, revealOutput, error) {
	if mcpkit.Retrying(req) {
		return nil, revealOutput{}, errInClient
	}
	c, err := t.callerOf(ctx, req)
	if err != nil {
		return nil, revealOutput{}, err
	}
	if strings.TrimSpace(c.who.ClientID) == "" {
		return nil, revealOutput{}, errNoClient
	}
	if id := strings.TrimSpace(in.ConfirmationID); id != "" {
		return t.handOver(ctx, c, id)
	}
	keys, err := keysOf(in.Keys)
	if err != nil {
		return nil, revealOutput{}, err
	}
	g := target{Project: in.Project, Environment: in.Environment, Keys: keys, Client: c.who.ClientID}
	_, outcome, err := t.confirm.Request(ctx, req, kindReveal, g.encode(), revealQuestion(g))
	if err != nil {
		return nil, revealOutput{}, mcpkit.Failure(err, hintSecret)
	}
	return nil, revealOutput{
		Status: mcpkit.StatusPending, Project: g.Project, Environment: g.Environment,
		ConfirmURL: t.confirm.Link(outcome.Action.ID), ConfirmationID: outcome.Action.ID,
		Message: onPageOnly(outcome.Action.ExpiresAt) + collectAfterConfirm,
	}, nil
}

// onPageOnly is the text for the model while a reveal waits on the confirm page.
func onPageOnly(expires time.Time) string {
	return "Nothing happened yet: Casier sends secret values only after the user confirms in their browser. " +
		"Give the user confirm_url and ask them to confirm there. Do not call this tool again for the same keys; " +
		"the request stays open until " + expires.UTC().Format(time.RFC3339) + "."
}

// handOver collects a confirmed reveal for the client that asked and reads its values
// through the reveal route, one audited read per key. A secret written since the
// confirmation is refused rather than handed over in a version the user did not see.
// The confirmation is spent only once the values were read.
func (t *toolset) handOver(ctx context.Context, c caller, id string) (*sdk.CallToolResult, revealOutput, error) {
	var out revealOutput
	err := collect(ctx, t.Actions, c.who, id, func(action mcpkit.Action) error {
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
