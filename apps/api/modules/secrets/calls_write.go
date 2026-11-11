package secrets

import "context"

// Create is POST .../secrets.
func (c Calls) Create(ctx context.Context, projectSlug, envSlug string, req *CreateRequest) (*SecretResponse, error) {
	return c.controller.create(ctx, projectSlug, envSlug, req)
}

// Update is PUT .../secrets/{key}: a new version of the value.
func (c Calls) Update(ctx context.Context, projectSlug, envSlug, key string, req *UpdateRequest) (*SecretResponse, error) {
	return c.controller.update(ctx, projectSlug, envSlug, key, req)
}

// UpdateTags is PUT .../secrets/{key}/tags.
func (c Calls) UpdateTags(ctx context.Context, projectSlug, envSlug, key string, req *UpdateTagsRequest) error {
	return c.controller.updateTags(ctx, projectSlug, envSlug, key, req)
}

// Rollback is POST .../secrets/{key}/rollback.
func (c Calls) Rollback(ctx context.Context, projectSlug, envSlug, key string, req *RollbackRequest) (*SecretResponse, error) {
	return c.controller.rollback(ctx, projectSlug, envSlug, key, req)
}

// Restore is POST .../secrets/{key}/restore.
func (c Calls) Restore(ctx context.Context, projectSlug, envSlug, key string) error {
	return c.controller.restore(ctx, projectSlug, envSlug, key)
}

// Delete is DELETE .../secrets/{key} without purge: it stays
// restorable for TrashRetentionDays.
func (c Calls) Delete(ctx context.Context, projectSlug, envSlug, key string) error {
	return c.controller.delete(ctx, projectSlug, envSlug, key, false)
}
