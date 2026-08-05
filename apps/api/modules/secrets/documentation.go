package secrets

import documentation "github.com/FacileStudio/Casier/apps/api/internal/documentation"

var Documentation = documentation.Module{
	Name:        "secrets",
	Description: "Secret management routes within an environment.",
	Routes: []documentation.Route{
		{
			Method:      "GET",
			Path:        "/projects/{slug}/environments/{envSlug}/secrets",
			Summary:     "List secrets",
			Description: "Returns the secrets in the environment the caller may read, with decrypted values, when each value was last changed, and how it grades against the project's rotation policy. A secret carrying per-secret ACLs is omitted unless the caller holds a read grant on it, or owns the project. The optional tag query param matches a whole tag.",
			Auth:        "bearer token required",
			PathParams: []documentation.Field{
				{Name: "slug", Type: "string", Description: "Project slug."},
				{Name: "envSlug", Type: "string", Description: "Environment slug."},
			},
			ResponseBody: "ListResponse",
			Errors: []documentation.Error{
				{Status: 403, Code: "permission_denied", Description: "API token is not scoped to this project or environment."},
				{Status: 404, Code: "not_found", Description: "Project or environment not found."},
			},
		},
		{
			Method:      "GET",
			Path:        "/projects/{slug}/secrets/rotation",
			Summary:     "Rotation summary",
			Description: "Counts the secrets that are overdue or nearly due for rotation in each environment of the project. Returns zeroes when the project's rotation policy is disabled.",
			Auth:        "bearer token required",
			PathParams: []documentation.Field{
				{Name: "slug", Type: "string", Description: "Project slug."},
			},
			ResponseBody: "RotationSummaryResponse",
			Errors: []documentation.Error{
				{Status: 404, Code: "not_found", Description: "Project not found."},
			},
		},
		{
			Method:      "POST",
			Path:        "/projects/{slug}/environments/{envSlug}/secrets",
			Summary:     "Create a secret",
			Description: "Creates a new encrypted secret. Requires owner, admin, or member role, and an API token carrying the write permission.",
			Auth:        "bearer token required",
			PathParams: []documentation.Field{
				{Name: "slug", Type: "string", Description: "Project slug."},
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
			Path:        "/projects/{slug}/environments/{envSlug}/secrets/{key}",
			Summary:     "Update a secret",
			Description: "Updates the secret value and increments the version. Requires owner, admin, or member role, an API token carrying the write permission, and a write grant when the secret has per-secret ACLs.",
			Auth:        "bearer token required",
			PathParams: []documentation.Field{
				{Name: "slug", Type: "string", Description: "Project slug."},
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
			Path:        "/projects/{slug}/environments/{envSlug}/secrets/{key}",
			Summary:     "Delete a secret",
			Description: "Deletes a secret. Requires owner, admin, or member role, an API token carrying the write permission, and a write grant when the secret has per-secret ACLs.",
			Auth:        "bearer token required",
			PathParams: []documentation.Field{
				{Name: "slug", Type: "string", Description: "Project slug."},
				{Name: "envSlug", Type: "string", Description: "Environment slug."},
				{Name: "key", Type: "string", Description: "Secret key."},
			},
			Errors: []documentation.Error{
				{Status: 403, Code: "permission_denied", Description: "Insufficient permissions."},
				{Status: 404, Code: "not_found", Description: "Secret not found."},
			},
		},
		{
			Method:      "POST",
			Path:        "/projects/{slug}/environments/{envSlug}/secrets/import",
			Summary:     "Import a .env file",
			Description: "Applies a whole .env file in one transaction: either every parsed line lands or none does. Quoted values are unescaped as the export route writes them. Lines without a '=', keys longer than 256 characters, and secrets the caller may not write are counted as skipped. Each created or updated secret produces an audit entry and a webhook event.",
			Auth:        "bearer token required",
			PathParams: []documentation.Field{
				{Name: "slug", Type: "string", Description: "Project slug."},
				{Name: "envSlug", Type: "string", Description: "Environment slug."},
			},
			RequestBody:  "ImportRequest",
			ResponseBody: "ImportResponse",
			Errors: []documentation.Error{
				{Status: 403, Code: "permission_denied", Description: "Insufficient permissions."},
				{Status: 404, Code: "not_found", Description: "Project or environment not found."},
			},
		},
		{
			Method:      "GET",
			Path:        "/projects/{slug}/environments/{envSlug}/secrets/export",
			Summary:     "Export a .env file",
			Description: "Renders the secrets the caller may read as a .env file. Values holding whitespace, quotes, backslashes, comment markers or line breaks are double-quoted and escaped so the import route reads them back byte-identical.",
			Auth:        "bearer token required",
			PathParams: []documentation.Field{
				{Name: "slug", Type: "string", Description: "Project slug."},
				{Name: "envSlug", Type: "string", Description: "Environment slug."},
			},
			ResponseBody: "ExportResponse",
			Errors: []documentation.Error{
				{Status: 403, Code: "permission_denied", Description: "API token is not scoped to this project or environment."},
				{Status: 404, Code: "not_found", Description: "Project or environment not found."},
			},
		},
	},
}
