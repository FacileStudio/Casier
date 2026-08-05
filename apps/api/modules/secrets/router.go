package secrets

import (
	"net/http"
	"strconv"

	"github.com/FacileStudio/Casier/apps/api/internal/middleware"
	"github.com/FacileStudio/Casier/apps/api/modules/auth"
	"github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(router chi.Router, service *Service, authService *auth.Service) {
	router.With(middleware.RequireAuth(authService)).Get("/projects/{slug}/secrets/diff", func(w http.ResponseWriter, request *http.Request) {
		projectSlug := chi.URLParam(request, "slug")
		from := request.URL.Query().Get("from")
		to := request.URL.Query().Get("to")
		if from == "" || to == "" {
			httpjson.WriteError(w, errors.Invalid("from and to query params are required"))
			return
		}
		resp, err := service.controller.diff(request.Context(), projectSlug, from, to)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})

	router.With(middleware.RequireAuth(authService)).Get("/projects/{slug}/secrets/rotation", func(w http.ResponseWriter, request *http.Request) {
		resp, err := service.controller.rotation(request.Context(), chi.URLParam(request, "slug"))
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})

	router.Route("/projects/{slug}/environments/{envSlug}/secrets", func(router chi.Router) {
		router.Use(middleware.RequireAuth(authService))

		router.Get("/", func(w http.ResponseWriter, request *http.Request) {
			projectSlug := chi.URLParam(request, "slug")
			envSlug := chi.URLParam(request, "envSlug")
			tagFilter := request.URL.Query().Get("tag")

			resp, err := service.controller.list(request.Context(), projectSlug, envSlug, tagFilter, boolParam(request, "reveal"))
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusOK, resp)
		})

		router.Get("/trash", func(w http.ResponseWriter, request *http.Request) {
			projectSlug := chi.URLParam(request, "slug")
			envSlug := chi.URLParam(request, "envSlug")

			resp, err := service.controller.trash(request.Context(), projectSlug, envSlug)
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusOK, resp)
		})

		router.Get("/{key}/reveal", func(w http.ResponseWriter, request *http.Request) {
			projectSlug := chi.URLParam(request, "slug")
			envSlug := chi.URLParam(request, "envSlug")
			key := chi.URLParam(request, "key")

			resp, err := service.controller.reveal(request.Context(), projectSlug, envSlug, key)
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusOK, resp)
		})

		router.Post("/{key}/restore", func(w http.ResponseWriter, request *http.Request) {
			projectSlug := chi.URLParam(request, "slug")
			envSlug := chi.URLParam(request, "envSlug")
			key := chi.URLParam(request, "key")

			if err := service.controller.restore(request.Context(), projectSlug, envSlug, key); err != nil {
				httpjson.WriteError(w, err)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		})

		router.Post("/import", func(w http.ResponseWriter, request *http.Request) {
			projectSlug := chi.URLParam(request, "slug")
			envSlug := chi.URLParam(request, "envSlug")

			var req ImportRequest
			if err := httpjson.DecodeJSON(w, request, &req); err != nil {
				httpjson.WriteError(w, err)
				return
			}

			resp, err := service.controller.importEnv(request.Context(), projectSlug, envSlug, &req)
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusOK, resp)
		})

		router.Get("/export", func(w http.ResponseWriter, request *http.Request) {
			projectSlug := chi.URLParam(request, "slug")
			envSlug := chi.URLParam(request, "envSlug")

			resp, err := service.controller.exportEnv(request.Context(), projectSlug, envSlug)
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusOK, resp)
		})

		router.Post("/", func(w http.ResponseWriter, request *http.Request) {
			projectSlug := chi.URLParam(request, "slug")
			envSlug := chi.URLParam(request, "envSlug")

			var req CreateRequest
			if err := httpjson.DecodeJSON(w, request, &req); err != nil {
				httpjson.WriteError(w, err)
				return
			}

			resp, err := service.controller.create(request.Context(), projectSlug, envSlug, &req)
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusCreated, resp)
		})

		router.Put("/{key}", func(w http.ResponseWriter, request *http.Request) {
			projectSlug := chi.URLParam(request, "slug")
			envSlug := chi.URLParam(request, "envSlug")
			key := chi.URLParam(request, "key")

			var req UpdateRequest
			if err := httpjson.DecodeJSON(w, request, &req); err != nil {
				httpjson.WriteError(w, err)
				return
			}

			resp, err := service.controller.update(request.Context(), projectSlug, envSlug, key, &req)
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusOK, resp)
		})

		router.Put("/{key}/tags", func(w http.ResponseWriter, request *http.Request) {
			projectSlug := chi.URLParam(request, "slug")
			envSlug := chi.URLParam(request, "envSlug")
			key := chi.URLParam(request, "key")

			var req UpdateTagsRequest
			if err := httpjson.DecodeJSON(w, request, &req); err != nil {
				httpjson.WriteError(w, err)
				return
			}

			if err := service.controller.updateTags(request.Context(), projectSlug, envSlug, key, &req); err != nil {
				httpjson.WriteError(w, err)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		})

		router.Get("/{key}/versions", func(w http.ResponseWriter, request *http.Request) {
			projectSlug := chi.URLParam(request, "slug")
			envSlug := chi.URLParam(request, "envSlug")
			key := chi.URLParam(request, "key")

			resp, err := service.controller.listVersions(request.Context(), projectSlug, envSlug, key)
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusOK, resp)
		})

		router.Post("/{key}/rollback", func(w http.ResponseWriter, request *http.Request) {
			projectSlug := chi.URLParam(request, "slug")
			envSlug := chi.URLParam(request, "envSlug")
			key := chi.URLParam(request, "key")

			var req RollbackRequest
			if err := httpjson.DecodeJSON(w, request, &req); err != nil {
				httpjson.WriteError(w, err)
				return
			}

			resp, err := service.controller.rollback(request.Context(), projectSlug, envSlug, key, &req)
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusOK, resp)
		})

		router.Get("/{key}/acls", func(w http.ResponseWriter, request *http.Request) {
			projectSlug := chi.URLParam(request, "slug")
			envSlug := chi.URLParam(request, "envSlug")
			key := chi.URLParam(request, "key")

			resp, err := service.controller.listACLs(request.Context(), projectSlug, envSlug, key)
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusOK, resp)
		})

		router.Put("/{key}/acls", func(w http.ResponseWriter, request *http.Request) {
			projectSlug := chi.URLParam(request, "slug")
			envSlug := chi.URLParam(request, "envSlug")
			key := chi.URLParam(request, "key")

			var req SetACLRequest
			if err := httpjson.DecodeJSON(w, request, &req); err != nil {
				httpjson.WriteError(w, err)
				return
			}
			if err := service.controller.setACL(request.Context(), projectSlug, envSlug, key, &req); err != nil {
				httpjson.WriteError(w, err)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		})

		router.Delete("/{key}/acls/{aclId}", func(w http.ResponseWriter, request *http.Request) {
			projectSlug := chi.URLParam(request, "slug")
			envSlug := chi.URLParam(request, "envSlug")
			key := chi.URLParam(request, "key")
			aclID, err := strconv.ParseInt(chi.URLParam(request, "aclId"), 10, 64)
			if err != nil {
				httpjson.WriteError(w, errors.Invalid("invalid acl id"))
				return
			}
			if err := service.controller.deleteACL(request.Context(), projectSlug, envSlug, key, aclID); err != nil {
				httpjson.WriteError(w, err)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		})

		router.Delete("/{key}", func(w http.ResponseWriter, request *http.Request) {
			projectSlug := chi.URLParam(request, "slug")
			envSlug := chi.URLParam(request, "envSlug")
			key := chi.URLParam(request, "key")

			if err := service.controller.delete(request.Context(), projectSlug, envSlug, key, boolParam(request, "purge")); err != nil {
				httpjson.WriteError(w, err)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		})
	})
}

// boolParam reads an opt-in query flag. An unparseable value is false, because
// every flag here widens what a request does - revealing plaintext, or
// destroying a secret outright - and a typo must never be read as consent.
func boolParam(request *http.Request, name string) bool {
	value, err := strconv.ParseBool(request.URL.Query().Get(name))
	return err == nil && value
}
