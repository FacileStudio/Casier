package environments

import (
	"context"
	stderrors "errors"
	"reflect"
	"testing"

	"github.com/FacileStudio/Casier/apps/api/internal/authcontext"
	"github.com/FacileStudio/tronc/errors"
)

// fakeResolver stands in for the projects module: it records the roles the
// controller demanded and answers with whatever the case under test wants, so
// the authorization decisions can be exercised without a database.
type fakeResolver struct {
	calls   [][]string
	results []error
}

func (resolver *fakeResolver) ResolveProjectID(ctx context.Context, slug string, userID int64, requiredRoles ...string) (int64, error) {
	resolver.calls = append(resolver.calls, requiredRoles)
	index := len(resolver.calls) - 1
	if index < len(resolver.results) && resolver.results[index] != nil {
		return 0, resolver.results[index]
	}
	return 1, nil
}

func controllerWithResolver(resolver *fakeResolver) *Controller {
	service := &Service{projectResolver: resolver}
	service.controller = newController(service)
	return service.controller
}

func codeOf(t *testing.T, err error) string {
	t.Helper()
	var appErr *errors.Error
	if !stderrors.As(err, &appErr) {
		t.Fatalf("error %v is not an app error", err)
	}
	return appErr.Code
}

func boolPtr(value bool) *bool { return &value }

func TestProtectionRolesExcludeEveryoneBelowAdmin(t *testing.T) {
	cases := map[string]bool{
		"owner":  true,
		"admin":  true,
		"member": false,
		"viewer": false,
		"":       false,
	}

	for role, want := range cases {
		t.Run(role, func(t *testing.T) {
			if got := containsRole(protectionRoles, role); got != want {
				t.Fatalf("protectionRoles allows %q = %v, want %v", role, got, want)
			}
			if role != "owner" && containsRole(protectedDeleteRoles, role) {
				t.Fatalf("protectedDeleteRoles must be owner only, but allows %q", role)
			}
		})
	}
}

func containsRole(roles []string, role string) bool {
	for _, allowed := range roles {
		if allowed == role {
			return true
		}
	}
	return false
}

func TestMutationsDemandProtectionRoles(t *testing.T) {
	denied := errors.Forbidden("insufficient permissions")

	cases := []struct {
		name string
		run  func(controller *Controller) error
		want [][]string
	}{
		{
			name: "create",
			run: func(controller *Controller) error {
				_, err := controller.create(contextWithScope(nil), "app", &CreateRequest{Name: "Prod", Protected: true})
				return err
			},
			want: [][]string{protectionRoles},
		},
		{
			name: "update",
			run: func(controller *Controller) error {
				_, err := controller.update(contextWithScope(nil), "app", "prod", &UpdateRequest{Name: "Prod", Protected: boolPtr(false)})
				return err
			},
			want: [][]string{protectionRoles},
		},
		{
			name: "delete",
			run: func(controller *Controller) error {
				return controller.delete(contextWithScope(nil), "app", "prod")
			},
			want: [][]string{protectionRoles},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			resolver := &fakeResolver{results: []error{denied}}
			err := c.run(controllerWithResolver(resolver))
			if err == nil {
				t.Fatal("a member was allowed through")
			}
			if code := codeOf(t, err); code != "permission_denied" {
				t.Fatalf("error code = %q, want permission_denied", code)
			}
			if !reflect.DeepEqual(resolver.calls, c.want) {
				t.Fatalf("resolver calls = %#v, want %#v", resolver.calls, c.want)
			}
		})
	}
}

func TestDeleteProbesOwnership(t *testing.T) {
	cases := []struct {
		name     string
		resolved error
		want     bool
		wantErr  bool
	}{
		{"owner", nil, true, false},
		{"admin is not an owner", errors.Forbidden("insufficient permissions"), false, false},
		{"a real failure is not a demotion", errors.Internal("db down", nil), false, true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			resolver := &fakeResolver{results: []error{c.resolved}}
			controller := controllerWithResolver(resolver)

			got, err := controller.callerIsOwner(contextWithScope(nil), "app", 1)
			if c.wantErr {
				if err == nil {
					t.Fatal("an internal failure was swallowed")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != c.want {
				t.Fatalf("callerIsOwner = %v, want %v", got, c.want)
			}
			if want := [][]string{protectedDeleteRoles}; !reflect.DeepEqual(resolver.calls, want) {
				t.Fatalf("resolver calls = %#v, want %#v", resolver.calls, want)
			}
		})
	}
}

func TestResolveProtectionLeavesAbsentFieldAlone(t *testing.T) {
	cases := []struct {
		name      string
		requested *bool
		current   bool
		want      bool
	}{
		{"absent keeps protection on", nil, true, true},
		{"absent keeps protection off", nil, false, false},
		{"explicit true protects", boolPtr(true), false, true},
		{"explicit false unprotects", boolPtr(false), true, false},
		{"explicit true on a protected environment is a no-op", boolPtr(true), true, true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := resolveProtection(c.requested, c.current); got != c.want {
				t.Fatalf("resolveProtection = %v, want %v", got, c.want)
			}
		})
	}
}

func TestProtectionChanges(t *testing.T) {
	cases := []struct {
		name      string
		requested *bool
		current   bool
		want      bool
	}{
		{"absent changes nothing", nil, true, false},
		{"restating the stored value changes nothing", boolPtr(true), true, false},
		{"protecting is a change", boolPtr(true), false, true},
		{"unprotecting is a change", boolPtr(false), true, true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := protectionChanges(c.requested, c.current); got != c.want {
				t.Fatalf("protectionChanges = %v, want %v", got, c.want)
			}
		})
	}
}

func TestAuthorizeProtectionChangeRefusesMachineTokens(t *testing.T) {
	fullToken := &authcontext.TokenScope{AllEnvironments: true, Permissions: []string{"*"}}

	cases := []struct {
		name      string
		ctx       context.Context
		requested *bool
		current   bool
		wantCode  string
	}{
		{"session may unprotect", contextWithScope(nil), boolPtr(false), true, ""},
		{"session may protect", contextWithScope(nil), boolPtr(true), false, ""},
		{"widest token may not unprotect", contextWithScope(fullToken), boolPtr(false), true, "permission_denied"},
		{"widest token may not protect", contextWithScope(fullToken), boolPtr(true), false, "permission_denied"},
		{"token that names no change passes", contextWithScope(fullToken), nil, true, ""},
		{"token restating the stored value passes", contextWithScope(fullToken), boolPtr(true), true, ""},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := authorizeProtectionChange(c.ctx, c.requested, c.current)
			if c.wantCode == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if code := codeOf(t, err); code != c.wantCode {
				t.Fatalf("error code = %q, want %q", code, c.wantCode)
			}
		})
	}
}

func TestAuthorizeProtectedDelete(t *testing.T) {
	fullToken := &authcontext.TokenScope{AllEnvironments: true, Permissions: []string{"*"}}

	cases := []struct {
		name          string
		ctx           context.Context
		protected     bool
		callerIsOwner bool
		wantCode      string
	}{
		{"admin deletes an unprotected environment", contextWithScope(nil), false, false, ""},
		{"token deletes an unprotected environment", contextWithScope(fullToken), false, false, ""},
		{"owner deletes a protected environment", contextWithScope(nil), true, true, ""},
		{"admin cannot delete a protected environment", contextWithScope(nil), true, false, "permission_denied"},
		{"token cannot delete a protected environment", contextWithScope(fullToken), true, true, "permission_denied"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := authorizeProtectedDelete(c.ctx, c.protected, c.callerIsOwner)
			if c.wantCode == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if code := codeOf(t, err); code != c.wantCode {
				t.Fatalf("error code = %q, want %q", code, c.wantCode)
			}
		})
	}
}

func TestCallerIsMachineToken(t *testing.T) {
	cases := []struct {
		name string
		ctx  context.Context
		want bool
	}{
		{"no identity", context.Background(), false},
		{"user session", contextWithScope(nil), false},
		{"read only token", contextWithScope(&authcontext.TokenScope{Permissions: []string{"read"}}), true},
		{"write token", contextWithScope(&authcontext.TokenScope{AllEnvironments: true, Permissions: []string{"write"}}), true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := callerIsMachineToken(c.ctx); got != c.want {
				t.Fatalf("callerIsMachineToken = %v, want %v", got, c.want)
			}
		})
	}
}
