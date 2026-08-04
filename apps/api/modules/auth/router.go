package auth

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"casier/internal/env"
	"casier/internal/httpjson"
	"casier/internal/middleware"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(router chi.Router, service *Service, appEnv env.Config) {
	oidcEnabled := appEnv.OIDC != nil
	authLimiter := middleware.NewRateLimiter(5, 1*time.Minute)

	router.Route("/auth", func(router chi.Router) {
		router.Get("/config", func(w http.ResponseWriter, r *http.Request) {
			httpjson.WriteJSON(w, http.StatusOK, map[string]bool{
				"sso_only":     appEnv.SSOOnly,
				"oidc_enabled": oidcEnabled,
			})
		})

		if !appEnv.SSOOnly {
			router.With(middleware.RateLimit(authLimiter)).Post("/register", func(w http.ResponseWriter, request *http.Request) {
				var req RegisterRequest
				if err := httpjson.DecodeJSON(w, request, &req); err != nil {
					httpjson.WriteError(w, err)
					return
				}
				resp, err := service.controller.register(request.Context(), &req)
				if err != nil {
					httpjson.WriteError(w, err)
					return
				}
				SetSessionCookie(w, resp.Token)
				httpjson.WriteJSON(w, http.StatusCreated, resp)
			})

			router.With(middleware.RateLimit(authLimiter)).Post("/login", func(w http.ResponseWriter, request *http.Request) {
				var req LoginRequest
				if err := httpjson.DecodeJSON(w, request, &req); err != nil {
					httpjson.WriteError(w, err)
					return
				}
				resp, err := service.controller.login(request.Context(), &req)
				if err != nil {
					httpjson.WriteError(w, err)
					return
				}
				SetSessionCookie(w, resp.Token)
				httpjson.WriteJSON(w, http.StatusOK, resp)
			})
		}

		router.With(middleware.RateLimit(authLimiter)).Post("/forgot-password", func(w http.ResponseWriter, request *http.Request) {
			var req ForgotPasswordRequest
			if err := httpjson.DecodeJSON(w, request, &req); err != nil {
				httpjson.WriteError(w, err)
				return
			}
			if err := service.controller.forgotPassword(request.Context(), &req); err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
		})

		router.With(middleware.RateLimit(authLimiter)).Post("/reset-password", func(w http.ResponseWriter, request *http.Request) {
			var req ResetPasswordRequest
			if err := httpjson.DecodeJSON(w, request, &req); err != nil {
				httpjson.WriteError(w, err)
				return
			}
			if err := service.controller.resetPassword(request.Context(), &req); err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
		})

		router.With(middleware.RequireAuth(service)).Post("/logout", func(w http.ResponseWriter, request *http.Request) {
			authorization := request.Header.Get("Authorization")
			if authorization == "" {
				if c, err := request.Cookie(SessionCookieName); err == nil {
					authorization = "Bearer " + c.Value
				}
			}
			if err := service.logoutUser(request.Context(), authorization); err != nil {
				httpjson.WriteError(w, err)
				return
			}
			ClearSessionCookie(w)
			w.WriteHeader(http.StatusNoContent)
		})

		router.With(middleware.RequireAuth(service)).Get("/me", func(w http.ResponseWriter, request *http.Request) {
			resp, err := service.controller.me(request.Context())
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusOK, resp)
		})

		if oidcEnabled {
			oidc, err := newOIDCHandler(context.Background(), appEnv.OIDC, service)
			if err != nil {
				slog.Error("failed to initialize OIDC provider", slog.Any("error", err))
			} else {
				router.Get("/oidc", oidc.login)
				router.Get("/oidc/callback", oidc.callback)
				router.With(middleware.RequireAuth(service)).Post("/sync-profile", oidc.syncProfile)
			}
		}
	})
}
