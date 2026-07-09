package spaces

import (
	"net/http"
	"strconv"

	"clef/internal/errors"
	"clef/internal/httpjson"
	"clef/internal/middleware"
	"clef/modules/auth"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(router chi.Router, service *Service, authService *auth.Service) {
	router.Route("/spaces", func(router chi.Router) {
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

			router.Get("/members", func(w http.ResponseWriter, request *http.Request) {
				slug := chi.URLParam(request, "slug")
				resp, err := service.controller.listMembers(request.Context(), slug)
				if err != nil {
					httpjson.WriteError(w, err)
					return
				}
				httpjson.WriteJSON(w, http.StatusOK, resp)
			})

			router.Post("/members", func(w http.ResponseWriter, request *http.Request) {
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

			router.Delete("/members/{userId}", func(w http.ResponseWriter, request *http.Request) {
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
				w.WriteHeader(http.StatusNoContent)
			})
		})
	})
}
