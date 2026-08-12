package environments

import (
	"context"
	stderrors "errors"
	"strconv"

	"github.com/FacileStudio/Casier/apps/api/internal/authcontext"
	"github.com/FacileStudio/tronc/errors"
)

// Controller adapts the environment Service to HTTP request handlers.
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

	projectID, err := controller.service.projectResolver.ResolveProjectID(ctx, projectSlug, userID, protectionRoles...)
	if err != nil {
		return nil, err
	}

	if req.Name == "" {
		return nil, errors.Invalid("name is required")
	}

	return controller.service.createEnvironment(ctx, projectID, req.Name, req.Protected)
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

	projectID, err := controller.service.projectResolver.ResolveProjectID(ctx, projectSlug, userID, protectionRoles...)
	if err != nil {
		return nil, err
	}

	if req.Name == "" {
		return nil, errors.Invalid("name is required")
	}

	return controller.service.updateEnvironment(ctx, projectID, envSlug, req.Name, req.Protected)
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

	projectID, err := controller.service.projectResolver.ResolveProjectID(ctx, projectSlug, userID, protectionRoles...)
	if err != nil {
		return err
	}

	callerIsOwner, err := controller.callerIsOwner(ctx, projectSlug, userID)
	if err != nil {
		return err
	}

	return controller.service.deleteEnvironment(ctx, projectID, envSlug, callerIsOwner)
}

// callerIsOwner asks the membership question the project resolver only answers
// by refusing: it returns an id, never a role, so owner-ness is probed rather
// than read. Only a permission_denied means "not the owner"; any other failure
// is a real error and is propagated.
func (controller *Controller) callerIsOwner(ctx context.Context, projectSlug string, userID int64) (bool, error) {
	if _, err := controller.service.projectResolver.ResolveProjectID(ctx, projectSlug, userID, protectedDeleteRoles...); err != nil {
		var appErr *errors.Error
		if stderrors.As(err, &appErr) && appErr.Code == "permission_denied" {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// requireWriteScope rejects a read-only API token before any mutation runs.
func requireWriteScope(ctx context.Context) error {
	if !scopeAllowsWrite(ctx) {
		return errors.Forbidden("token does not have write permission")
	}
	return nil
}

// requireProtectionScope rejects any API token, however broadly scoped, from
// touching an environment's protection. Write permission buys secret writes,
// not the right to remove the guard standing in front of them.
func requireProtectionScope(ctx context.Context) error {
	if callerIsMachineToken(ctx) {
		return errors.Forbidden("api tokens cannot change environment protection")
	}
	return nil
}

// authorizeProtectionChange runs on update. The owner/admin half of the rule is
// enforced by the resolver call in update; what is left here is the token half,
// and only when the flag actually moves — restating the stored value is not a
// change, and an absent field is not one either.
func authorizeProtectionChange(ctx context.Context, requested *bool, current bool) error {
	if !protectionChanges(requested, current) {
		return nil
	}
	return requireProtectionScope(ctx)
}

// authorizeProtectedDelete runs on delete. Removing a protected environment is
// lifting its protection and destroying its contents at once, so it needs a
// human caller who owns the project; an unprotected environment keeps the
// existing owner-or-admin rule.
func authorizeProtectedDelete(ctx context.Context, protected, callerIsOwner bool) error {
	if !protected {
		return nil
	}
	if err := requireProtectionScope(ctx); err != nil {
		return err
	}
	if !callerIsOwner {
		return errors.Forbidden("only a project owner can delete a protected environment")
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
