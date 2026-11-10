package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/FacileStudio/Casier/apps/api/internal/authcontext"
	"github.com/FacileStudio/Casier/apps/api/internal/crypto"
	"github.com/FacileStudio/Casier/apps/api/internal/people"
	"github.com/FacileStudio/porte/kit"
	"github.com/FacileStudio/porte/session"
	"github.com/FacileStudio/tronc/errors"

	"gorm.io/gorm"
)

// Service implements authentication: porte sessions and OIDC for people, and
// scoped API tokens for machines and the CLI.
type Service struct {
	orm        *gorm.DB
	keyring    *crypto.Keyring
	logger     *slog.Logger
	sessions   *session.Manager
	users      kit.Resolver
	controller *Controller
}

// NewService wires an auth service onto the database and the porte session
// manager. users resolves the account behind a session: the UserStore.
func NewService(orm *gorm.DB, keyring *crypto.Keyring, logger *slog.Logger, sessions *session.Manager, users kit.Resolver) *Service {
	service := &Service{
		orm:      orm,
		keyring:  keyring,
		logger:   logger,
		sessions: sessions,
		users:    users,
	}
	service.controller = newController(service)
	return service
}

// RequireAuth accepts a user session only, through porte/kit: porte checks
// the cookie or bearer session and the CSRF rule, and the kit reads the
// account. A session whose account is gone answers 401. The identity is then
// carried as Casier's own, whose user id is a string and whose scope is nil.
func (service *Service) RequireAuth(next http.Handler) http.Handler {
	return kit.RequireAuth(service.sessions, service.users)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		identity, _ := kit.FromContext(r.Context())
		ctx := authcontext.WithIdentity(r.Context(), authcontext.Identity{
			UserID:   strconv.FormatInt(identity.UserID, 10),
			Email:    identity.Email,
			ClientIP: authcontext.ClientIP(r),
		})
		next.ServeHTTP(w, r.WithContext(ctx))
	}))
}

// profile loads the signed-in user as the client renders them. A session whose
// user row is gone answers 401, the same as RequireAuth would on the next call.
func (service *Service) profile(ctx context.Context, userID string) (*MeResponse, error) {
	var me MeResponse
	result := service.orm.WithContext(ctx).
		Table("users u").
		Select(people.Columns+", u.locale").
		Where("u.id = ?", userID).
		Scan(&me)
	if result.Error != nil {
		return nil, errors.Internal("failed to load the account", result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, errors.Unauthorized("missing auth")
	}
	me.ID = userID
	return &me, nil
}

func (service *Service) StartSessionCleanup(ctx context.Context) {
	go func() {
		sweepLoop(ctx, service.sessions, service.logger)
	}()
}

func (service *Service) AuthenticateAPI(ctx context.Context, authorization string) (string, *authcontext.TokenScope, error) {
	token := normalizeBearer(authorization)
	if !strings.HasPrefix(token, "casier_") {
		return "", nil, errors.Unauthorized("invalid API token")
	}

	hashed := hashToken(token)
	var out apiTokenAuthRow
	err := service.orm.WithContext(ctx).
		Table("api_tokens at").
		Select("u.id as user_id, at.expires_at as expires_at, at.id as token_id, at.project_id as project_id, at.all_environments as all_environments, at.permissions as permissions").
		Joins("join users u on u.id = at.created_by").
		Where("at.token_hash = ? AND at.revoked_at IS NULL", hashed).
		Scan(&out).Error
	if err != nil {
		return "", nil, errors.Internal("failed to validate API token", err)
	}
	if out.UserID == 0 {
		return "", nil, errors.Unauthorized("invalid API token")
	}
	if time.Now().After(out.ExpiresAt) {
		return "", nil, errors.Unauthorized("expired API token")
	}

	var environments []string
	if !out.AllEnvironments {
		service.orm.WithContext(ctx).
			Table("api_token_environments te").
			Select("e.slug").
			Joins("join environments e on e.id = te.environment_id").
			Where("te.token_id = ?", out.TokenID).
			Scan(&environments)
	}

	return strconv.FormatInt(out.UserID, 10), &authcontext.TokenScope{
		TokenID:         out.TokenID,
		ProjectID:       out.ProjectID,
		AllEnvironments: out.AllEnvironments,
		Environments:    environments,
		Permissions:     authcontext.ParseScopeList(out.Permissions),
	}, nil
}

type apiTokenScopeRow struct {
	TokenID         int64
	ProjectID       int64
	AllEnvironments bool
	Environments    []string
	Permissions     string
}

type apiTokenAuthRow struct {
	UserID          int64
	ExpiresAt       time.Time
	TokenID         int64
	ProjectID       int64
	AllEnvironments bool
	Permissions     string
}

func scopeFromTokenRow(row apiTokenScopeRow) *authcontext.TokenScope {
	return &authcontext.TokenScope{
		TokenID:         row.TokenID,
		ProjectID:       row.ProjectID,
		AllEnvironments: row.AllEnvironments,
		Environments:    row.Environments,
		Permissions:     authcontext.ParseScopeList(row.Permissions),
	}
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func normalizeBearer(authorization string) string {
	value := strings.TrimSpace(authorization)
	if len(value) >= 7 && strings.EqualFold(value[:7], "bearer ") {
		return strings.TrimSpace(value[7:])
	}
	return value
}
