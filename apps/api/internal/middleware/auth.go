package middleware

import (
	"context"
	"net"
	"net/http"

	"github.com/FacileStudio/Casier/apps/api/internal/authcontext"
	"github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"
)

type Authenticator interface {
	Authenticate(context context.Context, authorization string) (string, any, error)
}

func RequireAuth(authService Authenticator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
			authorization := request.Header.Get("Authorization")
			if authorization == "" {
				if c, err := request.Cookie("casier_session"); err == nil && c.Value != "" {
					authorization = "Bearer " + c.Value
				}
			}
			userID, rawData, err := authService.Authenticate(request.Context(), authorization)
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			data, ok := rawData.(interface {
				GetEmail() string
				GetScope() *authcontext.TokenScope
			})
			if !ok {
				httpjson.WriteError(w, errors.Unauthorized("missing auth"))
				return
			}
			if data == nil {
				httpjson.WriteError(w, errors.Unauthorized("missing auth"))
				return
			}

			clientIP, _, _ := net.SplitHostPort(request.RemoteAddr)
			if clientIP == "" {
				clientIP = request.RemoteAddr
			}

			authContext := authcontext.WithIdentity(request.Context(), authcontext.Identity{
				UserID:   userID,
				Email:    data.GetEmail(),
				ClientIP: clientIP,
				Scope:    data.GetScope(),
			})
			next.ServeHTTP(w, request.WithContext(authContext))
		})
	}
}

// RequireUserSession rejects API-token callers. Project administration and
// access grants are a human's job: a CI credential moves secrets, it never
// decides who may read them. Shared by the projects and secrets modules so the
// two cannot drift on what "administration" excludes.
func RequireUserSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		if authcontext.ScopeFromContext(request.Context()) != nil {
			httpjson.WriteError(w, errors.Forbidden("API tokens cannot perform this operation"))
			return
		}
		next.ServeHTTP(w, request)
	})
}
