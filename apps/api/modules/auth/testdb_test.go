package auth

import (
	"testing"

	"github.com/FacileStudio/Casier/apps/api/internal/testdb"

	"gorm.io/gorm"
)

func resetTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	connection, ok := testdb.URL()
	if !ok {
		testdb.Announce()
		t.Skip(testdb.SkipReason)
	}
	db, err := testdb.Open(connection)
	if err != nil {
		t.Fatalf("testdb.Open: %v", err)
	}
	if err := testdb.Truncate(db); err != nil {
		t.Fatalf("testdb.Truncate: %v", err)
	}
	return db
}
