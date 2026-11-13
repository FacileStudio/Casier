package mcp

import (
	"context"
	stderrors "errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

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

// TestAFailedReadKeepsTheConfirmation checks collect spends a confirmation only after
// its read succeeded, so a failed read can be retried.
func TestAFailedReadKeepsTheConfirmation(t *testing.T) {
	f := newFixture(t)
	id := confirmedReveal(t, f, "DATABASE_URL")
	owner, client := f.askedBy(t, id)
	failed := stderrors.New("read failed")
	if err := collect(context.Background(), f.sqlDB, owner, client, id, func(mcpkit.Action) error { return failed }); !stderrors.Is(err, failed) {
		t.Fatalf("want the read's error, got %v", err)
	}
	if got := f.outcomeOf(t, id); got != mcpkit.OutcomeDone {
		t.Fatalf("a failed read must not spend the confirmation, got %q", got)
	}
	var out revealOutput
	call(t, f.session(t, ownerSubject), toolReveal, collectArgs(id), &out)
	if out.Status != statusRevealed || len(out.Secrets) != 1 {
		t.Fatalf("want the value after the failed read, got %+v", out)
	}
}

// TestTwoCollectionsHandOverOnce checks a collection racing one in progress waits for
// it and is then refused, so values leave Casier once per confirmation.
func TestTwoCollectionsHandOverOnce(t *testing.T) {
	f := newFixture(t)
	id := confirmedReveal(t, f, "DATABASE_URL")
	owner, client := f.askedBy(t, id)
	ctx := context.Background()
	entered, release := make(chan struct{}), make(chan struct{})
	first, second := make(chan error, 1), make(chan error, 1)
	go func() {
		first <- collect(ctx, f.sqlDB, owner, client, id, func(mcpkit.Action) error { close(entered); <-release; return nil })
	}()
	<-entered
	go func() { second <- collect(ctx, f.sqlDB, owner, client, id, func(mcpkit.Action) error { return nil }) }()
	select {
	case err := <-second:
		t.Fatalf("the second collection must wait for the first, got %v", err)
	case <-time.After(300 * time.Millisecond):
	}
	close(release)
	if err := <-first; err != nil {
		t.Fatalf("the first collection must succeed, got %v", err)
	}
	if err := <-second; !stderrors.Is(err, ErrNotConfirmed) {
		t.Fatalf("the second collection must be refused, got %v", err)
	}
}

// askedBy is the person and the client an action was asked by.
func (f *fixture) askedBy(t *testing.T, id string) (string, string) {
	t.Helper()
	var user, client string
	if err := f.sqlDB.QueryRow(`SELECT user_id, client_id FROM mcp_actions WHERE id = $1`, id).Scan(&user, &client); err != nil {
		t.Fatal(err)
	}
	return user, client
}
