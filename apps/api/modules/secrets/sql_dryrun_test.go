package secrets

import (
	"strings"
	"testing"
	"time"

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

// TestTrashListingSQL pins the one query that has to see past the soft-delete
// scope. Without Unscoped, GORM appends deleted_at IS NULL and the trash lists
// nothing at all, forever, silently.
func TestTrashListingSQL(t *testing.T) {
	orm := dryRunDB(t)
	var records []schemas.Secret
	sql := orm.Unscoped().
		Where("environment_id = ? AND deleted_at IS NOT NULL", int64(1)).
		Order("deleted_at desc").
		Find(&records).Statement.SQL.String()

	want := `SELECT * FROM "secrets" WHERE environment_id = $1 AND deleted_at IS NOT NULL ORDER BY deleted_at desc`
	if sql != want {
		t.Fatalf("trash listing SQL = %q, want %q", sql, want)
	}
}

// TestPurgeSelectionSQL checks that the sweep is bounded by the cutoff and is
// not filtered back down to live rows by the soft-delete scope.
func TestPurgeSelectionSQL(t *testing.T) {
	orm := dryRunDB(t)
	var ids []int64
	statement := orm.Unscoped().
		Model(&schemas.Secret{}).
		Where("deleted_at IS NOT NULL AND deleted_at <= ?", trashCutoff(time.Now().UTC())).
		Pluck("id", &ids).Statement

	sql := statement.SQL.String()
	want := `SELECT "id" FROM "secrets" WHERE deleted_at IS NOT NULL AND deleted_at <= $1`
	if sql != want {
		t.Fatalf("purge selection SQL = %q, want %q", sql, want)
	}
	if strings.Contains(sql, "IS NULL") {
		t.Fatalf("purge selection is still scoped to live rows: %q", sql)
	}
}

// TestListSecretsStaysScoped is the mirror image: every ordinary read must keep
// the soft-delete scope, so a trashed secret cannot reappear in a listing.
func TestListSecretsStaysScoped(t *testing.T) {
	orm := dryRunDB(t)
	var records []schemas.Secret
	sql := orm.Where("environment_id = ?", int64(1)).
		Order("key asc").
		Find(&records).Statement.SQL.String()

	if !strings.Contains(sql, `"secrets"."deleted_at" IS NULL`) {
		t.Fatalf("secret listing lost the soft-delete scope: %q", sql)
	}
}
