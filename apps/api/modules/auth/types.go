package auth

import "github.com/FacileStudio/Casier/apps/api/internal/authcontext"

// AuthConfigResponse tells the client which flows a deployment offers.
type AuthConfigResponse struct {
	SSOOnly              bool `json:"sso_only"`
	OIDCEnabled          bool `json:"oidc_enabled"`
	PasswordResetEnabled bool `json:"password_reset_enabled"`
}

// RegisterRequest is the body of a local password registration.
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginRequest is the body of a local password login.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse carries the session token issued to a successful caller.
type AuthResponse struct {
	UserID string `json:"user_id"`
	Token  string `json:"token"`
}

// MeResponse describes the signed-in user.
type MeResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

// ForgotPasswordRequest asks for a password reset email for an address.
type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

// ResetPasswordRequest carries the reset token and the new password.
type ResetPasswordRequest struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

// Data is the authenticated caller payload: the email, plus the token scope
// when the request came in on a machine token.
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
