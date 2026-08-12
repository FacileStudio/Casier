// Package testdb connects a test to a real PostgreSQL database.
//
// It is deliberately Postgres and nothing else. Casier's SQL is Postgres SQL —
// string_to_array, LATERAL, ILIKE, bool_or, catalog queries in the migrations —
// so a test against SQLite would pass while proving nothing about what ships.
package testdb

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/FacileStudio/Casier/apps/api/internal/database"
	"github.com/FacileStudio/Casier/apps/api/schemas"

	"gorm.io/gorm"
)

// EnvVar names the connection string the database-backed tests read.
const EnvVar = "TEST_DATABASE_URL"

// SkipReason is the message a skipped test prints. It spells out the whole
// command because a gate that rots does so silently: whoever sees this message
// has to be able to turn the tests on without reading any other file.
const SkipReason = EnvVar + ` is not set, so the HTTP authorization matrix did not run.

Start the repo-local Postgres and point the tests at a scratch database:

    sh scripts/postgres-up.sh
    createdb -p 5433 casier_test
    export ` + EnvVar + `='postgres://localhost:5433/casier_test?sslmode=disable'

'sh scripts/check.sh' does all of that for you when pg_ctl is on PATH.`

// URL returns the configured connection string, or false when the tests should
// skip.
func URL() (string, bool) {
	value := strings.TrimSpace(os.Getenv(EnvVar))
	return value, value != ""
}

// Announce writes the skip banner to stderr. t.Skip only shows under -v, and a
// gate nobody notices is off is worse than no gate at all.
func Announce() {
	fmt.Fprintf(os.Stderr, "\n=== SKIPPED: %s ===\n%s\n\n", EnvVar, SkipReason)
}

// Open connects to url and brings the schema up to date exactly as main.go
// does, so the tests run against the migrations that ship.
//
// Each test package gets its own Postgres schema, named after its test binary.
// `go test ./...` runs packages concurrently against one TEST_DATABASE_URL, and
// two packages truncating the same tables deadlock and steal each other's
// fixtures. Isolating by schema keeps that concurrency instead of forcing the
// whole suite through -p 1.
// Open returns a scoped, migrated connection. The search_path has to travel in
// the connection string rather than a SET, because GORM hands out pooled
// connections and a SET would bind one of them — every other query in the
// package would quietly run against public instead of the test's schema.
func Open(url string) (*gorm.DB, error) {
	schema := schemaName()

	admin, err := database.Open(url)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", EnvVar, err)
	}
	if err := admin.Exec(`CREATE SCHEMA IF NOT EXISTS "` + schema + `"`).Error; err != nil {
		return nil, fmt.Errorf("create schema %s: %w", schema, err)
	}
	if handle, err := admin.DB(); err == nil {
		_ = handle.Close()
	}

	scoped, err := database.Open(withSearchPath(url, schema))
	if err != nil {
		return nil, fmt.Errorf("open %s scoped to %s: %w", EnvVar, schema, err)
	}
	if err := schemas.Migrate(scoped); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return scoped, nil
}

func withSearchPath(url, schema string) string {
	separator := "?"
	if strings.Contains(url, "?") {
		separator = "&"
	}
	return url + separator + "search_path=" + schema
}

// schemaName derives a stable, unique schema per test binary. os.Args[0] is the
// compiled binary, whose base name is the package under test.
func schemaName() string {
	base := strings.TrimSuffix(filepath.Base(os.Args[0]), ".test")
	safe := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			return r
		}
		if r >= 'A' && r <= 'Z' {
			return r + 32
		}
		return '_'
	}, base)
	if safe == "" {
		safe = "anonymous"
	}
	return "casier_test_" + safe
}

// Truncate empties every table in the current schema and restarts the identity
// sequences, so a re-seeded fixture always lands on the same primary keys.
//
// Truncating and re-seeding is used rather than wrapping each case in a
// transaction because the service layer opens transactions of its own on the
// shared handle, and its fire-and-forget goroutines (last_used_at, the session
// sliding window) would use that same connection concurrently.
func Truncate(db *gorm.DB) error {
	var tables []string
	err := db.Raw(`SELECT tablename FROM pg_tables WHERE schemaname = current_schema()`).Scan(&tables).Error
	if err != nil {
		return fmt.Errorf("list tables: %w", err)
	}
	if len(tables) == 0 {
		return nil
	}

	quoted := make([]string, len(tables))
	for i, table := range tables {
		quoted[i] = `"` + table + `"`
	}
	statement := "TRUNCATE TABLE " + strings.Join(quoted, ", ") + " RESTART IDENTITY CASCADE"
	if err := db.Exec(statement).Error; err != nil {
		return fmt.Errorf("truncate: %w", err)
	}
	return nil
}
