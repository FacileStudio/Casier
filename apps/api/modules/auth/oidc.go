package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/FacileStudio/Casier/apps/api/internal/authcontext"
	"github.com/FacileStudio/Casier/apps/api/internal/env"
	"github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"

	gooidc "github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

const (
	oidcStateCookie = "oidc_state"
	oidcCLICookie   = "oidc_cli"
	oidcStateTTL    = 10 * time.Minute
)

type oidcHandler struct {
	provider   *gooidc.Provider
	verifier   *gooidc.IDTokenVerifier
	oauth2Cfg  oauth2.Config
	service    *Service
	successURL string
}

func newOIDCHandler(ctx context.Context, cfg *env.OIDCConfig, service *Service) (*oidcHandler, error) {
	provider, err := gooidc.NewProvider(ctx, cfg.Issuer)
	if err != nil {
		return nil, err
	}
	oauth2Cfg := oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.RedirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{gooidc.ScopeOpenID, "email", "profile", "offline_access"},
	}
	verifier := provider.Verifier(&gooidc.Config{ClientID: cfg.ClientID})
	return &oidcHandler{
		provider:   provider,
		verifier:   verifier,
		oauth2Cfg:  oauth2Cfg,
		service:    service,
		successURL: cfg.SuccessURL,
	}, nil
}

// emailIsVerified decides whether an OIDC identity may be trusted to name a
// local account. Identity crosses the wire by email here, so an IdP that lets a
// user set an unverified address is an account-takeover path: claim the email of
// an existing owner and the callback hands back their session.
//
// The claim is a pointer so that "absent" and "false" stay distinguishable. An
// explicit false is a provider telling us the address is unproven, and it is
// refused. Absent means the provider publishes no verification signal at all —
// there is nothing to act on, and failing closed there would lock out every
// deployment whose IdP omits the claim.
func emailIsVerified(verified *bool) bool {
	return verified == nil || *verified
}

func (h *oidcHandler) login(w http.ResponseWriter, r *http.Request) {
	state, err := randomState()
	if err != nil {
		httpjson.WriteError(w, errors.Internal("failed to generate state", err))
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     oidcStateCookie,
		Value:    state,
		Path:     "/",
		MaxAge:   int(oidcStateTTL.Seconds()),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	cli, isCLI := parseCLICallback(r.URL.Query().Get("cli_port"), r.URL.Query().Get("cli_state"))
	if isCLI {
		http.SetCookie(w, &http.Cookie{
			Name:     oidcCLICookie,
			Value:    cli.encode(),
			Path:     "/",
			MaxAge:   int(oidcStateTTL.Seconds()),
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
		})
	}

	http.Redirect(w, r, h.oauth2Cfg.AuthCodeURL(state), http.StatusFound)
}

func (h *oidcHandler) callback(w http.ResponseWriter, r *http.Request) {
	stateCookie, err := r.Cookie(oidcStateCookie)
	if err != nil || stateCookie.Value != r.URL.Query().Get("state") {
		httpjson.WriteError(w, errors.Invalid("invalid oauth2 state"))
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     oidcStateCookie,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
	})

	oauth2Token, err := h.oauth2Cfg.Exchange(r.Context(), r.URL.Query().Get("code"))
	if err != nil {
		httpjson.WriteError(w, errors.Internal("failed to exchange code", err))
		return
	}

	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		httpjson.WriteError(w, errors.Internal("missing id_token in response", nil))
		return
	}

	idToken, err := h.verifier.Verify(r.Context(), rawIDToken)
	if err != nil {
		httpjson.WriteError(w, errors.Unauthorized("invalid id_token"))
		return
	}

	var claims struct {
		Email         string `json:"email"`
		EmailVerified *bool  `json:"email_verified"`
		Name          string `json:"name"`
	}
	if err := idToken.Claims(&claims); err != nil {
		httpjson.WriteError(w, errors.Internal("failed to parse claims", err))
		return
	}
	if claims.Email == "" {
		httpjson.WriteError(w, errors.Invalid("OIDC provider did not return an email"))
		return
	}
	if !emailIsVerified(claims.EmailVerified) {
		httpjson.WriteError(w, errors.Unauthorized("email is not verified"))
		return
	}

	_, token, err := h.service.upsertOIDCUser(r.Context(), claims.Email, claims.Name, oauth2Token)
	if err != nil {
		httpjson.WriteError(w, err)
		return
	}

	SetSessionCookie(w, token)

	if cli, isCLI := h.takeCLICallback(w, r); isCLI {
		cliToken, err := h.service.exchangeCLIToken(r.Context(), claims.Email)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		http.Redirect(w, r, cli.redirectURL(cliToken), http.StatusFound)
		return
	}

	http.Redirect(w, r, h.successURL, http.StatusFound)
}

func (h *oidcHandler) takeCLICallback(w http.ResponseWriter, r *http.Request) (cliCallback, bool) {
	cookie, err := r.Cookie(oidcCLICookie)
	if err != nil {
		return cliCallback{}, false
	}
	http.SetCookie(w, &http.Cookie{
		Name:     oidcCLICookie,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
	})
	return decodeCLICallback(cookie.Value)
}

func (h *oidcHandler) cliExchange(w http.ResponseWriter, r *http.Request) {
	var req CLIExchangeRequest
	if err := httpjson.DecodeJSON(w, r, &req); err != nil {
		httpjson.WriteError(w, err)
		return
	}
	if req.IDToken == "" {
		httpjson.WriteError(w, errors.Invalid("id_token is required"))
		return
	}

	idToken, err := h.verifier.Verify(r.Context(), req.IDToken)
	if err != nil {
		httpjson.WriteError(w, errors.Unauthorized("invalid id_token"))
		return
	}

	var claims struct {
		Email         string `json:"email"`
		EmailVerified *bool  `json:"email_verified"`
	}
	if err := idToken.Claims(&claims); err != nil || claims.Email == "" {
		httpjson.WriteError(w, errors.Unauthorized("invalid id_token"))
		return
	}
	if !emailIsVerified(claims.EmailVerified) {
		httpjson.WriteError(w, errors.Unauthorized("email is not verified"))
		return
	}

	token, err := h.service.exchangeCLIToken(r.Context(), claims.Email)
	if err != nil {
		httpjson.WriteError(w, err)
		return
	}
	httpjson.WriteJSON(w, http.StatusOK, CLIExchangeResponse{Token: token})
}

func (h *oidcHandler) syncProfile(w http.ResponseWriter, r *http.Request) {
	identity, ok := authcontext.IdentityFromContext(r.Context())
	if !ok {
		httpjson.WriteError(w, errors.Unauthorized("missing auth"))
		return
	}

	synced, err := h.service.SyncOIDCProfile(r.Context(), identity.UserID, h.provider, h.oauth2Cfg)
	if err != nil {
		httpjson.WriteError(w, err)
		return
	}
	httpjson.WriteJSON(w, http.StatusOK, map[string]bool{"synced": synced})
}

func randomState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

type cliCallback struct {
	port  int
	state string
}

func parseCLICallback(port, state string) (cliCallback, bool) {
	parsedPort, err := strconv.Atoi(port)
	if err != nil || parsedPort < 1024 || parsedPort > 65535 {
		return cliCallback{}, false
	}
	if state == "" || len(state) > 64 || !isURLSafeToken(state) {
		return cliCallback{}, false
	}
	return cliCallback{port: parsedPort, state: state}, true
}

func decodeCLICallback(value string) (cliCallback, bool) {
	port, state, found := strings.Cut(value, ":")
	if !found {
		return cliCallback{}, false
	}
	return parseCLICallback(port, state)
}

func (c cliCallback) encode() string {
	return strconv.Itoa(c.port) + ":" + c.state
}

func (c cliCallback) redirectURL(token string) string {
	return fmt.Sprintf(
		"http://127.0.0.1:%d/callback?token=%s&state=%s",
		c.port,
		url.QueryEscape(token),
		url.QueryEscape(c.state),
	)
}

func isURLSafeToken(value string) bool {
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_':
		default:
			return false
		}
	}
	return true
}
