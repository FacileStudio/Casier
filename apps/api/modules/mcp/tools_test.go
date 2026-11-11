package mcp

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/FacileStudio/tronc/mcp/mcptest"
)

// plaintexts are every value the fixture stores. No tool but a confirmed reveal may
// ever answer one of them.
var plaintexts = []string{"postgres://plaintext-db", "plaintext-locked", "plaintext-new", "plaintext-v2"}

// leaks reports the first plaintext value found in v's JSON.
func leaks(v any) string {
	raw, _ := json.Marshal(v)
	for _, value := range plaintexts {
		if strings.Contains(string(raw), value) {
			return value
		}
	}
	return ""
}

// TestNoToolButRevealReturnsAValue drives every read and write tool and checks no
// answer carries a plaintext value.
func TestNoToolButRevealReturnsAValue(t *testing.T) {
	f := newFixture(t)
	session := f.session(t, ownerSubject)
	calls := []struct {
		tool string
		args map[string]any
	}{
		{toolProjects, map[string]any{}},
		{toolSecrets, prod(map[string]any{"response_format": "detailed"})},
		{toolSet, prod(map[string]any{"key": "NEW_KEY", "value": "plaintext-new", "tags": []string{"api"}})},
		{toolSet, prod(map[string]any{"key": "DATABASE_URL", "value": "plaintext-v2"})},
		{toolSecret, prod(map[string]any{"key": "DATABASE_URL"})},
		{toolDiff, map[string]any{"project": projectSlug, "from": "staging", "to": envSlug}},
		{toolTag, prod(map[string]any{"key": "DATABASE_URL", "tags": []string{"db"}})},
		{toolRollback, prod(map[string]any{"key": "DATABASE_URL", "version": 1})},
		{toolTrash, prod(nil)},
		{toolAudit, map[string]any{"project": projectSlug}},
	}
	for _, c := range calls {
		result := mcptest.Call(t, session, c.tool, c.args)
		if result.IsError {
			t.Fatalf("%s failed: %s", c.tool, mcptest.Text(result))
		}
		if value := leaks(result); value != "" {
			t.Errorf("%s answered the plaintext %q", c.tool, value)
		}
	}
}

// TestSetSecretCreatesThenVersions checks set_secret creates a missing key and writes a
// new version of an existing one.
func TestSetSecretCreatesThenVersions(t *testing.T) {
	f := newFixture(t)
	session := f.session(t, ownerSubject)
	var created, updated writeOutput
	call(t, session, toolSet, prod(map[string]any{"key": "NEW_KEY", "value": "plaintext-new"}), &created)
	call(t, session, toolSet, prod(map[string]any{"key": "NEW_KEY", "value": "plaintext-v2"}), &updated)
	if created.Status != statusCreated || updated.Status != statusUpdated || updated.Secret == nil || updated.Secret.Version != 2 {
		t.Fatalf("want created then version 2, got %+v then %+v", created, updated)
	}
}

// TestAnACLHidesASecretFromEveryTool checks a member locked out of one secret cannot
// list, inspect or ask to reveal it, exactly as in the web app.
func TestAnACLHidesASecretFromEveryTool(t *testing.T) {
	f := newFixture(t)
	session := f.session(t, memberSubject)
	var list secretsOutput
	call(t, session, toolSecrets, prod(nil), &list)
	for _, secret := range list.Secrets {
		if secret.Key == lockedKey {
			t.Fatal("the locked secret is listed to a member it is hidden from")
		}
	}
	refused(t, session, toolSecret, prod(map[string]any{"key": lockedKey}))
	refused(t, session, toolReveal, prod(map[string]any{"keys": []string{lockedKey}}))
	if n := f.count(t, `SELECT count(*) FROM mcp_actions`); n != 0 {
		t.Fatalf("a reveal of a hidden secret must not open an action, got %d", n)
	}
}

// TestAStrangerSeesNoProject checks a person outside the project reaches nothing.
func TestAStrangerSeesNoProject(t *testing.T) {
	f := newFixture(t)
	session := f.session(t, strangerSubject)
	var out projectsOutput
	call(t, session, toolProjects, map[string]any{}, &out)
	if len(out.Projects) != 0 {
		t.Fatalf("a stranger must see no project, got %+v", out.Projects)
	}
	refused(t, session, toolSecrets, prod(nil))
	refused(t, session, toolSet, prod(map[string]any{"key": "X", "value": "y"}))
}

// TestListProjectsNamesEnvironments checks the first call gives every slug the other
// tools take.
func TestListProjectsNamesEnvironments(t *testing.T) {
	f := newFixture(t)
	var out projectsOutput
	call(t, f.session(t, ownerSubject), toolProjects, map[string]any{}, &out)
	if len(out.Projects) != 1 || out.Projects[0].Slug != projectSlug || out.Projects[0].Role != "owner" || len(out.Projects[0].Environments) != 3 {
		t.Fatalf("want the web project with its three environments, got %+v", out.Projects)
	}
}
