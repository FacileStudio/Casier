// Package docs names the route-registry types each module fills in.
//
// They are aliases for tronc/apiref, so a module's documentation.go stays
// written against a local package while apiref owns the shape and renders it as
// OpenAPI behind the reference page at /docs.
package docs

import "github.com/FacileStudio/tronc/apiref"

type (
	// Response names the registry that aggregates a module's declared routes.
	Response = apiref.Registry

	// Module is one module's route group.
	Module = apiref.Module

	// Route is one declared endpoint.
	Route = apiref.Route

	// Field is one parameter or body field.
	Field = apiref.Field

	// Error is one documented failure response.
	Error = apiref.Error
)
