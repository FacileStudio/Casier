package environments

import (
	"context"
	stderrors "errors"
	"regexp"
	"strings"

	"github.com/FacileStudio/Casier/apps/api/internal/authcontext"
	"github.com/FacileStudio/Casier/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"

	"gorm.io/gorm"
)

var slugRegexp = regexp.MustCompile(`[^a-z0-9]+`)

// protectionRoles are the project roles allowed to flip an environment's
// protected flag, in either direction. A member who could unprotect production
// and then write to it would have bypassed the feature entirely, so this is the
// check the whole flag rests on.
var protectionRoles = []string{"owner", "admin"}

// protectedDeleteRoles gates deletion of a protected environment. Deleting one
// is lifting its protection and destroying its contents in a single call, so it
// asks for more than an admin.
var protectedDeleteRoles = []string{"owner"}

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
		envs[i] = newEnvironmentResponse(r)
	}
	return envs, nil
}

func (service *Service) createEnvironment(ctx context.Context, projectID int64, name string, protected bool) (*EnvironmentResponse, error) {
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
		Protected: protected,
	}
	if err := service.orm.WithContext(ctx).Create(record).Error; err != nil {
		if stderrors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, errors.Conflict("environment slug already exists in this project")
		}
		return nil, errors.Internal("failed to create environment", err)
	}

	response := newEnvironmentResponse(*record)
	return &response, nil
}

// updateEnvironment leaves the slug alone on purpose: the CLI, client config and
// every scoped token grant address an environment by slug, so a rename would
// silently strand them.
func (service *Service) updateEnvironment(ctx context.Context, projectID int64, envSlug, name string, protected *bool) (*EnvironmentResponse, error) {
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

	if err := authorizeProtectionChange(ctx, protected, record.Protected); err != nil {
		return nil, err
	}

	record.Name = name
	record.Protected = resolveProtection(protected, record.Protected)
	if err := service.orm.WithContext(ctx).Save(&record).Error; err != nil {
		return nil, errors.Internal("failed to update environment", err)
	}

	response := newEnvironmentResponse(record)
	return &response, nil
}

// deleteEnvironment takes callerIsOwner rather than a role string because the
// project resolver only answers with an id; the controller asks the membership
// question and this decides what to do with the answer, inside the same
// transaction that reads the protected flag.
func (service *Service) deleteEnvironment(ctx context.Context, projectID int64, envSlug string, callerIsOwner bool) error {
	return service.orm.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var env schemas.Environment
		err := tx.Where("project_id = ? AND slug = ?", projectID, envSlug).First(&env).Error
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return errors.NotFound("environment not found")
		}
		if err != nil {
			return errors.Internal("failed to find environment", err)
		}

		if err := authorizeProtectedDelete(ctx, env.Protected, callerIsOwner); err != nil {
			return err
		}

		var secretIDs []int64
		if err := tx.Model(&schemas.Secret{}).
			Where("environment_id = ?", env.ID).
			Pluck("id", &secretIDs).Error; err != nil {
			return errors.Internal("failed to list secrets", err)
		}

		if len(secretIDs) > 0 {
			if err := tx.Where("secret_id IN ?", secretIDs).Delete(&schemas.SecretVersion{}).Error; err != nil {
				return errors.Internal("failed to delete secret versions", err)
			}

			if err := tx.Where("secret_id IN ?", secretIDs).Delete(&schemas.SecretACL{}).Error; err != nil {
				return errors.Internal("failed to delete secret acls", err)
			}

			if err := tx.Where("id IN ?", secretIDs).Delete(&schemas.Secret{}).Error; err != nil {
				return errors.Internal("failed to delete secrets", err)
			}
		}

		if err := tx.Delete(&env).Error; err != nil {
			return errors.Internal("failed to delete environment", err)
		}

		return nil
	})
}

// resolveEnvironmentID is the choke point every secret read and write funnels
// through, so it is where an API token's environment scope is enforced. A token
// outside its scope gets the same not-found answer as a missing environment,
// otherwise the error alone would enumerate the project's environments.
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

	if !scopeAllowsEnvironment(ctx, record.Slug) {
		return 0, errors.NotFound("environment not found")
	}

	return record.ID, nil
}

// scopeAllowsEnvironment reports whether the calling API token may touch the
// environment. A user session carries no scope and is always allowed; project
// roles bound it instead.
func scopeAllowsEnvironment(ctx context.Context, envSlug string) bool {
	return authcontext.ScopeFromContext(ctx).AllowsEnvironment(envSlug)
}

// scopeAllowsWrite reports whether the calling API token carries write
// permission. A user session carries no scope and is always allowed.
func scopeAllowsWrite(ctx context.Context) bool {
	return authcontext.ScopeFromContext(ctx).AllowsWrite()
}

// callerIsMachineToken reports whether the request arrived on a `casier_…`
// token rather than a user session. Protection is a human decision: a CI
// credential holding write permission must not be able to lift the guard that
// exists to contain it, no matter how broad its scope.
func callerIsMachineToken(ctx context.Context) bool {
	return authcontext.ScopeFromContext(ctx) != nil
}

// protectionChanges reports whether the request actually flips the flag. An
// absent field, or one that restates the stored value, changes nothing and must
// not trip the protection checks.
func protectionChanges(requested *bool, current bool) bool {
	return requested != nil && *requested != current
}

// resolveProtection folds the optional request field onto the stored value. A
// nil pointer means the client never mentioned protection, which is the case an
// older client that PUTs only {"name": …} lands in.
func resolveProtection(requested *bool, current bool) bool {
	if requested == nil {
		return current
	}
	return *requested
}

func newEnvironmentResponse(record schemas.Environment) EnvironmentResponse {
	return EnvironmentResponse{
		ID:        record.ID,
		Name:      record.Name,
		Slug:      record.Slug,
		Position:  record.Position,
		Protected: record.Protected,
		CreatedAt: record.CreatedAt,
		UpdatedAt: record.UpdatedAt,
	}
}

func (service *Service) ResolveEnvironmentID(ctx context.Context, projectID int64, envSlug string) (int64, error) {
	return service.resolveEnvironmentID(ctx, projectID, envSlug)
}

func slugify(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))
	s = slugRegexp.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}
