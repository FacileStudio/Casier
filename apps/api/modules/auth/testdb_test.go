package auth

import (
	"context"
	"testing"

	"github.com/FacileStudio/Casier/apps/api/internal/testdb"
	"github.com/FacileStudio/porte/kit"

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

func newTestUserStore(t *testing.T, db *gorm.DB) *kit.Store {
	t.Helper()
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("db.DB: %v", err)
	}
	store, err := NewUserStore(context.Background(), sqlDB, nil)
	if err != nil {
		t.Fatalf("NewUserStore: %v", err)
	}
	return store
}
