package mcp

import (
	"context"
	"database/sql"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/FacileStudio/Casier/apps/api/internal/audit"
	"github.com/FacileStudio/Casier/apps/api/internal/authcontext"
	"github.com/FacileStudio/Casier/apps/api/internal/crypto"
	"github.com/FacileStudio/Casier/apps/api/internal/testdb"
	"github.com/FacileStudio/Casier/apps/api/internal/webhook"
	"github.com/FacileStudio/Casier/apps/api/modules/auth"
	"github.com/FacileStudio/Casier/apps/api/modules/environments"
	"github.com/FacileStudio/Casier/apps/api/modules/projects"
	"github.com/FacileStudio/Casier/apps/api/modules/secrets"
	"github.com/FacileStudio/Casier/apps/api/schemas"
	"github.com/FacileStudio/porte"
	"github.com/FacileStudio/porte/mcpauth"
	mcpkit "github.com/FacileStudio/tronc/mcp"
	"github.com/FacileStudio/tronc/mcp/mcptest"
	"github.com/go-chi/chi/v5"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Subjects of the test people: the owner calls the tools, the member shares the
// project under an ACL that hides one secret, the stranger is in no project.
const (
	ownerSubject    = "owner-sub"
	memberSubject   = "member-sub"
	strangerSubject = "stranger-sub"
	projectSlug     = "web"
	envSlug         = "prod"
	lockedKey       = "LOCKED_KEY"
)

// identities maps the test subjects to their users and nothing else.
type identities struct {
	porte.IdentityStore
	users map[string]int64
}

// Find resolves only the test subjects.
func (s identities) Find(_ context.Context, provider, sub string) (porte.StoredIdentity, error) {
	id, ok := s.users[sub]
	if !ok {
		return porte.StoredIdentity{}, porte.ErrNotFound
	}
	return porte.StoredIdentity{UserID: id, Provider: provider, Subject: sub}, nil
}

// fixture is a running MCP endpoint over Casier's real services and database.
type fixture struct {
	db       *gorm.DB
	sqlDB    *sql.DB
	srv      *httptest.Server
	issuer   *mcptest.Issuer
	service  *Service
	deps     Deps
	owner    int64
	member   int64
	stranger int64
	resource string
}

// newFixture starts the endpoint on any free port.
func newFixture(t *testing.T) *fixture {
	t.Helper()
	return newFixtureOn(t, nil, nil)
}

// newFixtureOn is newFixture on listener, nil for any free port, with front wrapping
// the routes when it is set.
func newFixtureOn(t *testing.T, listener net.Listener, front func(*fixture, http.Handler) http.Handler) *fixture {
	t.Helper()
	f := &fixture{issuer: mcptest.NewIssuer(t)}
	f.db, f.sqlDB = openDB(t)
	f.deps = newDeps(t, f.db, f.sqlDB)
	f.seed(t)
	var handler http.Handler
	f.srv = httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { handler.ServeHTTP(w, r) }))
	if listener != nil {
		f.srv.Listener = listener
	}
	f.srv.Start()
	t.Cleanup(f.srv.Close)
	f.resource = f.srv.URL + "/mcp"
	f.service = f.build(t, listener != nil)
	router := chi.NewRouter()
	RegisterRoutes(router, f.service)
	handler = router
	if front != nil {
		handler = front(f, router)
	}
	return f
}

// openDB opens the migrated test database, mcp_actions included, emptied.
func openDB(t *testing.T) (*gorm.DB, *sql.DB) {
	t.Helper()
	url, ok := testdb.URL()
	if !ok {
		testdb.Announce()
		t.Skip(testdb.SkipReason)
	}
	db, err := testdb.Open(url)
	if err != nil {
		t.Fatal(err)
	}
	if err := testdb.Truncate(db); err != nil {
		t.Fatal(err)
	}
	db.Logger = logger.Discard
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	return db, sqlDB
}

// newDeps wires Casier's services as main does.
func newDeps(t *testing.T, db *gorm.DB, sqlDB *sql.DB) Deps {
	t.Helper()
	quiet := slog.New(slog.NewTextHandler(io.Discard, nil))
	key := make([]byte, 32)
	auditLog := audit.NewLogger(db, quiet)
	projectsService := projects.NewService(db, crypto.NewKeyring(key), auditLog)
	environmentsService := environments.NewService(db, projectsService)
	secretsService := secrets.NewService(db, quiet, projectsService, environmentsService, projectsService, webhook.NewDispatcher(db, quiet), auditLog)
	users, err := auth.NewUserStore(context.Background(), sqlDB, quiet)
	if err != nil {
		t.Fatalf("auth.NewUserStore: %v", err)
	}
	return Deps{
		Projects: projectsService, Environments: environmentsService, Secrets: secretsService,
		Accounts: users, Actions: sqlDB,
	}
}

// seed creates the people and the owner's web project, where the member may read
// every secret but the locked one.
func (f *fixture) seed(t *testing.T) {
	t.Helper()
	f.owner = f.user(t, "owner@example.test", "Owner")
	f.member = f.user(t, "member@example.test", "Member")
	f.stranger = f.user(t, "stranger@example.test", "Stranger")
	project, err := f.deps.Projects.Calls().Create(f.as(f.owner), &projects.CreateRequest{Name: "Web"})
	if err != nil {
		t.Fatal(err)
	}
	if err := f.db.Create(&schemas.ProjectMember{ProjectID: project.ID, UserID: f.member, Role: "member"}).Error; err != nil {
		t.Fatal(err)
	}
	f.secret(t, "DATABASE_URL", "postgres://plaintext-db")
	locked := f.secret(t, lockedKey, "plaintext-locked")
	if err := f.db.Exec(`INSERT INTO secret_acls (secret_id, user_id, can_read, can_write) VALUES (?, ?, false, false)`, locked, f.member).Error; err != nil {
		t.Fatal(err)
	}
}

// user creates one person.
func (f *fixture) user(t *testing.T, email, name string) int64 {
	t.Helper()
	u := schemas.User{Email: email, Name: name}
	if err := f.db.Create(&u).Error; err != nil {
		t.Fatal(err)
	}
	return u.ID
}

// as is a context carrying userID as a web session, for seeding.
func (f *fixture) as(userID int64) context.Context {
	return authcontext.WithIdentity(context.Background(), authcontext.Identity{UserID: strconv.FormatInt(userID, 10)})
}

// secret creates a secret in the owner's prod environment and returns its id.
func (f *fixture) secret(t testing.TB, key, value string) int64 {
	t.Helper()
	created, err := f.deps.Secrets.Calls().Create(f.as(f.owner), projectSlug, envSlug, &secrets.CreateRequest{Key: key, Value: value})
	if err != nil {
		t.Fatal(err)
	}
	return created.ID
}

// build makes the MCP service behind the fake Registre.
func (f *fixture) build(t *testing.T, serving bool) *Service {
	t.Helper()
	verifier, err := mcpauth.Verifier(context.Background(), mcpauth.Config{
		Issuer: f.issuer.URL, Resource: f.resource, ScopePrefix: "casier",
		Identities: identities{users: map[string]int64{ownerSubject: f.owner, memberSubject: f.member, strangerSubject: f.stranger}},
	})
	if err != nil {
		t.Fatal(err)
	}
	perMinute := 0
	if serving {
		perMinute = 1_000_000
	}
	service, err := NewService(f.deps, Config{
		Resource: f.resource, Issuer: f.issuer.URL, Version: "test", StateKey: []byte("state-key"), Verifier: verifier,
		Pending: mcpkit.PostgresPending{DB: f.sqlDB}, PerMinute: perMinute,
	})
	if err != nil {
		t.Fatal(err)
	}
	return service
}

// token is a token for subject on this endpoint with scopes, or both when none is
// given.
func (f *fixture) token(t testing.TB, subject string, scopes ...string) string {
	if len(scopes) == 0 {
		scopes = []string{"casier.read", "casier.write"}
	}
	return f.issuer.Token(t, f.resource, subject, scopes...)
}

// session opens subject's client session with every scope and no elicitation.
func (f *fixture) session(t *testing.T, subject string) *sdk.ClientSession {
	t.Helper()
	return mcptest.Connect(t, f.resource, f.token(t, subject), mcptest.ClientOptions{})
}
