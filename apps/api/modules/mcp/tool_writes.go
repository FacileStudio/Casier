package mcp

import (
	"context"
	"slices"

	"github.com/FacileStudio/Casier/apps/api/modules/environments"
	"github.com/FacileStudio/Casier/apps/api/modules/projects"
	"github.com/FacileStudio/Casier/apps/api/modules/secrets"
	"github.com/FacileStudio/tronc/errors"
	mcpkit "github.com/FacileStudio/tronc/mcp"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

var errSecretNotFound = errors.NotFound("secret not found")

// createProject is casier_create_project; the caller becomes its owner.
func (t *toolset) createProject(ctx context.Context, req *sdk.CallToolRequest, in createProjectInput) (*sdk.CallToolResult, writeOutput, error) {
	c, err := t.callerOf(ctx, req)
	if err != nil {
		return nil, writeOutput{}, err
	}
	project, err := t.Projects.Calls().Create(c.ctx, &projects.CreateRequest{Name: in.Name, Description: in.Description})
	if err != nil {
		return nil, writeOutput{}, mcpkit.Failure(err, "")
	}
	envs, err := t.Environments.Calls().List(c.ctx, project.Slug)
	if err != nil {
		return nil, writeOutput{}, mcpkit.Failure(err, hintProject)
	}
	out := projectOf(*project, envs.Environments)
	return nil, writeOutput{Status: statusCreated, Project: &out}, nil
}

// createEnvironment is casier_create_environment.
func (t *toolset) createEnvironment(ctx context.Context, req *sdk.CallToolRequest, in createEnvironmentInput) (*sdk.CallToolResult, writeOutput, error) {
	c, err := t.callerOf(ctx, req)
	if err != nil {
		return nil, writeOutput{}, err
	}
	env, err := t.Environments.Calls().Create(c.ctx, in.Project, &environments.CreateRequest{Name: in.Name, Protected: in.Protected})
	if err != nil {
		return nil, writeOutput{}, mcpkit.Failure(err, hintProject)
	}
	out := envOf(*env)
	return nil, writeOutput{Status: statusCreated, Environment: &out}, nil
}

// setSecret is casier_set_secret: a new version of an existing secret, or a new
// secret. The answer never echoes the value.
func (t *toolset) setSecret(ctx context.Context, req *sdk.CallToolRequest, in setInput) (*sdk.CallToolResult, writeOutput, error) {
	c, err := t.callerOf(ctx, req)
	if err != nil {
		return nil, writeOutput{}, err
	}
	calls, status := t.Secrets.Calls(), statusUpdated
	secret, err := calls.Update(c.ctx, in.Project, in.Environment, in.Key, &secrets.UpdateRequest{Value: in.Value})
	if isNotFound(err) {
		status = statusCreated
		secret, err = calls.Create(c.ctx, in.Project, in.Environment, &secrets.CreateRequest{Key: in.Key, Value: in.Value, Tags: in.Tags})
	}
	if err != nil {
		return nil, writeOutput{}, mcpkit.Failure(err, hintProject)
	}
	out := secretOf(*secret, false)
	return nil, writeOutput{Status: status, Secret: &out}, nil
}

// tagSecret is casier_tag_secret: replaces a secret's tags, value untouched.
func (t *toolset) tagSecret(ctx context.Context, req *sdk.CallToolRequest, in tagInput) (*sdk.CallToolResult, writeOutput, error) {
	c, err := t.callerOf(ctx, req)
	if err != nil {
		return nil, writeOutput{}, err
	}
	if err := t.Secrets.Calls().UpdateTags(c.ctx, in.Project, in.Environment, in.Key, &secrets.UpdateTagsRequest{Tags: in.Tags}); err != nil {
		return nil, writeOutput{}, mcpkit.Failure(err, hintSecret)
	}
	return t.written(c, in.secretRef, statusTagged)
}

// rollbackSecret is casier_rollback_secret: an earlier version's value becomes the
// newest version, so the rollback itself can be rolled back.
func (t *toolset) rollbackSecret(ctx context.Context, req *sdk.CallToolRequest, in rollbackInput) (*sdk.CallToolResult, writeOutput, error) {
	c, err := t.callerOf(ctx, req)
	if err != nil {
		return nil, writeOutput{}, err
	}
	secret, err := t.Secrets.Calls().Rollback(c.ctx, in.Project, in.Environment, in.Key, &secrets.RollbackRequest{Version: in.Version})
	if err != nil {
		return nil, writeOutput{}, mcpkit.Failure(err, "Call casier_get_secret for the versions of that secret.")
	}
	out := secretOf(*secret, false)
	return nil, writeOutput{Status: statusRolled, Secret: &out}, nil
}

// restoreSecret is casier_restore_secret: a trashed secret comes back with its history.
func (t *toolset) restoreSecret(ctx context.Context, req *sdk.CallToolRequest, in secretRef) (*sdk.CallToolResult, writeOutput, error) {
	c, err := t.callerOf(ctx, req)
	if err != nil {
		return nil, writeOutput{}, err
	}
	if err := t.Secrets.Calls().Restore(c.ctx, in.Project, in.Environment, in.Key); err != nil {
		return nil, writeOutput{}, mcpkit.Failure(err, hintTrash)
	}
	return t.written(c, in, statusRestored)
}

// written answers a write with the secret's metadata as it is now, when the secret
// list still shows it. The write itself succeeded either way.
func (t *toolset) written(c caller, ref secretRef, status string) (*sdk.CallToolResult, writeOutput, error) {
	out := writeOutput{Status: status}
	if list, err := t.Secrets.Calls().List(c.ctx, ref.Project, ref.Environment, ""); err == nil {
		if index := slices.IndexFunc(list.Secrets, func(s secrets.SecretResponse) bool { return s.Key == ref.Key }); index >= 0 {
			secret := secretOf(list.Secrets[index], false)
			out.Secret = &secret
		}
	}
	return nil, out, nil
}
