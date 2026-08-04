package auth

import documentation "casier/internal/documentation"

var Documentation = documentation.Module{
	Name:        "auth",
	Description: "Authentication routes.",
	Routes: []documentation.Route{
		{
			Method:       "POST",
			Path:         "/auth/register",
			Summary:      "Register a new user",
			Description:  "Creates a user account and returns an auth token.",
			Auth:         "public",
			RequestBody:  "RegisterRequest",
			ResponseBody: "AuthResponse",
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Invalid JSON body or invalid registration input."},
				{Status: 409, Code: "already_exists", Description: "A user with the same email already exists."},
				{Status: 500, Code: "internal", Description: "Unexpected server error."},
			},
		},
		{
			Method:       "POST",
			Path:         "/auth/login",
			Summary:      "Authenticate a user",
			Description:  "Authenticates credentials and returns an auth token.",
			Auth:         "public",
			RequestBody:  "LoginRequest",
			ResponseBody: "AuthResponse",
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Invalid JSON body or invalid login input."},
				{Status: 401, Code: "unauthenticated", Description: "Email or password is invalid."},
				{Status: 500, Code: "internal", Description: "Unexpected server error."},
			},
		},
		{
			Method:       "POST",
			Path:         "/auth/cli",
			Summary:      "Exchange an OIDC ID token for a CLI auth token",
			Description:  "Validates an OIDC ID token against the configured issuer and returns an auth token for an existing user. Only registered when OIDC is configured.",
			Auth:         "public",
			RequestBody:  "CLIExchangeRequest",
			ResponseBody: "CLIExchangeResponse",
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Invalid JSON body or missing id_token."},
				{Status: 401, Code: "unauthenticated", Description: "ID token is invalid, expired, or not issued for this client."},
				{Status: 403, Code: "permission_denied", Description: "No account exists for this identity."},
				{Status: 404, Code: "not_found", Description: "OIDC is not configured."},
				{Status: 500, Code: "internal", Description: "Unexpected server error."},
			},
		},
		{
			Method:      "POST",
			Path:        "/auth/logout",
			Summary:     "Log out",
			Description: "Invalidates the current session token.",
			Auth:        "bearer token required",
			Errors: []documentation.Error{
				{Status: 401, Code: "unauthenticated", Description: "Authorization header is missing or invalid."},
				{Status: 500, Code: "internal", Description: "Unexpected server error."},
			},
		},
		{
			Method:       "GET",
			Path:         "/auth/me",
			Summary:      "Get current user",
			Description:  "Returns the authenticated user's profile.",
			Auth:         "bearer token required",
			ResponseBody: "MeResponse",
			Errors: []documentation.Error{
				{Status: 401, Code: "unauthenticated", Description: "Authorization header is missing or invalid."},
			},
		},
	},
}
