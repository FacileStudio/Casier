package mcp

import (
	"context"
	"net/http"
	"strconv"

	"github.com/FacileStudio/porte"
	"github.com/FacileStudio/tronc/errors"
	mcpkit "github.com/FacileStudio/tronc/mcp"
)

const webActor = "casier-web"

// pendingUser is the signed-in person of a confirm page request. Only the person's own
// session passes: a delegated token must never confirm what an agent asked for.
func pendingUser(r *http.Request) (string, error) {
	authenticated, ok := porte.From(r.Context())
	if !ok || authenticated.UserID == 0 {
		return "", errors.Unauthorized("missing auth")
	}
	if authenticated.SessionID == 0 || authenticated.Actor != "" {
		return "", mcpkit.ErrDelegated
	}
	return strconv.FormatInt(authenticated.UserID, 10), nil
}

// describe adds the secrets an action names, as they are now, to a confirm page item.
func (t *toolset) describe(ctx context.Context, item mcpkit.PendingItem, a mcpkit.Action) PendingItem {
	out := PendingItem{PendingItem: item}
	g, err := targetOf(a)
	if err != nil {
		return out
	}
	out.Project, out.Environment = g.Project, g.Environment
	found, err := t.lookup(ctx, a, g)
	if err != nil {
		return out
	}
	for _, secret := range found {
		out.Secrets = append(out.Secrets, PendingSecret{Key: secret.Key, Version: secret.Version, UpdatedAt: stamp(secret.UpdatedAt)})
	}
	return out
}

// done is the confirm page's answer once an action ran: the secret is in the trash, or
// the reveal is approved for the client that asked.
func done(_ context.Context, o mcpkit.Outcome) mcpkit.PendingResult {
	if o.Action.Kind == kindReveal {
		return mcpkit.PendingResult{Status: statusApproved}
	}
	return mcpkit.PendingResult{Status: statusTrashed}
}
