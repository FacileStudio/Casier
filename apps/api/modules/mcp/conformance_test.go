package mcp

import (
	"fmt"
	"testing"

	"github.com/FacileStudio/tronc/mcp/mcptest"
)

// TestConformance runs tronc's conformance kit on the endpoint, with the reveal and the
// delete as the sensitive tools.
func TestConformance(t *testing.T) {
	f := newFixture(t)
	mcptest.Run(t, mcptest.Config{
		URL:   f.resource,
		Token: func(t testing.TB, scopes ...string) string { return f.token(t, ownerSubject, scopes...) },
		Confirm: []mcptest.ConfirmCase{
			{Tool: toolDelete, Setup: f.freshTrash},
			{Tool: toolReveal, Setup: f.freshReveal},
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

// freshReveal adds a secret and reports how many times its value was handed over.
func (f *fixture) freshReveal(t testing.TB) (map[string]any, func() int) {
	key := fmt.Sprintf("REVEAL_%d", f.count(t, `SELECT count(*) FROM secrets`))
	f.secret(t, key, "plaintext-reveal")
	return prod(map[string]any{"keys": []string{key}}), func() int {
		return f.count(t, `SELECT count(*) FROM mcp_actions WHERE outcome = $1 AND target LIKE $2`, outcomeRevealed, `%"`+key+`"%`)
	}
}
