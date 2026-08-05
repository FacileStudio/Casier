package environments

import (
	"context"
	"strconv"

	"github.com/FacileStudio/Casier/apps/api/internal/authcontext"
	"github.com/FacileStudio/tronc/errors"
)

type Controller struct {
	service *Service
}

func newController(service *Service) *Controller {
	return &Controller{service: service}
}

func (controller *Controller) list(ctx context.Context, projectSlug string) (*ListResponse, error) {
	userID, err := extractUserID(ctx)
	if err != nil {
		return nil, err
	}

	projectID, err := controller.service.projectResolver.ResolveProjectID(ctx, projectSlug, userID)
	if err != nil {
		return nil, err
	}

	envs, err := controller.service.listEnvironments(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return &ListResponse{Environments: envs}, nil
}

func (controller *Controller) create(ctx context.Context, projectSlug string, req *CreateRequest) (*EnvironmentResponse, error) {
	userID, err := extractUserID(ctx)
	if err != nil {
		return nil, err
	}

	if err := requireWriteScope(ctx); err != nil {
		return nil, err
	}

	projectID, err := controller.service.projectResolver.ResolveProjectID(ctx, projectSlug, userID, "owner", "admin")
	if err != nil {
		return nil, err
	}

	if req.Name == "" {
		return nil, errors.Invalid("name is required")
	}

	return controller.service.createEnvironment(ctx, projectID, req.Name)
}

func (controller *Controller) update(ctx context.Context, projectSlug, envSlug string, req *UpdateRequest) (*EnvironmentResponse, error) {
	userID, err := extractUserID(ctx)
	if err != nil {
		return nil, err
	}

	if err := requireWriteScope(ctx); err != nil {
		return nil, err
	}

	if !scopeAllowsEnvironment(ctx, envSlug) {
		return nil, errors.NotFound("environment not found")
	}

	projectID, err := controller.service.projectResolver.ResolveProjectID(ctx, projectSlug, userID, "owner", "admin")
	if err != nil {
		return nil, err
	}

	if req.Name == "" {
		return nil, errors.Invalid("name is required")
	}

	return controller.service.updateEnvironment(ctx, projectID, envSlug, req.Name)
}

func (controller *Controller) delete(ctx context.Context, projectSlug, envSlug string) error {
	userID, err := extractUserID(ctx)
	if err != nil {
		return err
	}

	if err := requireWriteScope(ctx); err != nil {
		return err
	}

	if !scopeAllowsEnvironment(ctx, envSlug) {
		return errors.NotFound("environment not found")
	}

	projectID, err := controller.service.projectResolver.ResolveProjectID(ctx, projectSlug, userID, "owner", "admin")
	if err != nil {
		return err
	}

	return controller.service.deleteEnvironment(ctx, projectID, envSlug)
}

// requireWriteScope rejects a read-only API token before any mutation runs.
func requireWriteScope(ctx context.Context) error {
	if !scopeAllowsWrite(ctx) {
		return errors.Forbidden("token does not have write permission")
	}
	return nil
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
