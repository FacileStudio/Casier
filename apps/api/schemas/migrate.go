package schemas

import "gorm.io/gorm"

func Migrate(db *gorm.DB) error {
	if err := renameSpacesToProjects(db); err != nil {
		return err
	}

	return db.AutoMigrate(
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
		&Webhook{},
		&AuditLog{},
	)
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
