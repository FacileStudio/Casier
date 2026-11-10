package auth

import (
	"context"
	"database/sql"
	stderrors "errors"
	"strings"

	"github.com/FacileStudio/Casier/apps/api/schemas"
	"github.com/FacileStudio/porte"
	"github.com/FacileStudio/porte/kit"
	"github.com/FacileStudio/tronc/errors"

	"gorm.io/gorm"
)

// UserStore is the porte.UserStore adapter that resolves a user table lookup
// into the identity porte hands back.
//
// The sign-in write stays here rather than in porte/kit's Store: Casier's
// linking rule compares the provider with any trailing slash ignored, so an
// account holding a subject of the issuer is never handed to a second subject
// of the same issuer spelled with or without the slash. kit.Store compares the
// provider exactly. Reading the profile and the picture backfill are the kit's.
type UserStore struct {
	orm      *gorm.DB
	profiles *kit.Store
}

// NewUserStore wires a UserStore onto the database. sqlDB is orm's own handle.
func NewUserStore(orm *gorm.DB, sqlDB *sql.DB) *UserStore {
	return &UserStore{orm: orm, profiles: &kit.Store{DB: sqlDB}}
}

// Identity reads the profile behind a user id porte authenticated, and
// answers Unauthorized when the account is gone. It is the kit.Resolver
// RequireAuth hydrates the request with.
func (s *UserStore) Identity(ctx context.Context, userID int64) (kit.Identity, error) {
	return s.profiles.Identity(ctx, userID)
}

// BackfillPictures gives users who signed in before Casier stored the picture
// claim the URL their next sign-in would write, <issuer>/avatars/<subject>. It
// fills empty rows only, ignores a trailing slash on either side, and returns
// how many rows it filled.
func (s *UserStore) BackfillPictures(ctx context.Context, issuer string) (int64, error) {
	return s.profiles.BackfillPictures(ctx, issuer)
}

// UpsertFromOIDC matches on (provider, subject) first, falls back to a verified
// email, and creates the account when neither finds anything.
//
// The email fallback is what links a pre-existing account, one created by the
// password sign-up this app used to have, to a Facile identity signing in for
// the first time. It follows porte's own linking rules: the address must be
// verified by the provider, and the account must not already hold another
// subject of the same issuer, so a recycled address never hands one person's
// vault to someone else.
//
// The stored email is rewritten only when the provider marks it verified, the
// suite rule porte/kit applies. The picture claim is written on every
// sign-in, so the stored URL follows the provider. Registre sends the same stable URL for everyone whether or not they
// set a picture, and the client falls back to initials when it answers 404.
// The locale claim is stored the same way and is what the client renders in.
func (s *UserStore) UpsertFromOIDC(ctx context.Context, claims porte.Claims) (int64, error) {
	email := strings.ToLower(strings.TrimSpace(claims.Email))
	if email == "" {
		return 0, errors.Invalid("the identity provider returned no email")
	}

	user, err := s.resolve(ctx, claims, email)
	if err != nil {
		return 0, err
	}

	updates := map[string]any{"avatar_url": kit.PictureURL(claims.Picture)}
	if claims.EmailVerified {
		updates["email"] = email
	}
	if displayName := claims.DisplayName(); displayName != "" {
		updates["name"] = displayName
	}
	if claims.Locale != "" {
		updates["locale"] = claims.Locale
	}
	if err := s.orm.WithContext(ctx).Model(&schemas.User{}).Where("id = ?", user.ID).Updates(updates).Error; err != nil {
		return 0, errors.Internal("failed to update the account", err)
	}
	return user.ID, nil
}

func (s *UserStore) resolve(ctx context.Context, claims porte.Claims, email string) (schemas.User, error) {
	var user schemas.User
	err := s.orm.WithContext(ctx).
		Joins("JOIN porte_identities i ON i.user_id = users.id AND i.provider = ? AND i.subject = ?", claims.Provider, claims.Subject).
		First(&user).Error
	switch {
	case err == nil:
		return user, nil
	case stderrors.Is(err, gorm.ErrRecordNotFound):
		return s.linkOrCreate(ctx, claims, email)
	default:
		return user, errors.Internal("failed to resolve the identity", err)
	}
}

func (s *UserStore) linkOrCreate(ctx context.Context, claims porte.Claims, email string) (schemas.User, error) {
	var user schemas.User
	err := s.orm.WithContext(ctx).Where("email = ?", email).First(&user).Error
	switch {
	case err == nil:
		return user, s.checkLink(ctx, user.ID, claims)
	case stderrors.Is(err, gorm.ErrRecordNotFound):
		user = schemas.User{Email: email, Name: claims.DisplayName()}
		if createErr := s.orm.WithContext(ctx).Create(&user).Error; createErr != nil {
			return user, errors.Internal("failed to create the account", createErr)
		}
		return user, nil
	default:
		return user, errors.Internal("failed to look up the account", err)
	}
}

func (s *UserStore) checkLink(ctx context.Context, userID int64, claims porte.Claims) error {
	if !claims.EmailVerified {
		return errors.Conflict("an account with this email already exists and the identity provider did not verify the address")
	}
	var others int64
	err := s.orm.WithContext(ctx).Table("porte_identities").
		Where("user_id = ? AND rtrim(provider, '/') = ? AND subject <> ?", userID, strings.TrimRight(claims.Provider, "/"), claims.Subject).
		Count(&others).Error
	if err != nil {
		return errors.Internal("failed to look up the account", err)
	}
	if others > 0 {
		return errors.Conflict("an account with this email is already linked to another identity")
	}
	return nil
}
