package auth

import (
	"context"
	"net/http"
	"strconv"

	"github.com/FacileStudio/Casier/apps/api/internal/authcontext"
	"github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"
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

func (c *Controller) handleMe(w http.ResponseWriter, r *http.Request) {
	resp, err := c.me(r.Context())
	if err != nil {
		httpjson.WriteError(w, err)
		return
	}
	httpjson.WriteJSON(w, http.StatusOK, resp)
}

func intToStr(id int64) string {
	return strconv.FormatInt(id, 10)
}
