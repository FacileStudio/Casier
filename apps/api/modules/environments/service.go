package environments

import (
	"context"
	stderrors "errors"
	"regexp"
	"strings"

	"clef/internal/errors"
	"clef/schemas"

	"gorm.io/gorm"
)

var slugRegexp = regexp.MustCompile(`[^a-z0-9]+`)

type SpaceResolver interface {
	ResolveSpaceID(ctx context.Context, slug string, userID int64, requiredRoles ...string) (int64, error)
}

type Service struct {
	orm           *gorm.DB
	spaceResolver SpaceResolver
	controller    *Controller
}

func NewService(orm *gorm.DB, spaceResolver SpaceResolver) *Service {
	service := &Service{orm: orm, spaceResolver: spaceResolver}
	service.controller = newController(service)
	return service
}

func (service *Service) listEnvironments(ctx context.Context, spaceID int64) ([]EnvironmentResponse, error) {
	var records []schemas.Environment
	err := service.orm.WithContext(ctx).
		Where("space_id = ?", spaceID).
		Order("position asc").
		Find(&records).Error
	if err != nil {
		return nil, errors.Internal("failed to list environments", err)
	}

	envs := make([]EnvironmentResponse, len(records))
	for i, r := range records {
		envs[i] = EnvironmentResponse{
			ID:        r.ID,
			Name:      r.Name,
			Slug:      r.Slug,
			Position:  r.Position,
			CreatedAt: r.CreatedAt,
			UpdatedAt: r.UpdatedAt,
		}
	}
	return envs, nil
}

func (service *Service) createEnvironment(ctx context.Context, spaceID int64, name string) (*EnvironmentResponse, error) {
	slug := slugify(name)
	if slug == "" {
		return nil, errors.Invalid("name must contain at least one alphanumeric character")
	}

	var maxPos struct{ Position int }
	service.orm.WithContext(ctx).
		Table("environments").
		Select("COALESCE(MAX(position), -1) as position").
		Where("space_id = ?", spaceID).
		Scan(&maxPos)

	record := &schemas.Environment{
		SpaceID:  spaceID,
		Name:     name,
		Slug:     slug,
		Position: maxPos.Position + 1,
	}
	if err := service.orm.WithContext(ctx).Create(record).Error; err != nil {
		if stderrors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, errors.Conflict("environment slug already exists in this space")
		}
		return nil, errors.Internal("failed to create environment", err)
	}

	return &EnvironmentResponse{
		ID:        record.ID,
		Name:      record.Name,
		Slug:      record.Slug,
		Position:  record.Position,
		CreatedAt: record.CreatedAt,
		UpdatedAt: record.UpdatedAt,
	}, nil
}

func (service *Service) updateEnvironment(ctx context.Context, spaceID int64, envSlug, name string) (*EnvironmentResponse, error) {
	var record schemas.Environment
	err := service.orm.WithContext(ctx).
		Where("space_id = ? AND slug = ?", spaceID, envSlug).
		First(&record).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.NotFound("environment not found")
	}
	if err != nil {
		return nil, errors.Internal("failed to find environment", err)
	}

	record.Name = name
	if err := service.orm.WithContext(ctx).Save(&record).Error; err != nil {
		return nil, errors.Internal("failed to update environment", err)
	}

	return &EnvironmentResponse{
		ID:        record.ID,
		Name:      record.Name,
		Slug:      record.Slug,
		Position:  record.Position,
		CreatedAt: record.CreatedAt,
		UpdatedAt: record.UpdatedAt,
	}, nil
}

func (service *Service) deleteEnvironment(ctx context.Context, spaceID int64, envSlug string) error {
	return service.orm.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var env schemas.Environment
		err := tx.Where("space_id = ? AND slug = ?", spaceID, envSlug).First(&env).Error
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return errors.NotFound("environment not found")
		}
		if err != nil {
			return errors.Internal("failed to find environment", err)
		}

		if err := tx.Where("environment_id = ?", env.ID).Delete(&schemas.Secret{}).Error; err != nil {
			return errors.Internal("failed to delete secrets", err)
		}

		if err := tx.Delete(&env).Error; err != nil {
			return errors.Internal("failed to delete environment", err)
		}

		return nil
	})
}

func (service *Service) resolveEnvironmentID(ctx context.Context, spaceID int64, envSlug string) (int64, error) {
	var record schemas.Environment
	err := service.orm.WithContext(ctx).
		Where("space_id = ? AND slug = ?", spaceID, envSlug).
		First(&record).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return 0, errors.NotFound("environment not found")
	}
	if err != nil {
		return 0, errors.Internal("failed to find environment", err)
	}
	return record.ID, nil
}

func (service *Service) ResolveEnvironmentID(ctx context.Context, spaceID int64, envSlug string) (int64, error) {
	return service.resolveEnvironmentID(ctx, spaceID, envSlug)
}

func slugify(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))
	s = slugRegexp.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}
