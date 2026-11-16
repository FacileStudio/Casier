package env

import (
	"fmt"
	"os"

	"github.com/FacileStudio/porte"
	troncenv "github.com/FacileStudio/tronc/env"
)

// Config is the process configuration: the tronc core plus Casier's OIDC and
// service accounts. Sign-in goes through Registre only, so there is no
// password or SMTP setting.
type Config struct {
	troncenv.Core
	EncryptionKey string
	OIDC          porte.Config
	AdminEmails   []string

	// MCPResource is the canonical URI of the remote MCP server,
	// https://casier.facile.studio/mcp in production, and the RFC 8707
	// resource Registre binds its tokens to, from MCP_RESOURCE. Empty leaves
	// /mcp unmounted.
	MCPResource string
}

// Load reads and validates the environment. A production boot without
// OIDC_ISSUER is refused: with no password sign-in left, such a deployment
// would have no way in at all.
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
	cfg.MCPResource = troncenv.String("MCP_RESOURCE", "")

	defaults := porte.Config{AcceptLegacyCookie: true}
	if len(core.CORSAllowedOrigins) > 0 {
		defaults.SuccessURL = core.CORSAllowedOrigins[0]
	}
	if cfg.OIDC, err = porte.ConfigFromEnv(os.Getenv, defaults); err != nil {
		return Config{}, err
	}
	if !cfg.OIDC.Enabled() && core.IsProduction() {
		return Config{}, fmt.Errorf("OIDC_ISSUER is required in production: Casier signs in through Registre only")
	}

	return cfg, nil
}
