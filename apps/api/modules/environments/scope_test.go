package environments

import (
	"context"
	stderrors "errors"
	"testing"

	"github.com/FacileStudio/Casier/apps/api/internal/authcontext"
	"github.com/FacileStudio/tronc/errors"
)

func contextWithScope(scope *authcontext.TokenScope) context.Context {
	return authcontext.WithIdentity(context.Background(), authcontext.Identity{UserID: "1", Scope: scope})
}

func TestScopeAllowsEnvironment(t *testing.T) {
	cases := []struct {
		name    string
		ctx     context.Context
		envSlug string
		want    bool
	}{
		{"no identity at all", context.Background(), "prod", true},
		{"user session has no scope", contextWithScope(nil), "prod", true},
		{"token scoped to every environment", contextWithScope(&authcontext.TokenScope{AllEnvironments: true}), "prod", true},
		{"token whose environments were all deleted reaches none", contextWithScope(&authcontext.TokenScope{}), "prod", false},
		{"a literal star is a slug, not a wildcard", contextWithScope(&authcontext.TokenScope{Environments: []string{"*"}}), "prod", false},
		{"scoped token on its own environment", contextWithScope(&authcontext.TokenScope{Environments: []string{"dev"}}), "dev", true},
		{"scoped token on another environment", contextWithScope(&authcontext.TokenScope{Environments: []string{"dev"}}), "prod", false},
		{"multi environment token", contextWithScope(&authcontext.TokenScope{Environments: []string{"dev", "staging"}}), "staging", true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := scopeAllowsEnvironment(c.ctx, c.envSlug); got != c.want {
				t.Fatalf("scopeAllowsEnvironment(%q) = %v, want %v", c.envSlug, got, c.want)
			}
		})
	}
}

func TestScopeAllowsWrite(t *testing.T) {
	cases := []struct {
		name string
		ctx  context.Context
		want bool
	}{
		{"no identity at all", context.Background(), true},
		{"user session has no scope", contextWithScope(nil), true},
		{"read only token", contextWithScope(&authcontext.TokenScope{Permissions: []string{"read"}}), false},
		{"token with no permissions", contextWithScope(&authcontext.TokenScope{}), false},
		{"write token", contextWithScope(&authcontext.TokenScope{Permissions: []string{"read", "write"}}), true},
		{"wildcard token", contextWithScope(&authcontext.TokenScope{Permissions: []string{"*"}}), true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := scopeAllowsWrite(c.ctx); got != c.want {
				t.Fatalf("scopeAllowsWrite() = %v, want %v", got, c.want)
			}
		})
	}
}

func TestRequireWriteScope(t *testing.T) {
	if err := requireWriteScope(contextWithScope(nil)); err != nil {
		t.Fatalf("user session rejected: %v", err)
	}

	err := requireWriteScope(contextWithScope(&authcontext.TokenScope{Permissions: []string{"read"}}))
	if err == nil {
		t.Fatal("read-only token was allowed to write")
	}
	var appErr *errors.Error
	if !stderrors.As(err, &appErr) || appErr.Code != "permission_denied" {
		t.Fatalf("error = %v, want permission_denied", err)
	}
}

func TestSlugify(t *testing.T) {
	cases := map[string]string{
		"Production":   "production",
		"  Staging  ":  "staging",
		"Pre Prod #2":  "pre-prod-2",
		"!!!":          "",
		"dev/eu-west1": "dev-eu-west1",
	}

	for input, want := range cases {
		if got := slugify(input); got != want {
			t.Errorf("slugify(%q) = %q, want %q", input, got, want)
		}
	}
}
