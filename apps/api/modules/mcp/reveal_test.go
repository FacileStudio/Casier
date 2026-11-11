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

// onPage calls a confirm page route as userID's web session, or as a delegated token
// when actor is set.
func onPage(f *fixture, method, path string, userID int64, actor string) *httptest.ResponseRecorder {
	router := chi.NewRouter()
	f.service.page.Routes(router)
	req := httptest.NewRequest(method, path, nil)
	ctx := porte.WithIdentity(req.Context(), porte.Identity{UserID: userID, SessionID: 1, Actor: actor})
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req.WithContext(ctx))
	return recorder
}

// collectArgs is the second call of a link confirmation.
func collectArgs(id string) map[string]any {
	return prod(map[string]any{"confirmation_id": id})
}

// TestARevealConfirmedOnThePageIsCollectedOnce checks the no-elicitation path: nothing
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
