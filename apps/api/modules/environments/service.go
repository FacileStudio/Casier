package environments

import (
	"context"
	stderrors "errors"
	"regexp"
	"strings"

	"github.com/FacileStudio/Casier/apps/api/internal/errors"
	"github.com/FacileStudio/Casier/apps/api/schemas"

	"gorm.io/gorm"
)

var slugRegexp = regexp.MustCompile(`[^a-z0-9]+`)

type ProjectResolver interface {
	ResolveProjectID(ctx context.Context, slug string, userID int64, requiredRoles ...string) (int64, error)
}

type Service struct {
	orm             *gorm.DB
	projectResolver ProjectResolver
	controller      *Controller
}

func NewService(orm *gorm.DB, projectResolver ProjectResolver) *Service {
	service := &Service{orm: orm, projectResolver: projectResolver}
	service.controller = newController(service)
	return service
}

func (service *Service) listEnvironments(ctx context.Context, projectID int64) ([]EnvironmentResponse, error) {
	var records []schemas.Environment
	err := service.orm.WithContext(ctx).
		Where("project_id = ?", projectID).
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

func (service *Service) createEnvironment(ctx context.Context, projectID int64, name string) (*EnvironmentResponse, error) {
	slug := slugify(name)
	if slug == "" {
		return nil, errors.Invalid("name must contain at least one alphanumeric character")
	}

	var maxPos struct{ Position int }
	service.orm.WithContext(ctx).
		Table("environments").
		Select("COALESCE(MAX(position), -1) as position").
		Where("project_id = ?", projectID).
		Scan(&maxPos)

	record := &schemas.Environment{
		ProjectID: projectID,
		Name:      name,
		Slug:      slug,
		Position:  maxPos.Position + 1,
	}
	if err := service.orm.WithContext(ctx).Create(record).Error; err != nil {
		if stderrors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, errors.Conflict("environment slug already exists in this project")
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

func (service *Service) updateEnvironment(ctx context.Context, projectID int64, envSlug, name string) (*EnvironmentResponse, error) {
	var record schemas.Environment
	err := service.orm.WithContext(ctx).
		Where("project_id = ? AND slug = ?", projectID, envSlug).
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

func (service *Service) deleteEnvironment(ctx context.Context, projectID int64, envSlug string) error {
	return service.orm.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var env schemas.Environment
		err := tx.Where("project_id = ? AND slug = ?", projectID, envSlug).First(&env).Error
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

func (service *Service) resolveEnvironmentID(ctx context.Context, projectID int64, envSlug string) (int64, error) {
	var record schemas.Environment
	err := service.orm.WithContext(ctx).
		Where("project_id = ? AND slug = ?", projectID, envSlug).
		First(&record).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return 0, errors.NotFound("environment not found")
	}
	if err != nil {
		return 0, errors.Internal("failed to find environment", err)
	}
	return record.ID, nil
}

func (service *Service) ResolveEnvironmentID(ctx context.Context, projectID int64, envSlug string) (int64, error) {
	return service.resolveEnvironmentID(ctx, projectID, envSlug)
}

func slugify(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))
	s = slugRegexp.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}
