package mcp

import (
	"context"
	stderrors "errors"
	"fmt"
	"strings"

	mcpkit "github.com/FacileStudio/tronc/mcp"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

const waitingReveal = "Nothing was revealed: your client cannot ask the user to confirm. Give the user confirm_url " +
	"and ask them to confirm there. Once they have, call casier_reveal_secrets with confirmation_id within 15 " +
	"minutes to read the values, once. Do not ask again for the same keys; the request stays open for 7 days."

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
func (t *toolset) handOver(ctx context.Context, c caller, id string) (*sdk.CallToolResult, revealOutput, error) {
	action, err := collect(ctx, t.Actions, c.who.UserID, c.who.ClientID, id)
	if err != nil {
		return nil, revealOutput{}, mcpkit.Failure(err, "")
	}
	g, err := targetOf(action)
	if err != nil {
		return nil, revealOutput{}, mcpkit.Failure(err, "")
	}
	digest, err := t.fingerprint(ctx, action)
	if err == nil && digest != action.Digest {
		err = mcpkit.ErrChanged
	}
	if err != nil {
		return nil, revealOutput{}, mcpkit.Failure(err, hintSecret)
	}
	out := revealOutput{Status: statusRevealed, Project: g.Project, Environment: g.Environment, Secrets: make([]revealedOut, 0, len(g.Keys))}
	for _, key := range g.Keys {
		secret, err := t.Secrets.Calls().Reveal(c.ctx, g.Project, g.Environment, key)
		if err != nil {
			return nil, revealOutput{}, mcpkit.Failure(err, hintSecret)
		}
		out.Secrets = append(out.Secrets, revealedOut{Key: secret.Key, Version: secret.Version, Value: *secret.Value})
	}
	out.Message = "The user confirmed. Use these values only for what they asked; do not repeat them in chat, files or logs."
	return nil, out, nil
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
