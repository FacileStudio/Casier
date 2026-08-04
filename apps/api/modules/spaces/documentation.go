package spaces

import documentation "casier/internal/documentation"

var Documentation = documentation.Module{
	Name:        "spaces",
	Description: "Space management and membership routes.",
	Routes: []documentation.Route{
		{
			Method:       "GET",
			Path:         "/spaces",
			Summary:      "List spaces",
			Description:  "Returns all spaces the authenticated user belongs to.",
			Auth:         "bearer token required",
			ResponseBody: "ListResponse",
			Errors: []documentation.Error{
				{Status: 401, Code: "unauthenticated", Description: "Authorization header is missing or invalid."},
				{Status: 500, Code: "internal", Description: "Unexpected server error."},
			},
		},
		{
			Method:       "POST",
			Path:         "/spaces",
			Summary:      "Create a space",
			Description:  "Creates a new space with default environments (dev, staging, prod). The creator becomes the owner.",
			Auth:         "bearer token required",
			RequestBody:  "CreateRequest",
			ResponseBody: "SpaceResponse",
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Invalid JSON body or missing name."},
				{Status: 409, Code: "already_exists", Description: "A space with the same slug already exists."},
				{Status: 500, Code: "internal", Description: "Unexpected server error."},
			},
		},
		{
			Method:      "GET",
			Path:        "/spaces/{slug}",
			Summary:     "Get a space",
			Description: "Returns a single space by slug.",
			Auth:        "bearer token required",
			PathParams: []documentation.Field{
				{Name: "slug", Type: "string", Description: "Space slug."},
			},
			ResponseBody: "SpaceResponse",
			Errors: []documentation.Error{
				{Status: 404, Code: "not_found", Description: "Space not found or user is not a member."},
			},
		},
		{
			Method:      "PUT",
			Path:        "/spaces/{slug}",
			Summary:     "Update a space",
			Description: "Updates space name and description. Requires owner or admin role.",
			Auth:        "bearer token required",
			PathParams: []documentation.Field{
				{Name: "slug", Type: "string", Description: "Space slug."},
			},
			RequestBody:  "UpdateRequest",
			ResponseBody: "SpaceResponse",
			Errors: []documentation.Error{
				{Status: 403, Code: "permission_denied", Description: "Insufficient permissions."},
				{Status: 404, Code: "not_found", Description: "Space not found."},
			},
		},
		{
			Method:      "DELETE",
			Path:        "/spaces/{slug}",
			Summary:     "Delete a space",
			Description: "Deletes a space. Owner only.",
			Auth:        "bearer token required",
			PathParams: []documentation.Field{
				{Name: "slug", Type: "string", Description: "Space slug."},
			},
			Errors: []documentation.Error{
				{Status: 403, Code: "permission_denied", Description: "Only the owner can delete a space."},
				{Status: 404, Code: "not_found", Description: "Space not found."},
			},
		},
		{
			Method:      "GET",
			Path:        "/spaces/{slug}/members",
			Summary:     "List members",
			Description: "Returns all members of the space.",
			Auth:        "bearer token required",
			PathParams: []documentation.Field{
				{Name: "slug", Type: "string", Description: "Space slug."},
			},
			ResponseBody: "MembersResponse",
			Errors: []documentation.Error{
				{Status: 404, Code: "not_found", Description: "Space not found."},
			},
		},
		{
			Method:      "POST",
			Path:        "/spaces/{slug}/members",
			Summary:     "Add a member",
			Description: "Adds a user to the space by email. Requires owner or admin role.",
			Auth:        "bearer token required",
			PathParams: []documentation.Field{
				{Name: "slug", Type: "string", Description: "Space slug."},
			},
			RequestBody: "AddMemberRequest",
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Invalid input."},
				{Status: 403, Code: "permission_denied", Description: "Insufficient permissions."},
				{Status: 404, Code: "not_found", Description: "User or space not found."},
				{Status: 409, Code: "already_exists", Description: "User is already a member."},
			},
		},
		{
			Method:      "DELETE",
			Path:        "/spaces/{slug}/members/{userId}",
			Summary:     "Remove a member",
			Description: "Removes a user from the space. Requires owner or admin role. Returns a warning reminding to rotate credentials the removed member may have fetched.",
			Auth:        "bearer token required",
			PathParams: []documentation.Field{
				{Name: "slug", Type: "string", Description: "Space slug."},
				{Name: "userId", Type: "int64", Description: "User ID to remove."},
			},
			ResponseBody: "RemoveMemberResponse",
			Errors: []documentation.Error{
				{Status: 403, Code: "permission_denied", Description: "Insufficient permissions."},
				{Status: 404, Code: "not_found", Description: "Member not found."},
			},
		},
	},
}
