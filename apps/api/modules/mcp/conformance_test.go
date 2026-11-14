package mcp

import (
	"fmt"
	"testing"

	"github.com/FacileStudio/tronc/mcp/mcptest"
)

// TestConformance runs tronc's conformance kit on the endpoint, with the delete as the
// sensitive tool. The reveal confirms on the page only, which reveal_test.go covers.
func TestConformance(t *testing.T) {
	f := newFixture(t)
	mcptest.Run(t, mcptest.Config{
		URL:   f.resource,
		Token: func(t testing.TB, scopes ...string) string { return f.token(t, ownerSubject, scopes...) },
		Confirm: []mcptest.ConfirmCase{
			{Tool: toolDelete, Setup: f.freshTrash},
		},
	})
}

// freshTrash adds a secret and reports 1 once it is in the trash.
func (f *fixture) freshTrash(t testing.TB) (map[string]any, func() int) {
	key := fmt.Sprintf("TRASH_%d", f.count(t, `SELECT count(*) FROM secrets`))
	id := f.secret(t, key, "plaintext-trash")
	return prod(map[string]any{"key": key}), func() int {
		return f.count(t, `SELECT count(*) FROM secrets WHERE id = $1 AND deleted_at IS NOT NULL`, id)
	}
}
