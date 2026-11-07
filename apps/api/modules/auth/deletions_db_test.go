package auth

import (
	"context"
	"testing"

	"gorm.io/gorm"
)

const (
	deleteIssuer = "https://sso.facile.studio"
	deletedID    = "deleted-subject"
)

type erasureFixture struct {
	gone, kept, vault, shared int64
}

func exec(t *testing.T, db *gorm.DB, query string, args ...any) {
	t.Helper()
	if err := db.Exec(query, args...).Error; err != nil {
		t.Fatalf("%s: %v", query, err)
	}
}

func insertID(t *testing.T, db *gorm.DB, query string, args ...any) int64 {
	t.Helper()
	var id int64
	if err := db.Raw(query+" RETURNING id", args...).Scan(&id).Error; err != nil {
		t.Fatalf("%s: %v", query, err)
	}
	return id
}

func seedErasure(t *testing.T, db *gorm.DB) erasureFixture {
	t.Helper()
	f := erasureFixture{}
	f.gone = insertID(t, db, `INSERT INTO users (email, name) VALUES ('gone@facile.studio', 'Gone')`)
	f.kept = insertID(t, db, `INSERT INTO users (email, name) VALUES ('kept@facile.studio', 'Kept')`)
	exec(t, db, `INSERT INTO porte_identities (user_id, provider, subject) VALUES (?, ?, ?)`, f.gone, deleteIssuer, deletedID)
	exec(t, db, `INSERT INTO porte_identities (user_id, provider, subject) VALUES (?, ?, 'kept-subject')`, f.kept, deleteIssuer)
	exec(t, db, `INSERT INTO porte_sessions (token_hash, user_id) VALUES ('h1', ?)`, f.gone)

	f.vault = insertID(t, db, `INSERT INTO projects (name, slug) VALUES ('Vault', 'vault')`)
	exec(t, db, `INSERT INTO project_members (project_id, user_id, role) VALUES (?, ?, 'owner')`, f.vault, f.gone)
	vaultEnv := insertID(t, db, `INSERT INTO environments (project_id, name, slug) VALUES (?, 'Dev', 'dev')`, f.vault)
	exec(t, db, `INSERT INTO secrets (environment_id, key) VALUES (?, 'PERSONAL')`, vaultEnv)
	exec(t, db, `INSERT INTO audit_logs (project_id, user_id, user_email, action) VALUES (?, ?, 'gone@facile.studio', 'read')`, f.vault, intToStr(f.gone))

	f.shared = insertID(t, db, `INSERT INTO projects (name, slug) VALUES ('Shared', 'shared')`)
	exec(t, db, `INSERT INTO project_members (project_id, user_id, role, created_at) VALUES (?, ?, 'owner', now())`, f.shared, f.gone)
	exec(t, db, `INSERT INTO project_members (project_id, user_id, role, created_at) VALUES (?, ?, 'member', now())`, f.shared, f.kept)
	sharedEnv := insertID(t, db, `INSERT INTO environments (project_id, name, slug) VALUES (?, 'Prod', 'prod')`, f.shared)
	secret := insertID(t, db, `INSERT INTO secrets (environment_id, key, deleted_by) VALUES (?, 'TEAM_KEY', ?)`, sharedEnv, intToStr(f.gone))
	exec(t, db, `INSERT INTO secret_versions (secret_id, version, changed_by) VALUES (?, 1, ?)`, secret, intToStr(f.gone))
	exec(t, db, `INSERT INTO secret_acls (secret_id, user_id) VALUES (?, ?)`, secret, f.gone)
	exec(t, db, `INSERT INTO api_tokens (token_hash, name, project_id, created_by) VALUES ('tok-gone', 'ci', ?, ?)`, f.shared, f.gone)
	exec(t, db, `INSERT INTO api_tokens (token_hash, name, project_id, created_by) VALUES ('tok-kept', 'ci', ?, ?)`, f.shared, f.kept)
	exec(t, db, `INSERT INTO invites (token, project_id, role, created_by) VALUES ('inv', ?, 'member', ?)`, f.shared, f.gone)
	exec(t, db, `INSERT INTO webhooks (project_id, url, created_by) VALUES (?, 'https://hook.test', ?)`, f.shared, f.gone)
	exec(t, db, `INSERT INTO audit_logs (project_id, user_id, user_email, action, client_ip) VALUES (?, ?, 'gone@facile.studio', 'write', '203.0.113.7')`, f.shared, intToStr(f.gone))
	return f
}

func count(t *testing.T, db *gorm.DB, query string, args ...any) int64 {
	t.Helper()
	var n int64
	if err := db.Raw(query, args...).Scan(&n).Error; err != nil {
		t.Fatalf("%s: %v", query, err)
	}
	return n
}

type erasureCheck struct {
	what  string
	query string
	args  []any
	want  int64
}

func erasureChecks(f erasureFixture) []erasureCheck {
	return []erasureCheck{
		{"the account", `SELECT count(*) FROM users WHERE id = ?`, []any{f.gone}, 0},
		{"its identity and sessions", `SELECT count(*) FROM porte_identities WHERE user_id = ?`, []any{f.gone}, 0},
		{"the personal vault", `SELECT count(*) FROM projects WHERE id = ?`, []any{f.vault}, 0},
		{"the personal vault secrets", `SELECT count(*) FROM secrets WHERE key = 'PERSONAL'`, nil, 0},
		{"the personal vault audit rows", `SELECT count(*) FROM audit_logs WHERE project_id = ?`, []any{f.vault}, 0},
		{"the shared space", `SELECT count(*) FROM projects WHERE id = ?`, []any{f.shared}, 1},
		{"the shared space secret", `SELECT count(*) FROM secrets WHERE key = 'TEAM_KEY'`, nil, 1},
		{"the shared secret history", `SELECT count(*) FROM secret_versions`, nil, 1},
		{"its tokens", `SELECT count(*) FROM api_tokens WHERE token_hash = 'tok-gone'`, nil, 0},
		{"the other member's token", `SELECT count(*) FROM api_tokens WHERE token_hash = 'tok-kept'`, nil, 1},
		{"its invites", `SELECT count(*) FROM invites`, nil, 0},
		{"its grants", `SELECT count(*) FROM secret_acls`, nil, 0},
		{"the successor owner", `SELECT count(*) FROM project_members WHERE project_id = ? AND user_id = ? AND role = 'owner'`, []any{f.shared, f.kept}, 1},
		{"the shared webhook", `SELECT count(*) FROM webhooks WHERE created_by = 0`, nil, 1},
		{"anonymised audit authors", `SELECT count(*) FROM audit_logs WHERE project_id = ? AND user_email = ? AND user_id = ''`, []any{f.shared, DeletedUserName}, 1},
		{"the erased person's address", `SELECT count(*) FROM audit_logs WHERE client_ip <> ''`, nil, 0},
		{"anonymised history authors", `SELECT count(*) FROM secret_versions WHERE changed_by = ''`, nil, 1},
		{"anonymised trash authors", `SELECT count(*) FROM secrets WHERE deleted_by = ''`, nil, 1},
		{"the other person", `SELECT count(*) FROM users WHERE id = ?`, []any{f.kept}, 1},
	}
}

func assertErased(t *testing.T, db *gorm.DB, f erasureFixture) {
	t.Helper()
	for _, check := range erasureChecks(f) {
		if got := count(t, db, check.query, check.args...); got != check.want {
			t.Errorf("%s: count = %d, want %d", check.what, got, check.want)
		}
	}
}

// TestErasePersonDeletesThePersonalVaultAndKeepsTheSpace is the GDPR contract:
// the person's own data goes, shared space secrets stay with the space, and a
// second notice for the same person changes nothing.
func TestErasePersonDeletesThePersonalVaultAndKeepsTheSpace(t *testing.T) {
	db := resetTestDB(t)
	f := seedErasure(t, db)
	ctx := context.Background()

	if err := ErasePerson(ctx, db, deleteIssuer+"/", deletedID); err != nil {
		t.Fatalf("ErasePerson: %v", err)
	}
	assertErased(t, db, f)

	if err := ErasePerson(ctx, db, deleteIssuer, deletedID); err != nil {
		t.Fatalf("a second ErasePerson must be a no-op, got: %v", err)
	}
	assertErased(t, db, f)

	if err := ErasePerson(ctx, db, deleteIssuer, "never-signed-in"); err != nil {
		t.Fatalf("an unknown subject must be a no-op, got: %v", err)
	}
}

func TestDeletionStateRoundTrips(t *testing.T) {
	db := resetTestDB(t)
	state := deletionState{orm: db}
	ctx := context.Background()

	if got, err := state.Load(ctx); err != nil || got != nil {
		t.Fatalf("Load before any save = (%q, %v), want (nil, nil)", got, err)
	}
	for _, blob := range []string{`{"a":1}`, `{"a":2}`} {
		if err := state.Save(ctx, []byte(blob)); err != nil {
			t.Fatalf("Save: %v", err)
		}
	}
	if got, err := state.Load(ctx); err != nil || string(got) != `{"a":2}` {
		t.Fatalf("Load = (%q, %v), want the second blob", got, err)
	}
}
