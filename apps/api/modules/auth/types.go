package auth

import "github.com/FacileStudio/Casier/apps/api/internal/authcontext"

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

// GetScope returns the API token scope the caller authenticated with, or nil
// for a user session, which is bounded by project roles alone.
func (d *Data) GetScope() *authcontext.TokenScope {
	if d == nil {
		return nil
	}
	return d.Scope
}

type CLIExchangeRequest struct {
	IDToken string `json:"id_token"`
}

type CLIExchangeResponse struct {
	Token string `json:"token"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

type ResetPasswordRequest struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

const SessionCookieName = "casier_session"
