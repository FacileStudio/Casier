package auth

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/FacileStudio/Casier/apps/api/internal/authcontext"
	"github.com/FacileStudio/tronc/errors"
)

type Controller struct {
	service *Service
}

func newController(service *Service) *Controller {
	return &Controller{service: service}
}

func (c *Controller) register(ctx context.Context, w http.ResponseWriter, r *http.Request, req *RegisterRequest) (*AuthResponse, error) {
	email := strings.TrimSpace(strings.ToLower(req.Email))
	if email == "" || !strings.Contains(email, "@") {
		return nil, errors.Invalid("invalid email")
	}
	if len(req.Password) < 8 {
		return nil, errors.Invalid("password must be at least 8 characters")
	}

	userID, token, err := c.service.passwords.Register(ctx, w, r, email, "", req.Password)
	if err != nil {
		return nil, err
	}
	return &AuthResponse{UserID: intToStr(userID), Token: token}, nil
}

func (c *Controller) login(ctx context.Context, w http.ResponseWriter, r *http.Request, req *LoginRequest) (*AuthResponse, error) {
	email := strings.TrimSpace(strings.ToLower(req.Email))
	if email == "" || req.Password == "" {
		return nil, errors.Invalid("email and password required")
	}

	userID, token, err := c.service.passwords.Login(ctx, w, r, email, req.Password)
	if err != nil {
		return nil, err
	}
	return &AuthResponse{UserID: intToStr(userID), Token: token}, nil
}

func (c *Controller) me(ctx context.Context) (*MeResponse, error) {
	identity, ok := authcontext.IdentityFromContext(ctx)
	if !ok {
		return nil, errors.Unauthorized("missing auth")
	}
	return &MeResponse{ID: identity.UserID, Email: identity.Email}, nil
}

func intToStr(id int64) string {
	return strconv.FormatInt(id, 10)
}
