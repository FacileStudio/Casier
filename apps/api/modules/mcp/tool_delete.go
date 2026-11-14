package mcp

import (
	"context"
	"fmt"

	mcpkit "github.com/FacileStudio/tronc/mcp"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// deleteSecret is casier_delete_secret: the secret moves to the trash once the user
// confirms.
func (t *toolset) deleteSecret(ctx context.Context, req *sdk.CallToolRequest, in secretRef) (*sdk.CallToolResult, deleteOutput, error) {
	if mcpkit.Retrying(req) {
		return t.resumeDelete(ctx, req)
	}
	if _, err := t.callerOf(ctx, req); err != nil {
		return nil, deleteOutput{}, err
	}
	g := target{Project: in.Project, Environment: in.Environment, Keys: []string{in.Key}}
	result, outcome, err := t.confirm.Request(ctx, req, kindDelete, g.encode(), deleteQuestion(g))
	if err != nil {
		return nil, deleteOutput{}, mcpkit.Failure(err, hintSecret)
	}
	if result != nil {
		return result, deleteOutput{}, nil
	}
	return nil, deleteOutput{Status: outcome.Status, Key: in.Key, ConfirmURL: outcome.ConfirmURL, Message: outcome.Message}, nil
}

// resumeDelete handles the retry that carries the person's answer.
func (t *toolset) resumeDelete(ctx context.Context, req *sdk.CallToolRequest) (*sdk.CallToolResult, deleteOutput, error) {
	if _, err := t.callerOf(ctx, req); err != nil {
		return nil, deleteOutput{}, err
	}
	outcome, err := t.confirm.Resume(ctx, req, kindDelete)
	if err != nil {
		return nil, deleteOutput{}, mcpkit.Failure(err, "Call casier_delete_secret again to ask the user afresh.")
	}
	if outcome.Status == mcpkit.StatusDeclined {
		return nil, deleteOutput{Status: mcpkit.StatusDeclined, Message: outcome.Message}, nil
	}
	g, _ := targetOf(outcome.Action)
	key := ""
	if len(g.Keys) > 0 {
		key = g.Keys[0]
	}
	return nil, deleteOutput{Status: statusTrashed, Key: key, Message: "The user confirmed. The secret is in the trash; casier_restore_secret brings it back within 30 days."}, nil
}

// deleteQuestion is what the person is asked before a secret goes to the trash.
func deleteQuestion(g target) string {
	return fmt.Sprintf("Delete the secret %s from %s / %s? It moves to the trash and can be restored for 30 days.",
		g.Keys[0], g.Project, g.Environment)
}
