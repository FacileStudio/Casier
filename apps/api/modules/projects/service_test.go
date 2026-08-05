package projects

import (
	"context"
	"testing"

	"github.com/FacileStudio/Casier/apps/api/internal/authcontext"
	"github.com/FacileStudio/tronc/errors"
)

func TestIsAdminEmail(t *testing.T) {
	cases := []struct {
		name    string
		admins  []string
		email   string
		allowed bool
	}{
		{name: "exact match", admins: []string{"ops@facile.studio"}, email: "ops@facile.studio", allowed: true},
		{name: "case insensitive config", admins: []string{"OPS@Facile.Studio"}, email: "ops@facile.studio", allowed: true},
		{name: "case insensitive caller", admins: []string{"ops@facile.studio"}, email: "Ops@Facile.Studio", allowed: true},
		{name: "whitespace in config", admins: []string{"  ops@facile.studio\t"}, email: "ops@facile.studio", allowed: true},
		{name: "whitespace in caller", admins: []string{"ops@facile.studio"}, email: " ops@facile.studio ", allowed: true},
		{name: "second entry", admins: []string{"a@facile.studio", "b@facile.studio"}, email: "b@facile.studio", allowed: true},
		{name: "non admin", admins: []string{"ops@facile.studio"}, email: "intern@facile.studio", allowed: false},
		{name: "empty list denies", admins: nil, email: "ops@facile.studio", allowed: false},
		{name: "empty list denies empty caller", admins: nil, email: "", allowed: false},
		{name: "empty caller", admins: []string{"ops@facile.studio"}, email: "", allowed: false},
		{name: "blank caller", admins: []string{"ops@facile.studio"}, email: "   ", allowed: false},
		{name: "blank config entry", admins: []string{"  "}, email: "", allowed: false},
		{name: "substring is not a match", admins: []string{"ops@facile.studio"}, email: "ops@facile.studio.evil.com", allowed: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isAdminEmail(tc.admins, tc.email); got != tc.allowed {
				t.Fatalf("isAdminEmail(%q, %q) = %v, want %v", tc.admins, tc.email, got, tc.allowed)
			}
		})
	}
}

func TestEnsureScopeAllowsProject(t *testing.T) {
	sessionContext := context.Background()
	scopedContext := authcontext.WithIdentity(context.Background(), authcontext.Identity{
		UserID: "1",
		Email:  "ci@facile.studio",
		Scope:  &authcontext.TokenScope{TokenID: 7, ProjectID: 42},
	})
	unscopedContext := authcontext.WithIdentity(context.Background(), authcontext.Identity{
		UserID: "1",
		Email:  "human@facile.studio",
	})

	cases := []struct {
		name      string
		ctx       context.Context
		projectID int64
		wantErr   bool
	}{
		{name: "no identity", ctx: sessionContext, projectID: 42},
		{name: "user session", ctx: unscopedContext, projectID: 99},
		{name: "token on its own project", ctx: scopedContext, projectID: 42},
		{name: "token on another project", ctx: scopedContext, projectID: 43, wantErr: true},
		{name: "token on unresolved project", ctx: scopedContext, projectID: 0, wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ensureScopeAllowsProject(tc.ctx, tc.projectID)
			if !tc.wantErr {
				if err != nil {
					t.Fatalf("ensureScopeAllowsProject: unexpected error %v", err)
				}
				return
			}
			if err == nil {
				t.Fatal("ensureScopeAllowsProject: expected an error")
			}
			if got := errors.Status(err); got != 404 {
				t.Fatalf("status = %d, want 404 so tokens cannot enumerate projects", got)
			}
		})
	}
}
