package auth

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/FacileStudio/porte/directory"

	"gorm.io/gorm"
)

// Deletions erases the people Registre erased. It polls Registre's deletion
// list with the app's own OIDC client credentials and accepts Registre's
// signed push on directory.DeletionsPath.
type Deletions struct {
	receiver *directory.Deletions
}

// NewDeletions builds the receiver over Casier's tables. erase is porte's own
// rows eraser, pg.Store.EraseIdentity.
func NewDeletions(orm *gorm.DB, client directory.Client, erase func(context.Context, string, string) error, logger *slog.Logger) *Deletions {
	return &Deletions{receiver: &directory.Deletions{
		Client: client,
		OnPersonDeleted: func(ctx context.Context, provider, subject string) error {
			return ErasePerson(ctx, orm, provider, subject)
		},
		Erase:  erase,
		State:  deletionState{orm: orm},
		Logger: logger,
	}}
}

// Run polls until ctx ends. Start it in exactly one replica.
func (d *Deletions) Run(ctx context.Context) {
	d.receiver.Run(ctx)
}

// Handler verifies and applies Registre's signed user.deleted push.
func (d *Deletions) Handler() http.Handler {
	return d.receiver.Handler()
}

// deletionState keeps the receiver's cursor and handled subjects in one row.
type deletionState struct {
	orm *gorm.DB
}

// Load returns the saved state, or nil before the first save.
func (s deletionState) Load(ctx context.Context) ([]byte, error) {
	var rows []struct{ State []byte }
	err := s.orm.WithContext(ctx).Raw(`SELECT state FROM directory_deletion_state WHERE id = 1`).Scan(&rows).Error
	if err != nil || len(rows) == 0 {
		return nil, err
	}
	return rows[0].State, nil
}

// Save replaces the saved state.
func (s deletionState) Save(ctx context.Context, state []byte) error {
	return s.orm.WithContext(ctx).Exec(
		`INSERT INTO directory_deletion_state (id, state, updated_at) VALUES (1, ?, now())
		 ON CONFLICT (id) DO UPDATE SET state = EXCLUDED.state, updated_at = now()`, state).Error
}
