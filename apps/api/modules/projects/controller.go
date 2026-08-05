package projects

import (
	"context"

	"casier/internal/authcontext"
	"casier/internal/errors"
	"casier/schemas"
)

type Controller struct {
	service *Service
}

func newController(service *Service) *Controller {
	return &Controller{service: service}
}

func (controller *Controller) list(ctx context.Context) (*ListResponse, error) {
	identity, ok := authcontext.IdentityFromContext(ctx)
	if !ok {
		return nil, errors.Unauthorized("missing auth")
	}
	userID, err := parseUserID(identity.UserID)
	if err != nil {
		return nil, err
	}

	projects, err := controller.service.listProjects(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &ListResponse{Projects: projects}, nil
}

func (controller *Controller) create(ctx context.Context, req *CreateRequest) (*ProjectResponse, error) {
	identity, ok := authcontext.IdentityFromContext(ctx)
	if !ok {
		return nil, errors.Unauthorized("missing auth")
	}
	userID, err := parseUserID(identity.UserID)
	if err != nil {
		return nil, err
	}

	if req.Name == "" {
		return nil, errors.Invalid("name is required")
	}

	return controller.service.createProject(ctx, userID, req.Name, req.Description)
}

func (controller *Controller) get(ctx context.Context, slug string) (*ProjectResponse, error) {
	identity, ok := authcontext.IdentityFromContext(ctx)
	if !ok {
		return nil, errors.Unauthorized("missing auth")
	}
	userID, err := parseUserID(identity.UserID)
	if err != nil {
		return nil, err
	}

	return controller.service.getProject(ctx, slug, userID)
}

func (controller *Controller) update(ctx context.Context, slug string, req *UpdateRequest) (*ProjectResponse, error) {
	identity, ok := authcontext.IdentityFromContext(ctx)
	if !ok {
		return nil, errors.Unauthorized("missing auth")
	}
	userID, err := parseUserID(identity.UserID)
	if err != nil {
		return nil, err
	}

	if req.Name == "" {
		return nil, errors.Invalid("name is required")
	}
	if req.RotationDays != nil && (*req.RotationDays < 0 || *req.RotationDays > schemas.MaxRotationDays) {
		return nil, errors.Invalid("rotation_days must be between 0 and 3650")
	}

	return controller.service.updateProject(ctx, slug, userID, req.Name, req.Description, req.RotationDays)
}

func (controller *Controller) delete(ctx context.Context, slug string) error {
	identity, ok := authcontext.IdentityFromContext(ctx)
	if !ok {
		return errors.Unauthorized("missing auth")
	}
	userID, err := parseUserID(identity.UserID)
	if err != nil {
		return err
	}

	return controller.service.deleteProject(ctx, slug, userID)
}

func (controller *Controller) listMembers(ctx context.Context, slug string) (*MembersResponse, error) {
	identity, ok := authcontext.IdentityFromContext(ctx)
	if !ok {
		return nil, errors.Unauthorized("missing auth")
	}
	userID, err := parseUserID(identity.UserID)
	if err != nil {
		return nil, err
	}

	members, err := controller.service.listMembers(ctx, slug, userID)
	if err != nil {
		return nil, err
	}
	return &MembersResponse{Members: members}, nil
}

func (controller *Controller) addMember(ctx context.Context, slug string, req *AddMemberRequest) error {
	identity, ok := authcontext.IdentityFromContext(ctx)
	if !ok {
		return errors.Unauthorized("missing auth")
	}
	userID, err := parseUserID(identity.UserID)
	if err != nil {
		return err
	}

	if req.Email == "" {
		return errors.Invalid("email is required")
	}
	if !isValidRole(req.Role) {
		return errors.Invalid("role must be one of: admin, member, viewer")
	}

	return controller.service.addMember(ctx, slug, userID, req.Email, req.Role)
}

func (controller *Controller) removeMember(ctx context.Context, slug string, targetUserID int64) error {
	identity, ok := authcontext.IdentityFromContext(ctx)
	if !ok {
		return errors.Unauthorized("missing auth")
	}
	userID, err := parseUserID(identity.UserID)
	if err != nil {
		return err
	}

	return controller.service.removeMember(ctx, slug, userID, targetUserID, identity.Email, identity.ClientIP)
}

func (controller *Controller) createInvite(ctx context.Context, slug string, req *CreateInviteRequest) (*InviteResponse, error) {
	identity, ok := authcontext.IdentityFromContext(ctx)
	if !ok {
		return nil, errors.Unauthorized("missing auth")
	}
	userID, err := parseUserID(identity.UserID)
	if err != nil {
		return nil, err
	}
	if !isValidRole(req.Role) {
		return nil, errors.Invalid("role must be one of: admin, member, viewer")
	}
	return controller.service.createInvite(ctx, slug, userID, req.Role, req.MaxUses)
}

func (controller *Controller) listInvites(ctx context.Context, slug string) (*InvitesResponse, error) {
	identity, ok := authcontext.IdentityFromContext(ctx)
	if !ok {
		return nil, errors.Unauthorized("missing auth")
	}
	userID, err := parseUserID(identity.UserID)
	if err != nil {
		return nil, err
	}
	invites, err := controller.service.listInvites(ctx, slug, userID)
	if err != nil {
		return nil, err
	}
	return &InvitesResponse{Invites: invites}, nil
}

func (controller *Controller) acceptInvite(ctx context.Context, token string) error {
	identity, ok := authcontext.IdentityFromContext(ctx)
	if !ok {
		return errors.Unauthorized("missing auth")
	}
	userID, err := parseUserID(identity.UserID)
	if err != nil {
		return err
	}
	return controller.service.acceptInvite(ctx, token, userID)
}

func (controller *Controller) deleteInvite(ctx context.Context, slug string, inviteID int64) error {
	identity, ok := authcontext.IdentityFromContext(ctx)
	if !ok {
		return errors.Unauthorized("missing auth")
	}
	userID, err := parseUserID(identity.UserID)
	if err != nil {
		return err
	}
	return controller.service.deleteInvite(ctx, slug, userID, inviteID)
}

func (controller *Controller) listAuditLogs(ctx context.Context, slug, action string, limit int) (*AuditLogsResponse, error) {
	identity, ok := authcontext.IdentityFromContext(ctx)
	if !ok {
		return nil, errors.Unauthorized("missing auth")
	}
	userID, err := parseUserID(identity.UserID)
	if err != nil {
		return nil, err
	}
	logs, err := controller.service.listAuditLogs(ctx, slug, userID, action, limit)
	if err != nil {
		return nil, err
	}
	return &AuditLogsResponse{Logs: logs}, nil
}

func (controller *Controller) createWebhook(ctx context.Context, slug string, req *CreateWebhookRequest) (*WebhookResponse, error) {
	identity, ok := authcontext.IdentityFromContext(ctx)
	if !ok {
		return nil, errors.Unauthorized("missing auth")
	}
	userID, err := parseUserID(identity.UserID)
	if err != nil {
		return nil, err
	}
	if req.URL == "" {
		return nil, errors.Invalid("url is required")
	}
	return controller.service.createWebhook(ctx, slug, userID, req.URL, req.Events, req.Secret)
}

func (controller *Controller) listWebhooks(ctx context.Context, slug string) (*WebhooksResponse, error) {
	identity, ok := authcontext.IdentityFromContext(ctx)
	if !ok {
		return nil, errors.Unauthorized("missing auth")
	}
	userID, err := parseUserID(identity.UserID)
	if err != nil {
		return nil, err
	}
	webhooks, err := controller.service.listWebhooks(ctx, slug, userID)
	if err != nil {
		return nil, err
	}
	return &WebhooksResponse{Webhooks: webhooks}, nil
}

func (controller *Controller) deleteWebhook(ctx context.Context, slug string, webhookID int64) error {
	identity, ok := authcontext.IdentityFromContext(ctx)
	if !ok {
		return errors.Unauthorized("missing auth")
	}
	userID, err := parseUserID(identity.UserID)
	if err != nil {
		return err
	}
	return controller.service.deleteWebhook(ctx, slug, userID, webhookID)
}

func (controller *Controller) createApiToken(ctx context.Context, slug string, req *CreateApiTokenRequest) (*ApiTokenResponse, error) {
	identity, ok := authcontext.IdentityFromContext(ctx)
	if !ok {
		return nil, errors.Unauthorized("missing auth")
	}
	userID, err := parseUserID(identity.UserID)
	if err != nil {
		return nil, err
	}
	if req.Name == "" {
		return nil, errors.Invalid("name is required")
	}
	return controller.service.createApiToken(ctx, slug, userID, req.Name, req.Environments, req.Permissions)
}

func (controller *Controller) listApiTokens(ctx context.Context, slug string) (*ApiTokensResponse, error) {
	identity, ok := authcontext.IdentityFromContext(ctx)
	if !ok {
		return nil, errors.Unauthorized("missing auth")
	}
	userID, err := parseUserID(identity.UserID)
	if err != nil {
		return nil, err
	}
	tokens, err := controller.service.listApiTokens(ctx, slug, userID)
	if err != nil {
		return nil, err
	}
	return &ApiTokensResponse{Tokens: tokens}, nil
}

func (controller *Controller) revokeApiToken(ctx context.Context, slug string, tokenID int64) error {
	identity, ok := authcontext.IdentityFromContext(ctx)
	if !ok {
		return errors.Unauthorized("missing auth")
	}
	userID, err := parseUserID(identity.UserID)
	if err != nil {
		return err
	}
	return controller.service.revokeApiToken(ctx, slug, userID, tokenID)
}

func isValidRole(role string) bool {
	switch role {
	case "admin", "member", "viewer":
		return true
	default:
		return false
	}
}
