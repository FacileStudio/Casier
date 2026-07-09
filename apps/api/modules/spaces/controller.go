package spaces

import (
	"context"

	"clef/internal/authcontext"
	"clef/internal/errors"
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

	spaces, err := controller.service.listSpaces(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &ListResponse{Spaces: spaces}, nil
}

func (controller *Controller) create(ctx context.Context, req *CreateRequest) (*SpaceResponse, error) {
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

	return controller.service.createSpace(ctx, userID, req.Name, req.Description)
}

func (controller *Controller) get(ctx context.Context, slug string) (*SpaceResponse, error) {
	identity, ok := authcontext.IdentityFromContext(ctx)
	if !ok {
		return nil, errors.Unauthorized("missing auth")
	}
	userID, err := parseUserID(identity.UserID)
	if err != nil {
		return nil, err
	}

	return controller.service.getSpace(ctx, slug, userID)
}

func (controller *Controller) update(ctx context.Context, slug string, req *UpdateRequest) (*SpaceResponse, error) {
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

	return controller.service.updateSpace(ctx, slug, userID, req.Name, req.Description)
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

	return controller.service.deleteSpace(ctx, slug, userID)
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

	return controller.service.removeMember(ctx, slug, userID, targetUserID)
}

func isValidRole(role string) bool {
	switch role {
	case "admin", "member", "viewer":
		return true
	default:
		return false
	}
}
