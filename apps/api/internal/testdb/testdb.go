// Package testdb holds Casier's configuration for tronc/testdb, shared by the
// test packages that open the migrated Postgres schema.
package testdb

import (
	"context"

	"github.com/FacileStudio/Casier/apps/api/migrations"

	"github.com/FacileStudio/tronc/migrate"
	trtestdb "github.com/FacileStudio/tronc/testdb"
	"gorm.io/gorm"
)

// Bootstrap is the command the skip banner names.
const Bootstrap = "sh scripts/postgres-up.sh && createdb -p 5433 casier_test"

// Config opens one schema per test binary and migrates it as main.go does.
var Config = trtestdb.Config{
	Prefix: "casier_test",
	Migrate: func(db *gorm.DB) error {
		sqlDB, err := db.DB()
		if err != nil {
			return err
		}
		lockID, err := trtestdb.LockID(db)
		if err != nil {
			return err
		}
		return migrate.Run(context.Background(), migrate.Config{DB: sqlDB, FS: migrations.FS, LockID: lockID})
	},
}
