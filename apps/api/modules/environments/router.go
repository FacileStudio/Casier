package environments

import (
	"net/http"

	"clef/internal/httpjson"
	"clef/internal/middleware"
	"clef/modules/auth"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(router chi.Router, service *Service, authService *auth.Service) {
	router.Route("/spaces/{slug}/environments", func(router chi.Router) {
		router.Use(middleware.RequireAuth(authService))

		router.Get("/", func(w http.ResponseWriter, request *http.Request) {
			slug := chi.URLParam(request, "slug")
			resp, err := service.controller.list(request.Context(), slug)
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusOK, resp)
		})

		router.Post("/", func(w http.ResponseWriter, request *http.Request) {
			slug := chi.URLParam(request, "slug")
			var req CreateRequest
			if err := httpjson.DecodeJSON(w, request, &req); err != nil {
				httpjson.WriteError(w, err)
				return
			}

			resp, err := service.controller.create(request.Context(), slug, &req)
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusCreated, resp)
		})

		router.Put("/{envSlug}", func(w http.ResponseWriter, request *http.Request) {
			slug := chi.URLParam(request, "slug")
			envSlug := chi.URLParam(request, "envSlug")
			var req UpdateRequest
			if err := httpjson.DecodeJSON(w, request, &req); err != nil {
				httpjson.WriteError(w, err)
				return
			}

			resp, err := service.controller.update(request.Context(), slug, envSlug, &req)
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusOK, resp)
		})

		router.Delete("/{envSlug}", func(w http.ResponseWriter, request *http.Request) {
			slug := chi.URLParam(request, "slug")
			envSlug := chi.URLParam(request, "envSlug")
			if err := service.controller.delete(request.Context(), slug, envSlug); err != nil {
				httpjson.WriteError(w, err)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		})
	})
}
