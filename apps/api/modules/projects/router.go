package projects

import (
	"net/http"
	"strconv"

	"github.com/FacileStudio/Casier/apps/api/internal/authcontext"
	"github.com/FacileStudio/Casier/apps/api/internal/crypto"
	"github.com/FacileStudio/Casier/apps/api/internal/env"
	"github.com/FacileStudio/Casier/apps/api/internal/middleware"
	"github.com/FacileStudio/Casier/apps/api/modules/auth"
	"github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"

	"github.com/go-chi/chi/v5"
)

// RegisterRoutes mounts the project endpoints.
func RegisterRoutes(router chi.Router, service *Service, authService *auth.Service, appEnv env.Config) {
	router.With(middleware.RequireAuth(authService), middleware.RequireUserSession).Post("/admin/rotate-key", func(w http.ResponseWriter, request *http.Request) {
		identity, ok := authcontext.IdentityFromContext(request.Context())
		if !ok {
			httpjson.WriteError(w, errors.Unauthorized("missing auth"))
			return
		}
		if !isAdminEmail(appEnv.AdminEmails, identity.Email) {
			httpjson.WriteError(w, errors.Forbidden("root key rotation is restricted to instance admins"))
			return
		}

		var req RotateKeyRequest
		if err := httpjson.DecodeJSON(w, request, &req); err != nil {
			httpjson.WriteError(w, err)
			return
		}

		newKey, err := crypto.KeyFromBase64(req.NewKey)
		if err != nil {
			httpjson.WriteError(w, errors.Invalid("invalid new key: must be base64-encoded 32 bytes"))
			return
		}

		count, err := service.RotateRootKey(request.Context(), newKey, identity.UserID, identity.Email, identity.ClientIP)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}

		httpjson.WriteJSON(w, http.StatusOK, RotateKeyResponse{Rotated: count, Warning: RotateKeyWarning})
	})

	router.With(middleware.RequireAuth(authService), middleware.RequireUserSession).Post("/invites/{token}/accept", func(w http.ResponseWriter, request *http.Request) {
		token := chi.URLParam(request, "token")
		if err := service.controller.acceptInvite(request.Context(), token); err != nil {
			httpjson.WriteError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	router.Route("/projects", func(router chi.Router) {
		router.Use(middleware.RequireAuth(authService))

		router.Get("/", func(w http.ResponseWriter, request *http.Request) {
			resp, err := service.controller.list(request.Context())
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusOK, resp)
		})

		router.Post("/", func(w http.ResponseWriter, request *http.Request) {
			var req CreateRequest
			if err := httpjson.DecodeJSON(w, request, &req); err != nil {
				httpjson.WriteError(w, err)
				return
			}

			resp, err := service.controller.create(request.Context(), &req)
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusCreated, resp)
		})

		router.Route("/{slug}", func(router chi.Router) {
			router.Get("/", func(w http.ResponseWriter, request *http.Request) {
				slug := chi.URLParam(request, "slug")
				resp, err := service.controller.get(request.Context(), slug)
				if err != nil {
					httpjson.WriteError(w, err)
					return
				}
				httpjson.WriteJSON(w, http.StatusOK, resp)
			})

			router.Put("/", func(w http.ResponseWriter, request *http.Request) {
				slug := chi.URLParam(request, "slug")
				var req UpdateRequest
				if err := httpjson.DecodeJSON(w, request, &req); err != nil {
					httpjson.WriteError(w, err)
					return
				}

				resp, err := service.controller.update(request.Context(), slug, &req)
				if err != nil {
					httpjson.WriteError(w, err)
					return
				}
				httpjson.WriteJSON(w, http.StatusOK, resp)
			})

			router.Delete("/", func(w http.ResponseWriter, request *http.Request) {
				slug := chi.URLParam(request, "slug")
				if err := service.controller.delete(request.Context(), slug); err != nil {
					httpjson.WriteError(w, err)
					return
				}
				w.WriteHeader(http.StatusNoContent)
			})

			router.With(middleware.RequireUserSession).Get("/members", func(w http.ResponseWriter, request *http.Request) {
				slug := chi.URLParam(request, "slug")
				resp, err := service.controller.listMembers(request.Context(), slug)
				if err != nil {
					httpjson.WriteError(w, err)
					return
				}
				httpjson.WriteJSON(w, http.StatusOK, resp)
			})

			router.With(middleware.RequireUserSession).Post("/members", func(w http.ResponseWriter, request *http.Request) {
				slug := chi.URLParam(request, "slug")
				var req AddMemberRequest
				if err := httpjson.DecodeJSON(w, request, &req); err != nil {
					httpjson.WriteError(w, err)
					return
				}

				if err := service.controller.addMember(request.Context(), slug, &req); err != nil {
					httpjson.WriteError(w, err)
					return
				}
				w.WriteHeader(http.StatusNoContent)
			})

			router.With(middleware.RequireUserSession).Get("/invites", func(w http.ResponseWriter, request *http.Request) {
				slug := chi.URLParam(request, "slug")
				resp, err := service.controller.listInvites(request.Context(), slug)
				if err != nil {
					httpjson.WriteError(w, err)
					return
				}
				httpjson.WriteJSON(w, http.StatusOK, resp)
			})

			router.With(middleware.RequireUserSession).Post("/invites", func(w http.ResponseWriter, request *http.Request) {
				slug := chi.URLParam(request, "slug")
				var req CreateInviteRequest
				if err := httpjson.DecodeJSON(w, request, &req); err != nil {
					httpjson.WriteError(w, err)
					return
				}
				resp, err := service.controller.createInvite(request.Context(), slug, &req)
				if err != nil {
					httpjson.WriteError(w, err)
					return
				}
				httpjson.WriteJSON(w, http.StatusCreated, resp)
			})

			router.With(middleware.RequireUserSession).Delete("/invites/{inviteId}", func(w http.ResponseWriter, request *http.Request) {
				slug := chi.URLParam(request, "slug")
				inviteID, err := strconv.ParseInt(chi.URLParam(request, "inviteId"), 10, 64)
				if err != nil {
					httpjson.WriteError(w, errors.Invalid("invalid invite id"))
					return
				}
				if err := service.controller.deleteInvite(request.Context(), slug, inviteID); err != nil {
					httpjson.WriteError(w, err)
					return
				}
				w.WriteHeader(http.StatusNoContent)
			})

			router.Get("/audit", func(w http.ResponseWriter, request *http.Request) {
				slug := chi.URLParam(request, "slug")
				action := request.URL.Query().Get("action")
				limit := 50
				if l := request.URL.Query().Get("limit"); l != "" {
					if parsed, err := strconv.Atoi(l); err == nil {
						limit = parsed
					}
				}
				resp, err := service.controller.listAuditLogs(request.Context(), slug, action, limit)
				if err != nil {
					httpjson.WriteError(w, err)
					return
				}
				httpjson.WriteJSON(w, http.StatusOK, resp)
			})

			router.With(middleware.RequireUserSession).Get("/webhooks", func(w http.ResponseWriter, request *http.Request) {
				slug := chi.URLParam(request, "slug")
				resp, err := service.controller.listWebhooks(request.Context(), slug)
				if err != nil {
					httpjson.WriteError(w, err)
					return
				}
				httpjson.WriteJSON(w, http.StatusOK, resp)
			})

			router.With(middleware.RequireUserSession).Post("/webhooks", func(w http.ResponseWriter, request *http.Request) {
				slug := chi.URLParam(request, "slug")
				var req CreateWebhookRequest
				if err := httpjson.DecodeJSON(w, request, &req); err != nil {
					httpjson.WriteError(w, err)
					return
				}
				resp, err := service.controller.createWebhook(request.Context(), slug, &req)
				if err != nil {
					httpjson.WriteError(w, err)
					return
				}
				httpjson.WriteJSON(w, http.StatusCreated, resp)
			})

			router.With(middleware.RequireUserSession).Delete("/webhooks/{webhookId}", func(w http.ResponseWriter, request *http.Request) {
				slug := chi.URLParam(request, "slug")
				webhookID, err := strconv.ParseInt(chi.URLParam(request, "webhookId"), 10, 64)
				if err != nil {
					httpjson.WriteError(w, errors.Invalid("invalid webhook id"))
					return
				}
				if err := service.controller.deleteWebhook(request.Context(), slug, webhookID); err != nil {
					httpjson.WriteError(w, err)
					return
				}
				w.WriteHeader(http.StatusNoContent)
			})

			router.With(middleware.RequireUserSession).Get("/tokens", func(w http.ResponseWriter, request *http.Request) {
				slug := chi.URLParam(request, "slug")
				resp, err := service.controller.listApiTokens(request.Context(), slug)
				if err != nil {
					httpjson.WriteError(w, err)
					return
				}
				httpjson.WriteJSON(w, http.StatusOK, resp)
			})

			router.With(middleware.RequireUserSession).Post("/tokens", func(w http.ResponseWriter, request *http.Request) {
				slug := chi.URLParam(request, "slug")
				var req CreateApiTokenRequest
				if err := httpjson.DecodeJSON(w, request, &req); err != nil {
					httpjson.WriteError(w, err)
					return
				}
				resp, err := service.controller.createApiToken(request.Context(), slug, &req)
				if err != nil {
					httpjson.WriteError(w, err)
					return
				}
				httpjson.WriteJSON(w, http.StatusCreated, resp)
			})

			router.With(middleware.RequireUserSession).Delete("/tokens/{tokenId}", func(w http.ResponseWriter, request *http.Request) {
				slug := chi.URLParam(request, "slug")
				tokenID, err := strconv.ParseInt(chi.URLParam(request, "tokenId"), 10, 64)
				if err != nil {
					httpjson.WriteError(w, errors.Invalid("invalid token id"))
					return
				}
				if err := service.controller.revokeApiToken(request.Context(), slug, tokenID); err != nil {
					httpjson.WriteError(w, err)
					return
				}
				w.WriteHeader(http.StatusNoContent)
			})

			router.With(middleware.RequireUserSession).Delete("/members/{userId}", func(w http.ResponseWriter, request *http.Request) {
				slug := chi.URLParam(request, "slug")
				targetUserID, err := strconv.ParseInt(chi.URLParam(request, "userId"), 10, 64)
				if err != nil {
					httpjson.WriteError(w, errors.Invalid("invalid user id"))
					return
				}

				if err := service.controller.removeMember(request.Context(), slug, targetUserID); err != nil {
					httpjson.WriteError(w, err)
					return
				}
				httpjson.WriteJSON(w, http.StatusOK, RemoveMemberResponse{
					Warning: "Removed members keep any secrets they already fetched. Rotate affected credentials.",
				})
			})
		})
	})
}
