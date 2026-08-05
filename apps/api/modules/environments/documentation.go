package environments

import documentation "github.com/FacileStudio/Casier/apps/api/internal/documentation"

var Documentation = documentation.Module{
	Name:        "environments",
	Description: "Environment management routes within a project.",
	Routes: []documentation.Route{
		{
			Method:      "GET",
			Path:        "/projects/{slug}/environments",
			Summary:     "List environments",
			Description: "Returns all environments in the project, ordered by position.",
			Auth:        "bearer token required",
			PathParams: []documentation.Field{
				{Name: "slug", Type: "string", Description: "Project slug."},
			},
			ResponseBody: "ListResponse",
			Errors: []documentation.Error{
				{Status: 404, Code: "not_found", Description: "Project not found."},
			},
		},
		{
			Method:      "POST",
			Path:        "/projects/{slug}/environments",
			Summary:     "Create an environment",
			Description: "Creates a new environment in the project. Requires owner or admin role, and a token with write permission.",
			Auth:        "bearer token required",
			PathParams: []documentation.Field{
				{Name: "slug", Type: "string", Description: "Project slug."},
			},
			RequestBody:  "CreateRequest",
			ResponseBody: "EnvironmentResponse",
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Invalid input."},
				{Status: 403, Code: "permission_denied", Description: "Insufficient permissions, or a read-only API token."},
				{Status: 409, Code: "already_exists", Description: "Environment slug already exists in this project."},
			},
		},
		{
			Method:      "PUT",
			Path:        "/projects/{slug}/environments/{envSlug}",
			Summary:     "Update an environment",
			Description: "Updates the environment name. Requires owner or admin role, and a token scoped to the environment with write permission.",
			Auth:        "bearer token required",
			PathParams: []documentation.Field{
				{Name: "slug", Type: "string", Description: "Project slug."},
				{Name: "envSlug", Type: "string", Description: "Environment slug."},
			},
			RequestBody:  "UpdateRequest",
			ResponseBody: "EnvironmentResponse",
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Invalid input."},
				{Status: 403, Code: "permission_denied", Description: "Insufficient permissions, or a read-only API token."},
				{Status: 404, Code: "not_found", Description: "Environment not found, or outside the API token's environment scope."},
			},
		},
		{
			Method:      "DELETE",
			Path:        "/projects/{slug}/environments/{envSlug}",
			Summary:     "Delete an environment",
			Description: "Deletes an environment and all its secrets, versions and ACLs. Requires owner or admin role, and a token scoped to the environment with write permission.",
			Auth:        "bearer token required",
			PathParams: []documentation.Field{
				{Name: "slug", Type: "string", Description: "Project slug."},
				{Name: "envSlug", Type: "string", Description: "Environment slug."},
			},
			Errors: []documentation.Error{
				{Status: 403, Code: "permission_denied", Description: "Insufficient permissions, or a read-only API token."},
				{Status: 404, Code: "not_found", Description: "Environment not found, or outside the API token's environment scope."},
			},
		},
	},
}
