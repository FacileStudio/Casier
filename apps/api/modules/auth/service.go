package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	stderrors "errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"casier/internal/authcrypto"
	"casier/internal/crypto"
	"casier/internal/errors"
	"casier/schemas"

	gooidc "github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

type Service struct {
	orm        *gorm.DB
	encKey     []byte
	logger     *slog.Logger
	controller *Controller
}

func NewService(orm *gorm.DB, encKey []byte, logger *slog.Logger) *Service {
	service := &Service{orm: orm, encKey: encKey, logger: logger}
	service.controller = newController(service)
	return service
}

func (service *Service) encryptToken(token string) ([]byte, []byte, error) {
	if token == "" {
		return nil, nil, nil
	}
	return crypto.Encrypt([]byte(token), service.encKey)
}

func (service *Service) decryptToken(ciphertext, nonce []byte) string {
	if len(ciphertext) == 0 {
		return ""
	}
	plaintext, err := crypto.Decrypt(ciphertext, nonce, service.encKey)
	if err != nil {
		return ""
	}
	return string(plaintext)
}

func (service *Service) registerUser(ctx context.Context, email string, password string) (string, string, error) {
	hash, err := authcrypto.HashPassword(password)
	if err != nil {
		return "", "", errors.Invalid("invalid password")
	}

	record := &schemas.User{
		Email:        email,
		PasswordHash: hash,
	}
	if err := service.orm.WithContext(ctx).Create(record).Error; err != nil {
		if stderrors.Is(err, gorm.ErrDuplicatedKey) {
			return "", "", errors.Conflict("email already registered")
		}
		return "", "", errors.Internal("failed to create user", err)
	}

	token, err := authcrypto.NewToken()
	if err != nil {
		return "", "", errors.Internal("failed to create session", err)
	}
	if err := service.insertSession(ctx, token, record.ID); err != nil {
		return "", "", err
	}

	return strconv.FormatInt(record.ID, 10), token, nil
}

func (service *Service) loginUser(ctx context.Context, email string, password string) (string, string, error) {
	var record schemas.User
	err := service.orm.WithContext(ctx).Where("email = ?", email).First(&record).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		authcrypto.VerifyPassword(password, authcrypto.DummyHash)
		return "", "", errors.Unauthorized("invalid credentials")
	}
	if err != nil {
		return "", "", errors.Internal("failed to read user", err)
	}
	if !authcrypto.VerifyPassword(password, record.PasswordHash) {
		return "", "", errors.Unauthorized("invalid credentials")
	}

	token, err := authcrypto.NewToken()
	if err != nil {
		return "", "", errors.Internal("failed to create session", err)
	}
	if err := service.insertSession(ctx, token, record.ID); err != nil {
		return "", "", err
	}

	return strconv.FormatInt(record.ID, 10), token, nil
}

func (service *Service) logoutUser(ctx context.Context, authorization string) error {
	token := normalizeBearer(authorization)
	if token == "" {
		return errors.Unauthorized("missing auth token")
	}

	result := service.orm.WithContext(ctx).Where("token = ?", hashToken(token)).Delete(&schemas.Session{})
	if result.Error != nil {
		return errors.Internal("failed to delete session", result.Error)
	}
	return nil
}

func (service *Service) insertSession(ctx context.Context, token string, userID int64) error {
	record := &schemas.Session{
		Token:     hashToken(token),
		UserID:    userID,
		ExpiresAt: time.Now().Add(schemas.SessionSlidingWindow),
	}
	if err := service.orm.WithContext(ctx).Create(record).Error; err != nil {
		return errors.Internal("failed to persist session", err)
	}
	return nil
}

func SetSessionCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   int(schemas.SessionMaxLifetime.Seconds()),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}

func ClearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}

func normalizeBearer(authorization string) string {
	value := strings.TrimSpace(authorization)
	if len(value) >= 7 && strings.EqualFold(value[:7], "bearer ") {
		return strings.TrimSpace(value[7:])
	}
	return value
}

func (service *Service) authenticateRequest(ctx context.Context, authorization string) (string, *Data, error) {
	token := normalizeBearer(authorization)
	if token == "" {
		return "", nil, errors.Unauthorized("missing auth token")
	}

	hashed := hashToken(token)
	var out struct {
		UserID    int64
		Email     string
		ExpiresAt time.Time
		CreatedAt time.Time
	}
	err := service.orm.WithContext(ctx).
		Table("sessions s").
		Select("u.id as user_id, u.email as email, s.expires_at as expires_at, s.created_at as created_at").
		Joins("join users u on u.id = s.user_id").
		Where("s.token = ?", hashed).
		Scan(&out).Error
	if err != nil {
		return "", nil, errors.Internal("failed to validate auth token", err)
	}
	if out.UserID == 0 {
		return "", nil, errors.Unauthorized("invalid auth token")
	}

	now := time.Now()
	if now.After(out.ExpiresAt) {
		return "", nil, errors.Unauthorized("expired auth token")
	}
	if now.Sub(out.CreatedAt) > schemas.SessionMaxLifetime {
		return "", nil, errors.Unauthorized("session expired")
	}

	newExpiry := now.Add(schemas.SessionSlidingWindow)
	absoluteMax := out.CreatedAt.Add(schemas.SessionMaxLifetime)
	if newExpiry.After(absoluteMax) {
		newExpiry = absoluteMax
	}
	go func() {
		service.orm.Model(&schemas.Session{}).
			Where("token = ?", hashed).
			Update("expires_at", newExpiry)
	}()

	return strconv.FormatInt(out.UserID, 10), &Data{Email: out.Email}, nil
}

func (service *Service) Authenticate(ctx context.Context, authorization string) (string, any, error) {
	token := normalizeBearer(authorization)
	if strings.HasPrefix(token, "casier_") {
		return service.authenticateApiToken(ctx, token)
	}
	return service.authenticateRequest(ctx, authorization)
}

func (service *Service) authenticateApiToken(ctx context.Context, token string) (string, *Data, error) {
	hashed := hashToken(token)
	var out struct {
		UserID    int64
		Email     string
		ExpiresAt time.Time
	}
	err := service.orm.WithContext(ctx).
		Table("api_tokens at").
		Select("u.id as user_id, u.email as email, at.expires_at as expires_at").
		Joins("join users u on u.id = at.created_by").
		Where("at.token_hash = ?", hashed).
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

	go func() {
		service.orm.Model(&schemas.ApiToken{}).
			Where("token_hash = ?", hashed).
			Update("last_used_at", time.Now())
	}()

	return strconv.FormatInt(out.UserID, 10), &Data{Email: out.Email}, nil
}

func (service *Service) StartSessionCleanup(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(15 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				result := service.orm.Where("expires_at < ?", time.Now()).Delete(&schemas.Session{})
				if result.Error != nil {
					service.logger.Error("session cleanup failed", slog.Any("error", result.Error))
				} else if result.RowsAffected > 0 {
					service.logger.Info("purged expired sessions", slog.Int64("count", result.RowsAffected))
				}
			}
		}
	}()
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func (service *Service) forgotPassword(ctx context.Context, email string) error {
	var user schemas.User
	err := service.orm.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	if err != nil {
		return errors.Internal("failed to look up user", err)
	}
	if user.PasswordHash == "" {
		return nil
	}

	token, err := authcrypto.NewToken()
	if err != nil {
		return errors.Internal("failed to generate reset token", err)
	}

	reset := &schemas.PasswordReset{
		Token:     hashToken(token),
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}
	if err := service.orm.WithContext(ctx).Create(reset).Error; err != nil {
		return errors.Internal("failed to create reset token", err)
	}

	service.logger.Info("password reset requested",
		slog.String("email", email),
		slog.String("reset_token", token),
	)

	return nil
}

func (service *Service) resetPassword(ctx context.Context, token string, newPassword string) error {
	hashed := hashToken(token)
	var reset schemas.PasswordReset
	err := service.orm.WithContext(ctx).Where("token = ?", hashed).First(&reset).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return errors.Invalid("invalid or expired reset token")
	}
	if err != nil {
		return errors.Internal("failed to validate reset token", err)
	}
	if reset.Used || time.Now().After(reset.ExpiresAt) {
		return errors.Invalid("invalid or expired reset token")
	}

	hash, err := authcrypto.HashPassword(newPassword)
	if err != nil {
		return errors.Invalid("invalid password")
	}

	return service.orm.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&schemas.User{}).Where("id = ?", reset.UserID).Update("password_hash", hash).Error; err != nil {
			return errors.Internal("failed to update password", err)
		}
		if err := tx.Model(&schemas.PasswordReset{}).Where("token = ?", hashed).Update("used", true).Error; err != nil {
			return errors.Internal("failed to mark token used", err)
		}
		if err := tx.Where("user_id = ?", reset.UserID).Delete(&schemas.Session{}).Error; err != nil {
			return errors.Internal("failed to clear sessions", err)
		}
		return nil
	})
}

func (service *Service) upsertOIDCUser(ctx context.Context, email string, name string, oauth2Token *oauth2.Token) (string, string, error) {
	var record schemas.User
	err := service.orm.WithContext(ctx).Where("email = ?", email).First(&record).Error
	if err != nil && !stderrors.Is(err, gorm.ErrRecordNotFound) {
		return "", "", errors.Internal("failed to look up user", err)
	}

	isNew := stderrors.Is(err, gorm.ErrRecordNotFound)
	if isNew {
		record = schemas.User{Email: email, Name: name}
		if err := service.orm.WithContext(ctx).Create(&record).Error; err != nil {
			return "", "", errors.Internal("failed to create user", err)
		}
	} else if name != "" {
		record.Name = name
	}

	atEnc, atNonce, err := service.encryptToken(oauth2Token.AccessToken)
	if err != nil {
		return "", "", errors.Internal("failed to encrypt access token", err)
	}
	rtEnc, rtNonce, err := service.encryptToken(oauth2Token.RefreshToken)
	if err != nil {
		return "", "", errors.Internal("failed to encrypt refresh token", err)
	}
	record.OIDCAccessTokenEncrypted = atEnc
	record.OIDCAccessTokenNonce = atNonce
	record.OIDCRefreshTokenEncrypted = rtEnc
	record.OIDCRefreshTokenNonce = rtNonce
	record.OIDCTokenExpiry = oauth2Token.Expiry
	record.ProfileSyncedAt = time.Now()
	if err := service.orm.WithContext(ctx).Save(&record).Error; err != nil {
		return "", "", errors.Internal("failed to save user", err)
	}

	token, err := authcrypto.NewToken()
	if err != nil {
		return "", "", errors.Internal("failed to create session", err)
	}
	if err := service.insertSession(ctx, token, record.ID); err != nil {
		return "", "", err
	}
	return strconv.FormatInt(record.ID, 10), token, nil
}

func (service *Service) SyncOIDCProfile(ctx context.Context, userID string, provider *gooidc.Provider, oauth2Cfg oauth2.Config) (bool, error) {
	var record schemas.User
	if err := service.orm.WithContext(ctx).Where("id = ?", userID).First(&record).Error; err != nil {
		return false, errors.Internal("failed to load user", err)
	}

	if len(record.OIDCAccessTokenEncrypted) == 0 {
		return false, nil
	}

	if time.Since(record.ProfileSyncedAt) < 5*time.Minute {
		return false, nil
	}

	storedToken := &oauth2.Token{
		AccessToken:  service.decryptToken(record.OIDCAccessTokenEncrypted, record.OIDCAccessTokenNonce),
		RefreshToken: service.decryptToken(record.OIDCRefreshTokenEncrypted, record.OIDCRefreshTokenNonce),
		Expiry:       record.OIDCTokenExpiry,
	}

	tokenSource := oauth2Cfg.TokenSource(ctx, storedToken)

	userInfo, err := provider.UserInfo(ctx, tokenSource)
	if err != nil {
		service.logger.Warn("OIDC profile sync failed, clearing tokens", slog.String("user_id", userID), slog.Any("error", err))
		record.OIDCAccessTokenEncrypted = nil
		record.OIDCAccessTokenNonce = nil
		record.OIDCRefreshTokenEncrypted = nil
		record.OIDCRefreshTokenNonce = nil
		record.OIDCTokenExpiry = time.Time{}
		if err := service.orm.WithContext(ctx).Save(&record).Error; err != nil {
			return false, errors.Internal("failed to clear OIDC tokens", err)
		}
		return false, nil
	}

	var claims struct {
		Name string `json:"name"`
	}
	if err := userInfo.Claims(&claims); err != nil {
		service.logger.Warn("failed to parse UserInfo claims", slog.String("user_id", userID), slog.Any("error", err))
		return false, nil
	}

	if claims.Name != "" {
		record.Name = claims.Name
	}

	newToken, tokenErr := tokenSource.Token()
	if tokenErr == nil {
		atEnc, atNonce, encErr := service.encryptToken(newToken.AccessToken)
		if encErr == nil {
			record.OIDCAccessTokenEncrypted = atEnc
			record.OIDCAccessTokenNonce = atNonce
		}
		rtEnc, rtNonce, encErr := service.encryptToken(newToken.RefreshToken)
		if encErr == nil {
			record.OIDCRefreshTokenEncrypted = rtEnc
			record.OIDCRefreshTokenNonce = rtNonce
		}
		record.OIDCTokenExpiry = newToken.Expiry
	}

	record.ProfileSyncedAt = time.Now()
	if err := service.orm.WithContext(ctx).Save(&record).Error; err != nil {
		return false, errors.Internal("failed to save synced profile", err)
	}

	service.logger.Info("synced OIDC profile", slog.String("user_id", userID))
	return true, nil
}
