package env

import (
	"fmt"

	"github.com/FacileStudio/Casier/apps/api/internal/mailer"
	troncenv "github.com/FacileStudio/tronc/env"
)

type OIDCConfig struct {
	Issuer       string
	ClientID     string
	ClientSecret string
	RedirectURL  string
	SuccessURL   string
}

type Config struct {
	troncenv.Core
	EncryptionKey string
	OIDC          *OIDCConfig
	SMTP          *mailer.SMTPConfig
	PublicURL     string
	SSOOnly       bool
	AdminEmails   []string
}

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

	if cfg.SSOOnly, err = troncenv.Bool("SSO_ONLY", false); err != nil {
		return Config{}, err
	}

	cfg.AdminEmails = troncenv.List("CASIER_ADMIN_EMAILS")

	// The public origin is what a link in an email has to point at. Deployments
	// almost never set it twice, so it falls back to the first CORS origin the
	// same way OIDC_SUCCESS_URL does.
	cfg.PublicURL = troncenv.String("PUBLIC_URL", "")
	if cfg.PublicURL == "" && len(cfg.CORSAllowedOrigins) > 0 {
		cfg.PublicURL = cfg.CORSAllowedOrigins[0]
	}

	if issuer := troncenv.String("OIDC_ISSUER", ""); issuer != "" {
		clientID := troncenv.String("OIDC_CLIENT_ID", "")
		clientSecret := troncenv.String("OIDC_CLIENT_SECRET", "")
		redirectURL := troncenv.String("OIDC_REDIRECT_URL", "")
		if clientID == "" || clientSecret == "" || redirectURL == "" {
			return Config{}, fmt.Errorf("OIDC_CLIENT_ID, OIDC_CLIENT_SECRET, and OIDC_REDIRECT_URL are required when OIDC_ISSUER is set")
		}
		successURL := troncenv.String("OIDC_SUCCESS_URL", "")
		if successURL == "" && len(cfg.CORSAllowedOrigins) > 0 {
			successURL = cfg.CORSAllowedOrigins[0]
		}
		cfg.OIDC = &OIDCConfig{
			Issuer:       issuer,
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			SuccessURL:   successURL,
		}
	}

	// SMTP_HOST is the switch: unset means no mailer and a password reset flow
	// that refuses rather than silently drops mail. Set means the rest has to
	// add up now, at startup, instead of at the first reset request.
	if host := troncenv.String("SMTP_HOST", ""); host != "" {
		port, err := troncenv.Int("SMTP_PORT", mailer.DefaultPort)
		if err != nil {
			return Config{}, err
		}
		from := troncenv.String("SMTP_FROM", "")
		if from == "" {
			return Config{}, fmt.Errorf("SMTP_FROM is required when SMTP_HOST is set")
		}
		if cfg.PublicURL == "" {
			return Config{}, fmt.Errorf("PUBLIC_URL or CORS_ALLOWED_ORIGINS is required when SMTP_HOST is set, a password reset link needs a public origin")
		}
		cfg.SMTP = &mailer.SMTPConfig{
			Host:     host,
			Port:     port,
			Username: troncenv.String("SMTP_USERNAME", ""),
			Password: troncenv.String("SMTP_PASSWORD", ""),
			From:     from,
		}
	}

	return cfg, nil
}
