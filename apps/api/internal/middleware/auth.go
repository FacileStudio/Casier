package middleware

import (
	"context"
	"net"
	"net/http"

	"clef/internal/authcontext"
	"clef/internal/errors"
	"clef/internal/httpjson"
)

type Authenticator interface {
	Authenticate(context context.Context, authorization string) (string, any, error)
}

func RequireAuth(authService Authenticator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
			authorization := request.Header.Get("Authorization")
			if authorization == "" {
				if c, err := request.Cookie("clef_session"); err == nil && c.Value != "" {
					authorization = "Bearer " + c.Value
				}
			}
			userID, rawData, err := authService.Authenticate(request.Context(), authorization)
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			data, ok := rawData.(interface{ GetEmail() string })
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
			})
			next.ServeHTTP(w, request.WithContext(authContext))
		})
	}
}
