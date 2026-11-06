package auth

import (
	"context"
	"strings"

	"gorm.io/gorm"
)

// DeletedUserName is the author shown on audit rows that a deleted person wrote
// in a space that outlives them.
const DeletedUserName = "Deleted user"

const personalProjects = `SELECT pm.project_id FROM project_members pm WHERE pm.user_id = @id
	AND NOT EXISTS (SELECT 1 FROM project_members o WHERE o.project_id = pm.project_id AND o.user_id <> @id)`

const promoteSuccessorOwner = `UPDATE project_members SET role = 'owner' WHERE id IN (
	SELECT DISTINCT ON (pm.project_id) pm.id FROM project_members pm
	WHERE pm.user_id <> @id
	  AND pm.project_id IN (SELECT project_id FROM project_members WHERE user_id = @id AND role = 'owner')
	  AND NOT EXISTS (SELECT 1 FROM project_members o
	                  WHERE o.project_id = pm.project_id AND o.user_id <> @id AND o.role = 'owner')
	ORDER BY pm.project_id, (pm.role = 'admin') DESC, pm.created_at, pm.id)`

func eraseSteps() []string {
	return []string{
		`DELETE FROM audit_logs WHERE project_id IN (` + personalProjects + `)`,
		`DELETE FROM projects WHERE id IN (` + personalProjects + `)`,
		promoteSuccessorOwner,
		`DELETE FROM secret_acls WHERE user_id = @id`,
		`DELETE FROM project_members WHERE user_id = @id`,
		`DELETE FROM api_tokens WHERE created_by = @id`,
		`DELETE FROM invites WHERE created_by = @id`,
		`UPDATE webhooks SET created_by = 0 WHERE created_by = @id`,
		`UPDATE audit_logs SET user_id = '', user_email = @name WHERE user_id = @text_id`,
		`UPDATE secret_versions SET changed_by = '' WHERE changed_by = @text_id`,
		`UPDATE secrets SET deleted_by = '' WHERE deleted_by = @text_id`,
		`DELETE FROM users WHERE id = @id`,
	}
}

// ErasePerson deletes the account behind a Registre identity and what it owns.
//
// A project the person was the only member of is their personal vault: it goes
// with its environments, secrets, versions, tokens and audit rows. A project
// with other members belongs to them. It keeps its secrets, the person's
// membership, grants, invites and API tokens are removed, authorship on its
// audit rows and secret history is anonymised, and when the person was its last
// owner the longest-standing admin, or else the longest-standing member, becomes
// owner. An account that is already gone is a no-op, so a replayed notice
// changes nothing.
func ErasePerson(ctx context.Context, orm *gorm.DB, provider, subject string) error {
	var account struct{ ID int64 }
	err := orm.WithContext(ctx).Raw(
		`SELECT u.id FROM users u JOIN porte_identities i ON i.user_id = u.id
		 WHERE rtrim(i.provider, '/') = ? AND i.subject = ?`,
		strings.TrimRight(provider, "/"), subject).Scan(&account).Error
	if err != nil || account.ID == 0 {
		return err
	}
	args := map[string]any{"id": account.ID, "text_id": intToStr(account.ID), "name": DeletedUserName}
	return orm.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, statement := range eraseSteps() {
			if err := tx.Exec(statement, args).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
