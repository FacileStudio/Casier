package mcp

import mcpkit "github.com/FacileStudio/tronc/mcp"

// PendingItem is one action waiting on the confirm page, with the secrets it names as
// they are now. It never carries a value.
type PendingItem struct {
	mcpkit.PendingItem
	Project     string          `json:"project,omitempty"`
	Environment string          `json:"environment,omitempty"`
	Secrets     []PendingSecret `json:"secrets,omitempty"`
}

// PendingSecret is one secret an action names.
type PendingSecret struct {
	Key       string `json:"key"`
	Version   int    `json:"version"`
	UpdatedAt string `json:"updated_at"`
}
