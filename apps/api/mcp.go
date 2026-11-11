package main

import (
	"context"
	"log/slog"

	"github.com/FacileStudio/Casier/apps/api/internal/env"
	"github.com/FacileStudio/Casier/apps/api/modules/mcp"
	"github.com/FacileStudio/porte/mcpauth"
	"github.com/FacileStudio/porte/oidc"
	portepg "github.com/FacileStudio/porte/pg"
	mcpkit "github.com/FacileStudio/tronc/mcp"
)

// newMCP builds the MCP server when MCP_RESOURCE is set and SSO is on, and nothing
// otherwise. deps.Actions is the database the pending actions live in. The verifier
// runs Registre discovery here, so a wrong issuer fails at boot like the OIDC kit does.
// A person's first MCP call creates their account through the OIDC kit, Casier's own
// sign-in write path, as a first web login would.
func newMCP(ctx context.Context, appEnv env.Config, kit *oidc.Kit, deps mcp.Deps, appLogger *slog.Logger) (*mcp.Service, error) {
	if appEnv.OIDC == nil || appEnv.OIDC.MCPResource == "" || kit == nil {
		return nil, nil
	}
	verifier, err := mcpauth.Verifier(ctx, mcpauth.Config{
		Issuer: appEnv.OIDC.Issuer, Resource: appEnv.OIDC.MCPResource, ScopePrefix: "casier",
		Identities: portepg.New(deps.Actions).Identities(), Provision: true, Accounts: kit, Logger: appLogger,
	})
	if err != nil {
		return nil, err
	}
	service, err := mcp.NewService(deps, mcp.Config{
		Resource: appEnv.OIDC.MCPResource, Issuer: appEnv.OIDC.Issuer, Version: version,
		Origins: appEnv.CORSAllowedOrigins, StateKey: appEnv.OIDC.MCPStateKey(), Verifier: verifier,
		Pending: mcpkit.PostgresPending{DB: deps.Actions}, Logger: appLogger,
	})
	if err != nil {
		return nil, err
	}
	appLogger.Info("mcp server on", slog.String("resource", appEnv.OIDC.MCPResource))
	return service, nil
}
