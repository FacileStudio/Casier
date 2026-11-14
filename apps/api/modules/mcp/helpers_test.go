package mcp

import (
	"net/http"
	"testing"

	mcpkit "github.com/FacileStudio/tronc/mcp"
	"github.com/FacileStudio/tronc/mcp/mcptest"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// count runs a count query.
func (f *fixture) count(t testing.TB, query string, args ...any) int {
	t.Helper()
	var n int
	if err := f.sqlDB.QueryRow(query, args...).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

// call runs a tool and decodes its result into out, failing on an error result.
func call(t *testing.T, session *sdk.ClientSession, tool string, args map[string]any, out any) {
	t.Helper()
	mcptest.Decode(t, mcptest.Call(t, session, tool, args), out)
}

// refused runs a tool and returns its error text, failing when it succeeded.
func refused(t *testing.T, session *sdk.ClientSession, tool string, args map[string]any) string {
	t.Helper()
	result := mcptest.Call(t, session, tool, args)
	if !result.IsError {
		t.Fatalf("%s must fail, got %s", tool, mcptest.Text(result))
	}
	return mcptest.Text(result)
}

// prod names the owner's prod environment.
func prod(extra map[string]any) map[string]any {
	args := map[string]any{"project": projectSlug, "environment": envSlug}
	for k, v := range extra {
		args[k] = v
	}
	return args
}

// confirmedReveal asks to reveal keys from a client that cannot confirm, confirms it on
// the page, and returns the confirmation id.
func confirmedReveal(t *testing.T, f *fixture, keys ...string) string {
	t.Helper()
	var asked revealOutput
	call(t, f.session(t, ownerSubject), toolReveal, prod(map[string]any{"keys": keys}), &asked)
	if got := onPage(f, http.MethodPost, "/"+asked.ConfirmationID+"/confirm", f.owner, ""); got.Code != http.StatusOK {
		t.Fatalf("confirm: %d %s", got.Code, got.Body)
	}
	return asked.ConfirmationID
}

// outcomeOf is the stored outcome of one action.
func (f *fixture) outcomeOf(t *testing.T, id string) string {
	t.Helper()
	var outcome string
	if err := f.sqlDB.QueryRow(`SELECT outcome FROM mcp_actions WHERE id = $1`, id).Scan(&outcome); err != nil {
		t.Fatal(err)
	}
	return outcome
}

// askedBy is the person and the client an action was asked by.
func (f *fixture) askedBy(t *testing.T, id string) mcpkit.Caller {
	t.Helper()
	var who mcpkit.Caller
	if err := f.sqlDB.QueryRow(`SELECT user_id, client_id FROM mcp_actions WHERE id = $1`, id).Scan(&who.UserID, &who.ClientID); err != nil {
		t.Fatal(err)
	}
	return who
}
