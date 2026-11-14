package mcp

import (
	"strings"
	"testing"

	mcpkit "github.com/FacileStudio/tronc/mcp"
	"github.com/FacileStudio/tronc/mcp/mcptest"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// TestAnElicitingClientIsStillSentToThePage checks a client that can ask the person is
// never asked: it gets the confirm link, and nothing is confirmed.
func TestAnElicitingClientIsStillSentToThePage(t *testing.T) {
	f := newFixture(t)
	asked := 0
	session := mcptest.Connect(t, f.resource, f.token(t, ownerSubject), mcptest.ClientOptions{Elicit: mcptest.Answering("accept", &asked)})
	var out revealOutput
	call(t, session, toolReveal, prod(map[string]any{"keys": []string{"DATABASE_URL"}}), &out)
	if asked != 0 || out.Status != mcpkit.StatusPending || !strings.HasSuffix(out.ConfirmURL, "action="+out.ConfirmationID) || len(out.Secrets) != 0 {
		t.Fatalf("want a link and no question, got %d questions and %+v", asked, out)
	}
	if !strings.Contains(out.Message, "until ") || f.outcomeOf(t, out.ConfirmationID) != "" {
		t.Fatalf("want an open action and its expiry, got %q and %q", out.Message, f.outcomeOf(t, out.ConfirmationID))
	}
}

// TestAnInClientAnswerNeverReveals checks an accept sent with the retry, as a client
// could forge it, confirms nothing and hands nothing over. It retries with the signed
// state when the server issued one, and with the action id otherwise.
func TestAnInClientAnswerNeverReveals(t *testing.T) {
	f := newFixture(t)
	asked := 0
	session := mcptest.Connect(t, f.resource, f.token(t, ownerSubject), mcptest.ClientOptions{Elicit: mcptest.Answering("accept", &asked), Manual: true})
	args := prod(map[string]any{"keys": []string{"DATABASE_URL"}})
	question := mcptest.Call(t, session, toolReveal, args)
	var out revealOutput
	if !question.NeedsInput() {
		mcptest.Decode(t, question, &out)
	}
	state := question.RequestState
	if state == "" {
		state = out.ConfirmationID
	}
	accept := sdk.InputResponseMap{mcpkit.ConfirmInput: &sdk.ElicitResult{Action: "accept", Content: map[string]any{mcpkit.ConfirmInput: true}}}
	forged, err := session.CallTool(t.Context(), &sdk.CallToolParams{Name: toolReveal, Arguments: args, RequestState: state, InputResponses: accept})
	if err != nil || !forged.IsError || strings.Contains(mcptest.Text(forged), "plaintext") {
		t.Fatalf("an in-client answer must be refused, got %v and %+v", err, forged)
	}
	if n := f.count(t, `SELECT count(*) FROM mcp_actions WHERE outcome <> ''`); n != 0 {
		t.Fatalf("an in-client answer must leave every action open, got %d resolved", n)
	}
}

// TestATokenWithoutAClientCannotReveal checks a token naming no OAuth client is
// refused before any action opens, so clients without an id never share reveals.
func TestATokenWithoutAClientCannotReveal(t *testing.T) {
	f := newFixture(t)
	token := f.issuer.Mint(t, map[string]any{"aud": []string{f.resource}, "sub": ownerSubject, "scope": "casier.read casier.write", "client_id": nil})
	text := refused(t, mcptest.Connect(t, f.resource, token, mcptest.ClientOptions{}), toolReveal, prod(map[string]any{"keys": []string{"DATABASE_URL"}}))
	if !strings.Contains(text, "names no OAuth client") || f.count(t, `SELECT count(*) FROM mcp_actions`) != 0 {
		t.Fatalf("want a refusal and no action, got %q", text)
	}
}
