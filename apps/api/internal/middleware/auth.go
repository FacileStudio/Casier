package middleware

import (
	"context"
	"net/http"

	"github.com/FacileStudio/Casier/apps/api/internal/authcontext"
	"github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"
)

// Authenticator lets the middleware authenticate either a session, an API
// token, or a user identity — the one seam the porte kit and the token store
// both satisfy.
type Authenticator interface {
	RequireAuth(next http.Handler) http.Handler
	IdentityForUser(ctx context.Context, userID int64) (string, error)
	AuthenticateAPI(ctx context.Context, authorization string) (string, *authcontext.TokenScope, error)
}

// RequireAuth accepts either a user session or a scoped API token, exactly as
// RequireAuthOrToken does; it is that alias rather than a session-only guard.
func RequireAuth(authService Authenticator) func(http.Handler) http.Handler {
	return RequireAuthOrToken(authService)
}

// RequireAuthOrToken accepts either a user session (via cookie/bearer) or
// a scoped API token. API token callers get a TokenScope on the context;
// session callers get a nil scope which means unlimited by project role.
func RequireAuthOrToken(authService Authenticator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authorization := r.Header.Get("Authorization")
			if authorization != "" {
				userID, scope, err := authService.AuthenticateAPI(r.Context(), authorization)
				if err == nil {
					ctx := authcontext.WithIdentity(r.Context(), authcontext.Identity{
						UserID:   userID,
						ClientIP: authcontext.ClientIP(r),
						Scope:    scope,
					})
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}

			authService.RequireAuth(next).ServeHTTP(w, r)
		})
	}
}

// RequireUserSession rejects a scoped API token, so the route is reachable
// only by a real logged-in user rather than a machine credential.
func RequireUserSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if authcontext.ScopeFromContext(r.Context()) != nil {
			httpjson.WriteError(w, errors.Forbidden("API tokens cannot perform this operation"))
			return
		}
		next.ServeHTTP(w, r)
	})
}
