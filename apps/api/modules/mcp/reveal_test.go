package mcp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/FacileStudio/Casier/apps/api/modules/auth"
	"github.com/FacileStudio/Casier/apps/api/modules/secrets"
	"github.com/FacileStudio/porte"
	mcpkit "github.com/FacileStudio/tronc/mcp"
	"github.com/FacileStudio/tronc/mcp/mcptest"
	"github.com/go-chi/chi/v5"
)

// askReveal asks to reveal DATABASE_URL from a client that cannot confirm and returns
// the answer.
func askReveal(t *testing.T, f *fixture) revealOutput {
	t.Helper()
	var out revealOutput
	call(t, f.session(t, ownerSubject), toolReveal, prod(map[string]any{"keys": []string{"DATABASE_URL"}}), &out)
	if out.Status != "pending_confirmation" || out.ConfirmationID == "" || !strings.HasSuffix(out.ConfirmURL, "action="+out.ConfirmationID) || len(out.Secrets) != 0 {
		t.Fatalf("want a pending reveal with a link and no value, got %+v", out)
	}
	return out
}

// onPage calls a confirm page route as userID's browser session, or as a delegated
// token when actor is set.
func onPage(f *fixture, method, path string, userID int64, actor string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	return asPerson(f, req, porte.Identity{UserID: userID, SessionID: 1, Actor: actor, BrowserSession: true})
}

// asPerson serves req on the confirm page routes as identity.
func asPerson(f *fixture, req *http.Request, identity porte.Identity) *httptest.ResponseRecorder {
	router := chi.NewRouter()
	f.service.page.Routes(router)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req.WithContext(porte.WithIdentity(req.Context(), identity)))
	return recorder
}

// TestOnlyABrowserSessionConfirms checks a CLI session, or a browser session sent with
// a bearer header, cannot confirm a reveal, and the action stays open.
func TestOnlyABrowserSessionConfirms(t *testing.T) {
	f := newFixture(t)
	asked := askReveal(t, f)
	path := "/" + asked.ConfirmationID + "/confirm"
	cli := asPerson(f, httptest.NewRequest(http.MethodPost, path, nil), porte.Identity{UserID: f.owner, SessionID: 1})
	withBearer := httptest.NewRequest(http.MethodPost, path, nil)
	withBearer.Header.Set("Authorization", "Bearer x")
	bearer := asPerson(f, withBearer, porte.Identity{UserID: f.owner, SessionID: 1, BrowserSession: true})
	if cli.Code != http.StatusForbidden || bearer.Code != http.StatusForbidden {
		t.Fatalf("want 403 for a CLI session and a bearer header, got %d and %d", cli.Code, bearer.Code)
	}
	if got := f.outcomeOf(t, asked.ConfirmationID); got != "" {
		t.Fatalf("a refused confirmation must leave the action open, got %q", got)
	}
}

// TestTwoClientsAskingTheSameKeysGetTheirOwnReveal checks the request dedupe never
// hands one client an action only another client can collect.
func TestTwoClientsAskingTheSameKeysGetTheirOwnReveal(t *testing.T) {
	f := newFixture(t)
	first := askReveal(t, f)
	again := askReveal(t, f)
	other := f.issuer.Mint(t, map[string]any{"aud": []string{f.resource}, "sub": ownerSubject, "scope": "casier.read casier.write", "client_id": "https://other.example/client"})
	var theirs revealOutput
	call(t, mcptest.Connect(t, f.resource, other, mcptest.ClientOptions{}), toolReveal, prod(map[string]any{"keys": []string{"DATABASE_URL"}}), &theirs)
	if again.ConfirmationID != first.ConfirmationID || theirs.ConfirmationID == "" || theirs.ConfirmationID == first.ConfirmationID {
		t.Fatalf("want one action per client, got %q, %q and %q", first.ConfirmationID, again.ConfirmationID, theirs.ConfirmationID)
	}
}

// collectArgs is the second call of a link confirmation.
func collectArgs(id string) map[string]any {
	return prod(map[string]any{"confirmation_id": id})
}

// TestARevealConfirmedOnThePageIsCollectedOnce checks the page path: nothing
// before the confirmation, the values once after it, never on the page itself.
func TestARevealConfirmedOnThePageIsCollectedOnce(t *testing.T) {
	f := newFixture(t)
	asked := askReveal(t, f)
	session := f.session(t, ownerSubject)
	refused(t, session, toolReveal, collectArgs(asked.ConfirmationID))
	if got := onPage(f, http.MethodPost, "/"+asked.ConfirmationID+"/confirm", f.owner, "bulle"); got.Code != http.StatusForbidden {
		t.Fatalf("a delegated token must not confirm, got %d", got.Code)
	}
	list := onPage(f, http.MethodGet, "/", f.owner, "")
	if list.Code != http.StatusOK || !strings.Contains(list.Body.String(), `"key":"DATABASE_URL"`) || strings.Contains(list.Body.String(), "plaintext") {
		t.Fatalf("want the action with its key and no value, got %d %s", list.Code, list.Body)
	}
	page := onPage(f, http.MethodPost, "/"+asked.ConfirmationID+"/confirm", f.owner, "")
	if page.Code != http.StatusOK || !strings.Contains(page.Body.String(), `"status":"approved"`) || strings.Contains(page.Body.String(), "plaintext") {
		t.Fatalf("want approved and no value on the page, got %d %s", page.Code, page.Body)
	}
	var out revealOutput
	call(t, session, toolReveal, collectArgs(asked.ConfirmationID), &out)
	if out.Status != statusRevealed || len(out.Secrets) != 1 || out.Secrets[0].Value != "postgres://plaintext-db" {
		t.Fatalf("want the value once, got %+v", out)
	}
	refused(t, session, toolReveal, collectArgs(asked.ConfirmationID))
}

// TestOnlyTheClientThatAskedCollects checks a confirmed reveal is bound to the person
// and the client that asked.
func TestOnlyTheClientThatAskedCollects(t *testing.T) {
	f := newFixture(t)
	asked := askReveal(t, f)
	onPage(f, http.MethodPost, "/"+asked.ConfirmationID+"/confirm", f.owner, "")
	other := f.issuer.Mint(t, map[string]any{"aud": []string{f.resource}, "sub": ownerSubject, "scope": "casier.read casier.write", "client_id": "https://other.example/client"})
	refused(t, mcptest.Connect(t, f.resource, other, mcptest.ClientOptions{}), toolReveal, collectArgs(asked.ConfirmationID))
	refused(t, f.session(t, memberSubject), toolReveal, collectArgs(asked.ConfirmationID))
	var out revealOutput
	call(t, f.session(t, ownerSubject), toolReveal, collectArgs(asked.ConfirmationID), &out)
	if out.Status != statusRevealed {
		t.Fatalf("the asking client must still collect, got %+v", out)
	}
}

// TestAValueWrittenAfterTheQuestionIsNotHandedOver checks a confirmation covers the
// version the person was asked about, not a later one.
func TestAValueWrittenAfterTheQuestionIsNotHandedOver(t *testing.T) {
	f := newFixture(t)
	asked := askReveal(t, f)
	onPage(f, http.MethodPost, "/"+asked.ConfirmationID+"/confirm", f.owner, "")
	if _, err := f.deps.Secrets.Calls().Update(f.as(f.owner), projectSlug, envSlug, "DATABASE_URL", &secrets.UpdateRequest{Value: "plaintext-v2"}); err != nil {
		t.Fatal(err)
	}
	if text := refused(t, f.session(t, ownerSubject), toolReveal, collectArgs(asked.ConfirmationID)); strings.Contains(text, "plaintext") {
		t.Fatalf("a refusal must not carry a value: %s", text)
	}
	if got := f.outcomeOf(t, asked.ConfirmationID); got != mcpkit.OutcomeDone {
		t.Fatalf("a refused collection must not spend the confirmation, got %q", got)
	}
}

// TestADeleteConfirmedOnThePageTrashes checks the delete link path keeps the secret
// restorable.
func TestADeleteConfirmedOnThePageTrashes(t *testing.T) {
	f := newFixture(t)
	args, trashed := f.freshTrash(t)
	var out deleteOutput
	call(t, f.session(t, ownerSubject), toolDelete, args, &out)
	id := out.ConfirmURL[strings.LastIndex(out.ConfirmURL, "=")+1:]
	if got := onPage(f, http.MethodPost, "/"+id+"/confirm", f.owner, ""); got.Code != http.StatusOK || trashed() != 1 {
		t.Fatalf("want the secret trashed, got %d %s", got.Code, got.Body)
	}
	var restored writeOutput
	call(t, f.session(t, ownerSubject), toolRestore, args, &restored)
	if restored.Status != statusRestored || trashed() != 0 {
		t.Fatalf("want the secret restored, got %+v", restored)
	}
}

// TestErasureDropsPendingActions checks a person's erasure takes their MCP actions.
func TestErasureDropsPendingActions(t *testing.T) {
	f := newFixture(t)
	askReveal(t, f)
	owner := strconv.FormatInt(f.owner, 10)
	if err := f.db.Exec(`INSERT INTO porte_identities (user_id, provider, subject) VALUES (?, 'https://sso.example', 'erase-me')`, f.owner).Error; err != nil {
		t.Fatal(err)
	}
	if err := auth.ErasePerson(context.Background(), f.db, "https://sso.example", "erase-me"); err != nil {
		t.Fatal(err)
	}
	if n := f.count(t, `SELECT count(*) FROM mcp_actions WHERE user_id = $1`, owner); n != 0 {
		t.Fatalf("want no action left for the erased person, got %d", n)
	}
}

// TestARevealStopsAtTheResultBudget checks values that do not fit in one answer are
// named as omitted, not sent, and the confirmation is spent on what was sent.
func TestARevealStopsAtTheResultBudget(t *testing.T) {
	f := newFixture(t)
	big := strings.Repeat("v", 30_000)
	for _, key := range []string{"BIG_A", "BIG_B", "BIG_C"} {
		f.secret(t, key, big)
	}
	id := confirmedReveal(t, f, "BIG_A", "BIG_B", "BIG_C")
	session := f.session(t, ownerSubject)
	result := mcptest.Call(t, session, toolReveal, collectArgs(id))
	var out revealOutput
	mcptest.Decode(t, result, &out)
	if out.Status != statusRevealed || len(out.Secrets) != 2 || out.Secrets[1].Key != "BIG_B" || len(out.Omitted) != 1 || out.Omitted[0] != "BIG_C" {
		t.Fatalf("want BIG_A and BIG_B sent and BIG_C omitted, got %d secrets, omitted %v", len(out.Secrets), out.Omitted)
	}
	if size := mcpkit.EncodedSize(out); size > mcpkit.ResultBudget {
		t.Fatalf("the answer is %d bytes, over the %d budget", size, mcpkit.ResultBudget)
	}
	if f.outcomeOf(t, id) != outcomeRevealed {
		t.Fatal("a reveal that sent values must spend its confirmation")
	}
	refused(t, session, toolReveal, collectArgs(id))
}

// TestAValueTooLargeToSendKeepsTheConfirmation checks a reveal whose first value alone
// overflows the budget sends nothing and leaves the confirmation unspent.
func TestAValueTooLargeToSendKeepsTheConfirmation(t *testing.T) {
	f := newFixture(t)
	f.secret(t, "QUOTED", strings.Repeat(`"`, 40_000))
	id := confirmedReveal(t, f, "QUOTED", "DATABASE_URL")
	text := refused(t, f.session(t, ownerSubject), toolReveal, collectArgs(id))
	if !strings.Contains(text, "too large") || strings.Contains(text, "plaintext") || strings.Contains(text, `""`) {
		t.Fatalf("want a size refusal without values, got %s", text)
	}
	if got := f.outcomeOf(t, id); got != mcpkit.OutcomeDone {
		t.Fatalf("a reveal that sent nothing must stay confirmed, got %q", got)
	}
}
