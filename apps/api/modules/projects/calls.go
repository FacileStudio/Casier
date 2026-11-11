package projects

import "context"

// Calls are the project routes' operations for an in-process caller, such as
// the MCP server, run through the route's own controller method for the
// identity on ctx.
type Calls struct {
	controller *Controller
}

// Calls returns the in-process operations of the project routes.
func (service *Service) Calls() Calls {
	return Calls{controller: service.controller}
}

// List is GET /projects: every project the caller is a member of.
func (c Calls) List(ctx context.Context) (*ListResponse, error) {
	return c.controller.list(ctx)
}

// Create is POST /projects; the caller becomes its owner.
func (c Calls) Create(ctx context.Context, req *CreateRequest) (*ProjectResponse, error) {
	return c.controller.create(ctx, req)
}

// AuditLogs is GET /projects/{slug}/audit, newest first.
func (c Calls) AuditLogs(ctx context.Context, slug, action string, limit int) (*AuditLogsResponse, error) {
	return c.controller.listAuditLogs(ctx, slug, action, limit)
}
