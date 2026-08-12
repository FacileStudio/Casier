package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/FacileStudio/Casier/apps/api/internal/authcontext"
	"github.com/FacileStudio/Casier/apps/api/internal/crypto"
	"github.com/FacileStudio/Casier/apps/api/internal/mailer"
	"github.com/FacileStudio/Casier/apps/api/schemas"
	"github.com/FacileStudio/porte"
	"github.com/FacileStudio/porte/local"
	"github.com/FacileStudio/porte/session"
	"github.com/FacileStudio/tronc/errors"

	"gorm.io/gorm"
)

// Service implements authentication: sessions, passwords via porte, OIDC and
// password reset, with a mailer seam for the reset delivery.
type Service struct {
	orm          *gorm.DB
	keyring      *crypto.Keyring
	logger       *slog.Logger
	mailer       mailer.Mailer
	publicURL    string
	sessions     *session.Manager
	passwords    *local.Kit
	identities   porte.IdentityStore
	controller   *Controller
	deliverReset func(ctx context.Context, email string)
}

// NewService wires an auth service onto the database and the porte kits.
func NewService(orm *gorm.DB, keyring *crypto.Keyring, logger *slog.Logger, mail mailer.Mailer, publicURL string, sessions *session.Manager, passwords *local.Kit, identities porte.IdentityStore) *Service {
	service := &Service{
		orm:        orm,
		keyring:    keyring,
		logger:     logger,
		mailer:     mail,
		publicURL:  publicURL,
		sessions:   sessions,
		passwords:  passwords,
		identities: identities,
	}
	service.controller = newController(service)
	service.deliverReset = service.sendPasswordReset
	return service
}

func (service *Service) RequireAuth(next http.Handler) http.Handler {
	return service.sessions.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		identity, ok := porte.From(r.Context())
		if !ok {
			http.Error(w, "", http.StatusUnauthorized)
			return
		}

		userID := strconv.FormatInt(identity.UserID, 10)
		email, err := service.IdentityForUser(r.Context(), identity.UserID)
		if err != nil {
			http.Error(w, "", http.StatusUnauthorized)
			return
		}

		_, port, _ := net.SplitHostPort(r.RemoteAddr)
		_ = port
		clientIP := r.RemoteAddr
		if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
			clientIP = strings.Split(fwd, ",")[0]
		}
		ctx := authcontext.WithIdentity(r.Context(), authcontext.Identity{
			UserID:   userID,
			Email:    email,
			ClientIP: clientIP,
		})
		next.ServeHTTP(w, r.WithContext(ctx))
	}))
}

func (service *Service) IdentityForUser(ctx context.Context, userID int64) (string, error) {
	var out struct{ Email string }
	err := service.orm.WithContext(ctx).
		Model(&schemas.User{}).
		Select("email").
		Where("id = ?", userID).
		Scan(&out).Error
	if err != nil {
		return "", err
	}
	return out.Email, nil
}

func (service *Service) registerUser(ctx context.Context, email, password string) (int64, string, error) {
	w := &noopResponseWriter{}
	r, _ := http.NewRequestWithContext(ctx, "POST", "/", nil)
	return service.passwords.Register(ctx, w, r, email, "", password)
}

func (service *Service) loginUser(ctx context.Context, email, password string) (int64, string, error) {
	w := &noopResponseWriter{}
	r, _ := http.NewRequestWithContext(ctx, "POST", "/", nil)
	return service.passwords.Login(ctx, w, r, email, password)
}

func (service *Service) authenticateRequest(ctx context.Context, bearer string) (int64, string, error) {
	token := strings.TrimPrefix(bearer, "Bearer ")
	r, _ := http.NewRequestWithContext(ctx, "GET", "/", nil)
	r.Header.Set("Authorization", "Bearer "+token)
	w := &noopResponseWriter{}
	id, err := service.sessions.Authenticate(w, r)
	if err != nil {
		return 0, "", err
	}
	return id.UserID, "", nil
}

type noopResponseWriter struct{}

func (w *noopResponseWriter) Header() http.Header         { return http.Header{} }
func (w *noopResponseWriter) Write(b []byte) (int, error) { return len(b), nil }
func (w *noopResponseWriter) WriteHeader(int)             {}

func (service *Service) ConfigExtra() func() map[string]any {
	return func() map[string]any {
		return map[string]any{
			"password_reset_enabled": service.mailer != nil,
		}
	}
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
