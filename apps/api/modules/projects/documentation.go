package projects

import documentation "github.com/FacileStudio/Casier/apps/api/internal/documentation"

var Documentation = documentation.Module{
	Name:        "projects",
	Description: "Project management and membership routes.",
	Routes: []documentation.Route{
		{
			Method:       "GET",
			Path:         "/projects",
			Summary:      "List projects",
			Description:  "Returns all projects the authenticated user belongs to.",
			Auth:         "bearer token required",
			ResponseBody: "ListResponse",
			Errors: []documentation.Error{
				{Status: 401, Code: "unauthenticated", Description: "Authorization header is missing or invalid."},
				{Status: 500, Code: "internal", Description: "Unexpected server error."},
			},
		},
		{
			Method:       "POST",
			Path:         "/projects",
			Summary:      "Create a project",
			Description:  "Creates a new project with default environments (dev, staging, prod). The creator becomes the owner.",
			Auth:         "bearer token required",
			RequestBody:  "CreateRequest",
			ResponseBody: "ProjectResponse",
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Invalid JSON body or missing name."},
				{Status: 409, Code: "already_exists", Description: "A project with the same slug already exists."},
				{Status: 500, Code: "internal", Description: "Unexpected server error."},
			},
		},
		{
			Method:      "GET",
			Path:        "/projects/{slug}",
			Summary:     "Get a project",
			Description: "Returns a single project by slug.",
			Auth:        "bearer token required",
			PathParams: []documentation.Field{
				{Name: "slug", Type: "string", Description: "Project slug."},
			},
			ResponseBody: "ProjectResponse",
			Errors: []documentation.Error{
				{Status: 404, Code: "not_found", Description: "Project not found or user is not a member."},
			},
		},
		{
			Method:      "PUT",
			Path:        "/projects/{slug}",
			Summary:     "Update a project",
			Description: "Updates project name, description and rotation policy. Requires owner or admin role. Omitting rotation_days leaves the current policy in place; sending 0 disables rotation warnings.",
			Auth:        "bearer token required",
			PathParams: []documentation.Field{
				{Name: "slug", Type: "string", Description: "Project slug."},
			},
			RequestBody:  "UpdateRequest",
			ResponseBody: "ProjectResponse",
			Errors: []documentation.Error{
				{Status: 403, Code: "permission_denied", Description: "Insufficient permissions."},
				{Status: 404, Code: "not_found", Description: "Project not found."},
			},
		},
		{
			Method:      "DELETE",
			Path:        "/projects/{slug}",
			Summary:     "Delete a project",
			Description: "Deletes a project. Owner only.",
			Auth:        "bearer token required",
			PathParams: []documentation.Field{
				{Name: "slug", Type: "string", Description: "Project slug."},
			},
			Errors: []documentation.Error{
				{Status: 403, Code: "permission_denied", Description: "Only the owner can delete a project."},
				{Status: 404, Code: "not_found", Description: "Project not found."},
			},
		},
		{
			Method:      "GET",
			Path:        "/projects/{slug}/members",
			Summary:     "List members",
			Description: "Returns all members of the project.",
			Auth:        "bearer token required",
			PathParams: []documentation.Field{
				{Name: "slug", Type: "string", Description: "Project slug."},
			},
			ResponseBody: "MembersResponse",
			Errors: []documentation.Error{
				{Status: 404, Code: "not_found", Description: "Project not found."},
			},
		},
		{
			Method:      "POST",
			Path:        "/projects/{slug}/members",
			Summary:     "Add a member",
			Description: "Adds a user to the project by email. Requires owner or admin role.",
			Auth:        "bearer token required",
			PathParams: []documentation.Field{
				{Name: "slug", Type: "string", Description: "Project slug."},
			},
			RequestBody: "AddMemberRequest",
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Invalid input."},
				{Status: 403, Code: "permission_denied", Description: "Insufficient permissions."},
				{Status: 404, Code: "not_found", Description: "User or project not found."},
				{Status: 409, Code: "already_exists", Description: "User is already a member."},
			},
		},
		{
			Method:      "DELETE",
			Path:        "/projects/{slug}/members/{userId}",
			Summary:     "Remove a member",
			Description: "Removes a user from the project. Requires owner or admin role. Returns a warning reminding to rotate credentials the removed member may have fetched.",
			Auth:        "bearer token required",
			PathParams: []documentation.Field{
				{Name: "slug", Type: "string", Description: "Project slug."},
				{Name: "userId", Type: "int64", Description: "User ID to remove."},
			},
			ResponseBody: "RemoveMemberResponse",
			Errors: []documentation.Error{
				{Status: 403, Code: "permission_denied", Description: "Insufficient permissions."},
				{Status: 404, Code: "not_found", Description: "Member not found."},
			},
		},
		{
			Method:      "GET",
			Path:        "/projects/{slug}/invites",
			Summary:     "List invites",
			Description: "Returns the project's pending invites. Refused for machine tokens.",
			Auth:        "bearer token required, user session only",
			PathParams: []documentation.Field{
				{Name: "slug", Type: "string", Description: "Project slug."},
			},
			ResponseBody: "InvitesResponse",
			Errors: []documentation.Error{
				{Status: 403, Code: "permission_denied", Description: "Caller is an API token."},
				{Status: 404, Code: "not_found", Description: "Project not found."},
			},
		},
		{
			Method:      "POST",
			Path:        "/projects/{slug}/invites",
			Summary:     "Invite a user",
			Description: "Creates an invite for an email address at a given role. Requires owner or admin, and is refused for machine tokens.",
			Auth:        "bearer token required, user session only",
			PathParams: []documentation.Field{
				{Name: "slug", Type: "string", Description: "Project slug."},
			},
			RequestBody:  "CreateInviteRequest",
			ResponseBody: "InviteResponse",
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Invalid input."},
				{Status: 403, Code: "permission_denied", Description: "Insufficient permissions, or the caller is an API token."},
				{Status: 404, Code: "not_found", Description: "Project not found."},
				{Status: 409, Code: "already_exists", Description: "The email is already a member or already invited."},
			},
		},
		{
			Method:      "DELETE",
			Path:        "/projects/{slug}/invites/{inviteId}",
			Summary:     "Revoke an invite",
			Description: "Deletes a pending invite. Requires owner or admin, and is refused for machine tokens. Answers 204.",
			Auth:        "bearer token required, user session only",
			PathParams: []documentation.Field{
				{Name: "slug", Type: "string", Description: "Project slug."},
				{Name: "inviteId", Type: "int64", Description: "Invite ID to revoke."},
			},
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Invite ID is not an integer."},
				{Status: 403, Code: "permission_denied", Description: "Insufficient permissions, or the caller is an API token."},
				{Status: 404, Code: "not_found", Description: "Invite not found."},
			},
		},
		{
			Method:      "POST",
			Path:        "/invites/{token}/accept",
			Summary:     "Accept an invite",
			Description: "Joins the inviting project as the authenticated user, at the role the invite carries. Refused for machine tokens. Answers 204.",
			Auth:        "bearer token required, user session only",
			PathParams: []documentation.Field{
				{Name: "token", Type: "string", Description: "Invite token from the invitation link."},
			},
			Errors: []documentation.Error{
				{Status: 403, Code: "permission_denied", Description: "The invite was issued to another email, or the caller is an API token."},
				{Status: 404, Code: "not_found", Description: "Invite not found or expired."},
				{Status: 409, Code: "already_exists", Description: "The user is already a member."},
			},
		},
		{
			Method:      "GET",
			Path:        "/projects/{slug}/tokens",
			Summary:     "List API tokens",
			Description: "Returns the project's machine tokens as metadata only; the secret value is shown once, at creation. Refused for machine tokens.",
			Auth:        "bearer token required, user session only",
			PathParams: []documentation.Field{
				{Name: "slug", Type: "string", Description: "Project slug."},
			},
			ResponseBody: "ApiTokensResponse",
			Errors: []documentation.Error{
				{Status: 403, Code: "permission_denied", Description: "Caller is an API token."},
				{Status: 404, Code: "not_found", Description: "Project not found."},
			},
		},
		{
			Method:      "POST",
			Path:        "/projects/{slug}/tokens",
			Summary:     "Mint an API token",
			Description: "Creates a casier_ machine token bounded by a permission level, an environment grant and an expiry (90 days by default, 730 at most). The raw token is returned once and only its SHA-256 hash is stored. Refused for machine tokens, so a credential cannot mint another.",
			Auth:        "bearer token required, user session only",
			PathParams: []documentation.Field{
				{Name: "slug", Type: "string", Description: "Project slug."},
			},
			RequestBody:  "CreateApiTokenRequest",
			ResponseBody: "ApiTokenResponse",
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Invalid permission, expiry or environment list."},
				{Status: 403, Code: "permission_denied", Description: "Insufficient permissions, or the caller is an API token."},
				{Status: 404, Code: "not_found", Description: "Project or environment not found."},
			},
		},
		{
			Method:      "DELETE",
			Path:        "/projects/{slug}/tokens/{tokenId}",
			Summary:     "Revoke an API token",
			Description: "Revokes a machine token. The row is kept rather than deleted so audit entries naming it still resolve. Refused for machine tokens. Answers 204.",
			Auth:        "bearer token required, user session only",
			PathParams: []documentation.Field{
				{Name: "slug", Type: "string", Description: "Project slug."},
				{Name: "tokenId", Type: "int64", Description: "API token ID to revoke."},
			},
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Token ID is not an integer."},
				{Status: 403, Code: "permission_denied", Description: "Insufficient permissions, or the caller is an API token."},
				{Status: 404, Code: "not_found", Description: "Token not found."},
			},
		},
		{
			Method:      "GET",
			Path:        "/projects/{slug}/webhooks",
			Summary:     "List webhooks",
			Description: "Returns the project's webhook endpoints. Refused for machine tokens.",
			Auth:        "bearer token required, user session only",
			PathParams: []documentation.Field{
				{Name: "slug", Type: "string", Description: "Project slug."},
			},
			ResponseBody: "WebhooksResponse",
			Errors: []documentation.Error{
				{Status: 403, Code: "permission_denied", Description: "Caller is an API token."},
				{Status: 404, Code: "not_found", Description: "Project not found."},
			},
		},
		{
			Method:      "POST",
			Path:        "/projects/{slug}/webhooks",
			Summary:     "Register a webhook",
			Description: "Registers an endpoint notified when the project's secrets change. Requires owner or admin, and is refused for machine tokens.",
			Auth:        "bearer token required, user session only",
			PathParams: []documentation.Field{
				{Name: "slug", Type: "string", Description: "Project slug."},
			},
			RequestBody:  "CreateWebhookRequest",
			ResponseBody: "WebhookResponse",
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Invalid URL or event list."},
				{Status: 403, Code: "permission_denied", Description: "Insufficient permissions, or the caller is an API token."},
				{Status: 404, Code: "not_found", Description: "Project not found."},
			},
		},
		{
			Method:      "DELETE",
			Path:        "/projects/{slug}/webhooks/{webhookId}",
			Summary:     "Delete a webhook",
			Description: "Removes a webhook endpoint. Requires owner or admin, and is refused for machine tokens. Answers 204.",
			Auth:        "bearer token required, user session only",
			PathParams: []documentation.Field{
				{Name: "slug", Type: "string", Description: "Project slug."},
				{Name: "webhookId", Type: "int64", Description: "Webhook ID to delete."},
			},
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Webhook ID is not an integer."},
				{Status: 403, Code: "permission_denied", Description: "Insufficient permissions, or the caller is an API token."},
				{Status: 404, Code: "not_found", Description: "Webhook not found."},
			},
		},
		{
			Method:      "GET",
			Path:        "/projects/{slug}/audit",
			Summary:     "List audit entries",
			Description: "Returns the project's audit log, most recent first. Accepts action to filter and limit to cap the page, defaulting to 50 entries.",
			Auth:        "bearer token required",
			PathParams: []documentation.Field{
				{Name: "slug", Type: "string", Description: "Project slug."},
			},
			ResponseBody: "AuditLogsResponse",
			Errors: []documentation.Error{
				{Status: 403, Code: "permission_denied", Description: "Insufficient permissions."},
				{Status: 404, Code: "not_found", Description: "Project not found."},
			},
		},
		{
			Method:       "POST",
			Path:         "/admin/rotate-key",
			Summary:      "Rotate the root encryption key",
			Description:  "Re-encrypts every project data key under a new base64 32-byte root key, in a single transaction. Restricted to the emails listed in CASIER_ADMIN_EMAILS and refused outright when that list is empty or when the caller is an API token. The response warns that ENCRYPTION_KEY must be updated in the environment before the next restart.",
			Auth:         "bearer token required, instance admin user session",
			RequestBody:  "RotateKeyRequest",
			ResponseBody: "RotateKeyResponse",
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Key is not base64-encoded 32 bytes, or is identical to the current key."},
				{Status: 403, Code: "permission_denied", Description: "Caller is not an instance admin, or is an API token."},
				{Status: 500, Code: "internal", Description: "Rotation failed and was rolled back; the old key is still in force."},
			},
		},
	},
}
