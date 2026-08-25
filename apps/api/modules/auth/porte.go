package auth

import (
	"context"
	stderrors "errors"
	"strings"

	"github.com/FacileStudio/Casier/apps/api/schemas"
	"github.com/FacileStudio/porte"
	"github.com/FacileStudio/tronc/errors"

	"gorm.io/gorm"
)

// UserStore is the porte.register adapter that resolves a user table lookup
// into the identity porte hands back.
type UserStore struct {
	orm *gorm.DB
}

// NewUserStore wires a UserStore onto the database.
func NewUserStore(orm *gorm.DB) *UserStore {
	return &UserStore{orm: orm}
}

// UpsertFromOIDC matches on (provider, subject) first, falls back to a verified
// email, and creates the account when neither finds anything.
//
// The email fallback is what links a pre-existing account to a subject signing
// in for the first time, which is exactly what an identity provider change
// produces: both halves of the key move, so every existing user arrives here
// looking new. It is refused for an unverified address because matching on one
// is an account takeover primitive.
//
// This guard used to be inverted — it consulted the email only when the address
// was NOT verified, and had no linking branch at all, so a verified address that
// missed on (provider, subject) fell straight through to Create and died on the
// unique index. That is every existing user, on their first login after the
// registre cutover.
func (s *UserStore) UpsertFromOIDC(ctx context.Context, claims porte.Claims) (int64, error) {
	email := strings.ToLower(strings.TrimSpace(claims.Email))
	if email == "" {
		return 0, errors.Invalid("the identity provider returned no email")
	}

	var user schemas.User
	err := s.orm.WithContext(ctx).
		Joins("JOIN porte_identities i ON i.user_id = users.id AND i.provider = ? AND i.subject = ?", claims.Provider, claims.Subject).
		First(&user).Error
	switch {
	case err == nil:
	case stderrors.Is(err, gorm.ErrRecordNotFound):
		err = s.orm.WithContext(ctx).Where("email = ?", email).First(&user).Error
		switch {
		case err == nil:
			if !claims.EmailVerified {
				return 0, errors.Conflict("an account with this email already exists and the identity provider did not verify the address")
			}
		case stderrors.Is(err, gorm.ErrRecordNotFound):
			user = schemas.User{Email: email, Name: claims.DisplayName()}
			if createErr := s.orm.WithContext(ctx).Create(&user).Error; createErr != nil {
				return 0, errors.Internal("failed to create the account", createErr)
			}
		default:
			return 0, errors.Internal("failed to look up the account", err)
		}
	default:
		return 0, errors.Internal("failed to resolve the identity", err)
	}

	updates := map[string]any{"email": email}
	if displayName := claims.DisplayName(); displayName != "" {
		updates["name"] = displayName
	}
	if err := s.orm.WithContext(ctx).Model(&schemas.User{}).Where("id = ?", user.ID).Updates(updates).Error; err != nil {
		return 0, errors.Internal("failed to update the account", err)
	}
	return user.ID, nil
}

func (s *UserStore) CreateFromPassword(ctx context.Context, email, name string) (int64, error) {
	user := schemas.User{Email: email, Name: name}
	if err := s.orm.WithContext(ctx).Create(&user).Error; err != nil {
		return 0, err
	}
	return user.ID, nil
}

func (s *UserStore) FindByEmail(ctx context.Context, email string) (int64, error) {
	var user schemas.User
	err := s.orm.WithContext(ctx).Select("id").Where("email = ?", email).First(&user).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return 0, porte.ErrNotFound
	}
	return user.ID, err
}

func (s *UserStore) CountUsers(ctx context.Context) (int64, error) {
	var count int64
	err := s.orm.WithContext(ctx).Model(&schemas.User{}).Count(&count).Error
	return count, err
}
