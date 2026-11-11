package environments

import "context"

// Calls are the environment routes' operations for an in-process caller, such
// as the MCP server, run through the route's own controller method for the
// identity on ctx.
type Calls struct {
	controller *Controller
}

// Calls returns the in-process operations of the environment routes.
func (service *Service) Calls() Calls {
	return Calls{controller: service.controller}
}

// List is GET /projects/{slug}/environments, in their display order.
func (c Calls) List(ctx context.Context, projectSlug string) (*ListResponse, error) {
	return c.controller.list(ctx, projectSlug)
}

// Create is POST /projects/{slug}/environments.
func (c Calls) Create(ctx context.Context, projectSlug string, req *CreateRequest) (*EnvironmentResponse, error) {
	return c.controller.create(ctx, projectSlug, req)
}
