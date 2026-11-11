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

// collect marks one of userID's confirmed reveals, asked for by clientID, as handed
// over, and returns it. The single UPDATE is what makes a confirmation single use.
func collect(ctx context.Context, db *sql.DB, userID, clientID, id string) (mcpkit.Action, error) {
	var a mcpkit.Action
	err := db.QueryRowContext(ctx, `UPDATE mcp_actions SET outcome = $5
		WHERE id = $1 AND user_id = $2 AND client_id = $3 AND kind = $4 AND outcome = $6
		  AND resolved_at > now() - interval '`+collectWindow+`'
		RETURNING id, user_id, kind, target, digest, client_id`,
		id, userID, clientID, kindReveal, outcomeRevealed, mcpkit.OutcomeDone,
	).Scan(&a.ID, &a.UserID, &a.Kind, &a.Target, &a.Digest, &a.ClientID)
	switch {
	case stderrors.Is(err, sql.ErrNoRows):
		return mcpkit.Action{}, ErrNotConfirmed
	case err != nil:
		return mcpkit.Action{}, errors.Internal("failed to collect a confirmed reveal", err)
	}
	return a, nil
}
