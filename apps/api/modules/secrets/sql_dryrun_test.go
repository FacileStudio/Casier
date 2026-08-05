package secrets

import (
	"testing"

	"github.com/FacileStudio/Casier/apps/api/schemas"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func dryRunDB(t *testing.T) *gorm.DB {
	t.Helper()
	orm, err := gorm.Open(postgres.New(postgres.Config{DSN: "postgres://casier@127.0.0.1:5432/casier"}), &gorm.Config{
		DryRun:               true,
		DisableAutomaticPing: true,
	})
	if err != nil {
		t.Skipf("no postgres dialector available: %v", err)
	}
	return orm
}

func TestTagFilterSQL(t *testing.T) {
	orm := dryRunDB(t)
	var records []schemas.Secret
	statement := orm.Where("environment_id = ?", int64(1)).
		Where(tagFilterClause, tagFilterPattern("pr%d")).
		Order("key asc").
		Find(&records).Statement

	sql := statement.SQL.String()
	want := `SELECT * FROM "secrets" WHERE environment_id = $1 AND ',' || tags || ',' LIKE $2 ESCAPE '\' AND "secrets"."deleted_at" IS NULL ORDER BY key asc`
	if sql != want {
		t.Fatalf("tag filter SQL = %q, want %q", sql, want)
	}
	if got := statement.Vars[1]; got != `%,pr\%d,%` {
		t.Fatalf("tag filter pattern bound as %v", got)
	}
}

func TestSecretGrantsSQL(t *testing.T) {
	orm := dryRunDB(t)
	var rows []struct {
		SecretID int64
		CanRead  bool
		CanWrite bool
	}
	statement := orm.Table("secret_acls").
		Select("secret_id, bool_or(user_id = ? and can_read) as can_read, bool_or(user_id = ? and can_write) as can_write", int64(7), int64(7)).
		Where("secret_id IN ?", []int64{1, 2}).
		Group("secret_id").
		Find(&rows).Statement

	sql := statement.SQL.String()
	want := `SELECT secret_id, bool_or(user_id = $1 and can_read) as can_read, bool_or(user_id = $2 and can_write) as can_write FROM "secret_acls" WHERE secret_id IN ($3,$4) GROUP BY "secret_id"`
	if sql != want {
		t.Fatalf("secret grants SQL = %q, want %q", sql, want)
	}
}
