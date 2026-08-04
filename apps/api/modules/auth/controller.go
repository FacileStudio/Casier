package auth

import (
	"context"
	"strings"

	"casier/internal/authcontext"
	"casier/internal/errors"
)

type Controller struct {
	service *Service
}

func newController(service *Service) *Controller {
	return &Controller{service: service}
}

func (controller *Controller) register(ctx context.Context, req *RegisterRequest) (*AuthResponse, error) {
	email := strings.TrimSpace(strings.ToLower(req.Email))
	if email == "" || !strings.Contains(email, "@") {
		return nil, errors.Invalid("invalid email")
	}
	if len(req.Password) < 8 {
		return nil, errors.Invalid("password must be at least 8 characters")
	}

	userID, token, err := controller.service.registerUser(ctx, email, req.Password)
	if err != nil {
		return nil, err
	}
	return &AuthResponse{UserID: userID, Token: token}, nil
}

func (controller *Controller) login(ctx context.Context, req *LoginRequest) (*AuthResponse, error) {
	email := strings.TrimSpace(strings.ToLower(req.Email))
	if email == "" || req.Password == "" {
		return nil, errors.Invalid("email and password required")
	}

	userID, token, err := controller.service.loginUser(ctx, email, req.Password)
	if err != nil {
		return nil, err
	}
	return &AuthResponse{UserID: userID, Token: token}, nil
}

func (controller *Controller) forgotPassword(ctx context.Context, req *ForgotPasswordRequest) error {
	email := strings.TrimSpace(strings.ToLower(req.Email))
	if email == "" || !strings.Contains(email, "@") {
		return errors.Invalid("invalid email")
	}
	return controller.service.forgotPassword(ctx, email)
}

func (controller *Controller) resetPassword(ctx context.Context, req *ResetPasswordRequest) error {
	if req.Token == "" {
		return errors.Invalid("token is required")
	}
	if len(req.Password) < 8 {
		return errors.Invalid("password must be at least 8 characters")
	}
	return controller.service.resetPassword(ctx, req.Token, req.Password)
}

func (controller *Controller) me(ctx context.Context) (*MeResponse, error) {
	identity, ok := authcontext.IdentityFromContext(ctx)
	if !ok {
		return nil, errors.Unauthorized("missing auth")
	}
	return &MeResponse{ID: identity.UserID, Email: identity.Email}, nil
}
