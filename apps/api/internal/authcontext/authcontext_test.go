package authcontext

import (
	"context"
	"reflect"
	"testing"
)

func TestParseScopeList(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		want []string
	}{
		{"empty means unset", "", nil},
		{"star is preserved", "*", []string{"*"}},
		{"single value", "dev", []string{"dev"}},
		{"trims whitespace", " dev , staging ", []string{"dev", "staging"}},
		{"drops empty entries", "dev,,staging,", []string{"dev", "staging"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			got := ParseScopeList(testCase.raw)
			if !reflect.DeepEqual(got, testCase.want) {
				t.Fatalf("ParseScopeList(%q) = %#v, want %#v", testCase.raw, got, testCase.want)
			}
		})
	}
}

func TestNilScopeAllowsEverything(t *testing.T) {
	var scope *TokenScope

	if !scope.AllowsProject(42) {
		t.Error("a user session must not be restricted to a project")
	}
	if !scope.AllowsEnvironment("prod") {
		t.Error("a user session must not be restricted to an environment")
	}
	if !scope.AllowsWrite() {
		t.Error("a user session must not be denied writes")
	}
}

func TestScopeAllowsProject(t *testing.T) {
	scope := &TokenScope{ProjectID: 7}

	if !scope.AllowsProject(7) {
		t.Error("token must reach the project it was minted for")
	}
	if scope.AllowsProject(8) {
		t.Error("token must not reach another project")
	}
	if scope.AllowsProject(0) {
		t.Error("an unresolved project id must not be allowed")
	}
}

func TestScopeAllowsEnvironment(t *testing.T) {
	cases := []struct {
		name         string
		all          bool
		environments string
		slug         string
		want         bool
	}{
		{"all environments reaches prod", true, "", "prod", true},
		{"no grants left reaches nothing", false, "", "prod", false},
		{"a star is a slug, not a wildcard", false, "*", "prod", false},
		{"listed slug allowed", false, "dev,staging", "dev", true},
		{"second listed slug allowed", false, "dev,staging", "staging", true},
		{"unlisted slug denied", false, "dev,staging", "prod", false},
		{"prefix is not a match", false, "dev", "development", false},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			scope := &TokenScope{AllEnvironments: testCase.all, Environments: ParseScopeList(testCase.environments)}
			if got := scope.AllowsEnvironment(testCase.slug); got != testCase.want {
				t.Fatalf("environments=%q slug=%q: got %v, want %v", testCase.environments, testCase.slug, got, testCase.want)
			}
		})
	}
}

func TestScopeAllowsWrite(t *testing.T) {
	cases := []struct {
		permissions string
		want        bool
	}{
		{"read", false},
		{"read,write", true},
		{"write", true},
		{"", false},
		{"*", true},
		{"admin", false},
	}

	for _, testCase := range cases {
		t.Run(testCase.permissions, func(t *testing.T) {
			scope := &TokenScope{Permissions: ParseScopeList(testCase.permissions)}
			if got := scope.AllowsWrite(); got != testCase.want {
				t.Fatalf("permissions=%q: got %v, want %v", testCase.permissions, got, testCase.want)
			}
		})
	}
}

func TestScopeFromContext(t *testing.T) {
	if scope := ScopeFromContext(context.Background()); scope != nil {
		t.Fatal("a context with no identity must yield no scope")
	}

	sessionCtx := WithIdentity(context.Background(), Identity{UserID: "1", Email: "a@b.c"})
	if scope := ScopeFromContext(sessionCtx); scope != nil {
		t.Fatal("a session identity must yield no scope")
	}

	tokenCtx := WithIdentity(context.Background(), Identity{UserID: "1", Scope: &TokenScope{TokenID: 3, ProjectID: 9}})
	scope := ScopeFromContext(tokenCtx)
	if scope == nil || scope.TokenID != 3 || scope.ProjectID != 9 {
		t.Fatalf("token identity yielded %#v", scope)
	}
}
