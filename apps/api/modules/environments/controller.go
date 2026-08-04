package environments

import (
	"context"
	"strconv"

	"casier/internal/authcontext"
	"casier/internal/errors"
)

type Controller struct {
	service *Service
}

func newController(service *Service) *Controller {
	return &Controller{service: service}
}

func (controller *Controller) list(ctx context.Context, spaceSlug string) (*ListResponse, error) {
	userID, err := extractUserID(ctx)
	if err != nil {
		return nil, err
	}

	spaceID, err := controller.service.spaceResolver.ResolveSpaceID(ctx, spaceSlug, userID)
	if err != nil {
		return nil, err
	}

	envs, err := controller.service.listEnvironments(ctx, spaceID)
	if err != nil {
		return nil, err
	}
	return &ListResponse{Environments: envs}, nil
}

func (controller *Controller) create(ctx context.Context, spaceSlug string, req *CreateRequest) (*EnvironmentResponse, error) {
	userID, err := extractUserID(ctx)
	if err != nil {
		return nil, err
	}

	spaceID, err := controller.service.spaceResolver.ResolveSpaceID(ctx, spaceSlug, userID, "owner", "admin")
	if err != nil {
		return nil, err
	}

	if req.Name == "" {
		return nil, errors.Invalid("name is required")
	}

	return controller.service.createEnvironment(ctx, spaceID, req.Name)
}

func (controller *Controller) update(ctx context.Context, spaceSlug, envSlug string, req *UpdateRequest) (*EnvironmentResponse, error) {
	userID, err := extractUserID(ctx)
	if err != nil {
		return nil, err
	}

	spaceID, err := controller.service.spaceResolver.ResolveSpaceID(ctx, spaceSlug, userID, "owner", "admin")
	if err != nil {
		return nil, err
	}

	if req.Name == "" {
		return nil, errors.Invalid("name is required")
	}

	return controller.service.updateEnvironment(ctx, spaceID, envSlug, req.Name)
}

func (controller *Controller) delete(ctx context.Context, spaceSlug, envSlug string) error {
	userID, err := extractUserID(ctx)
	if err != nil {
		return err
	}

	spaceID, err := controller.service.spaceResolver.ResolveSpaceID(ctx, spaceSlug, userID, "owner", "admin")
	if err != nil {
		return err
	}

	return controller.service.deleteEnvironment(ctx, spaceID, envSlug)
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
