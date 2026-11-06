package auth

import "github.com/FacileStudio/Casier/apps/api/internal/authcontext"

// AuthConfigResponse tells the client which flows a deployment offers.
type AuthConfigResponse struct {
	SSOOnly     bool `json:"sso_only"`
	OIDCEnabled bool `json:"oidc_enabled"`
}

// MeResponse describes the signed-in user: the id plus the email, name,
// picture URL and locale the client draws them with.
type MeResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
	Locale    string `json:"locale"`
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
