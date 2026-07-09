package spaces

import (
	"context"
	"crypto/rand"
	stderrors "errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"clef/internal/crypto"
	"clef/internal/errors"
	"clef/schemas"

	"gorm.io/gorm"
)

var slugRegexp = regexp.MustCompile(`[^a-z0-9]+`)

type Service struct {
	orm        *gorm.DB
	encKey     []byte
	controller *Controller
}

func NewService(orm *gorm.DB, encKey []byte) *Service {
	service := &Service{orm: orm, encKey: encKey}
	service.controller = newController(service)
	return service
}

func (service *Service) listSpaces(ctx context.Context, userID int64) ([]SpaceResponse, error) {
	var results []struct {
		schemas.Space
		Role string
	}

	err := service.orm.WithContext(ctx).
		Table("spaces s").
		Select("s.*, sm.role").
		Joins("join space_members sm on sm.space_id = s.id").
		Where("sm.user_id = ?", userID).
		Order("s.created_at desc").
		Scan(&results).Error
	if err != nil {
		return nil, errors.Internal("failed to list spaces", err)
	}

	spaces := make([]SpaceResponse, len(results))
	for i, r := range results {
		spaces[i] = SpaceResponse{
			ID:          r.ID,
			Name:        r.Name,
			Slug:        r.Slug,
			Description: r.Description,
			Role:        r.Role,
			CreatedAt:   r.CreatedAt,
			UpdatedAt:   r.UpdatedAt,
		}
	}
	return spaces, nil
}

func (service *Service) createSpace(ctx context.Context, userID int64, name, description string) (*SpaceResponse, error) {
	slug := slugify(name)
	if slug == "" {
		return nil, errors.Invalid("name must contain at least one alphanumeric character")
	}

	dataKey := make([]byte, 32)
	if _, err := rand.Read(dataKey); err != nil {
		return nil, errors.Internal("failed to generate data key", err)
	}

	encDataKey, nonce, err := crypto.Encrypt(dataKey, service.encKey)
	if err != nil {
		return nil, errors.Internal("failed to encrypt data key", err)
	}

	space := &schemas.Space{
		Name:             name,
		Slug:             slug,
		Description:      description,
		EncryptedDataKey: encDataKey,
		DataKeyNonce:     nonce,
	}

	err = service.orm.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(space).Error; err != nil {
			if stderrors.Is(err, gorm.ErrDuplicatedKey) {
				return errors.Conflict("space slug already exists")
			}
			return errors.Internal("failed to create space", err)
		}

		member := &schemas.SpaceMember{
			SpaceID: space.ID,
			UserID:  userID,
			Role:    "owner",
		}
		if err := tx.Create(member).Error; err != nil {
			return errors.Internal("failed to add owner", err)
		}

		defaultEnvs := []schemas.Environment{
			{SpaceID: space.ID, Name: "Development", Slug: "dev", Position: 0},
			{SpaceID: space.ID, Name: "Staging", Slug: "staging", Position: 1},
			{SpaceID: space.ID, Name: "Production", Slug: "prod", Position: 2},
		}
		for i := range defaultEnvs {
			if err := tx.Create(&defaultEnvs[i]).Error; err != nil {
				return errors.Internal("failed to create default environment", err)
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &SpaceResponse{
		ID:          space.ID,
		Name:        space.Name,
		Slug:        space.Slug,
		Description: space.Description,
		Role:        "owner",
		CreatedAt:   space.CreatedAt,
		UpdatedAt:   space.UpdatedAt,
	}, nil
}

func (service *Service) getSpace(ctx context.Context, slug string, userID int64) (*SpaceResponse, error) {
	var result struct {
		schemas.Space
		Role string
	}

	err := service.orm.WithContext(ctx).
		Table("spaces s").
		Select("s.*, sm.role").
		Joins("join space_members sm on sm.space_id = s.id").
		Where("s.slug = ? AND sm.user_id = ?", slug, userID).
		Scan(&result).Error
	if err != nil {
		return nil, errors.Internal("failed to get space", err)
	}
	if result.ID == 0 {
		return nil, errors.NotFound("space not found")
	}

	return &SpaceResponse{
		ID:          result.ID,
		Name:        result.Name,
		Slug:        result.Slug,
		Description: result.Description,
		Role:        result.Role,
		CreatedAt:   result.CreatedAt,
		UpdatedAt:   result.UpdatedAt,
	}, nil
}

func (service *Service) updateSpace(ctx context.Context, slug string, userID int64, name, description string) (*SpaceResponse, error) {
	space, role, err := service.getSpaceWithRole(ctx, slug, userID)
	if err != nil {
		return nil, err
	}
	if role != "owner" && role != "admin" {
		return nil, errors.Forbidden("insufficient permissions")
	}

	space.Name = name
	space.Description = description
	if err := service.orm.WithContext(ctx).Save(space).Error; err != nil {
		return nil, errors.Internal("failed to update space", err)
	}

	return &SpaceResponse{
		ID:          space.ID,
		Name:        space.Name,
		Slug:        space.Slug,
		Description: space.Description,
		Role:        role,
		CreatedAt:   space.CreatedAt,
		UpdatedAt:   space.UpdatedAt,
	}, nil
}

func (service *Service) deleteSpace(ctx context.Context, slug string, userID int64) error {
	_, role, err := service.getSpaceWithRole(ctx, slug, userID)
	if err != nil {
		return err
	}
	if role != "owner" {
		return errors.Forbidden("only the owner can delete a space")
	}

	return service.orm.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var space schemas.Space
		if err := tx.Where("slug = ?", slug).First(&space).Error; err != nil {
			return errors.Internal("failed to find space for deletion", err)
		}

		var envIDs []int64
		tx.Model(&schemas.Environment{}).Where("space_id = ?", space.ID).Pluck("id", &envIDs)
		if len(envIDs) > 0 {
			if err := tx.Where("environment_id IN ?", envIDs).Delete(&schemas.Secret{}).Error; err != nil {
				return errors.Internal("failed to delete secrets", err)
			}
		}

		if err := tx.Where("space_id = ?", space.ID).Delete(&schemas.Environment{}).Error; err != nil {
			return errors.Internal("failed to delete environments", err)
		}

		if err := tx.Where("space_id = ?", space.ID).Delete(&schemas.SpaceMember{}).Error; err != nil {
			return errors.Internal("failed to delete members", err)
		}

		if err := tx.Delete(&space).Error; err != nil {
			return errors.Internal("failed to delete space", err)
		}

		return nil
	})
}

func (service *Service) listMembers(ctx context.Context, slug string, userID int64) ([]MemberResponse, error) {
	_, _, err := service.getSpaceWithRole(ctx, slug, userID)
	if err != nil {
		return nil, err
	}

	var results []struct {
		ID    int64
		Email string
		Role  string
	}

	err = service.orm.WithContext(ctx).
		Table("space_members sm").
		Select("u.id, u.email, sm.role").
		Joins("join users u on u.id = sm.user_id").
		Joins("join spaces s on s.id = sm.space_id").
		Where("s.slug = ?", slug).
		Order("sm.created_at asc").
		Scan(&results).Error
	if err != nil {
		return nil, errors.Internal("failed to list members", err)
	}

	members := make([]MemberResponse, len(results))
	for i, r := range results {
		members[i] = MemberResponse{ID: r.ID, Email: r.Email, Role: r.Role}
	}
	return members, nil
}

func (service *Service) addMember(ctx context.Context, slug string, callerID int64, email, role string) error {
	_, callerRole, err := service.getSpaceWithRole(ctx, slug, callerID)
	if err != nil {
		return err
	}
	if callerRole != "owner" && callerRole != "admin" {
		return errors.Forbidden("insufficient permissions")
	}

	var user schemas.User
	if err := service.orm.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return errors.NotFound("user not found")
		}
		return errors.Internal("failed to find user", err)
	}

	var space schemas.Space
	if err := service.orm.WithContext(ctx).Where("slug = ?", slug).First(&space).Error; err != nil {
		return errors.Internal("failed to find space", err)
	}

	member := &schemas.SpaceMember{
		SpaceID: space.ID,
		UserID:  user.ID,
		Role:    role,
	}
	if err := service.orm.WithContext(ctx).Create(member).Error; err != nil {
		if stderrors.Is(err, gorm.ErrDuplicatedKey) {
			return errors.Conflict("user is already a member")
		}
		return errors.Internal("failed to add member", err)
	}
	return nil
}

func (service *Service) removeMember(ctx context.Context, slug string, callerID int64, targetUserID int64) error {
	_, callerRole, err := service.getSpaceWithRole(ctx, slug, callerID)
	if err != nil {
		return err
	}
	if callerRole != "owner" && callerRole != "admin" {
		return errors.Forbidden("insufficient permissions")
	}

	var space schemas.Space
	if err := service.orm.WithContext(ctx).Where("slug = ?", slug).First(&space).Error; err != nil {
		return errors.Internal("failed to find space", err)
	}

	var target schemas.SpaceMember
	err = service.orm.WithContext(ctx).
		Where("space_id = ? AND user_id = ?", space.ID, targetUserID).
		First(&target).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return errors.NotFound("member not found")
	}
	if err != nil {
		return errors.Internal("failed to find member", err)
	}

	if target.Role == "owner" {
		var ownerCount int64
		service.orm.WithContext(ctx).
			Model(&schemas.SpaceMember{}).
			Where("space_id = ? AND role = ?", space.ID, "owner").
			Count(&ownerCount)
		if ownerCount <= 1 {
			return errors.Forbidden("cannot remove the last owner")
		}
	}

	if err := service.orm.WithContext(ctx).Delete(&target).Error; err != nil {
		return errors.Internal("failed to remove member", err)
	}
	return nil
}

func (service *Service) getSpaceWithRole(ctx context.Context, slug string, userID int64) (*schemas.Space, string, error) {
	var space schemas.Space
	if err := service.orm.WithContext(ctx).Where("slug = ?", slug).First(&space).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", errors.NotFound("space not found")
		}
		return nil, "", errors.Internal("failed to find space", err)
	}

	var member schemas.SpaceMember
	err := service.orm.WithContext(ctx).
		Where("space_id = ? AND user_id = ?", space.ID, userID).
		First(&member).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return nil, "", errors.NotFound("space not found")
	}
	if err != nil {
		return nil, "", errors.Internal("failed to check membership", err)
	}

	return &space, member.Role, nil
}

func (service *Service) resolveSpaceID(ctx context.Context, slug string, userID int64, requiredRoles ...string) (int64, error) {
	space, role, err := service.getSpaceWithRole(ctx, slug, userID)
	if err != nil {
		return 0, err
	}

	if len(requiredRoles) > 0 {
		allowed := false
		for _, r := range requiredRoles {
			if role == r {
				allowed = true
				break
			}
		}
		if !allowed {
			return 0, errors.Forbidden("insufficient permissions")
		}
	}

	return space.ID, nil
}

func (service *Service) ResolveSpaceID(ctx context.Context, slug string, userID int64, requiredRoles ...string) (int64, error) {
	return service.resolveSpaceID(ctx, slug, userID, requiredRoles...)
}

func (service *Service) GetSpaceDataKey(ctx context.Context, spaceID int64) ([]byte, error) {
	var space schemas.Space
	if err := service.orm.WithContext(ctx).
		Select("encrypted_data_key, data_key_nonce").
		Where("id = ?", spaceID).
		First(&space).Error; err != nil {
		return nil, fmt.Errorf("failed to load space data key: %w", err)
	}

	if len(space.EncryptedDataKey) == 0 {
		return nil, fmt.Errorf("space %d has no data key", spaceID)
	}

	return crypto.Decrypt(space.EncryptedDataKey, space.DataKeyNonce, service.encKey)
}

func slugify(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))
	s = slugRegexp.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}

func parseUserID(raw string) (int64, error) {
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, errors.Invalid("invalid user id")
	}
	return id, nil
}
