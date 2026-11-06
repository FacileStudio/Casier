package auth

import (
	"context"
	"strconv"

	"github.com/FacileStudio/Casier/apps/api/internal/authcontext"
	"github.com/FacileStudio/tronc/errors"
)

// Controller adapts the auth Service to HTTP request handlers.
type Controller struct {
	service *Service
}

func newController(service *Service) *Controller {
	return &Controller{service: service}
}

func (c *Controller) me(ctx context.Context) (*MeResponse, error) {
	identity, ok := authcontext.IdentityFromContext(ctx)
	if !ok {
		return nil, errors.Unauthorized("missing auth")
	}
	return c.service.profile(ctx, identity.UserID)
}

func intToStr(id int64) string {
	return strconv.FormatInt(id, 10)
}
