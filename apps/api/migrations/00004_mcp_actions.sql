-- The sensitive actions MCP clients asked for and the person has yet to confirm:
-- revealing secret values and deleting a secret. The table is tronc/mcp's
-- PendingSchema, which PostgresPending reads and writes. A confirmed reveal is
-- marked outcome = 'revealed' once its values were handed to the client.

-- +goose Up
CREATE TABLE IF NOT EXISTS mcp_actions (
	id          text        PRIMARY KEY,
	user_id     text        NOT NULL,
	kind        text        NOT NULL,
	target      text        NOT NULL,
	digest      text        NOT NULL,
	summary     text        NOT NULL DEFAULT '',
	client_id   text        NOT NULL DEFAULT '',
	created_at  timestamptz NOT NULL DEFAULT now(),
	expires_at  timestamptz NOT NULL,
	resolved_at timestamptz,
	outcome     text        NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS mcp_actions_open_idx
	ON mcp_actions (user_id, created_at DESC) WHERE resolved_at IS NULL;

-- +goose Down
DROP TABLE IF EXISTS mcp_actions;
