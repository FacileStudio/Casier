package auth

import (
	"context"
	"testing"

	dbconfig "github.com/FacileStudio/Casier/apps/api/internal/testdb"
	"github.com/FacileStudio/porte/kit"
	"github.com/FacileStudio/tronc/testdb"

	"gorm.io/gorm"
)

func resetTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	connection, ok := testdb.URL()
	if !ok {
		testdb.Announce(dbconfig.Bootstrap)
		t.Skip(testdb.SkipReason(dbconfig.Bootstrap))
	}
	db, err := testdb.Open(connection, dbconfig.Config)
	if err != nil {
		t.Fatalf("testdb.Open: %v", err)
	}
	if err := testdb.Truncate(db, dbconfig.Config); err != nil {
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
