package secrets

import (
	"context"
	"strconv"
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

func (controller *Controller) list(ctx context.Context, projectSlug, envSlug, tagFilter string) (*ListResponse, error) {
	identity, ok := authcontext.IdentityFromContext(ctx)
	if !ok {
		return nil, errors.Unauthorized("missing auth")
	}
	envID, err := controller.resolveEnv(ctx, projectSlug, envSlug)
	if err != nil {
		return nil, err
	}

	secrets, err := controller.service.listSecrets(ctx, envID, identity.UserID, identity.ClientIP, tagFilter)
	if err != nil {
		return nil, err
	}
	return &ListResponse{Secrets: secrets}, nil
}

func (controller *Controller) create(ctx context.Context, projectSlug, envSlug string, req *CreateRequest) (*SecretResponse, error) {
	identity, ok := authcontext.IdentityFromContext(ctx)
	if !ok {
		return nil, errors.Unauthorized("missing auth")
	}
	envID, err := controller.resolveEnvWithWrite(ctx, projectSlug, envSlug)
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

	return controller.service.createSecret(ctx, envID, key, req.Value, identity.UserID, identity.ClientIP, req.Tags)
}

func (controller *Controller) update(ctx context.Context, projectSlug, envSlug, key string, req *UpdateRequest) (*SecretResponse, error) {
	identity, ok := authcontext.IdentityFromContext(ctx)
	if !ok {
		return nil, errors.Unauthorized("missing auth")
	}
	envID, err := controller.resolveEnvWithWrite(ctx, projectSlug, envSlug)
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

func (controller *Controller) delete(ctx context.Context, projectSlug, envSlug, key string) error {
	identity, ok := authcontext.IdentityFromContext(ctx)
	if !ok {
		return errors.Unauthorized("missing auth")
	}
	envID, err := controller.resolveEnvWithWrite(ctx, projectSlug, envSlug)
	if err != nil {
		return err
	}

	return controller.service.deleteSecret(ctx, envID, key, identity.UserID, identity.ClientIP)
}

func (controller *Controller) importEnv(ctx context.Context, projectSlug, envSlug string, req *ImportRequest) (*ImportResponse, error) {
	identity, ok := authcontext.IdentityFromContext(ctx)
	if !ok {
		return nil, errors.Unauthorized("missing auth")
	}
	envID, err := controller.resolveEnvWithWrite(ctx, projectSlug, envSlug)
	if err != nil {
		return nil, err
	}
	return controller.service.importEnv(ctx, envID, req.Content, identity.UserID, identity.ClientIP)
}

func (controller *Controller) exportEnv(ctx context.Context, projectSlug, envSlug string) (*ExportResponse, error) {
	identity, ok := authcontext.IdentityFromContext(ctx)
	if !ok {
		return nil, errors.Unauthorized("missing auth")
	}
	envID, err := controller.resolveEnv(ctx, projectSlug, envSlug)
	if err != nil {
		return nil, err
	}
	content, err := controller.service.exportEnv(ctx, envID, identity.UserID, identity.ClientIP)
	if err != nil {
		return nil, err
	}
	return &ExportResponse{Content: content}, nil
}

func (controller *Controller) updateTags(ctx context.Context, projectSlug, envSlug, key string, req *UpdateTagsRequest) error {
	envID, err := controller.resolveEnvWithWrite(ctx, projectSlug, envSlug)
	if err != nil {
		return err
	}
	return controller.service.updateTags(ctx, envID, key, req.Tags)
}

func (controller *Controller) listVersions(ctx context.Context, projectSlug, envSlug, key string) (*VersionsResponse, error) {
	envID, err := controller.resolveEnv(ctx, projectSlug, envSlug)
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
	identity, ok := authcontext.IdentityFromContext(ctx)
	if !ok {
		return nil, errors.Unauthorized("missing auth")
	}
	envID, err := controller.resolveEnvWithWrite(ctx, projectSlug, envSlug)
	if err != nil {
		return nil, err
	}
	if req.Version < 1 {
		return nil, errors.Invalid("version must be >= 1")
	}
	return controller.service.rollbackSecret(ctx, envID, key, req.Version, identity.UserID, identity.ClientIP)
}

func (controller *Controller) diff(ctx context.Context, projectSlug, envSlugA, envSlugB string) (*DiffResponse, error) {
	identity, ok := authcontext.IdentityFromContext(ctx)
	if !ok {
		return nil, errors.Unauthorized("missing auth")
	}
	userID, err := extractUserID(ctx)
	if err != nil {
		return nil, err
	}

	projectID, err := controller.service.projectRes.ResolveProjectID(ctx, projectSlug, userID)
	if err != nil {
		return nil, err
	}

	envA, err := controller.service.envRes.ResolveEnvironmentID(ctx, projectID, envSlugA)
	if err != nil {
		return nil, err
	}
	envB, err := controller.service.envRes.ResolveEnvironmentID(ctx, projectID, envSlugB)
	if err != nil {
		return nil, err
	}

	resp, err := controller.service.diffEnvironments(ctx, envA, envB, identity.UserID, identity.ClientIP)
	if err != nil {
		return nil, err
	}
	resp.From = envSlugA
	resp.To = envSlugB
	return resp, nil
}

func (controller *Controller) listACLs(ctx context.Context, projectSlug, envSlug, key string) (*ACLsResponse, error) {
	envID, err := controller.resolveEnv(ctx, projectSlug, envSlug)
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
	envID, err := controller.resolveEnvWithWrite(ctx, projectSlug, envSlug)
	if err != nil {
		return err
	}
	if req.Email == "" {
		return errors.Invalid("email is required")
	}
	return controller.service.setACL(ctx, envID, key, req.Email, req.CanRead, req.CanWrite)
}

func (controller *Controller) deleteACL(ctx context.Context, projectSlug, envSlug, key string, aclID int64) error {
	envID, err := controller.resolveEnvWithWrite(ctx, projectSlug, envSlug)
	if err != nil {
		return err
	}
	return controller.service.deleteACL(ctx, envID, key, aclID)
}

func (controller *Controller) resolveEnv(ctx context.Context, projectSlug, envSlug string) (int64, error) {
	userID, err := extractUserID(ctx)
	if err != nil {
		return 0, err
	}

	projectID, err := controller.service.projectRes.ResolveProjectID(ctx, projectSlug, userID)
	if err != nil {
		return 0, err
	}

	return controller.service.envRes.ResolveEnvironmentID(ctx, projectID, envSlug)
}

func (controller *Controller) resolveEnvWithWrite(ctx context.Context, projectSlug, envSlug string) (int64, error) {
	userID, err := extractUserID(ctx)
	if err != nil {
		return 0, err
	}

	projectID, err := controller.service.projectRes.ResolveProjectID(ctx, projectSlug, userID, "owner", "admin", "member")
	if err != nil {
		return 0, err
	}

	return controller.service.envRes.ResolveEnvironmentID(ctx, projectID, envSlug)
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
