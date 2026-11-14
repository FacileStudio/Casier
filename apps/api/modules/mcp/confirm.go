package mcp

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/FacileStudio/Casier/apps/api/modules/secrets"
	"github.com/FacileStudio/tronc/errors"
	mcpkit "github.com/FacileStudio/tronc/mcp"
)

// The sensitive actions. A delete moves a secret to the trash. A reveal hands
// plaintext values to the MCP client, so it is confirmed like a delete even though
// it changes nothing.
const (
	kindDelete = "delete_secret"
	kindReveal = "reveal_secrets"
)

var errTarget = errors.NotFound("this action names no secret")

// target is what a pending action names: keys of one environment of one project. A
// reveal also names the client that asked, so another client asking for the same keys
// gets its own action instead of one only the first client can collect.
type target struct {
	Project     string   `json:"p"`
	Environment string   `json:"e"`
	Keys        []string `json:"k"`
	Client      string   `json:"c,omitempty"`
}

// kinds are the sensitive actions. The fingerprint covers the secrets as they were
// when the person was asked, so a value written in between closes the action instead
// of revealing or deleting a version the person did not see.
func (t *toolset) kinds() map[string]mcpkit.Kind {
	return map[string]mcpkit.Kind{
		kindDelete: {Tool: toolDelete, Fingerprint: t.fingerprint, Run: t.runDelete},
		kindReveal: {Tool: toolReveal, Fingerprint: t.fingerprint, Run: approve},
	}
}

// encode is the stored form of a target.
func (g target) encode() string {
	raw, _ := json.Marshal(g)
	return string(raw)
}

// targetOf reads an action's target.
func targetOf(a mcpkit.Action) (target, error) {
	var g target
	if err := json.Unmarshal([]byte(a.Target), &g); err != nil || g.Project == "" || g.Environment == "" || len(g.Keys) == 0 {
		return target{}, errTarget
	}
	return g, nil
}

// fingerprint digests the targeted secrets as they are now, through the person's own
// reach: a secret they can no longer see is missing.
func (t *toolset) fingerprint(ctx context.Context, a mcpkit.Action) (string, error) {
	g, err := targetOf(a)
	if err != nil {
		return "", err
	}
	found, err := t.lookup(ctx, a, g)
	if err != nil {
		return "", err
	}
	fields := []string{g.Project, g.Environment}
	for _, secret := range found {
		fields = append(fields, secret.Key, strconv.Itoa(secret.Version), secret.UpdatedAt.UTC().Format(time.RFC3339Nano))
	}
	return mcpkit.Digest(fields...), nil
}

// lookup is the metadata of every key a target names, for the action's person, in the
// target's order. A key that is gone or out of reach is not_found.
func (t *toolset) lookup(ctx context.Context, a mcpkit.Action, g target) ([]secrets.SecretResponse, error) {
	personCtx, err := t.as(ctx, a.UserID, a.ClientID)
	if err != nil {
		return nil, err
	}
	found := make([]secrets.SecretResponse, 0, len(g.Keys))
	for _, key := range g.Keys {
		secret, err := t.Secrets.Calls().Metadata(personCtx, g.Project, g.Environment, key)
		if err != nil {
			return nil, err
		}
		found = append(found, *secret)
	}
	return found, nil
}

// runDelete moves the targeted secret to the trash, as the person.
func (t *toolset) runDelete(ctx context.Context, a mcpkit.Action) error {
	g, err := targetOf(a)
	if err != nil {
		return err
	}
	personCtx, err := t.as(ctx, a.UserID, a.ClientID)
	if err != nil {
		return err
	}
	return t.Secrets.Calls().Delete(personCtx, g.Project, g.Environment, g.Keys[0])
}

// approve is a confirmed reveal's run: the confirmation is the grant, and the values
// are read by the client that asked, once, through collect.
func approve(context.Context, mcpkit.Action) error {
	return nil
}
