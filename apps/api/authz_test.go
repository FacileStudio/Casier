package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/FacileStudio/Casier/apps/api/internal/audit"
	"github.com/FacileStudio/Casier/apps/api/internal/crypto"
	"github.com/FacileStudio/Casier/apps/api/internal/env"
	"github.com/FacileStudio/Casier/apps/api/internal/testdb"
	"github.com/FacileStudio/Casier/apps/api/internal/webhook"
	"github.com/FacileStudio/Casier/apps/api/modules/auth"
	"github.com/FacileStudio/Casier/apps/api/modules/environments"
	"github.com/FacileStudio/Casier/apps/api/modules/mcp"
	"github.com/FacileStudio/Casier/apps/api/modules/projects"
	"github.com/FacileStudio/Casier/apps/api/modules/secrets"
	"github.com/FacileStudio/Casier/apps/api/schemas"

	"github.com/FacileStudio/porte/directory"
	"github.com/FacileStudio/porte/kit"
	"github.com/FacileStudio/porte/oidc"
	portepg "github.com/FacileStudio/porte/pg"
	"github.com/FacileStudio/porte/session"
	"github.com/FacileStudio/tronc/apiref"
	"github.com/FacileStudio/tronc/health"
	"github.com/FacileStudio/tronc/httpx"
	mcpkit "github.com/FacileStudio/tronc/mcp"
	"github.com/go-chi/chi/v5"
	sdkauth "github.com/modelcontextprotocol/go-sdk/auth"
	"gorm.io/gorm"
)

// The callers the matrix is written against. Every route case must name every
// one of them exactly once: an unnamed caller is an unanswered question, and
// unanswered questions are how a route ships without a guard.
const (
	anon      = "anonymous"
	badToken  = "garbage-token"
	cOwner    = "session-owner"
	cAdmin    = "session-admin"
	cMember   = "session-member"
	cViewer   = "session-viewer"
	cOutsider = "session-non-member"
	cRotator  = "session-instance-admin"
	tRead     = "token-read-only"
	tWrite    = "token-write"
	tDev      = "token-scoped-to-dev"
	tBeta     = "token-scoped-to-other-project"
	tExpired  = "token-expired"
	tRevoked  = "token-revoked"
)

var allCallers = []string{
	anon, badToken,
	cOwner, cAdmin, cMember, cViewer, cOutsider, cRotator,
	tRead, tWrite, tDev, tBeta, tExpired, tRevoked,
}

// rejectedAtTheDoor never reaches a handler: no credential, a credential that
// does not resolve, and the two states a machine token dies in.
var rejectedAtTheDoor = []string{anon, badToken, tExpired, tRevoked}

// liveTokens are the machine credentials that authenticate successfully, so
// they are the ones a route's own guards have to answer for.
var liveTokens = []string{tRead, tWrite, tDev, tBeta}

const (
	rotatorEmail  = "rotator@casier.test"
	outsiderEmail = "outsider@casier.test"
	memberEmail   = "member@casier.test"

	projectAlpha = "alpha"
	envDev       = "dev"
	envProd      = "prod"
	keyOpen      = "API_KEY"
	keyLocked    = "LOCKED"
	keyTrashed   = "TRASHED_KEY"
)

var (
	seedRootKey = base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{0x07}, 32))
	freshKey    = base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{0x09}, 32))
)

// bearers maps a caller to the credential it presents. An empty string sends
// no Authorization header at all.
var bearers = map[string]string{
	anon:      "",
	badToken:  "not-a-real-token",
	cOwner:    "session-token-owner",
	cAdmin:    "session-token-admin",
	cMember:   "session-token-member",
	cViewer:   "session-token-viewer",
	cOutsider: "session-token-outsider",
	cRotator:  "session-token-rotator",
	tRead:     "casier_test_readonly",
	tWrite:    "casier_test_write",
	tDev:      "casier_test_devonly",
	tBeta:     "casier_test_beta",
	tExpired:  "casier_test_expired",
	tRevoked:  "casier_test_revoked",
}

type expect struct {
	status int
	code   string
}

type outcome struct {
	expect
	callers []string
}

type routeCase struct {
	name    string
	method  string
	pattern string
	path    string
	body    string
	// skipReseed marks a case that cannot change state, so the fixture does not
	// have to be rebuilt between its callers.
	outcomes []outcome
}

func ok(status int, callers ...string) outcome {
	return outcome{expect{status, ""}, callers}
}

func fails(status int, code string, callers ...string) outcome {
	return outcome{expect{status, code}, callers}
}

func unauth(callers ...string) outcome {
	return fails(http.StatusUnauthorized, "unauthenticated", callers...)
}

func denied(callers ...string) outcome {
	return fails(http.StatusForbidden, "permission_denied", callers...)
}

func missing(callers ...string) outcome {
	return fails(http.StatusNotFound, "not_found", callers...)
}

type fixture struct {
	userIDs   map[string]int64
	inviteID  int64
	webhookID int64
	tokenID   int64
	aclID     int64
	keyring   *crypto.Keyring
}

var (
	testDB      *gorm.DB
	testRouter  chi.Router
	testKeyring *crypto.Keyring
	databaseOn  bool
)

func TestMain(m *testing.M) {
	url, configured := testdb.URL()
	if !configured {
		testdb.Announce()
	} else {
		db, err := testdb.Open(url)
		if err != nil {
			fmt.Fprintf(os.Stderr, "authz: %v\n", err)
			os.Exit(1)
		}
		testDB, databaseOn = db, true
	}

	key, err := crypto.KeyFromBase64(seedRootKey)
	if err != nil {
		fmt.Fprintf(os.Stderr, "authz: %v\n", err)
		os.Exit(1)
	}
	testKeyring = crypto.NewKeyring(key)

	var testSQLDB *sql.DB
	if databaseOn {
		if testSQLDB, err = testDB.DB(); err != nil {
			fmt.Fprintf(os.Stderr, "authz: %v\n", err)
			os.Exit(1)
		}
	}
	testRouter = buildTestRouter(testDB, testSQLDB, testKeyring)

	os.Exit(m.Run())
}

// testDirectory is a Registre nobody answers for. No seeded account has an
// identity under it, so the invitations route never reaches the network.
var testDirectory = directory.Client{Issuer: "https://sso.casier.test", ClientID: "casier", ClientSecret: "test-secret"}

// buildTestRouter assembles the router exactly as main.go's run() does, so
// every request in the matrix travels the middleware chain that ships. Testing
// the service layer directly is precisely how a route-level gap survives.
//
// It tolerates a nil database so the route-coverage test can walk the mounted
// routes without Postgres: no service touches the handle until a request
// arrives.
func buildTestRouter(db *gorm.DB, sqlDB *sql.DB, keyring *crypto.Keyring) chi.Router {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	appEnv := env.Config{
		EncryptionKey: seedRootKey,
		AdminEmails:   []string{rotatorEmail},
	}

	store := portepg.New(sqlDB)
	porteCfg := appEnv.Porte()
	sessions, err := session.New(porteCfg, session.Deps{Sessions: store.Sessions(), Logger: logger})
	if err != nil {
		panic(fmt.Sprintf("session.New: %v", err))
	}
	users, err := auth.NewUserStore(context.Background(), sqlDB, logger)
	if err != nil {
		panic(fmt.Sprintf("auth.NewUserStore: %v", err))
	}
	var oidcKit *oidc.Kit
	authService := auth.NewService(db, keyring, logger, sessions, users)
	deletions := auth.NewDeletions(db, testDirectory, nil, nil, logger)
	webhookDispatcher := webhook.NewDispatcher(db, logger)
	auditLogger := audit.NewLogger(db, logger)
	projectsService := projects.NewService(db, keyring, auditLogger)
	environmentsService := environments.NewService(db, projectsService)
	secretsService := secrets.NewService(db, logger, projectsService, environmentsService, projectsService, webhookDispatcher, auditLogger)

	router := httpx.NewRouter(httpx.Config{Logger: logger})
	health.Mount(router, health.DB(sqlDB))
	apiref.Mount(router, refConfig())

	router.Route("/api", func(r chi.Router) {
		if err := auth.RegisterRoutes(r, auth.Routes{
			Sessions: sessions, OIDC: oidcKit, Service: authService, Deletions: deletions.Handler(),
			Invitations: &kit.InvitationsConfig{Directory: testDirectory, Identities: store.Identities(), Logger: logger},
		}); err != nil {
			panic(fmt.Sprintf("auth.RegisterRoutes: %v", err))
		}
		projects.RegisterRoutes(r, projectsService, authService, appEnv)
		environments.RegisterRoutes(r, environmentsService, authService)
		secrets.RegisterRoutes(r, secretsService, authService)
		mcp.RegisterPendingRoutes(r, testMCP(users, sqlDB, projectsService, environmentsService, secretsService), authService)
	})

	router.Get("/spaces", redirectToProjects)
	router.Get("/spaces/*", redirectToProjects)

	return router
}

// testMCP is the MCP service behind the confirm page routes. Its verifier refuses every
// token: the matrix drives the page, which authenticates the person's own session and
// never an MCP token.
func testMCP(users mcp.Accounts, sqlDB *sql.DB, projectsService *projects.Service, environmentsService *environments.Service, secretsService *secrets.Service) *mcp.Service {
	service, err := mcp.NewService(mcp.Deps{
		Projects: projectsService, Environments: environmentsService, Secrets: secretsService,
		Accounts: users, Actions: sqlDB,
	}, mcp.Config{
		Resource: "https://casier.test/mcp", Issuer: "https://sso.casier.test", StateKey: []byte("state-key"),
		Verifier: func(context.Context, string, *http.Request) (*sdkauth.TokenInfo, error) {
			return nil, sdkauth.ErrInvalidToken
		},
		Pending: mcpkit.PostgresPending{DB: sqlDB},
	})
	if err != nil {
		panic(fmt.Sprintf("mcp.NewService: %v", err))
	}
	return service
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// seed builds the fixture the matrix asserts against: a project with dev and
// prod environments and a deliberately restricted LOCKED secret whose ACL names
// only the viewer. The registered webhook points at a port nothing listens on
// (127.0.0.1:1), a refused connection ends the goroutine immediately instead of
// hanging on DNS.
func seed(db *gorm.DB, keyring *crypto.Keyring) (fixture, error) {
	if err := testdb.Truncate(db); err != nil {
		return fixture{}, err
	}

	f := fixture{userIDs: map[string]int64{}, keyring: keyring}
	now := time.Now()

	users := []struct {
		caller string
		email  string
	}{
		{cOwner, "owner@casier.test"},
		{cAdmin, "admin@casier.test"},
		{cMember, memberEmail},
		{cViewer, "viewer@casier.test"},
		{cOutsider, outsiderEmail},
		{cRotator, rotatorEmail},
	}
	for _, u := range users {
		record := &schemas.User{Email: u.email}
		if err := db.Create(record).Error; err != nil {
			return f, fmt.Errorf("seed user %s: %w", u.email, err)
		}
		f.userIDs[u.email] = record.ID

		if err := db.Exec(`INSERT INTO porte_identities (user_id, provider, subject) VALUES (?, 'https://sso.casier.test', ?)`,
			record.ID, fmt.Sprintf("%d", record.ID)).Error; err != nil {
			return f, fmt.Errorf("seed identity %s: %w", u.email, err)
		}

		sessToken := hashToken(bearers[u.caller])
		if err := db.Exec(`INSERT INTO porte_sessions (token_hash, user_id, created_at, last_used_at, expires_at) VALUES (?, ?, now(), now(), ?)`,
			sessToken, record.ID, now.Add(time.Hour)).Error; err != nil {
			return f, fmt.Errorf("seed session %s: %w", u.email, err)
		}
	}

	alpha, alphaKey, err := seedProject(db, keyring, "Alpha", projectAlpha)
	if err != nil {
		return f, err
	}
	beta, betaKey, err := seedProject(db, keyring, "Beta", "beta")
	if err != nil {
		return f, err
	}

	memberships := []struct {
		project int64
		email   string
		role    string
	}{
		{alpha, "owner@casier.test", "owner"},
		{alpha, "admin@casier.test", "admin"},
		{alpha, memberEmail, "member"},
		{alpha, "viewer@casier.test", "viewer"},
		{beta, outsiderEmail, "owner"},
	}
	for _, m := range memberships {
		record := &schemas.ProjectMember{ProjectID: m.project, UserID: f.userIDs[m.email], Role: m.role}
		if err := db.Create(record).Error; err != nil {
			return f, fmt.Errorf("seed member %s: %w", m.email, err)
		}
	}

	alphaDev, err := seedEnvironment(db, alpha, "Development", envDev, 0)
	if err != nil {
		return f, err
	}
	if _, err := seedEnvironment(db, alpha, "Production", envProd, 1); err != nil {
		return f, err
	}
	betaDev, err := seedEnvironment(db, beta, "Development", envDev, 0)
	if err != nil {
		return f, err
	}

	openSecret, err := seedSecret(db, alphaDev, alphaKey, keyOpen, "s3cret", 2)
	if err != nil {
		return f, err
	}
	lockedSecret, err := seedSecret(db, alphaDev, alphaKey, keyLocked, "locked-value", 1)
	if err != nil {
		return f, err
	}
	trashedSecret, err := seedSecret(db, alphaDev, alphaKey, keyTrashed, "trashed-value", 1)
	if err != nil {
		return fixture{}, err
	}
	if err := db.Model(&schemas.Secret{}).Where("id = ?", trashedSecret).
		Updates(map[string]any{"deleted_at": time.Now().Add(-time.Hour), "deleted_by": "0"}).Error; err != nil {
		return fixture{}, fmt.Errorf("seed trashed secret: %w", err)
	}

	if _, err := seedSecret(db, betaDev, betaKey, "BETA_KEY", "beta-value", 1); err != nil {
		return f, err
	}
	_ = openSecret

	acl := &schemas.SecretACL{SecretID: lockedSecret, UserID: f.userIDs["viewer@casier.test"], CanRead: true, CanWrite: false}
	if err := db.Create(acl).Error; err != nil {
		return f, fmt.Errorf("seed acl: %w", err)
	}
	f.aclID = acl.ID

	invite := &schemas.Invite{
		Token:     "invite-token-alpha",
		ProjectID: alpha,
		Role:      "member",
		MaxUses:   0,
		ExpiresAt: now.Add(24 * time.Hour),
		CreatedBy: f.userIDs["owner@casier.test"],
	}
	if err := db.Create(invite).Error; err != nil {
		return f, fmt.Errorf("seed invite: %w", err)
	}
	f.inviteID = invite.ID

	hook := &schemas.Webhook{ProjectID: alpha, URL: "http://127.0.0.1:1/hook", Events: "*", Active: true, CreatedBy: f.userIDs["owner@casier.test"]}
	if err := db.Create(hook).Error; err != nil {
		return f, fmt.Errorf("seed webhook: %w", err)
	}
	f.webhookID = hook.ID

	apiTokens := []struct {
		caller      string
		project     int64
		allEnvs     bool
		envs        []int64
		permissions string
		expiresAt   time.Time
		revoked     bool
		creator     string
	}{
		{tRead, alpha, true, nil, "read", now.Add(24 * time.Hour), false, "owner@casier.test"},
		{tWrite, alpha, true, nil, "read,write", now.Add(24 * time.Hour), false, "owner@casier.test"},
		{tDev, alpha, false, []int64{alphaDev}, "read,write", now.Add(24 * time.Hour), false, "owner@casier.test"},
		{tBeta, beta, true, nil, "read,write", now.Add(24 * time.Hour), false, outsiderEmail},
		{tExpired, alpha, true, nil, "read,write", now.Add(-time.Hour), false, "owner@casier.test"},
		{tRevoked, alpha, true, nil, "read,write", now.Add(24 * time.Hour), true, "owner@casier.test"},
	}
	for _, spec := range apiTokens {
		record := &schemas.ApiToken{
			TokenHash:       hashToken(bearers[spec.caller]),
			Name:            spec.caller,
			ProjectID:       spec.project,
			AllEnvironments: spec.allEnvs,
			Permissions:     spec.permissions,
			CreatedBy:       f.userIDs[spec.creator],
			ExpiresAt:       spec.expiresAt,
		}
		if spec.revoked {
			revokedAt := now
			record.RevokedAt = &revokedAt
		}
		if err := db.Create(record).Error; err != nil {
			return f, fmt.Errorf("seed api token %s: %w", spec.caller, err)
		}
		if spec.caller == tRead {
			f.tokenID = record.ID
		}
		for _, envID := range spec.envs {
			grant := &schemas.ApiTokenEnvironment{TokenID: record.ID, EnvironmentID: envID}
			if err := db.Create(grant).Error; err != nil {
				return f, fmt.Errorf("seed token scope %s: %w", spec.caller, err)
			}
		}
	}

	return f, nil
}

func seedProject(db *gorm.DB, keyring *crypto.Keyring, name, slug string) (int64, []byte, error) {
	dataKey := bytes.Repeat([]byte(slug+"0123456789abcdef0123456789abcdef"), 1)[:32]
	encrypted, nonce, err := crypto.Encrypt(dataKey, keyring.Key())
	if err != nil {
		return 0, nil, fmt.Errorf("seed project %s: %w", slug, err)
	}
	record := &schemas.Project{
		Name:             name,
		Slug:             slug,
		EncryptedDataKey: encrypted,
		DataKeyNonce:     nonce,
		RotationDays:     schemas.DefaultRotationDays,
	}
	if err := db.Create(record).Error; err != nil {
		return 0, nil, fmt.Errorf("seed project %s: %w", slug, err)
	}
	return record.ID, dataKey, nil
}

func seedEnvironment(db *gorm.DB, projectID int64, name, slug string, position int) (int64, error) {
	record := &schemas.Environment{ProjectID: projectID, Name: name, Slug: slug, Position: position}
	if err := db.Create(record).Error; err != nil {
		return 0, fmt.Errorf("seed environment %s: %w", slug, err)
	}
	return record.ID, nil
}

// seedSecret writes a secret and one secret_versions row per version, so
// rollback has somewhere to roll back to and the rotation clock has an age.
func seedSecret(db *gorm.DB, envID int64, dataKey []byte, key, value string, versions int) (int64, error) {
	ciphertext, nonce, err := crypto.Encrypt([]byte(value), dataKey)
	if err != nil {
		return 0, fmt.Errorf("seed secret %s: %w", key, err)
	}
	record := &schemas.Secret{EnvironmentID: envID, Key: key, EncryptedValue: ciphertext, Nonce: nonce, Version: versions}
	if err := db.Create(record).Error; err != nil {
		return 0, fmt.Errorf("seed secret %s: %w", key, err)
	}
	for v := 1; v <= versions; v++ {
		version := &schemas.SecretVersion{SecretID: record.ID, Version: v, EncryptedValue: ciphertext, Nonce: nonce, ChangedBy: "0"}
		if err := db.Create(version).Error; err != nil {
			return 0, fmt.Errorf("seed secret version %s: %w", key, err)
		}
	}
	return record.ID, nil
}

// routeCases returns the full authorization matrix, grouped: unauthenticated
// surface, authenticated identity, instance-wide key rotation (gated on
// CASIER_ADMIN_EMAILS, closed to machine tokens and to every project role
// including owner), projects, environments, secrets, and the per-secret ACL
// layer.
//
// Known gaps, pinned as observed so a fix has to change the line deliberately:
// project creation carries no requireUserSession and no scope check, so a
// read-only CI token can mint a project of its own; rename has no AllowsWrite
// check, so a read-only token inherits its creator's owner role; delete is the
// same, and removes the project and every secret in it; the audit trail is
// open to every project role and to machine tokens, including one scoped to a
// single environment; and the environment listing is not filtered by the
// calling token's scope, so a dev-only token can enumerate prod.
//
// The secret endpoints carry their own rationale: revealing one value is a
// read gated exactly like a list (the separation it buys is in the audit
// trail, not in who may call it); restoring is a write, so a read-only token
// and a viewer are refused even though they can see the key in the trash; the
// LOCKED ACL names only the viewer, so admin and member lose a write their
// project role would otherwise allow while owner still bypasses it; the grant
// list is as sensitive as the secret it describes, so it answers only to
// callers who may read LOCKED itself, and admin and member get not_found
// rather than permission_denied because an ACL hides a secret rather than
// announcing it; and granting access is administration, not a secret write, so
// no machine token may do it whatever its permissions — every live token is
// refused identically, so the refusal says nothing about what exists.
func routeCases(f fixture) []routeCase {
	alphaPath := "/api/projects/" + projectAlpha
	devSecrets := alphaPath + "/environments/" + envDev + "/secrets"
	prodSecrets := alphaPath + "/environments/" + envProd + "/secrets"
	secretPattern := "/api/projects/{slug}/environments/{envSlug}/secrets"

	everyone := allCallers
	projectReaders := []string{cOwner, cAdmin, cMember, cViewer}
	projectWriters := []string{cOwner, cAdmin, cMember}
	projectAdmins := []string{cOwner, cAdmin}
	strangers := []string{cOutsider, cRotator}

	return []routeCase{
		{
			name: "liveness", method: "GET", pattern: "/health", path: "/health",
			outcomes: []outcome{ok(200, everyone...)},
		},
		{
			name: "readiness", method: "GET", pattern: "/ready", path: "/ready",
			outcomes: []outcome{ok(200, everyone...)},
		},
		{
			name: "liveness under /api", method: "GET", pattern: "/api/health", path: "/api/health",
			outcomes: []outcome{ok(200, everyone...)},
		},
		{
			name: "readiness under /api", method: "GET", pattern: "/api/ready", path: "/api/ready",
			outcomes: []outcome{ok(200, everyone...)},
		},
		{
			name: "api documentation", method: "GET", pattern: "/docs", path: "/docs",
			outcomes: []outcome{ok(200, everyone...)},
		},
		{
			name: "api specification", method: "GET", pattern: "/docs/openapi.json", path: "/docs/openapi.json",
			outcomes: []outcome{ok(200, everyone...)},
		},
		{
			name: "legacy spaces redirect", method: "GET", pattern: "/spaces", path: "/spaces",
			outcomes: []outcome{ok(301, everyone...)},
		},
		{
			name: "legacy spaces subtree redirect", method: "GET", pattern: "/spaces/*", path: "/spaces/alpha/secrets",
			outcomes: []outcome{ok(301, everyone...)},
		},
		{
			name: "auth config", method: "GET", pattern: "/api/auth/config", path: "/api/auth/config",
			outcomes: []outcome{ok(200, everyone...)},
		},
		{
			name: "registre deletion notice", method: "POST", pattern: "/api/registre/deletions", path: "/api/registre/deletions",
			outcomes: []outcome{ok(400, everyone...)},
		},

		{
			name: "logout", method: "POST", pattern: "/api/auth/logout", path: "/api/auth/logout",
			outcomes: []outcome{
				ok(200, everyone...),
			},
		},
		{
			name: "me", method: "GET", pattern: "/api/auth/me", path: "/api/auth/me",
			outcomes: []outcome{
				ok(200, cOwner, cAdmin, cMember, cViewer, cOutsider, cRotator),
				ok(200, liveTokens...),
				unauth(rejectedAtTheDoor...),
			},
		},
		{
			name: "invitations", method: "GET", pattern: "/api/me/invitations", path: "/api/me/invitations",
			outcomes: []outcome{
				ok(200, cOwner, cAdmin, cMember, cViewer, cOutsider, cRotator),
				unauth(liveTokens...),
				unauth(rejectedAtTheDoor...),
			},
		},

		{
			name: "mcp pending list", method: "GET", pattern: "/api/mcp/pending/", path: "/api/mcp/pending",
			outcomes: []outcome{
				ok(200, cOwner, cAdmin, cMember, cViewer, cOutsider, cRotator),
				unauth(liveTokens...),
				unauth(rejectedAtTheDoor...),
			},
		},
		{
			name: "mcp pending confirm", method: "POST", pattern: "/api/mcp/pending/{id}/confirm", path: "/api/mcp/pending/no-such-action/confirm",
			outcomes: []outcome{
				missing(cOwner, cAdmin, cMember, cViewer, cOutsider, cRotator),
				unauth(liveTokens...),
				unauth(rejectedAtTheDoor...),
			},
		},
		{
			name: "mcp pending dismiss", method: "POST", pattern: "/api/mcp/pending/{id}/dismiss", path: "/api/mcp/pending/no-such-action/dismiss",
			outcomes: []outcome{
				missing(cOwner, cAdmin, cMember, cViewer, cOutsider, cRotator),
				unauth(liveTokens...),
				unauth(rejectedAtTheDoor...),
			},
		},

		{
			name: "rotate root key", method: "POST", pattern: "/api/admin/rotate-key", path: "/api/admin/rotate-key",
			body: `{"new_key":"` + freshKey + `"}`,
			outcomes: []outcome{
				ok(200, cRotator),
				denied(cOwner, cAdmin, cMember, cViewer, cOutsider),
				denied(liveTokens...),
				unauth(rejectedAtTheDoor...),
			},
		},

		{
			name: "accept invite", method: "POST", pattern: "/api/invites/{token}/accept", path: "/api/invites/invite-token-alpha/accept",
			outcomes: []outcome{
				ok(204, cOutsider, cRotator),
				fails(409, "already_exists", cOwner, cAdmin, cMember, cViewer),
				denied(liveTokens...),
				unauth(rejectedAtTheDoor...),
			},
		},

		{
			name: "list projects", method: "GET", pattern: "/api/projects/", path: "/api/projects/",
			outcomes: []outcome{
				ok(200, cOwner, cAdmin, cMember, cViewer, cOutsider, cRotator),
				ok(200, liveTokens...),
				unauth(rejectedAtTheDoor...),
			},
		},
		{
			name: "create project", method: "POST", pattern: "/api/projects/", path: "/api/projects/",
			body: `{"name":"Gamma","description":""}`,
			outcomes: []outcome{
				ok(201, cOwner, cAdmin, cMember, cViewer, cOutsider, cRotator),
				ok(201, liveTokens...),
				unauth(rejectedAtTheDoor...),
			},
		},
		{
			name: "read project", method: "GET", pattern: "/api/projects/{slug}/", path: alphaPath + "/",
			outcomes: []outcome{
				ok(200, projectReaders...),
				ok(200, tRead, tWrite, tDev),
				missing(append(strangers, tBeta)...),
				unauth(rejectedAtTheDoor...),
			},
		},
		{
			name: "update project", method: "PUT", pattern: "/api/projects/{slug}/", path: alphaPath + "/",
			body: `{"name":"Alpha renamed","description":"x"}`,
			outcomes: []outcome{
				ok(200, projectAdmins...),
				denied(cMember, cViewer),
				ok(200, tRead, tWrite, tDev),
				missing(append(strangers, tBeta)...),
				unauth(rejectedAtTheDoor...),
			},
		},
		{
			name: "delete project", method: "DELETE", pattern: "/api/projects/{slug}/", path: alphaPath + "/",
			outcomes: []outcome{
				ok(204, cOwner),
				denied(cAdmin, cMember, cViewer),
				ok(204, tRead, tWrite, tDev),
				missing(append(strangers, tBeta)...),
				unauth(rejectedAtTheDoor...),
			},
		},
		{
			name: "list members", method: "GET", pattern: "/api/projects/{slug}/members", path: alphaPath + "/members",
			outcomes: []outcome{
				ok(200, projectReaders...),
				missing(strangers...),
				denied(liveTokens...),
				unauth(rejectedAtTheDoor...),
			},
		},
		{
			name: "add member", method: "POST", pattern: "/api/projects/{slug}/members", path: alphaPath + "/members",
			body: `{"email":"` + outsiderEmail + `","role":"member"}`,
			outcomes: []outcome{
				ok(204, projectAdmins...),
				denied(cMember, cViewer),
				missing(strangers...),
				denied(liveTokens...),
				unauth(rejectedAtTheDoor...),
			},
		},
		{
			name: "remove member", method: "DELETE", pattern: "/api/projects/{slug}/members/{userId}",
			path: alphaPath + "/members/" + strconv.FormatInt(f.userIDs["viewer@casier.test"], 10),
			outcomes: []outcome{
				ok(200, projectAdmins...),
				denied(cMember, cViewer),
				missing(strangers...),
				denied(liveTokens...),
				unauth(rejectedAtTheDoor...),
			},
		},
		{
			name: "list invites", method: "GET", pattern: "/api/projects/{slug}/invites", path: alphaPath + "/invites",
			outcomes: []outcome{
				ok(200, projectAdmins...),
				denied(cMember, cViewer),
				missing(strangers...),
				denied(liveTokens...),
				unauth(rejectedAtTheDoor...),
			},
		},
		{
			name: "create invite", method: "POST", pattern: "/api/projects/{slug}/invites", path: alphaPath + "/invites",
			body: `{"role":"member","max_uses":1}`,
			outcomes: []outcome{
				ok(201, projectAdmins...),
				denied(cMember, cViewer),
				missing(strangers...),
				denied(liveTokens...),
				unauth(rejectedAtTheDoor...),
			},
		},
		{
			name: "delete invite", method: "DELETE", pattern: "/api/projects/{slug}/invites/{inviteId}",
			path: alphaPath + "/invites/" + strconv.FormatInt(f.inviteID, 10),
			outcomes: []outcome{
				ok(204, projectAdmins...),
				denied(cMember, cViewer),
				missing(strangers...),
				denied(liveTokens...),
				unauth(rejectedAtTheDoor...),
			},
		},
		{
			name: "read audit log", method: "GET", pattern: "/api/projects/{slug}/audit", path: alphaPath + "/audit",
			outcomes: []outcome{
				ok(200, projectReaders...),
				ok(200, tRead, tWrite, tDev),
				missing(append(strangers, tBeta)...),
				unauth(rejectedAtTheDoor...),
			},
		},
		{
			name: "list webhooks", method: "GET", pattern: "/api/projects/{slug}/webhooks", path: alphaPath + "/webhooks",
			outcomes: []outcome{
				ok(200, projectAdmins...),
				denied(cMember, cViewer),
				missing(strangers...),
				denied(liveTokens...),
				unauth(rejectedAtTheDoor...),
			},
		},
		{
			name: "create webhook", method: "POST", pattern: "/api/projects/{slug}/webhooks", path: alphaPath + "/webhooks",
			body: `{"url":"http://127.0.0.1:1/new","events":"*","secret":""}`,
			outcomes: []outcome{
				ok(201, projectAdmins...),
				denied(cMember, cViewer),
				missing(strangers...),
				denied(liveTokens...),
				unauth(rejectedAtTheDoor...),
			},
		},
		{
			name: "delete webhook", method: "DELETE", pattern: "/api/projects/{slug}/webhooks/{webhookId}",
			path: alphaPath + "/webhooks/" + strconv.FormatInt(f.webhookID, 10),
			outcomes: []outcome{
				ok(204, projectAdmins...),
				denied(cMember, cViewer),
				missing(strangers...),
				denied(liveTokens...),
				unauth(rejectedAtTheDoor...),
			},
		},
		{
			name: "list api tokens", method: "GET", pattern: "/api/projects/{slug}/tokens", path: alphaPath + "/tokens",
			outcomes: []outcome{
				ok(200, projectAdmins...),
				denied(cMember, cViewer),
				missing(strangers...),
				denied(liveTokens...),
				unauth(rejectedAtTheDoor...),
			},
		},
		{
			name: "mint api token", method: "POST", pattern: "/api/projects/{slug}/tokens", path: alphaPath + "/tokens",
			body: `{"name":"ci","environments":"","permissions":"read"}`,
			outcomes: []outcome{
				ok(201, projectAdmins...),
				denied(cMember, cViewer),
				missing(strangers...),
				denied(liveTokens...),
				unauth(rejectedAtTheDoor...),
			},
		},
		{
			name: "revoke api token", method: "DELETE", pattern: "/api/projects/{slug}/tokens/{tokenId}",
			path: alphaPath + "/tokens/" + strconv.FormatInt(f.tokenID, 10),
			outcomes: []outcome{
				ok(204, projectAdmins...),
				denied(cMember, cViewer),
				missing(strangers...),
				denied(liveTokens...),
				unauth(rejectedAtTheDoor...),
			},
		},

		{
			name: "list environments", method: "GET", pattern: "/api/projects/{slug}/environments/", path: alphaPath + "/environments/",
			outcomes: []outcome{
				ok(200, projectReaders...),
				ok(200, tRead, tWrite, tDev),
				missing(append(strangers, tBeta)...),
				unauth(rejectedAtTheDoor...),
			},
		},
		{
			name: "create environment", method: "POST", pattern: "/api/projects/{slug}/environments/", path: alphaPath + "/environments/",
			body: `{"name":"QA"}`,
			outcomes: []outcome{
				ok(201, projectAdmins...),
				denied(cMember, cViewer),
				denied(tRead),
				ok(201, tWrite, tDev),
				missing(append(strangers, tBeta)...),
				unauth(rejectedAtTheDoor...),
			},
		},
		{
			name: "rename environment outside the token scope", method: "PUT", pattern: "/api/projects/{slug}/environments/{envSlug}",
			path: alphaPath + "/environments/" + envProd, body: `{"name":"Production"}`,
			outcomes: []outcome{
				ok(200, projectAdmins...),
				denied(cMember, cViewer),
				denied(tRead),
				ok(200, tWrite),
				missing(tDev),
				missing(append(strangers, tBeta)...),
				unauth(rejectedAtTheDoor...),
			},
		},
		{
			name: "delete environment outside the token scope", method: "DELETE", pattern: "/api/projects/{slug}/environments/{envSlug}",
			path: alphaPath + "/environments/" + envProd,
			outcomes: []outcome{
				ok(204, projectAdmins...),
				denied(cMember, cViewer),
				denied(tRead),
				ok(204, tWrite),
				missing(tDev),
				missing(append(strangers, tBeta)...),
				unauth(rejectedAtTheDoor...),
			},
		},

		{
			name: "diff across environments", method: "GET", pattern: "/api/projects/{slug}/secrets/diff",
			path: alphaPath + "/secrets/diff?from=" + envDev + "&to=" + envProd,
			outcomes: []outcome{
				ok(200, projectReaders...),
				ok(200, tRead, tWrite),
				denied(tDev),
				missing(append(strangers, tBeta)...),
				unauth(rejectedAtTheDoor...),
			},
		},
		{
			name: "rotation summary", method: "GET", pattern: "/api/projects/{slug}/secrets/rotation", path: alphaPath + "/secrets/rotation",
			outcomes: []outcome{
				ok(200, projectReaders...),
				ok(200, tRead, tWrite, tDev),
				missing(append(strangers, tBeta)...),
				unauth(rejectedAtTheDoor...),
			},
		},
		{
			name: "list secrets in the scoped environment", method: "GET", pattern: secretPattern + "/", path: devSecrets + "/",
			outcomes: []outcome{
				ok(200, projectReaders...),
				ok(200, tRead, tWrite, tDev),
				missing(append(strangers, tBeta)...),
				unauth(rejectedAtTheDoor...),
			},
		},
		{
			name: "list secrets outside the token scope", method: "GET", pattern: secretPattern + "/", path: prodSecrets + "/",
			outcomes: []outcome{
				ok(200, projectReaders...),
				ok(200, tRead, tWrite),
				denied(tDev),
				missing(append(strangers, tBeta)...),
				unauth(rejectedAtTheDoor...),
			},
		},
		{
			name: "create secret", method: "POST", pattern: secretPattern + "/", path: devSecrets + "/",
			body: `{"key":"NEW_KEY","value":"v","tags":[]}`,
			outcomes: []outcome{
				ok(201, projectWriters...),
				denied(cViewer),
				denied(tRead),
				ok(201, tWrite, tDev),
				missing(append(strangers, tBeta)...),
				unauth(rejectedAtTheDoor...),
			},
		},
		{
			name: "reveal a single secret", method: "GET", pattern: secretPattern + "/{key}/reveal", path: devSecrets + "/" + keyOpen + "/reveal",
			outcomes: []outcome{
				ok(200, projectReaders...),
				ok(200, tRead, tWrite, tDev),
				missing(append(strangers, tBeta)...),
				unauth(rejectedAtTheDoor...),
			},
		},
		{
			name: "list the trash", method: "GET", pattern: secretPattern + "/trash", path: devSecrets + "/trash",
			outcomes: []outcome{
				ok(200, projectReaders...),
				ok(200, tRead, tWrite, tDev),
				missing(append(strangers, tBeta)...),
				unauth(rejectedAtTheDoor...),
			},
		},
		{
			name: "restore a trashed secret", method: "POST", pattern: secretPattern + "/{key}/restore", path: devSecrets + "/" + keyTrashed + "/restore",
			outcomes: []outcome{
				ok(204, projectWriters...),
				denied(cViewer),
				denied(tRead),
				ok(204, tWrite, tDev),
				missing(append(strangers, tBeta)...),
				unauth(rejectedAtTheDoor...),
			},
		},
		{
			name: "import secrets", method: "POST", pattern: secretPattern + "/import", path: devSecrets + "/import",
			body: `{"content":"IMPORTED=1"}`,
			outcomes: []outcome{
				ok(200, projectWriters...),
				denied(cViewer),
				denied(tRead),
				ok(200, tWrite, tDev),
				missing(append(strangers, tBeta)...),
				unauth(rejectedAtTheDoor...),
			},
		},
		{
			name: "export secrets", method: "GET", pattern: secretPattern + "/export", path: devSecrets + "/export",
			outcomes: []outcome{
				ok(200, projectReaders...),
				ok(200, tRead, tWrite, tDev),
				missing(append(strangers, tBeta)...),
				unauth(rejectedAtTheDoor...),
			},
		},
		{
			name: "update an unrestricted secret", method: "PUT", pattern: secretPattern + "/{key}", path: devSecrets + "/" + keyOpen,
			body: `{"value":"rotated"}`,
			outcomes: []outcome{
				ok(200, projectWriters...),
				denied(cViewer),
				denied(tRead),
				ok(200, tWrite, tDev),
				missing(append(strangers, tBeta)...),
				unauth(rejectedAtTheDoor...),
			},
		},
		{
			name: "update an ACL-restricted secret", method: "PUT", pattern: secretPattern + "/{key}", path: devSecrets + "/" + keyLocked,
			body: `{"value":"rotated"}`,
			outcomes: []outcome{
				ok(200, cOwner),
				denied(cAdmin, cMember, cViewer),
				denied(tRead),
				ok(200, tWrite, tDev),
				missing(append(strangers, tBeta)...),
				unauth(rejectedAtTheDoor...),
			},
		},
		{
			name: "delete an unrestricted secret", method: "DELETE", pattern: secretPattern + "/{key}", path: devSecrets + "/" + keyOpen,
			outcomes: []outcome{
				ok(204, projectWriters...),
				denied(cViewer),
				denied(tRead),
				ok(204, tWrite, tDev),
				missing(append(strangers, tBeta)...),
				unauth(rejectedAtTheDoor...),
			},
		},
		{
			name: "delete an ACL-restricted secret", method: "DELETE", pattern: secretPattern + "/{key}", path: devSecrets + "/" + keyLocked,
			outcomes: []outcome{
				ok(204, cOwner),
				denied(cAdmin, cMember, cViewer),
				denied(tRead),
				ok(204, tWrite, tDev),
				missing(append(strangers, tBeta)...),
				unauth(rejectedAtTheDoor...),
			},
		},
		{
			name: "retag a secret", method: "PUT", pattern: secretPattern + "/{key}/tags", path: devSecrets + "/" + keyOpen + "/tags",
			body: `{"tags":["prod"]}`,
			outcomes: []outcome{
				ok(204, projectWriters...),
				denied(cViewer),
				denied(tRead),
				ok(204, tWrite, tDev),
				missing(append(strangers, tBeta)...),
				unauth(rejectedAtTheDoor...),
			},
		},
		{
			name: "list secret versions", method: "GET", pattern: secretPattern + "/{key}/versions", path: devSecrets + "/" + keyOpen + "/versions",
			outcomes: []outcome{
				ok(200, projectReaders...),
				ok(200, tRead, tWrite, tDev),
				missing(append(strangers, tBeta)...),
				unauth(rejectedAtTheDoor...),
			},
		},
		{
			name: "roll a secret back", method: "POST", pattern: secretPattern + "/{key}/rollback", path: devSecrets + "/" + keyOpen + "/rollback",
			body: `{"version":1}`,
			outcomes: []outcome{
				ok(200, projectWriters...),
				denied(cViewer),
				denied(tRead),
				ok(200, tWrite, tDev),
				missing(append(strangers, tBeta)...),
				unauth(rejectedAtTheDoor...),
			},
		},
		{
			name: "list secret ACLs", method: "GET", pattern: secretPattern + "/{key}/acls", path: devSecrets + "/" + keyLocked + "/acls",
			outcomes: []outcome{
				ok(200, cOwner, cViewer),
				missing(cAdmin, cMember),
				ok(200, tRead, tWrite, tDev),
				missing(append(strangers, tBeta)...),
				unauth(rejectedAtTheDoor...),
			},
		},
		{
			name: "grant a secret ACL", method: "PUT", pattern: secretPattern + "/{key}/acls", path: devSecrets + "/" + keyOpen + "/acls",
			body: `{"email":"` + memberEmail + `","can_read":true,"can_write":false}`,
			outcomes: []outcome{
				ok(204, projectAdmins...),
				denied(cMember, cViewer),
				denied(liveTokens...),
				missing(strangers...),
				unauth(rejectedAtTheDoor...),
			},
		},
		{
			name: "revoke a secret ACL", method: "DELETE", pattern: secretPattern + "/{key}/acls/{aclId}",
			path: devSecrets + "/" + keyLocked + "/acls/" + strconv.FormatInt(f.aclID, 10),
			outcomes: []outcome{
				ok(204, projectAdmins...),
				denied(cMember, cViewer),
				denied(liveTokens...),
				missing(strangers...),
				unauth(rejectedAtTheDoor...),
			},
		},
	}
}

// TestAuthorizationMatrix drives the real router over real rows: for every
// route, every caller, the exact status and the exact error code.
//
// The error code is asserted and not just the status because the codes carry a
// deliberate decision: an out-of-reach project answers not_found rather than
// permission_denied, so a token cannot map what it cannot touch. Pin it or
// someone will helpfully "fix" it into an oracle.
func TestAuthorizationMatrix(t *testing.T) {
	requireDatabase(t)

	f, err := seed(testDB, testKeyring)
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	address := 0
	for _, testCase := range routeCases(f) {
		t.Run(testCase.method+" "+testCase.pattern+" ("+testCase.name+")", func(t *testing.T) {
			answers, err := expand(testCase)
			if err != nil {
				t.Fatalf("%v", err)
			}

			for _, caller := range allCallers {
				t.Run(caller, func(t *testing.T) {
					fresh, err := seed(testDB, testKeyring)
					if err != nil {
						t.Fatalf("reseed: %v", err)
					}
					address++

					path := testCase.path
					for _, rebuilt := range routeCases(fresh) {
						if rebuilt.name == testCase.name && rebuilt.method == testCase.method {
							path = rebuilt.path
							break
						}
					}

					status, code := call(testCase.method, path, testCase.body, bearers[caller], address)
					want := answers[caller]
					if status != want.status || code != want.code {
						t.Errorf("got %d/%q, want %d/%q", status, code, want.status, want.code)
					}
				})
			}
		})
	}
}

// expand flattens a case's outcomes into one answer per caller and refuses a
// case that leaves a caller out or names one twice. That refusal is the whole
// reason the table is written as outcomes rather than as a sparse map.
func expand(testCase routeCase) (map[string]expect, error) {
	answers := make(map[string]expect, len(allCallers))
	for _, out := range testCase.outcomes {
		for _, caller := range out.callers {
			if _, duplicate := answers[caller]; duplicate {
				return nil, fmt.Errorf("%s %s: caller %q has two expectations", testCase.method, testCase.pattern, caller)
			}
			answers[caller] = out.expect
		}
	}
	var unanswered []string
	for _, caller := range allCallers {
		if _, ok := answers[caller]; !ok {
			unanswered = append(unanswered, caller)
		}
	}
	if len(unanswered) > 0 {
		return nil, fmt.Errorf("%s %s: no expectation for %s", testCase.method, testCase.pattern, strings.Join(unanswered, ", "))
	}
	return answers, nil
}

// call issues one request through the real router. Each request gets its own
// client address so the auth rate limiter, which is per IP and orthogonal to
// authorization, cannot make the matrix depend on its own execution order.
func call(method, path, body, bearer string, address int) (int, string) {
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	request := httptest.NewRequest(method, path, reader)
	request.RemoteAddr = fmt.Sprintf("10.%d.%d.%d:40000", address>>16&0xff, address>>8&0xff, address&0xff)
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	if bearer != "" {
		request.Header.Set("Authorization", "Bearer "+bearer)
	}

	recorder := httptest.NewRecorder()
	testRouter.ServeHTTP(recorder, request)

	var envelope struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	_ = json.Unmarshal(recorder.Body.Bytes(), &envelope)
	return recorder.Code, envelope.Error.Code
}

// TestEveryMountedRouteIsInTheMatrix walks the router chi actually serves and
// fails on any (method, pattern) the matrix does not answer for. It is the part
// that keeps the gate honest: a new endpoint cannot ship without someone
// writing down who is allowed to call it.
//
// It needs no database, so it guards the routes even where the matrix skips.
func TestEveryMountedRouteIsInTheMatrix(t *testing.T) {
	covered := map[string]bool{}
	for _, testCase := range routeCases(fixture{userIDs: map[string]int64{}}) {
		covered[testCase.method+" "+testCase.pattern] = true
	}

	mounted := map[string]bool{}
	router, ok := testRouter.(chi.Routes)
	if !ok {
		t.Fatal("the router is not walkable")
	}
	err := chi.Walk(router, func(method string, route string, _ http.Handler, _ ...func(http.Handler) http.Handler) error {
		mounted[method+" "+normalizeRoute(route)] = true
		return nil
	})
	if err != nil {
		t.Fatalf("walk: %v", err)
	}

	var uncovered, stale []string
	for route := range mounted {
		if !covered[route] {
			uncovered = append(uncovered, route)
		}
	}
	for route := range covered {
		if !mounted[route] {
			stale = append(stale, route)
		}
	}
	sort.Strings(uncovered)
	sort.Strings(stale)

	if len(uncovered) > 0 {
		t.Errorf("mounted but absent from the authorization matrix:\n  %s", strings.Join(uncovered, "\n  "))
	}
	if len(stale) > 0 {
		t.Errorf("in the authorization matrix but no longer mounted:\n  %s", strings.Join(stale, "\n  "))
	}
}

// normalizeRoute strips the artefacts chi.Walk leaves when it descends into a
// subrouter: a mount point contributes a "/*" that the served pattern does not
// have.
func normalizeRoute(route string) string {
	route = strings.ReplaceAll(route, "/*/", "/")
	if strings.HasSuffix(route, "/*") && route != "/spaces/*" {
		route = strings.TrimSuffix(route, "*")
	}
	return route
}

func requireDatabase(t *testing.T) {
	t.Helper()
	if !databaseOn {
		t.Skip(testdb.SkipReason)
	}
}

// TestSecretReferenceRespectsACLs pins the one hole that made the ACL layer
// decorative: ${ref:env/KEY} expansion resolved the referenced secret with only
// the token scope checked, never the per-secret ACL. A member who could not
// list, reveal, export or diff LOCKED could still read it by writing a
// reference to it into a secret they were allowed to create, then revealing
// that. The reference is a read and is gated like one, so the placeholder must
// survive expansion literally.
func TestSecretReferenceRespectsACLs(t *testing.T) {
	requireDatabase(t)

	if _, err := seed(testDB, testKeyring); err != nil {
		t.Fatalf("seed: %v", err)
	}

	devSecrets := "/api/projects/" + projectAlpha + "/environments/" + envDev + "/secrets"
	member := bearers[cMember]

	status, _ := call("POST", devSecrets, `{"key":"SIPHON","value":"${ref:`+envDev+`/`+keyLocked+`}"}`, member, 900)
	if status != 201 && status != 204 && status != 200 {
		t.Fatalf("member could not create the referencing secret: got %d", status)
	}

	status, body := callBody("GET", devSecrets+"/SIPHON/reveal", "", member, 901)
	if status != 200 {
		t.Fatalf("reveal: got %d, want 200", status)
	}
	if strings.Contains(body, "locked-value") {
		t.Fatalf("ACL bypassed: reference expansion leaked the locked value: %s", body)
	}
}

func callBody(method, path, body, bearer string, address int) (int, string) {
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	request := httptest.NewRequest(method, path, reader)
	request.RemoteAddr = fmt.Sprintf("10.%d.%d.%d:40000", address>>16&0xff, address>>8&0xff, address&0xff)
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	if bearer != "" {
		request.Header.Set("Authorization", "Bearer "+bearer)
	}

	recorder := httptest.NewRecorder()
	testRouter.ServeHTTP(recorder, request)
	return recorder.Code, recorder.Body.String()
}
