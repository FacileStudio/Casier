package mcp

import (
	"github.com/FacileStudio/tronc/apiref"
	mcpkit "github.com/FacileStudio/tronc/mcp"
)

// Documentation describes the confirm page routes for the API reference. The MCP endpoint
// itself is outside /api: POST /mcp speaks the Model Context Protocol to any MCP client
// holding a Registre access token for this server, and
// /.well-known/oauth-protected-resource/mcp describes how to get one.
var Documentation = apiref.Module{
	Name: "mcp",
	Description: "Secret reveals and deletions MCP clients asked for, waiting for the person to confirm them. A " +
		"client that cannot ask the person itself hands them a link to the pending page, which uses these routes. " +
		"The MCP endpoint is POST /mcp, outside /api.",
	Routes: pendingRoutes(),
}

// pendingRoutes documents the three confirm page routes.
func pendingRoutes() []apiref.Route {
	return mcpkit.PendingRoutes[PendingItem](mcpkit.PendingDoc{
		List: "Open actions, newest first, with the secrets they name as they are now, metadata only. state is " +
			"ready, changed (a secret was written after the request, so it cannot be confirmed) or missing.",
		Confirm: "Runs the action once. A delete moves the secret to the trash. A reveal lets the client that asked " +
			"read those values once within 15 minutes; the values never reach this page. A secret written since the " +
			"request is refused with 412 and the action closes.",
		Dismiss: "Closes the action without running it.",
		Changed: "A secret changed after the request.",
	})
}
