package authcontext

import (
	"context"
	"strings"
)

// TokenScope is the capability an API token carries. A nil scope means the
// caller is a user session, which is bounded by project roles alone; every
// method is nil-safe and answers "allowed" so session callers stay unaffected.
type TokenScope struct {
	TokenID      int64
	ProjectID    int64
	Environments []string
	Permissions  []string
}

func (scope *TokenScope) AllowsProject(projectID int64) bool {
	if scope == nil {
		return true
	}
	return scope.ProjectID == projectID
}

func (scope *TokenScope) AllowsEnvironment(slug string) bool {
	if scope == nil || len(scope.Environments) == 0 {
		return true
	}
	for _, allowed := range scope.Environments {
		if allowed == "*" || allowed == slug {
			return true
		}
	}
	return false
}

// AllowsWrite fails closed where AllowsEnvironment fails open, and the
// asymmetry is deliberate: an unset environment list is how every token minted
// before scoping existed is stored, so it has to mean "every environment",
// while an unset permission list must never be read as "may write".
func (scope *TokenScope) AllowsWrite() bool {
	if scope == nil {
		return true
	}
	for _, permission := range scope.Permissions {
		if permission == "write" || permission == "*" {
			return true
		}
	}
	return false
}

// ParseScopeList splits the comma-separated form the API token columns are
// stored in, preserving a literal "*" so both Allows* checks can see it.
func ParseScopeList(raw string) []string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil
	}

	parts := strings.Split(trimmed, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			values = append(values, part)
		}
	}
	return values
}

type Identity struct {
	UserID   string
	Email    string
	ClientIP string
	Scope    *TokenScope
}

type contextKey struct{}

func WithIdentity(parentContext context.Context, identity Identity) context.Context {
	return context.WithValue(parentContext, contextKey{}, identity)
}

func IdentityFromContext(parentContext context.Context) (Identity, bool) {
	identity, ok := parentContext.Value(contextKey{}).(Identity)
	return identity, ok
}

// ScopeFromContext returns the calling token's scope, or nil for a user session.
func ScopeFromContext(parentContext context.Context) *TokenScope {
	identity, ok := IdentityFromContext(parentContext)
	if !ok {
		return nil
	}
	return identity.Scope
}
