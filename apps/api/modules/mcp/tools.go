package mcp

import (
	"context"
	stderrors "errors"
	"strconv"

	"github.com/FacileStudio/Casier/apps/api/internal/authcontext"
	"github.com/FacileStudio/porte/mcpauth"
	mcpkit "github.com/FacileStudio/tronc/mcp"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	toolProjects    = "casier_list_projects"
	toolSecrets     = "casier_list_secrets"
	toolSecret      = "casier_get_secret"
	toolDiff        = "casier_diff_environments"
	toolTrash       = "casier_list_trash"
	toolAudit       = "casier_list_audit_events"
	toolNewProject  = "casier_create_project"
	toolNewEnv      = "casier_create_environment"
	toolSet         = "casier_set_secret"
	toolTag         = "casier_tag_secret"
	toolRollback    = "casier_rollback_secret"
	toolRestore     = "casier_restore_secret"
	toolReveal      = "casier_reveal_secrets"
	toolDelete      = "casier_delete_secret"
	hintProject     = "Call casier_list_projects for the project and environment slugs you can use."
	hintSecret      = "Call casier_list_secrets for the keys of that environment."
	hintTrash       = "Call casier_list_trash for the keys that can be restored."
	clientIPPrefix  = "mcp:"
	statusCreated   = "created"
	statusUpdated   = "updated"
	statusTagged    = "tagged"
	statusRestored  = "restored"
	statusRolled    = "rolled_back"
	statusRevealed  = "revealed"
	statusTrashed   = "trashed"
	statusApproved  = "approved"
	maxRevealedKeys = 50
)

var errNoCaller = stderrors.New("this request carries no valid access token. Reconnect Casier")

// toolset holds what the tool handlers share.
type toolset struct {
	Deps
	confirm *mcpkit.Confirmations
}

// caller is the person a call acts for: the context the REST controllers read, and the
// token's caller for the confirmations.
type caller struct {
	ctx context.Context
	who mcpkit.Caller
}

// callerOf resolves the person behind a call's verified token, refused the way
// RequireAuth refuses a session whose account is gone.
func (t *toolset) callerOf(ctx context.Context, req *sdk.CallToolRequest) (caller, error) {
	who, err := mcpkit.CallerOf(req)
	if err != nil {
		return caller{}, errNoCaller
	}
	if _, ok := mcpauth.Identity(req.Extra.TokenInfo); !ok {
		return caller{}, errNoCaller
	}
	personCtx, err := t.as(ctx, who.UserID, who.ClientID)
	if err != nil {
		return caller{}, err
	}
	return caller{ctx: personCtx, who: who}, nil
}

// as is ctx carrying the person userID as a user session, the identity the REST
// controllers authorize: no token scope, so the person's project roles and the secret
// ACLs decide. The audit trail records the MCP client in place of an address.
func (t *toolset) as(ctx context.Context, userID, clientID string) (context.Context, error) {
	id, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return nil, errNoCaller
	}
	account, err := t.Accounts.Identity(ctx, id)
	if err != nil {
		return nil, errNoCaller
	}
	return authcontext.WithIdentity(ctx, authcontext.Identity{
		UserID: userID, Email: account.Email, ClientIP: clientIPPrefix + clientID,
	}), nil
}
