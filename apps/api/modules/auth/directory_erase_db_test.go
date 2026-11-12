package auth

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/FacileStudio/porte/directory"
	portepg "github.com/FacileStudio/porte/pg"

	"gorm.io/gorm"
)

func seedReferences(t *testing.T, db *gorm.DB, f erasureFixture) {
	t.Helper()
	exec(t, db, `INSERT INTO porte_login_codes (code_hash, user_id, expires_at, provider, subject)
		VALUES ('c1', ?, now() + interval '1 minute', ?, ?)`, f.gone, deleteIssuer, deletedID)
	exec(t, db, `INSERT INTO api_token_environments (token_id, environment_id)
		SELECT t.id, e.id FROM api_tokens t, environments e WHERE t.token_hash = 'tok-gone' AND e.project_id = ?`, f.shared)
	exec(t, db, `INSERT INTO audit_logs (project_id, user_id, user_email, action, resource)
		VALUES (?, ?, 'kept@facile.studio', 'member.removed', ?)`, f.shared, intToStr(f.kept), "user:"+intToStr(f.gone))
	exec(t, db, `INSERT INTO audit_logs (project_id, user_id, user_email, action, client_ip)
		VALUES (0, ?, 'gone@facile.studio', 'instance.root_key_rotated', '198.51.100.4')`, intToStr(f.gone))
}

// eraseThroughDirectory runs what directory.Deletions runs for one notice:
// the app's OnPersonDeleted, then porte's own Erase, with the receiver's issuer.
func eraseThroughDirectory(t *testing.T, db *gorm.DB, issuer, subject string) {
	t.Helper()
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("db.DB: %v", err)
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	receiver := NewDeletions(db, directory.Client{Issuer: issuer}, portepg.New(sqlDB).EraseIdentity, portepg.New(sqlDB).HoldsIdentity, logger).receiver
	ctx := context.Background()
	if err := receiver.OnPersonDeleted(ctx, receiver.Client.Issuer, subject); err != nil {
		t.Fatalf("OnPersonDeleted(%s, %s): %v", issuer, subject, err)
	}
	if err := receiver.Erase(ctx, receiver.Client.Issuer, subject); err != nil {
		t.Fatalf("Erase(%s, %s): %v", issuer, subject, err)
	}
}

// assertUnreferenced checks every column that can name a person, in every
// table, for the erased account's id, address or subject.
func assertUnreferenced(t *testing.T, db *gorm.DB, f erasureFixture) {
	t.Helper()
	id, text := f.gone, intToStr(f.gone)
	for query, arg := range map[string]any{
		`SELECT count(*) FROM users WHERE id = ? OR email = 'gone@facile.studio'`:                                       id,
		`SELECT count(*) FROM porte_identities WHERE user_id = ? OR subject = 'deleted-subject'`:                        id,
		`SELECT count(*) FROM porte_sessions WHERE user_id = ?`:                                                         id,
		`SELECT count(*) FROM porte_login_codes WHERE user_id = ?`:                                                      id,
		`SELECT count(*) FROM project_members WHERE user_id = ?`:                                                        id,
		`SELECT count(*) FROM secret_acls WHERE user_id = ?`:                                                            id,
		`SELECT count(*) FROM api_tokens WHERE created_by = ?`:                                                          id,
		`SELECT count(*) FROM api_token_environments te JOIN api_tokens t ON t.id = te.token_id WHERE t.created_by = ?`: id,
		`SELECT count(*) FROM invites WHERE created_by = ?`:                                                             id,
		`SELECT count(*) FROM webhooks WHERE created_by = ?`:                                                            id,
		`SELECT count(*) FROM audit_logs WHERE user_id = ? OR resource = 'user:' || ? OR user_email = 'gone@facile.studio' OR client_ip IN ('203.0.113.7', '198.51.100.4')`: []any{text, text},
		`SELECT count(*) FROM secret_versions WHERE changed_by = ?`: text,
		`SELECT count(*) FROM secrets WHERE deleted_by = ?`:         text,
	} {
		args, ok := arg.([]any)
		if !ok {
			args = []any{arg}
		}
		if got := count(t, db, query, args...); got != 0 {
			t.Errorf("%s: %d rows still name the erased person", query, got)
		}
	}
}
