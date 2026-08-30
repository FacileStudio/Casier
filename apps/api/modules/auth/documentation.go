package auth

import documentation "github.com/FacileStudio/Casier/apps/api/internal/documentation"

var Documentation = documentation.Module{
	Name:        "auth",
	Description: "Authentication routes.",
	Routes: []documentation.Route{
		{
			Method:       "GET",
			Path:         "/auth/config",
			Summary:      "Describe the authentication modes",
			Description:  "Returns {sso_only, oidc_enabled, password_reset_enabled} so a client knows which sign-in paths exist before offering them. password_reset_enabled is true only when passwords are on and a mailer is configured, so the login screen never offers a reset that cannot be delivered.",
			Auth:         "public",
			ResponseBody: AuthConfigResponse{},
		},
		{
			Method:       "POST",
			Path:         "/auth/register",
			Summary:      "Register a new user",
			Description:  "Creates a user account and returns an auth token.",
			Auth:         "public",
			RequestBody:  RegisterRequest{},
			ResponseBody: AuthResponse{},
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
			RequestBody:  LoginRequest{},
			ResponseBody: AuthResponse{},
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Invalid JSON body or invalid login input."},
				{Status: 401, Code: "unauthenticated", Description: "Email or password is invalid."},
				{Status: 500, Code: "internal", Description: "Unexpected server error."},
			},
		},
		{
			Method:      "POST",
			Path:        "/auth/forgot-password",
			Summary:     "Request a password reset link",
			Description: "Emails a single-use reset link to the address if it belongs to a password account. The response is identical whether or not the address is registered, and delivery happens after the response, so it never reveals which. Only registered when password auth is enabled.",
			Auth:        "public",
			RequestBody: ForgotPasswordRequest{},
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Invalid JSON body or malformed email."},
				{Status: 429, Code: "rate_limited", Description: "Too many requests from this caller."},
				{Status: 503, Code: "unavailable", Description: "No SMTP relay is configured, so no link can be delivered."},
			},
		},
		{
			Method:      "POST",
			Path:        "/auth/reset-password",
			Summary:     "Set a new password from a reset link",
			Description: "Consumes the token from the reset link, sets the new password, invalidates every other outstanding reset token for that user, and deletes all of the user's sessions. Only registered when password auth is enabled.",
			Auth:        "public",
			RequestBody: ResetPasswordRequest{},
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Invalid JSON body, password shorter than 8 characters, or a token that is unknown, already used, or expired."},
				{Status: 429, Code: "rate_limited", Description: "Too many requests from this caller."},
				{Status: 503, Code: "unavailable", Description: "No SMTP relay is configured, so the reset flow is offline."},
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
			ResponseBody: MeResponse{},
			Errors: []documentation.Error{
				{Status: 401, Code: "unauthenticated", Description: "Authorization header is missing or invalid."},
			},
		},
	},
}
