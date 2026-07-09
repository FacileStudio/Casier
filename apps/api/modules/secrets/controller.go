package secrets

import (
	"context"
	"strconv"
	"strings"

	"clef/internal/authcontext"
	"clef/internal/errors"
)

type Controller struct {
	service *Service
}

func newController(service *Service) *Controller {
	return &Controller{service: service}
}

func (controller *Controller) list(ctx context.Context, spaceSlug, envSlug string) (*ListResponse, error) {
	identity, ok := authcontext.IdentityFromContext(ctx)
	if !ok {
		return nil, errors.Unauthorized("missing auth")
	}
	envID, err := controller.resolveEnv(ctx, spaceSlug, envSlug)
	if err != nil {
		return nil, err
	}

	secrets, err := controller.service.listSecrets(ctx, envID, identity.UserID, identity.ClientIP)
	if err != nil {
		return nil, err
	}
	return &ListResponse{Secrets: secrets}, nil
}

func (controller *Controller) create(ctx context.Context, spaceSlug, envSlug string, req *CreateRequest) (*SecretResponse, error) {
	identity, ok := authcontext.IdentityFromContext(ctx)
	if !ok {
		return nil, errors.Unauthorized("missing auth")
	}
	envID, err := controller.resolveEnvWithWrite(ctx, spaceSlug, envSlug)
	if err != nil {
		return nil, err
	}

	key := strings.TrimSpace(req.Key)
	if key == "" {
		return nil, errors.Invalid("key is required")
	}
	if len(key) > 256 {
		return nil, errors.Invalid("secret key exceeds maximum length of 256 characters")
	}

	const maxSecretValueSize = 64 * 1024
	if len(req.Value) > maxSecretValueSize {
		return nil, errors.Invalid("secret value exceeds maximum size of 64KB")
	}

	return controller.service.createSecret(ctx, envID, key, req.Value, identity.UserID, identity.ClientIP)
}

func (controller *Controller) update(ctx context.Context, spaceSlug, envSlug, key string, req *UpdateRequest) (*SecretResponse, error) {
	identity, ok := authcontext.IdentityFromContext(ctx)
	if !ok {
		return nil, errors.Unauthorized("missing auth")
	}
	envID, err := controller.resolveEnvWithWrite(ctx, spaceSlug, envSlug)
	if err != nil {
		return nil, err
	}

	if len(key) > 256 {
		return nil, errors.Invalid("secret key exceeds maximum length of 256 characters")
	}

	const maxSecretValueSize = 64 * 1024
	if len(req.Value) > maxSecretValueSize {
		return nil, errors.Invalid("secret value exceeds maximum size of 64KB")
	}

	return controller.service.updateSecret(ctx, envID, key, req.Value, identity.UserID, identity.ClientIP)
}

func (controller *Controller) delete(ctx context.Context, spaceSlug, envSlug, key string) error {
	identity, ok := authcontext.IdentityFromContext(ctx)
	if !ok {
		return errors.Unauthorized("missing auth")
	}
	envID, err := controller.resolveEnvWithWrite(ctx, spaceSlug, envSlug)
	if err != nil {
		return err
	}

	return controller.service.deleteSecret(ctx, envID, key, identity.UserID, identity.ClientIP)
}

func (controller *Controller) resolveEnv(ctx context.Context, spaceSlug, envSlug string) (int64, error) {
	userID, err := extractUserID(ctx)
	if err != nil {
		return 0, err
	}

	spaceID, err := controller.service.spaceRes.ResolveSpaceID(ctx, spaceSlug, userID)
	if err != nil {
		return 0, err
	}

	return controller.service.envRes.ResolveEnvironmentID(ctx, spaceID, envSlug)
}

func (controller *Controller) resolveEnvWithWrite(ctx context.Context, spaceSlug, envSlug string) (int64, error) {
	userID, err := extractUserID(ctx)
	if err != nil {
		return 0, err
	}

	spaceID, err := controller.service.spaceRes.ResolveSpaceID(ctx, spaceSlug, userID, "owner", "admin", "member")
	if err != nil {
		return 0, err
	}

	return controller.service.envRes.ResolveEnvironmentID(ctx, spaceID, envSlug)
}

func extractUserID(ctx context.Context) (int64, error) {
	identity, ok := authcontext.IdentityFromContext(ctx)
	if !ok {
		return 0, errors.Unauthorized("missing auth")
	}
	id, err := strconv.ParseInt(identity.UserID, 10, 64)
	if err != nil {
		return 0, errors.Invalid("invalid user id")
	}
	return id, nil
}
