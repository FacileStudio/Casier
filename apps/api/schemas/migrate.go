package schemas

import (
	"fmt"

	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	if err := renameSpacesToProjects(db); err != nil {
		return err
	}

	if err := renameResetTokenColumn(db); err != nil {
		return err
	}

	if err := db.AutoMigrate(
		&User{},
		&Session{},
		&Project{},
		&ProjectMember{},
		&Environment{},
		&Secret{},
		&SecretVersion{},
		&PasswordReset{},
		&Invite{},
		&SecretACL{},
		&ApiToken{},
		&ApiTokenEnvironment{},
		&Webhook{},
		&AuditLog{},
	); err != nil {
		return err
	}

	return bindTokenScopesToEnvironments(db)
}

// renameResetTokenColumn runs before AutoMigrate: the column already holds a
// SHA-256 digest, it was merely named `token`, and adding `token_hash`
// alongside a NOT NULL primary key would break every insert.
func renameResetTokenColumn(db *gorm.DB) error {
	migrator := db.Migrator()
	if !migrator.HasTable("password_resets") {
		return nil
	}
	if !migrator.HasColumn(&PasswordReset{}, "token") || migrator.HasColumn(&PasswordReset{}, "token_hash") {
		return nil
	}
	return db.Exec(`ALTER TABLE password_resets RENAME COLUMN token TO token_hash`).Error
}

// bindTokenScopesToEnvironments moves every token scope off the comma-separated
// `environments` column and onto rows that reference an environment by id. A
// token stored with an empty list meant "every environment" under the old
// semantics, so it becomes all_environments rather than a token that can reach
// nothing. Slugs that no longer resolve are dropped, which is the point: they
// were grants to an environment that does not exist.
func bindTokenScopesToEnvironments(db *gorm.DB) error {
	migrator := db.Migrator()
	if !migrator.HasTable("api_tokens") || !migrator.HasColumn(&ApiToken{}, "environments") {
		return nil
	}

	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(`
			UPDATE api_tokens SET all_environments = true
			WHERE coalesce(btrim(environments), '') = ''
		`).Error; err != nil {
			return fmt.Errorf("widen unscoped tokens: %w", err)
		}

		if err := tx.Exec(`
			INSERT INTO api_token_environments (token_id, environment_id)
			SELECT t.id, e.id
			FROM api_tokens t
			JOIN LATERAL unnest(string_to_array(t.environments, ',')) AS slug ON true
			JOIN environments e ON e.project_id = t.project_id AND e.slug = btrim(slug)
			WHERE coalesce(btrim(t.environments), '') <> ''
			ON CONFLICT DO NOTHING
		`).Error; err != nil {
			return fmt.Errorf("bind token scopes: %w", err)
		}

		return tx.Exec(`ALTER TABLE api_tokens DROP COLUMN environments`).Error
	})
}

func renameSpacesToProjects(db *gorm.DB) error {
	migrator := db.Migrator()
	if !migrator.HasTable("spaces") || migrator.HasTable("projects") {
		return nil
	}

	statements := []string{
		`ALTER TABLE spaces RENAME TO projects`,
		`ALTER TABLE space_members RENAME TO project_members`,
		`ALTER TABLE project_members RENAME COLUMN space_id TO project_id`,
		`ALTER TABLE environments RENAME COLUMN space_id TO project_id`,
		`ALTER TABLE invites RENAME COLUMN space_id TO project_id`,
		`ALTER TABLE api_tokens RENAME COLUMN space_id TO project_id`,
		`ALTER TABLE webhooks RENAME COLUMN space_id TO project_id`,
		`ALTER TABLE audit_logs RENAME COLUMN space_id TO project_id`,
		`DO $$
		DECLARE row record;
		BEGIN
			FOR row IN
				SELECT conrelid::regclass AS table_name, conname
				FROM pg_constraint
				WHERE connamespace = current_schema()::regnamespace
					AND conname LIKE '%space%'
			LOOP
				EXECUTE format(
					'ALTER TABLE %s RENAME CONSTRAINT %I TO %I',
					row.table_name, row.conname, replace(row.conname, 'space', 'project')
				);
			END LOOP;
		END $$`,
		`DO $$
		DECLARE row record;
		BEGIN
			FOR row IN
				SELECT indexname
				FROM pg_indexes
				WHERE schemaname = current_schema() AND indexname LIKE '%space%'
			LOOP
				EXECUTE format(
					'ALTER INDEX %I RENAME TO %I',
					row.indexname, replace(row.indexname, 'space', 'project')
				);
			END LOOP;
		END $$`,
	}

	return db.Transaction(func(tx *gorm.DB) error {
		for _, statement := range statements {
			if err := tx.Exec(statement).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
