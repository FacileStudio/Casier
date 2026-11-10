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

// Routes is what RegisterRoutes mounts. OIDC, Deletions and Invitations are
// nil when no issuer is configured.
//
// Deletions is Registre's signed user.deleted push, mounted at
// /api/registre/deletions, which is directory.DeletionsPath, outside the
// session middleware: the handler verifies the notice's signature itself.
// Invitations is porte/kit's user-menu count, behind a user session only.
type Routes struct {
	Sessions    *session.Manager
	OIDC        *oidc.Kit
	Service     *Service
	Deletions   http.Handler
	Invitations *kit.InvitationsConfig
}

// RegisterRoutes mounts the auth endpoints. Sign-in is Registre's OIDC flow
// only; the API-token path lives in middleware.RequireAuthOrToken and needs no
// route here. porte/kit mounts the session and OIDC routes. Without an OIDC
// kit, Casier answers GET /auth/config itself, because the CLI reads it to
// learn that SSO is off.
func RegisterRoutes(router chi.Router, routes Routes) error {
	if err := kit.Mount(router, kit.Config{Sessions: routes.Sessions, OIDC: routes.OIDC}); err != nil {
		return err
	}
	if routes.OIDC == nil {
		router.Get("/auth/config", func(w http.ResponseWriter, r *http.Request) {
			httpjson.WriteJSON(w, http.StatusOK, map[string]any{"sso_only": true, "oidc_enabled": false})
		})
	}
	if routes.Deletions != nil {
		router.Method(http.MethodPost, "/registre/deletions", routes.Deletions)
	}
	if routes.Invitations != nil {
		if err := kit.MountInvitations(router.With(routes.Service.RequireAuth), *routes.Invitations); err != nil {
			return err
		}
	}
	router.With(middleware.RequireAuthOrToken(routes.Service)).Get("/auth/me", routes.Service.controller.handleMe)
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
