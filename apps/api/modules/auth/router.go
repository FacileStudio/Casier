package auth

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/FacileStudio/Casier/apps/api/internal/env"
	"github.com/FacileStudio/Casier/apps/api/internal/middleware"
	"github.com/FacileStudio/porte/oidc"
	"github.com/FacileStudio/porte/session"
	"github.com/FacileStudio/tronc/httpjson"

	"github.com/go-chi/chi/v5"
)

// RegisterRoutes mounts the auth endpoints, gated by appEnv's SSO policy.
func RegisterRoutes(router chi.Router, sessions *session.Manager, kit *oidc.Kit, service *Service, appEnv env.Config) {
	authLimiter := middleware.NewRateLimiter(5, 1*time.Minute)

	sessions.Mount(router)

	if kit != nil {
		kit.Mount(router)
	} else {
		router.Get("/auth/config", func(w http.ResponseWriter, r *http.Request) {
			httpjson.WriteJSON(w, http.StatusOK, map[string]any{
				"sso_only":               appEnv.SSOOnly,
				"oidc_enabled":           false,
				"password_reset_enabled": service.mailer != nil,
			})
		})
	}

	router.Route("/auth", func(router chi.Router) {
		if !appEnv.SSOOnly && service.passwords != nil {
			router.With(middleware.RateLimit(authLimiter)).Post("/register", func(w http.ResponseWriter, r *http.Request) {
				var req RegisterRequest
				if err := httpjson.DecodeJSON(w, r, &req); err != nil {
					httpjson.WriteError(w, err)
					return
				}
				resp, err := service.controller.register(r.Context(), w, r, &req)
				if err != nil {
					httpjson.WriteError(w, err)
					return
				}
				httpjson.WriteJSON(w, http.StatusCreated, resp)
			})

			router.With(middleware.RateLimit(authLimiter)).Post("/login", func(w http.ResponseWriter, r *http.Request) {
				var req LoginRequest
				if err := httpjson.DecodeJSON(w, r, &req); err != nil {
					httpjson.WriteError(w, err)
					return
				}
				resp, err := service.controller.login(r.Context(), w, r, &req)
				if err != nil {
					httpjson.WriteError(w, err)
					return
				}
				httpjson.WriteJSON(w, http.StatusOK, resp)
			})

			router.With(middleware.RateLimit(authLimiter)).Post("/forgot-password", func(w http.ResponseWriter, r *http.Request) {
				var req ForgotPasswordRequest
				if err := httpjson.DecodeJSON(w, r, &req); err != nil {
					httpjson.WriteError(w, err)
					return
				}
				if err := service.forgotPassword(r.Context(), req.Email); err != nil {
					httpjson.WriteError(w, err)
					return
				}
				httpjson.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
			})

			router.With(middleware.RateLimit(authLimiter)).Post("/reset-password", func(w http.ResponseWriter, r *http.Request) {
				var req ResetPasswordRequest
				if err := httpjson.DecodeJSON(w, r, &req); err != nil {
					httpjson.WriteError(w, err)
					return
				}
				if err := service.resetPassword(r.Context(), req.Token, req.Password); err != nil {
					httpjson.WriteError(w, err)
					return
				}
				httpjson.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
			})
		}

		router.With(middleware.RequireAuthOrToken(service)).Get("/me", func(w http.ResponseWriter, r *http.Request) {
			resp, err := service.controller.me(r.Context())
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusOK, resp)
		})
	})
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
