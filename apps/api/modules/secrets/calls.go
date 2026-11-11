package secrets

import "context"

// Calls are the secret routes' operations for an in-process caller, such as
// the MCP server. Each one runs the route's own controller method for the
// identity on ctx, so roles, token scopes, ACLs, protected environments and
// the audit trail hold exactly as they do over HTTP. Listing never reveals a
// value: Reveal, one audited key at a time, is the only way to one.
type Calls struct {
	controller *Controller
}

// Calls returns the in-process operations of the secret routes.
func (service *Service) Calls() Calls {
	return Calls{controller: service.controller}
}

// List is GET .../secrets: metadata only, narrowed to tag when it is set.
func (c Calls) List(ctx context.Context, projectSlug, envSlug, tag string) (*ListResponse, error) {
	return c.controller.list(ctx, projectSlug, envSlug, tag, false)
}

// Reveal is GET .../secrets/{key}/reveal: one value, audited on its own.
func (c Calls) Reveal(ctx context.Context, projectSlug, envSlug, key string) (*SecretResponse, error) {
	return c.controller.reveal(ctx, projectSlug, envSlug, key)
}

// Versions is GET .../secrets/{key}/versions: who changed a secret and when.
func (c Calls) Versions(ctx context.Context, projectSlug, envSlug, key string) (*VersionsResponse, error) {
	return c.controller.listVersions(ctx, projectSlug, envSlug, key)
}

// Diff is GET .../diff. Its entries carry the values of both sides, so a
// caller that must not see values drops Left and Right.
func (c Calls) Diff(ctx context.Context, projectSlug, from, to string) (*DiffResponse, error) {
	return c.controller.diff(ctx, projectSlug, from, to)
}

// Trash is GET .../secrets/trash.
func (c Calls) Trash(ctx context.Context, projectSlug, envSlug string) (*TrashResponse, error) {
	return c.controller.trash(ctx, projectSlug, envSlug)
}

// Metadata is one secret's metadata through the read path's reach: a secret the
// caller may not read is not_found. It carries no value and writes no audit row,
// so a confirmation can fingerprint the secrets it names as often as it needs.
func (c Calls) Metadata(ctx context.Context, projectSlug, envSlug, key string) (*SecretResponse, error) {
	envID, person, err := c.controller.resolveEnv(ctx, projectSlug, envSlug)
	if err != nil {
		return nil, err
	}
	record, err := c.controller.service.readableSecret(ctx, envID, key, person)
	if err != nil {
		return nil, err
	}
	return &SecretResponse{
		ID: record.ID, Key: record.Key, Version: record.Version, Tags: splitTags(record.Tags),
		CreatedAt: record.CreatedAt, UpdatedAt: record.UpdatedAt,
	}, nil
}
