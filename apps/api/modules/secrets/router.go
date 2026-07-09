package secrets

import (
	"net/http"

	"clef/internal/httpjson"
	"clef/internal/middleware"
	"clef/modules/auth"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(router chi.Router, service *Service, authService *auth.Service) {
	router.Route("/spaces/{slug}/environments/{envSlug}/secrets", func(router chi.Router) {
		router.Use(middleware.RequireAuth(authService))

		router.Get("/", func(w http.ResponseWriter, request *http.Request) {
			spaceSlug := chi.URLParam(request, "slug")
			envSlug := chi.URLParam(request, "envSlug")

			resp, err := service.controller.list(request.Context(), spaceSlug, envSlug)
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusOK, resp)
		})

		router.Post("/", func(w http.ResponseWriter, request *http.Request) {
			spaceSlug := chi.URLParam(request, "slug")
			envSlug := chi.URLParam(request, "envSlug")

			var req CreateRequest
			if err := httpjson.DecodeJSON(w, request, &req); err != nil {
				httpjson.WriteError(w, err)
				return
			}

			resp, err := service.controller.create(request.Context(), spaceSlug, envSlug, &req)
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusCreated, resp)
		})

		router.Put("/{key}", func(w http.ResponseWriter, request *http.Request) {
			spaceSlug := chi.URLParam(request, "slug")
			envSlug := chi.URLParam(request, "envSlug")
			key := chi.URLParam(request, "key")

			var req UpdateRequest
			if err := httpjson.DecodeJSON(w, request, &req); err != nil {
				httpjson.WriteError(w, err)
				return
			}

			resp, err := service.controller.update(request.Context(), spaceSlug, envSlug, key, &req)
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusOK, resp)
		})

		router.Delete("/{key}", func(w http.ResponseWriter, request *http.Request) {
			spaceSlug := chi.URLParam(request, "slug")
			envSlug := chi.URLParam(request, "envSlug")
			key := chi.URLParam(request, "key")

			if err := service.controller.delete(request.Context(), spaceSlug, envSlug, key); err != nil {
				httpjson.WriteError(w, err)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		})
	})
}
