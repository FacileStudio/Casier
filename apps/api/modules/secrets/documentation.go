package secrets

import documentation "casier/internal/documentation"

var Documentation = documentation.Module{
	Name:        "secrets",
	Description: "Secret management routes within an environment.",
	Routes: []documentation.Route{
		{
			Method:      "GET",
			Path:        "/spaces/{slug}/environments/{envSlug}/secrets",
			Summary:     "List secrets",
			Description: "Returns all secrets in the environment with decrypted values.",
			Auth:        "bearer token required",
			PathParams: []documentation.Field{
				{Name: "slug", Type: "string", Description: "Space slug."},
				{Name: "envSlug", Type: "string", Description: "Environment slug."},
			},
			ResponseBody: "ListResponse",
			Errors: []documentation.Error{
				{Status: 404, Code: "not_found", Description: "Space or environment not found."},
			},
		},
		{
			Method:      "POST",
			Path:        "/spaces/{slug}/environments/{envSlug}/secrets",
			Summary:     "Create a secret",
			Description: "Creates a new encrypted secret. Requires owner, admin, or member role.",
			Auth:        "bearer token required",
			PathParams: []documentation.Field{
				{Name: "slug", Type: "string", Description: "Space slug."},
				{Name: "envSlug", Type: "string", Description: "Environment slug."},
			},
			RequestBody:  "CreateRequest",
			ResponseBody: "SecretResponse",
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Invalid input."},
				{Status: 403, Code: "permission_denied", Description: "Insufficient permissions."},
				{Status: 409, Code: "already_exists", Description: "Secret key already exists in this environment."},
			},
		},
		{
			Method:      "PUT",
			Path:        "/spaces/{slug}/environments/{envSlug}/secrets/{key}",
			Summary:     "Update a secret",
			Description: "Updates the secret value and increments the version. Requires owner, admin, or member role.",
			Auth:        "bearer token required",
			PathParams: []documentation.Field{
				{Name: "slug", Type: "string", Description: "Space slug."},
				{Name: "envSlug", Type: "string", Description: "Environment slug."},
				{Name: "key", Type: "string", Description: "Secret key."},
			},
			RequestBody:  "UpdateRequest",
			ResponseBody: "SecretResponse",
			Errors: []documentation.Error{
				{Status: 403, Code: "permission_denied", Description: "Insufficient permissions."},
				{Status: 404, Code: "not_found", Description: "Secret not found."},
			},
		},
		{
			Method:      "DELETE",
			Path:        "/spaces/{slug}/environments/{envSlug}/secrets/{key}",
			Summary:     "Delete a secret",
			Description: "Deletes a secret. Requires owner, admin, or member role.",
			Auth:        "bearer token required",
			PathParams: []documentation.Field{
				{Name: "slug", Type: "string", Description: "Space slug."},
				{Name: "envSlug", Type: "string", Description: "Environment slug."},
				{Name: "key", Type: "string", Description: "Secret key."},
			},
			Errors: []documentation.Error{
				{Status: 403, Code: "permission_denied", Description: "Insufficient permissions."},
				{Status: 404, Code: "not_found", Description: "Secret not found."},
			},
		},
	},
}
