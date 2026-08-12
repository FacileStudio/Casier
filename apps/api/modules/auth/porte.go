package auth

import (
	"context"
	stderrors "errors"
	"fmt"

	"github.com/FacileStudio/Casier/apps/api/schemas"
	"github.com/FacileStudio/porte"

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

func (s *UserStore) UpsertFromOIDC(ctx context.Context, claims porte.Claims) (int64, error) {
	var user schemas.User
	err := s.orm.WithContext(ctx).
		Joins("JOIN porte_identities i ON i.user_id = users.id AND i.provider = ? AND i.subject = ?", claims.Provider, claims.Subject).
		First(&user).Error
	if err == nil {
		if claims.Name != "" {
			user.Name = claims.Name
			s.orm.WithContext(ctx).Save(&user)
		}
		return user.ID, nil
	}
	if !stderrors.Is(err, gorm.ErrRecordNotFound) {
		return 0, err
	}

	if !claims.EmailVerified {
		err = s.orm.WithContext(ctx).Where("email = ?", claims.Email).First(&user).Error
		if err == nil {
			return 0, fmt.Errorf("email already exists and is not verified")
		}
		if !stderrors.Is(err, gorm.ErrRecordNotFound) {
			return 0, err
		}
	}

	user = schemas.User{Email: claims.Email, Name: claims.Name}
	if err := s.orm.WithContext(ctx).Create(&user).Error; err != nil {
		if stderrors.Is(err, gorm.ErrDuplicatedKey) {
			return 0, fmt.Errorf("email already exists and is not verified")
		}
		return 0, err
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
