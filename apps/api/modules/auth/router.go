package auth

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/FacileStudio/Casier/apps/api/internal/middleware"
	"github.com/FacileStudio/porte/kit"
	"github.com/FacileStudio/porte/oidc"
	"github.com/FacileStudio/porte/session"
	"github.com/FacileStudio/tronc/httpjson"

	"github.com/go-chi/chi/v5"
)

// RegisterRoutes mounts the auth endpoints. Sign-in is Registre's OIDC flow
// only; the API-token path lives in middleware.RequireAuthOrToken and needs no
// route here. porte/kit mounts the session and OIDC routes. Without an OIDC
// kit, Casier answers GET /auth/config itself, because the CLI reads it to
// learn that SSO is off.
//
// deletions is Registre's signed user.deleted push, nil when OIDC is not
// configured. It is mounted at /api/registre/deletions, which is
// directory.DeletionsPath, outside the session middleware: the handler
// verifies the notice's signature itself.
func RegisterRoutes(router chi.Router, sessions *session.Manager, oidcKit *oidc.Kit, service *Service, deletions http.Handler) error {
	if err := kit.Mount(router, kit.Config{Sessions: sessions, OIDC: oidcKit}); err != nil {
		return err
	}
	if oidcKit == nil {
		router.Get("/auth/config", func(w http.ResponseWriter, r *http.Request) {
			httpjson.WriteJSON(w, http.StatusOK, map[string]any{
				"sso_only":     true,
				"oidc_enabled": false,
			})
		})
	}

	if deletions != nil {
		router.Method(http.MethodPost, "/registre/deletions", deletions)
	}

	router.With(middleware.RequireAuthOrToken(service)).Get("/auth/me", func(w http.ResponseWriter, r *http.Request) {
		resp, err := service.controller.me(r.Context())
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})
	return nil
}

func sweepLoop(ctx context.Context, sessions *session.Manager, logger *slog.Logger) {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			n, err := sessions.Sweep(context.WithoutCancel(ctx))
			if err != nil {
				logger.Error("session sweep failed", slog.Any("error", err))
			} else if n > 0 {
				logger.Info("swept expired sessions", slog.Int64("count", n))
			}
		}
	}
}
