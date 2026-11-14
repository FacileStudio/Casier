package mcp

import (
	"net/http"

	mcpkit "github.com/FacileStudio/tronc/mcp"
	"github.com/go-chi/chi/v5"
)

// SessionGuard admits a person's own browser or CLI session only, never a casier_ API
// token. The auth service implements it.
type SessionGuard interface {
	RequireAuth(next http.Handler) http.Handler
}

// RegisterRoutes mounts the MCP endpoint at the resource's path and its RFC 9728
// metadata on router, the root router outside /api.
func RegisterRoutes(router chi.Router, service *Service) {
	service.server.Mount(router)
}

// RegisterPendingRoutes mounts the confirm page's API under /mcp/pending on the /api
// router, for the signed-in person. It sits behind the session guard rather than the
// token-or-session one, so a machine token can never confirm a reveal. A nil service,
// as in the docs test, still mounts the routes.
func RegisterPendingRoutes(router chi.Router, service *Service, sessions SessionGuard) {
	service.pending().Mount(router, sessions.RequireAuth)
}

// pending is the confirm page, nil on a nil service.
func (s *Service) pending() *mcpkit.PendingPage[PendingItem] {
	if s == nil {
		return nil
	}
	return s.page
}
