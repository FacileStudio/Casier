package secrets

import (
	"context"
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

func (controller *Controller) list(ctx context.Context, projectSlug, envSlug, tagFilter string) (*ListResponse, error) {
	envID, person, err := controller.resolveEnv(ctx, projectSlug, envSlug)
	if err != nil {
		return nil, err
	}

	secrets, rotationDays, err := controller.service.listSecrets(ctx, envID, person, strings.TrimSpace(tagFilter))
	if err != nil {
		return nil, err
	}
	return &ListResponse{Secrets: secrets, RotationDays: rotationDays}, nil
}

func (controller *Controller) rotation(ctx context.Context, projectSlug string) (*RotationSummaryResponse, error) {
	userID, err := extractUserID(ctx)
	if err != nil {
		return nil, err
	}
	projectID, err := controller.service.projectRes.ResolveProjectID(ctx, projectSlug, userID)
	if err != nil {
		return nil, err
	}
	return controller.service.rotationSummary(ctx, projectID)
}

func (controller *Controller) create(ctx context.Context, projectSlug, envSlug string, req *CreateRequest) (*SecretResponse, error) {
	envID, person, err := controller.resolveEnvWithWrite(ctx, projectSlug, envSlug)
	if err != nil {
		return nil, err
	}

	key := strings.TrimSpace(req.Key)
	if key == "" {
		return nil, errors.Invalid("key is required")
	}
	if len(key) > maxSecretKeyLength {
		return nil, errors.Invalid("secret key exceeds maximum length of 256 characters")
	}

	if len(req.Value) > maxSecretValueSize {
		return nil, errors.Invalid("secret value exceeds maximum size of 64KB")
	}

	return controller.service.createSecret(ctx, envID, key, req.Value, person, req.Tags)
}

func (controller *Controller) update(ctx context.Context, projectSlug, envSlug, key string, req *UpdateRequest) (*SecretResponse, error) {
	envID, person, err := controller.resolveEnvWithWrite(ctx, projectSlug, envSlug)
	if err != nil {
		return nil, err
	}

	if len(key) > maxSecretKeyLength {
		return nil, errors.Invalid("secret key exceeds maximum length of 256 characters")
	}

	if len(req.Value) > maxSecretValueSize {
		return nil, errors.Invalid("secret value exceeds maximum size of 64KB")
	}

	return controller.service.updateSecret(ctx, envID, key, req.Value, person)
}

func (controller *Controller) delete(ctx context.Context, projectSlug, envSlug, key string) error {
	envID, person, err := controller.resolveEnvWithWrite(ctx, projectSlug, envSlug)
	if err != nil {
		return err
	}

	return controller.service.deleteSecret(ctx, envID, key, person)
}

func (controller *Controller) importEnv(ctx context.Context, projectSlug, envSlug string, req *ImportRequest) (*ImportResponse, error) {
	envID, person, err := controller.resolveEnvWithWrite(ctx, projectSlug, envSlug)
	if err != nil {
		return nil, err
	}
	return controller.service.importEnv(ctx, envID, req.Content, person)
}

func (controller *Controller) exportEnv(ctx context.Context, projectSlug, envSlug string) (*ExportResponse, error) {
	envID, person, err := controller.resolveEnv(ctx, projectSlug, envSlug)
	if err != nil {
		return nil, err
	}
	content, err := controller.service.exportEnv(ctx, envID, person)
	if err != nil {
		return nil, err
	}
	return &ExportResponse{Content: content}, nil
}

func (controller *Controller) updateTags(ctx context.Context, projectSlug, envSlug, key string, req *UpdateTagsRequest) error {
	envID, person, err := controller.resolveEnvWithWrite(ctx, projectSlug, envSlug)
	if err != nil {
		return err
	}
	return controller.service.updateTags(ctx, envID, key, req.Tags, person)
}

func (controller *Controller) listVersions(ctx context.Context, projectSlug, envSlug, key string) (*VersionsResponse, error) {
	envID, _, err := controller.resolveEnv(ctx, projectSlug, envSlug)
	if err != nil {
		return nil, err
	}
	versions, err := controller.service.listVersions(ctx, envID, key)
	if err != nil {
		return nil, err
	}
	return &VersionsResponse{Versions: versions}, nil
}

func (controller *Controller) rollback(ctx context.Context, projectSlug, envSlug, key string, req *RollbackRequest) (*SecretResponse, error) {
	envID, person, err := controller.resolveEnvWithWrite(ctx, projectSlug, envSlug)
	if err != nil {
		return nil, err
	}
	if req.Version < 1 {
		return nil, errors.Invalid("version must be >= 1")
	}
	return controller.service.rollbackSecret(ctx, envID, key, req.Version, person)
}

func (controller *Controller) diff(ctx context.Context, projectSlug, envSlugA, envSlugB string) (*DiffResponse, error) {
	envA, person, err := controller.resolveEnv(ctx, projectSlug, envSlugA)
	if err != nil {
		return nil, err
	}
	envB, _, err := controller.resolveEnv(ctx, projectSlug, envSlugB)
	if err != nil {
		return nil, err
	}

	resp, err := controller.service.diffEnvironments(ctx, envA, envB, person)
	if err != nil {
		return nil, err
	}
	resp.From = envSlugA
	resp.To = envSlugB
	return resp, nil
}

func (controller *Controller) listACLs(ctx context.Context, projectSlug, envSlug, key string) (*ACLsResponse, error) {
	envID, _, err := controller.resolveEnv(ctx, projectSlug, envSlug)
	if err != nil {
		return nil, err
	}
	acls, err := controller.service.listACLs(ctx, envID, key)
	if err != nil {
		return nil, err
	}
	return &ACLsResponse{ACLs: acls}, nil
}

func (controller *Controller) setACL(ctx context.Context, projectSlug, envSlug, key string, req *SetACLRequest) error {
	envID, _, err := controller.resolveEnvWithRoles(ctx, projectSlug, envSlug, "owner", "admin")
	if err != nil {
		return err
	}
	if req.Email == "" {
		return errors.Invalid("email is required")
	}
	return controller.service.setACL(ctx, envID, key, req.Email, req.CanRead, req.CanWrite)
}

func (controller *Controller) deleteACL(ctx context.Context, projectSlug, envSlug, key string, aclID int64) error {
	envID, _, err := controller.resolveEnvWithRoles(ctx, projectSlug, envSlug, "owner", "admin")
	if err != nil {
		return err
	}
	return controller.service.deleteACL(ctx, envID, key, aclID)
}

func (controller *Controller) resolveEnv(ctx context.Context, projectSlug, envSlug string) (int64, caller, error) {
	return controller.resolveEnvFor(ctx, projectSlug, envSlug)
}

// resolveEnvWithWrite is the choke point every mutating secret route goes
// through: it rejects a read-only API token before any role is even looked up,
// then demands a role that may write in the project.
func (controller *Controller) resolveEnvWithWrite(ctx context.Context, projectSlug, envSlug string) (int64, caller, error) {
	return controller.resolveEnvWithRoles(ctx, projectSlug, envSlug, "owner", "admin", "member")
}

func (controller *Controller) resolveEnvWithRoles(ctx context.Context, projectSlug, envSlug string, requiredRoles ...string) (int64, caller, error) {
	if !authcontext.ScopeFromContext(ctx).AllowsWrite() {
		return 0, caller{}, errors.Forbidden("this API token is read-only")
	}
	return controller.resolveEnvFor(ctx, projectSlug, envSlug, requiredRoles...)
}

// resolveEnvFor resolves the environment a request targets and, with it, who is
// asking: the user, their role in the project and the scope of the token they
// used. An environment-scoped token cannot reach an environment it was not
// granted, whatever the user's role allows.
func (controller *Controller) resolveEnvFor(ctx context.Context, projectSlug, envSlug string, requiredRoles ...string) (int64, caller, error) {
	identity, ok := authcontext.IdentityFromContext(ctx)
	if !ok {
		return 0, caller{}, errors.Unauthorized("missing auth")
	}
	userID, err := strconv.ParseInt(identity.UserID, 10, 64)
	if err != nil {
		return 0, caller{}, errors.Invalid("invalid user id")
	}

	if !identity.Scope.AllowsEnvironment(envSlug) {
		return 0, caller{}, errors.Forbidden("this API token is not scoped to this environment")
	}

	projectID, err := controller.service.projectRes.ResolveProjectID(ctx, projectSlug, userID, requiredRoles...)
	if err != nil {
		return 0, caller{}, err
	}
	if !identity.Scope.AllowsProject(projectID) {
		return 0, caller{}, errors.Forbidden("this API token is not scoped to this project")
	}
	envID, err := controller.service.envRes.ResolveEnvironmentID(ctx, projectID, envSlug)
	if err != nil {
		return 0, caller{}, err
	}

	return envID, caller{
		userID:   userID,
		role:     controller.service.projectRole(ctx, projectID, userID),
		clientIP: identity.ClientIP,
		scope:    identity.Scope,
	}, nil
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
