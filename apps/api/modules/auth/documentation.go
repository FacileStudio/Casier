package auth

import documentation "github.com/FacileStudio/Casier/apps/api/internal/documentation"

var Documentation = documentation.Module{
	Name:        "auth",
	Description: "Authentication routes.",
	Routes: []documentation.Route{
		{
			Method:       "GET",
			Path:         "/auth/config",
			Summary:      "Describe the authentication mode",
			Description:  "Returns {sso_only, oidc_enabled}. sso_only is always true: sign-in goes through Registre only, and there is no password route.",
			Auth:         "public",
			ResponseBody: AuthConfigResponse{},
		},
		{
			Method:      "POST",
			Path:        "/registre/deletions",
			Summary:     "Receive a Registre deletion notice",
			Description: "Registre's signed user.deleted push. The body is a security event token (typ secevent+jwt, RS256) that the handler verifies against the issuer's keys; it erases the person's personal projects and account, keeps the secrets of spaces other members use, and anonymises the person's authorship there. Casier also polls Registre for the same notices, so a missed push is harmless.",
			Auth:        "signed security event token",
			Errors: []documentation.Error{
				{Status: 400, Code: "", Description: "The notice does not verify."},
				{Status: 503, Code: "", Description: "The issuer's keys cannot be fetched; Registre retries."},
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
