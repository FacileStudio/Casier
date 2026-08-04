package projects

import documentation "casier/internal/documentation"

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
			Description: "Updates project name and description. Requires owner or admin role.",
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
	},
}
