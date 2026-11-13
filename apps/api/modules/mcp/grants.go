package mcp

import (
	"context"
	"database/sql"
	stderrors "errors"

	"github.com/FacileStudio/tronc/errors"
	mcpkit "github.com/FacileStudio/tronc/mcp"
)

// outcomeRevealed marks a confirmed reveal whose values were handed over. It is
// Casier's own outcome on tronc's mcp_actions table: a done reveal is collected once,
// then never again.
const outcomeRevealed = "revealed"

// collectWindow is how long after the confirmation a client may collect a reveal.
const collectWindow = "15 minutes"

// ErrNotConfirmed is the refusal for a confirmation id that names no reveal the user
// confirmed for this client in the last 15 minutes, or one already collected.
var ErrNotConfirmed = errors.Failed("this reveal was not confirmed, was already collected, or its 15 minutes passed. " +
	"Call casier_reveal_secrets again without confirmation_id to ask the user afresh")

// collect hands one of userID's confirmed reveals, asked for by clientID, to read, and
// marks it collected only once read succeeds. The row stays locked meanwhile, so a
// second collection waits and then finds it spent; a failed read leaves the
// confirmation open for another try within the window.
func collect(ctx context.Context, db *sql.DB, userID, clientID, id string, read func(mcpkit.Action) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Internal("failed to collect a confirmed reveal", err)
	}
	defer func() { _ = tx.Rollback() }()
	var a mcpkit.Action
	err = tx.QueryRowContext(ctx, `SELECT id, user_id, kind, target, digest, client_id FROM mcp_actions
		WHERE id = $1 AND user_id = $2 AND client_id = $3 AND kind = $4 AND outcome = $5
		  AND resolved_at > now() - interval '`+collectWindow+`'
		FOR UPDATE`,
		id, userID, clientID, kindReveal, mcpkit.OutcomeDone,
	).Scan(&a.ID, &a.UserID, &a.Kind, &a.Target, &a.Digest, &a.ClientID)
	switch {
	case stderrors.Is(err, sql.ErrNoRows):
		return ErrNotConfirmed
	case err != nil:
		return errors.Internal("failed to collect a confirmed reveal", err)
	}
	if err := read(a); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE mcp_actions SET outcome = $2 WHERE id = $1`, a.ID, outcomeRevealed); err != nil {
		return errors.Internal("failed to collect a confirmed reveal", err)
	}
	if err := tx.Commit(); err != nil {
		return errors.Internal("failed to collect a confirmed reveal", err)
	}
	return nil
}
