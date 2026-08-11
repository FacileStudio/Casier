package auth

import "github.com/FacileStudio/Casier/apps/api/internal/authcontext"

type AuthConfigResponse struct {
	SSOOnly              bool `json:"sso_only"`
	OIDCEnabled          bool `json:"oidc_enabled"`
	PasswordResetEnabled bool `json:"password_reset_enabled"`
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	UserID string `json:"user_id"`
	Token  string `json:"token"`
}

type MeResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

type ResetPasswordRequest struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

type Data struct {
	Email string                  `json:"email"`
	Scope *authcontext.TokenScope `json:"-"`
}

func (d *Data) GetEmail() string {
	if d == nil {
		return ""
	}
	return d.Email
}

func (d *Data) GetScope() *authcontext.TokenScope {
	if d == nil {
		return nil
	}
	return d.Scope
}
