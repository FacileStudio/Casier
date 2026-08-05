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
			Description: "Returns the secrets in the environment the caller may read: keys, tags, versions, when each value was last changed, and how it grades against the project's rotation policy. Values are omitted entirely - the field is absent, not empty - unless reveal=true is passed, which is audited as secrets.reveal_list with the number of values revealed rather than as a plain secrets.list. A secret carrying per-secret ACLs is omitted unless the caller holds a read grant on it, or owns the project. The optional tag query param matches a whole tag.",
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
			Path:        "/projects/{slug}/environments/{envSlug}/secrets/{key}/reveal",
			Summary:     "Reveal one secret",
			Description: "Returns a single secret with its decrypted value, audited per key as secrets.reveal so reading one credential is never indistinguishable from dumping an environment. A secret the caller may not read answers not_found rather than permission_denied, matching the way per-secret ACLs hide a secret from every listing.",
			Auth:        "bearer token required",
			PathParams: []documentation.Field{
				{Name: "slug", Type: "string", Description: "Project slug."},
				{Name: "envSlug", Type: "string", Description: "Environment slug."},
				{Name: "key", Type: "string", Description: "Secret key."},
			},
			ResponseBody: "SecretResponse",
			Errors: []documentation.Error{
				{Status: 403, Code: "permission_denied", Description: "API token is not scoped to this project or environment."},
				{Status: 404, Code: "not_found", Description: "Secret not found, or hidden from the caller by a per-secret ACL."},
			},
		},
		{
			Method:      "GET",
			Path:        "/projects/{slug}/environments/{envSlug}/secrets/trash",
			Summary:     "List trashed secrets",
			Description: "Lists the secrets deleted from this environment that can still be restored, with who deleted each one, when, and the deadline after which the purge loop destroys it. Values are never returned here. A secret hidden from the caller by a per-secret ACL is hidden from the trash too.",
			Auth:        "bearer token required",
			PathParams: []documentation.Field{
				{Name: "slug", Type: "string", Description: "Project slug."},
				{Name: "envSlug", Type: "string", Description: "Environment slug."},
			},
			ResponseBody: "TrashResponse",
			Errors: []documentation.Error{
				{Status: 403, Code: "permission_denied", Description: "API token is not scoped to this project or environment."},
				{Status: 404, Code: "not_found", Description: "Project or environment not found."},
			},
		},
		{
			Method:      "POST",
			Path:        "/projects/{slug}/environments/{envSlug}/secrets/{key}/restore",
			Summary:     "Restore a trashed secret",
			Description: "Lifts a secret back out of the trash with its versions and per-secret ACLs intact. Requires owner, admin, or member role, an API token carrying the write permission, and a write grant when the secret has per-secret ACLs. Emits a secret.restored webhook event.",
			Auth:        "bearer token required",
			PathParams: []documentation.Field{
				{Name: "slug", Type: "string", Description: "Project slug."},
				{Name: "envSlug", Type: "string", Description: "Environment slug."},
				{Name: "key", Type: "string", Description: "Secret key."},
			},
			Errors: []documentation.Error{
				{Status: 403, Code: "permission_denied", Description: "Insufficient permissions, or the environment is protected and the caller is not an owner or admin."},
				{Status: 404, Code: "not_found", Description: "No secret with that key is in the trash."},
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
			Description: "Creates a new encrypted secret and echoes back its value. Requires owner, admin, or member role, and an API token carrying the write permission. A protected environment narrows that to owner and admin. A key that is currently in the trash still holds its slot in the unique index, so recreating it conflicts: the error says so and names restore or purge as the way out.",
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
				{Status: 409, Code: "already_exists", Description: "Secret key already exists in this environment, or is in the trash and must be restored or purged first."},
			},
		},
		{
			Method:      "PUT",
			Path:        "/projects/{slug}/environments/{envSlug}/secrets/{key}",
			Summary:     "Update a secret",
			Description: "Updates the secret value and increments the version. Requires owner, admin, or member role, an API token carrying the write permission, and a write grant when the secret has per-secret ACLs. A protected environment narrows the role to owner and admin.",
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
			Description: "Moves a secret to the trash, where it stays restorable for 30 days before the purge loop destroys it; its versions and per-secret ACLs are kept meanwhile, and its key stays reserved so a restore can never collide with a replacement. Emits a secret.trashed webhook event. Requires owner, admin, or member role, an API token carrying the write permission, and a write grant when the secret has per-secret ACLs. Passing purge=true skips the trash and destroys the secret, its versions and its ACLs in one transaction, freeing the key immediately; that is restricted to owner and admin and emits secret.purged.",
			Auth:        "bearer token required",
			PathParams: []documentation.Field{
				{Name: "slug", Type: "string", Description: "Project slug."},
				{Name: "envSlug", Type: "string", Description: "Environment slug."},
				{Name: "key", Type: "string", Description: "Secret key."},
			},
			Errors: []documentation.Error{
				{Status: 403, Code: "permission_denied", Description: "Insufficient permissions, purge attempted without owner or admin role, or the environment is protected and the caller is not an owner or admin."},
				{Status: 404, Code: "not_found", Description: "Secret not found."},
			},
		},
		{
			Method:      "POST",
			Path:        "/projects/{slug}/environments/{envSlug}/secrets/import",
			Summary:     "Import a .env file",
			Description: "Applies a whole .env file in one transaction: either every parsed line lands or none does. Quoted values are unescaped as the export route writes them. Lines without a '=', keys longer than 256 characters, and secrets the caller may not write are counted as skipped. Each created or updated secret produces an audit entry and a webhook event. A protected environment narrows the role to owner and admin. A line whose key sits in the trash conflicts and rolls the whole import back.",
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
			Description: "Renders the secrets the caller may read as a .env file. This is the bulk plaintext path and is audited as secrets.export, apart from both the metadata listing and the revealing one. Values holding whitespace, quotes, backslashes, comment markers or line breaks are double-quoted and escaped so the import route reads them back byte-identical.",
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
