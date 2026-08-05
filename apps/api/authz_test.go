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
	"github.com/FacileStudio/Casier/apps/api/internal/authcrypto"
	"github.com/FacileStudio/Casier/apps/api/internal/crypto"
	documentation "github.com/FacileStudio/Casier/apps/api/internal/documentation"
	"github.com/FacileStudio/Casier/apps/api/internal/env"
	"github.com/FacileStudio/Casier/apps/api/internal/mailer"
	"github.com/FacileStudio/Casier/apps/api/internal/testdb"
	"github.com/FacileStudio/Casier/apps/api/internal/webhook"
	"github.com/FacileStudio/Casier/apps/api/modules/auth"
	"github.com/FacileStudio/Casier/apps/api/modules/environments"
	"github.com/FacileStudio/Casier/apps/api/modules/projects"
	"github.com/FacileStudio/Casier/apps/api/modules/secrets"
	"github.com/FacileStudio/Casier/apps/api/schemas"

	"github.com/FacileStudio/tronc/health"
	"github.com/FacileStudio/tronc/httpjson"
	"github.com/FacileStudio/tronc/httpx"
	"github.com/go-chi/chi/v5"
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
	fixturePassword = "correct-horse-battery-staple"
	rotatorEmail    = "rotator@casier.test"
	outsiderEmail   = "outsider@casier.test"
	memberEmail     = "member@casier.test"

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
	testDB       *gorm.DB
	testRouter   http.Handler
	testKeyring  *crypto.Keyring
	passwordHash string
	databaseOn   bool
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
		hash, err := seedPasswordHash()
		if err != nil {
			fmt.Fprintf(os.Stderr, "authz: %v\n", err)
			os.Exit(1)
		}
		testDB, passwordHash, databaseOn = db, hash, true
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

type discardMailer struct{}

func (discardMailer) Send(context.Context, mailer.Message) error { return nil }

// buildTestRouter assembles the router exactly as main.go's run() does, so
// every request in the matrix travels the middleware chain that ships. Testing
// the service layer directly is precisely how a route-level gap survives.
//
// It tolerates a nil database so the route-coverage test can walk the mounted
// routes without Postgres: no service touches the handle until a request
// arrives.
func buildTestRouter(db *gorm.DB, sqlDB *sql.DB, keyring *crypto.Keyring) http.Handler {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	appEnv := env.Config{
		EncryptionKey: seedRootKey,
		AdminEmails:   []string{rotatorEmail},
	}

	// A stub mailer keeps the password reset routes live, so the matrix answers
	// who may call them rather than whether the deployment can send mail. The
	// no-mailer 503 is its own property and is covered in modules/auth.
	authService := auth.NewService(db, keyring, logger, discardMailer{}, "https://casier.test")
	webhookDispatcher := webhook.NewDispatcher(db, logger)
	auditLogger := audit.NewLogger(db, logger)
	projectsService := projects.NewService(db, keyring, auditLogger)
	environmentsService := environments.NewService(db, projectsService)
	secretsService := secrets.NewService(db, logger, projectsService, environmentsService, projectsService, webhookDispatcher, auditLogger)

	docs := documentation.Response{
		Modules: []documentation.Module{
			auth.Documentation,
			projects.Documentation,
			environments.Documentation,
			secrets.Documentation,
		},
	}

	router := httpx.NewRouter(httpx.Config{Logger: logger})
	health.Mount(router, health.DB(sqlDB))
	router.Get("/docs", func(w http.ResponseWriter, _ *http.Request) {
		httpjson.WriteJSON(w, http.StatusOK, docs)
	})

	router.Route("/api", func(r chi.Router) {
		auth.RegisterRoutes(r, authService, appEnv)
		projects.RegisterRoutes(r, projectsService, authService, appEnv)
		environments.RegisterRoutes(r, environmentsService, authService)
		secrets.RegisterRoutes(r, secretsService, authService)
	})

	router.Get("/spaces", redirectToProjects)
	router.Get("/spaces/*", redirectToProjects)

	return router
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// seedPasswordHash runs argon2 once per test binary; it is deliberately
// expensive, so the fixture reuses one hash for every user.
func seedPasswordHash() (string, error) {
	return authcrypto.HashPassword(fixturePassword)
}

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
		record := &schemas.User{Email: u.email, PasswordHash: passwordHash}
		if err := db.Create(record).Error; err != nil {
			return f, fmt.Errorf("seed user %s: %w", u.email, err)
		}
		f.userIDs[u.email] = record.ID

		session := &schemas.Session{
			Token:     hashToken(bearers[u.caller]),
			UserID:    record.ID,
			ExpiresAt: now.Add(time.Hour),
		}
		if err := db.Create(session).Error; err != nil {
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

	// LOCKED carries an ACL naming only the viewer, which makes it restricted
	// for everyone else: admin and member hold a project role that would
	// otherwise let them write it, and owner bypasses ACLs by design.
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

	// A port nothing listens on: the dispatcher fires in a goroutine, and a
	// refused connection ends it immediately instead of hanging on DNS.
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
		// Unauthenticated surface. Every caller gets the same answer, which is
		// the point: these routes must not vary with a credential.
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
			name: "register", method: "POST", pattern: "/api/auth/register", path: "/api/auth/register",
			body:     `{"email":"newcomer@casier.test","password":"a-long-enough-password"}`,
			outcomes: []outcome{ok(201, everyone...)},
		},
		{
			name: "login", method: "POST", pattern: "/api/auth/login", path: "/api/auth/login",
			body:     `{"email":"owner@casier.test","password":"` + fixturePassword + `"}`,
			outcomes: []outcome{ok(200, everyone...)},
		},
		{
			name: "forgot password", method: "POST", pattern: "/api/auth/forgot-password", path: "/api/auth/forgot-password",
			body:     `{"email":"owner@casier.test"}`,
			outcomes: []outcome{ok(200, everyone...)},
		},
		{
			name: "reset password with an unknown token", method: "POST", pattern: "/api/auth/reset-password", path: "/api/auth/reset-password",
			body:     `{"token":"not-a-token","password":"a-long-enough-password"}`,
			outcomes: []outcome{fails(400, "invalid_argument", everyone...)},
		},

		// Authenticated identity.
		{
			name: "logout", method: "POST", pattern: "/api/auth/logout", path: "/api/auth/logout",
			outcomes: []outcome{
				ok(204, cOwner, cAdmin, cMember, cViewer, cOutsider, cRotator),
				ok(204, liveTokens...),
				unauth(rejectedAtTheDoor...),
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

		// Instance-wide key rotation: gated on CASIER_ADMIN_EMAILS, closed to
		// machine tokens, and closed to every project role including owner.
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

		// Projects.
		{
			name: "list projects", method: "GET", pattern: "/api/projects/", path: "/api/projects/",
			outcomes: []outcome{
				ok(200, cOwner, cAdmin, cMember, cViewer, cOutsider, cRotator),
				ok(200, liveTokens...),
				unauth(rejectedAtTheDoor...),
			},
		},
		{
			// KNOWN GAP (reported, not fixed): project creation carries no
			// requireUserSession and no scope check, so a read-only CI token
			// can mint a project of its own. Pinned as observed so a fix has to
			// change this line deliberately.
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
			// KNOWN GAP (reported, not fixed): no requireUserSession and no
			// AllowsWrite check, so a read-only token inherits its creator's
			// owner role and can rename the project.
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
			// KNOWN GAP (reported, not fixed): as above, but this one deletes
			// the project and every secret in it.
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
			// KNOWN GAP (reported, not fixed): the audit trail is open to every
			// project role and to machine tokens, including one scoped to a
			// single environment.
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

		// Environments.
		{
			// KNOWN GAP (reported, not fixed): the listing is not filtered by
			// the calling token's environment scope, so a dev-only token can
			// enumerate prod.
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

		// Secrets.
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
			// Revealing one value is a read, gated exactly like a list: the
			// separation it buys is in the audit trail, not in who may call it.
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
			// Restoring is a write, so a read-only token and a viewer are
			// refused even though they can see the key sitting in the trash.
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
			// The per-secret ACL layer: LOCKED names only the viewer, so admin
			// and member lose a write their project role would otherwise allow,
			// and owner still bypasses it.
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
			// KNOWN GAP (reported, not fixed): who may read a secret's ACL list
			// is only "any project member", so a viewer and any live token see
			// the grants on a secret they cannot write.
			name: "list secret ACLs", method: "GET", pattern: secretPattern + "/{key}/acls", path: devSecrets + "/" + keyLocked + "/acls",
			outcomes: []outcome{
				ok(200, projectReaders...),
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
				denied(tRead),
				ok(204, tWrite, tDev),
				missing(append(strangers, tBeta)...),
				unauth(rejectedAtTheDoor...),
			},
		},
		{
			name: "revoke a secret ACL", method: "DELETE", pattern: secretPattern + "/{key}/acls/{aclId}",
			path: devSecrets + "/" + keyLocked + "/acls/" + strconv.FormatInt(f.aclID, 10),
			outcomes: []outcome{
				ok(204, projectAdmins...),
				denied(cMember, cViewer),
				denied(tRead),
				ok(204, tWrite, tDev),
				missing(append(strangers, tBeta)...),
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
