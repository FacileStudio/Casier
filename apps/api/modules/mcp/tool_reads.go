package mcp

import (
	"context"
	"slices"
	"strings"

	"github.com/FacileStudio/Casier/apps/api/modules/secrets"
	mcpkit "github.com/FacileStudio/tronc/mcp"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// secretsPerPage caps one page of casier_list_secrets before the size budget applies.
const secretsPerPage = 200

// page is the position casier_list_secrets's cursor encodes.
type page struct {
	Offset int `json:"o"`
}

// listProjects is casier_list_projects: the caller's projects, each with its
// environments, as the project list and its sidebar show them.
func (t *toolset) listProjects(ctx context.Context, req *sdk.CallToolRequest, in projectsInput) (*sdk.CallToolResult, projectsOutput, error) {
	c, err := t.callerOf(ctx, req)
	if err != nil {
		return nil, projectsOutput{}, err
	}
	list, err := t.Projects.Calls().List(c.ctx)
	if err != nil {
		return nil, projectsOutput{}, mcpkit.Failure(err, "")
	}
	query := strings.ToLower(strings.TrimSpace(in.Query))
	out := projectsOutput{Projects: []projectOut{}}
	for _, project := range list.Projects {
		if query != "" && !strings.Contains(strings.ToLower(project.Name+" "+project.Slug), query) {
			continue
		}
		envs, err := t.Environments.Calls().List(c.ctx, project.Slug)
		if err != nil {
			return nil, projectsOutput{}, mcpkit.Failure(err, hintProject)
		}
		out.Projects = append(out.Projects, projectOf(project, envs.Environments))
	}
	return nil, out, nil
}

// metadata lists an environment's secrets the caller may see, without values.
func (t *toolset) metadata(c caller, ref envRef, tag string) (*secrets.ListResponse, error) {
	list, err := t.Secrets.Calls().List(c.ctx, ref.Project, ref.Environment, tag)
	if err != nil {
		return nil, mcpkit.Failure(err, hintProject)
	}
	return list, nil
}

// listSecrets is casier_list_secrets: one environment's keys and their metadata.
func (t *toolset) listSecrets(ctx context.Context, req *sdk.CallToolRequest, in secretsInput) (*sdk.CallToolResult, secretsOutput, error) {
	detailed, err := mcpkit.Detailed(in.ResponseFormat)
	if err != nil {
		return nil, secretsOutput{}, err
	}
	var at page
	if err := mcpkit.DecodeCursor(in.Cursor, &at); err != nil {
		return nil, secretsOutput{}, err
	}
	c, err := t.callerOf(ctx, req)
	if err != nil {
		return nil, secretsOutput{}, err
	}
	list, err := t.metadata(c, in.envRef, in.Tag)
	if err != nil {
		return nil, secretsOutput{}, err
	}
	query := strings.ToLower(strings.TrimSpace(in.Query))
	kept := []secretOut{}
	for _, secret := range list.Secrets {
		if query == "" || strings.Contains(strings.ToLower(secret.Key), query) {
			kept = append(kept, secretOf(secret, detailed))
		}
	}
	out := secretsOutput{Project: in.Project, Environment: in.Environment, RotationDays: list.RotationDays, Total: len(kept)}
	return nil, pageOf(out, kept, at.Offset), nil
}

// pageOf fills out with the page of kept starting at offset, as many as fit the size
// budget, and the cursor of the next page.
func pageOf(out secretsOutput, kept []secretOut, offset int) secretsOutput {
	rest := kept[min(max(offset, 0), len(kept)):]
	rest = rest[:min(len(rest), secretsPerPage)]
	n := mcpkit.FitCount(len(rest), func(n int) int {
		out.Secrets = rest[:n]
		return mcpkit.EncodedSize(out)
	})
	out.Secrets = rest[:n]
	if next := offset + n; n > 0 && next < len(kept) {
		out.NextCursor = mcpkit.EncodeCursor(page{Offset: next})
	}
	return out
}

// getSecret is casier_get_secret: one secret's metadata and its version history, no
// values.
func (t *toolset) getSecret(ctx context.Context, req *sdk.CallToolRequest, in secretRef) (*sdk.CallToolResult, secretOutput, error) {
	c, err := t.callerOf(ctx, req)
	if err != nil {
		return nil, secretOutput{}, err
	}
	list, err := t.metadata(c, in.envRef, "")
	if err != nil {
		return nil, secretOutput{}, err
	}
	index := slices.IndexFunc(list.Secrets, func(s secrets.SecretResponse) bool { return s.Key == in.Key })
	if index < 0 {
		return nil, secretOutput{}, mcpkit.Failure(errSecretNotFound, hintSecret)
	}
	versions, err := t.Secrets.Calls().Versions(c.ctx, in.Project, in.Environment, in.Key)
	if err != nil {
		return nil, secretOutput{}, mcpkit.Failure(err, hintSecret)
	}
	out := secretOutput{Secret: secretOf(list.Secrets[index], true), Versions: make([]versionOut, len(versions.Versions))}
	for i, v := range versions.Versions {
		out.Versions[i] = versionOut{Version: v.Version, ChangedBy: nameOf(v.Actor, v.ChangedBy), CreatedAt: stamp(v.CreatedAt)}
	}
	return nil, out, nil
}

// diffEnvironments is casier_diff_environments: which keys are missing, equal or
// different across two environments. The REST diff carries both values; they are
// dropped here.
func (t *toolset) diffEnvironments(ctx context.Context, req *sdk.CallToolRequest, in diffInput) (*sdk.CallToolResult, diffOutput, error) {
	c, err := t.callerOf(ctx, req)
	if err != nil {
		return nil, diffOutput{}, err
	}
	diff, err := t.Secrets.Calls().Diff(c.ctx, in.Project, in.From, in.To)
	if err != nil {
		return nil, diffOutput{}, mcpkit.Failure(err, hintProject)
	}
	out := diffOutput{From: diff.From, To: diff.To, Entries: make([]diffOut, len(diff.Entries))}
	for i, entry := range diff.Entries {
		out.Entries[i] = diffOut{Key: entry.Key, Status: entry.Status}
	}
	return nil, out, nil
}

// listTrash is casier_list_trash: the secrets deleted from an environment that can
// still be restored.
func (t *toolset) listTrash(ctx context.Context, req *sdk.CallToolRequest, in envRef) (*sdk.CallToolResult, trashOutput, error) {
	c, err := t.callerOf(ctx, req)
	if err != nil {
		return nil, trashOutput{}, err
	}
	trash, err := t.Secrets.Calls().Trash(c.ctx, in.Project, in.Environment)
	if err != nil {
		return nil, trashOutput{}, mcpkit.Failure(err, hintProject)
	}
	out := trashOutput{RetentionDays: trash.RetentionDays, Secrets: make([]trashOut, len(trash.Secrets))}
	for i, s := range trash.Secrets {
		out.Secrets[i] = trashOut{
			Key: s.Key, Tags: s.Tags, DeletedAt: stamp(s.DeletedAt), DeletedBy: nameOf(s.Actor, s.DeletedBy), PurgeAt: stamp(s.PurgeAt),
		}
	}
	return nil, out, nil
}

// listAudit is casier_list_audit_events: a project's audit trail, newest first.
func (t *toolset) listAudit(ctx context.Context, req *sdk.CallToolRequest, in auditInput) (*sdk.CallToolResult, auditOutput, error) {
	c, err := t.callerOf(ctx, req)
	if err != nil {
		return nil, auditOutput{}, err
	}
	logs, err := t.Projects.Calls().AuditLogs(c.ctx, in.Project, strings.TrimSpace(in.Action), in.Limit)
	if err != nil {
		return nil, auditOutput{}, mcpkit.Failure(err, hintProject)
	}
	out := auditOutput{Events: make([]auditOut, len(logs.Logs))}
	for i, e := range logs.Logs {
		out.Events[i] = auditOut{
			Action: e.Action, Resource: e.Resource, Detail: e.Detail, Person: nameOf(e.Actor, e.UserEmail), CreatedAt: stamp(e.CreatedAt),
		}
	}
	return nil, out, nil
}
