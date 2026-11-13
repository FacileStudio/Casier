package auth

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/FacileStudio/porte/kit"
)

// NewUserStore builds the porte user store over Casier's users table with
// porte/kit: the sign-in write, the profile RequireAuth hydrates and the
// picture backfill. Run it after the migrations, since it reads the table.
//
// kit.NewStore adopts created_at and updated_at, which the baseline left
// without a default, so an account the kit creates holds every column an
// older account holds and Unwritten is empty. The kit's linking rules are the
// ones Casier kept on its own before: a verified address only, never an
// account that already holds a subject of the same issuer, and a provider
// matches with or without one trailing slash. The stored email is rewritten
// only when the provider marks it verified, because CASIER_ADMIN_EMAILS and
// invitations read it. The table has no admin column, so no account is ever
// an admin by signing in.
func NewUserStore(ctx context.Context, sqlDB *sql.DB, logger *slog.Logger) (*kit.Store, error) {
	return kit.NewStore(ctx, kit.Store{DB: sqlDB, Logger: logger})
}
