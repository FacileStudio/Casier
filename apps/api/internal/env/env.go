package env

import (
	"crypto/sha256"
	"fmt"

	"github.com/FacileStudio/porte"
	troncenv "github.com/FacileStudio/tronc/env"
)

// OIDCConfig is the OpenID Connect client configuration, present only when
// OIDC_ISSUER was set.
type OIDCConfig struct {
	Issuer       string
	ClientID     string
	ClientSecret string
	RedirectURL  string
	SuccessURL   string

	// CLIAudience is the client id the suite's CLI holds its device-grant
	// token under, and setting it is what mounts porte's device exchange so
	// one `facile login` serves this tool too. It is deliberately not
	// OIDC_CLIENT_ID: the token being traded was minted for the CLI, not for
	// Casier, and it is deliberately not the machine audience either, which
	// names this app and arms a different path.
	CLIAudience string

	// MCPResource is the canonical URI of the remote MCP server,
	// https://casier.facile.studio/mcp in production, and the RFC 8707
	// resource Registre binds its tokens to, from MCP_RESOURCE. Empty leaves
	// /mcp unmounted.
	MCPResource string
}

// MCPStateKey derives the key that signs MCP confirmation state from the client
// secret, so it is stable across restarts without a setting of its own.
func (o OIDCConfig) MCPStateKey() []byte {
	sum := sha256.Sum256([]byte("casier-mcp-state:" + o.ClientSecret))
	return sum[:]
}

// Config is the process configuration: the tronc core plus Casier's OIDC and
// service accounts. Sign-in goes through Registre only, so there is no
// password, SMTP or SSO_ONLY setting.
type Config struct {
	troncenv.Core
	EncryptionKey string
	OIDC          *OIDCConfig
	AdminEmails   []string
}

func (c Config) Porte() porte.Config {
	cfg := porte.Config{
		AcceptLegacyCookie: true,
	}
	if c.OIDC == nil {
		return cfg
	}
	cfg.Issuer = c.OIDC.Issuer
	cfg.ClientID = c.OIDC.ClientID
	cfg.ClientSecret = c.OIDC.ClientSecret
	cfg.RedirectURL = c.OIDC.RedirectURL
	cfg.SuccessURL = c.OIDC.SuccessURL
	cfg.CLIAudience = c.OIDC.CLIAudience
	return cfg
}

// Load reads and validates the environment, wiring the OIDC config through
// porte's own Config. A production boot without OIDC_ISSUER is refused: with
// no password sign-in left, such a deployment would have no way in at all.
func Load() (Config, error) {
	core, err := troncenv.LoadCore()
	if err != nil {
		return Config{}, err
	}
	if core.Port < 1 || core.Port > 65535 {
		return Config{}, fmt.Errorf("PORT must be a valid TCP port")
	}

	cfg := Config{Core: core}

	if cfg.EncryptionKey, err = troncenv.Required("ENCRYPTION_KEY"); err != nil {
		return Config{}, err
	}

	cfg.AdminEmails = troncenv.List("CASIER_ADMIN_EMAILS")

	if issuer := troncenv.String("OIDC_ISSUER", ""); issuer != "" {
		oidcConfig, err := loadOIDC(issuer, cfg.CORSAllowedOrigins)
		if err != nil {
			return Config{}, err
		}
		cfg.OIDC = oidcConfig
	}
	if cfg.OIDC == nil && core.IsProduction() {
		return Config{}, fmt.Errorf("OIDC_ISSUER is required in production: Casier signs in through Registre only")
	}

	return cfg, nil
}

func loadOIDC(issuer string, corsOrigins []string) (*OIDCConfig, error) {
	clientID := troncenv.String("OIDC_CLIENT_ID", "")
	clientSecret := troncenv.String("OIDC_CLIENT_SECRET", "")
	redirectURL := troncenv.String("OIDC_REDIRECT_URL", "")
	if clientID == "" || clientSecret == "" || redirectURL == "" {
		return nil, fmt.Errorf("OIDC_CLIENT_ID, OIDC_CLIENT_SECRET, and OIDC_REDIRECT_URL are required when OIDC_ISSUER is set")
	}
	successURL := troncenv.String("OIDC_SUCCESS_URL", "")
	if successURL == "" && len(corsOrigins) > 0 {
		successURL = corsOrigins[0]
	}
	return &OIDCConfig{
		Issuer:       issuer,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		SuccessURL:   successURL,
		CLIAudience:  troncenv.String("OIDC_CLI_AUDIENCE", ""),
		MCPResource:  troncenv.String("MCP_RESOURCE", ""),
	}, nil
}
