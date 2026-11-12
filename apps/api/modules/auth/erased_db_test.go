package auth

import (
	"context"
	stderrors "errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/FacileStudio/porte"
	"github.com/FacileStudio/porte/directory"
	"github.com/FacileStudio/porte/oidc"
	"github.com/FacileStudio/porte/oidc/jwt"
	portepg "github.com/FacileStudio/porte/pg"
	"github.com/FacileStudio/porte/session"
	"github.com/FacileStudio/tronc/mcp/mcptest"

	"gorm.io/gorm"
)

// erasedFixture is the receiver and the OIDC kit wired as main wires them,
// against a Registre that lists one erasure, sub-gone, on every poll.
type erasedFixture struct {
	db        *gorm.DB
	issuer    string
	deletions *Deletions
	kit       *oidc.Kit
	other     int64
}

func newErasedFixture(t *testing.T) erasedFixture {
	t.Helper()
	db := resetTestDB(t)
	exec(t, db, `DELETE FROM directory_deletion_state`)
	idp := mcptest.NewIssuer(t)
	deletedAt := time.Now().UTC().Add(-time.Minute).Format(time.RFC3339Nano)
	registre := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"deleted":[{"sub":"sub-gone","deleted_at":"` + deletedAt + `"}]}`))
	}))
	t.Cleanup(registre.Close)
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("db.DB: %v", err)
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	store := portepg.New(sqlDB)
	client := directory.Client{Issuer: idp.URL, HTTP: rewriteTo(registre.URL)}
	deletions := NewDeletions(db, client, store.EraseIdentity, store.HoldsIdentity, logger)
	cfg := porte.Config{Issuer: idp.URL, ClientID: "casier", ClientSecret: "secret", RedirectURL: "http://casier.test/api/auth/oidc/callback", SuccessURL: "http://casier.test"}
	sessions, err := session.New(cfg, session.Deps{Sessions: store.Sessions(), Logger: logger})
	if err != nil {
		t.Fatalf("session.New: %v", err)
	}
	kit, err := oidc.New(context.Background(), cfg, oidc.Deps{
		Users: newTestUserStore(t, db), Identities: store.Identities(), Sessions: sessions,
		Codes: store.LoginCodes(), Logger: logger, Erased: deletions,
	})
	if err != nil {
		t.Fatalf("oidc.New: %v", err)
	}
	f := erasedFixture{db: db, issuer: idp.URL, deletions: deletions, kit: kit}
	f.other = f.provision(t, "sub-other", "other@facile.studio")
	return f
}

func rewriteTo(target string) *http.Client {
	return &http.Client{Transport: roundTrip(func(r *http.Request) (*http.Response, error) {
		r = r.Clone(r.Context())
		r.URL.Scheme, r.URL.Host, r.Host = "http", target[len("http://"):], ""
		return http.DefaultTransport.RoundTrip(r)
	})}
}

type roundTrip func(*http.Request) (*http.Response, error)

func (f roundTrip) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func (f erasedFixture) provision(t *testing.T, subject, email string) int64 {
	t.Helper()
	identity, err := f.kit.ProvisionBearer(context.Background(), f.claims(subject, email))
	if err != nil {
		t.Fatalf("ProvisionBearer(%s): %v", subject, err)
	}
	return identity.UserID
}

func (f erasedFixture) claims(subject, email string) jwt.Claims {
	verified := true
	return jwt.Claims{Subject: subject, Email: email, Name: subject, Profile: jwt.Profile{EmailVerified: &verified}}
}

func (f erasedFixture) poll(t *testing.T) {
	t.Helper()
	if err := f.deletions.receiver.Poll(context.Background()); err != nil {
		t.Fatalf("poll: %v", err)
	}
}

func (f erasedFixture) held(t *testing.T, subject string) int64 {
	t.Helper()
	return count(t, f.db, `SELECT count(*) FROM porte_identities WHERE provider = ? AND subject = ?`, f.issuer, subject)
}

// An account re-created for a subject Registre already erased (here a row
// written past the guard, as a bearer minted before the erasure could) is
// erased again by the next poll of the same receiver, with no restart, and
// nobody else is touched.
func TestARecreatedAccountForAnErasedSubjectIsErasedAgain(t *testing.T) {
	f := newErasedFixture(t)
	f.provision(t, "sub-gone", "gone@facile.studio")
	f.poll(t)
	if f.held(t, "sub-gone") != 0 || count(t, f.db, `SELECT count(*) FROM users WHERE email = 'gone@facile.studio'`) != 0 {
		t.Fatalf("the first poll did not erase sub-gone")
	}

	exec(t, f.db, `INSERT INTO users (email, name) VALUES ('gone@facile.studio', 'again')`)
	exec(t, f.db, `INSERT INTO porte_identities (user_id, provider, subject)
		SELECT id, ?, 'sub-gone' FROM users WHERE email = 'gone@facile.studio'`, f.issuer)
	f.poll(t)
	if f.held(t, "sub-gone") != 0 || count(t, f.db, `SELECT count(*) FROM users WHERE email = 'gone@facile.studio'`) != 0 {
		t.Fatalf("the re-created account for sub-gone survived the next poll")
	}
	if f.held(t, "sub-other") != 1 || count(t, f.db, `SELECT count(*) FROM users WHERE id = ?`, f.other) != 1 {
		t.Fatalf("another account was touched")
	}
}

// Once the receiver handled an erasure, the OIDC kit refuses that subject an
// account on the MCP first-call path, and still provisions everyone else.
func TestProvisioningAnErasedSubjectIsRefused(t *testing.T) {
	f := newErasedFixture(t)
	f.provision(t, "sub-gone", "gone@facile.studio")
	f.poll(t)

	_, err := f.kit.ProvisionBearer(context.Background(), f.claims("sub-gone", "gone@facile.studio"))
	if !stderrors.Is(err, porte.ErrSubjectErased) || !stderrors.Is(err, oidc.ErrProvisionRefused) {
		t.Fatalf("ProvisionBearer(sub-gone) = %v, want a refusal wrapping ErrSubjectErased", err)
	}
	if f.held(t, "sub-gone") != 0 || count(t, f.db, `SELECT count(*) FROM users WHERE email = 'gone@facile.studio'`) != 0 {
		t.Fatalf("a refused provisioning still wrote an account")
	}
	if f.provision(t, "sub-new", "new@facile.studio") == 0 {
		t.Fatalf("a subject nobody erased was not provisioned")
	}
}
