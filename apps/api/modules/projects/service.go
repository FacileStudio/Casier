package projects

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	stderrors "errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/FacileStudio/Casier/apps/api/internal/audit"
	"github.com/FacileStudio/Casier/apps/api/internal/authcontext"
	"github.com/FacileStudio/Casier/apps/api/internal/crypto"
	"github.com/FacileStudio/Casier/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"

	"gorm.io/gorm"
)

// RotateKeyWarning is returned with every successful root key rotation: the new
// key only lives in this process until ENCRYPTION_KEY is updated too.
const RotateKeyWarning = "Update ENCRYPTION_KEY in the environment to the new key before restarting the API, otherwise it will start with the old key and be unable to decrypt any project data key."

var slugRegexp = regexp.MustCompile(`[^a-z0-9]+`)

// Service manages projects: creation, membership, invites, tokens, webhooks
// and the audit trail, under the shared keyring.
type Service struct {
	orm        *gorm.DB
	keyring    *crypto.Keyring
	audit      *audit.Logger
	controller *Controller
}

// NewService wires a project service onto the database.
func NewService(orm *gorm.DB, keyring *crypto.Keyring, auditLog *audit.Logger) *Service {
	service := &Service{orm: orm, keyring: keyring, audit: auditLog}
	service.controller = newController(service)
	return service
}

func (service *Service) listProjects(ctx context.Context, userID int64) ([]ProjectResponse, error) {
	var results []struct {
		schemas.Project
		Role string
	}

	query := service.orm.WithContext(ctx).
		Table("projects p").
		Select("p.*, pm.role").
		Joins("join project_members pm on pm.project_id = p.id").
		Where("pm.user_id = ?", userID)

	if scope := authcontext.ScopeFromContext(ctx); scope != nil {
		query = query.Where("p.id = ?", scope.ProjectID)
	}

	if err := query.Order("p.created_at desc").Scan(&results).Error; err != nil {
		return nil, errors.Internal("failed to list projects", err)
	}

	projects := make([]ProjectResponse, len(results))
	for i, r := range results {
		projects[i] = ProjectResponse{
			ID:           r.ID,
			Name:         r.Name,
			Slug:         r.Slug,
			Description:  r.Description,
			Role:         r.Role,
			RotationDays: r.RotationDays,
			CreatedAt:    r.CreatedAt,
			UpdatedAt:    r.UpdatedAt,
		}
	}
	return projects, nil
}

func (service *Service) createProject(ctx context.Context, userID int64, name, description string) (*ProjectResponse, error) {
	slug := slugify(name)
	if slug == "" {
		return nil, errors.Invalid("name must contain at least one alphanumeric character")
	}

	dataKey := make([]byte, 32)
	if _, err := rand.Read(dataKey); err != nil {
		return nil, errors.Internal("failed to generate data key", err)
	}

	encDataKey, nonce, err := crypto.Encrypt(dataKey, service.keyring.Key())
	if err != nil {
		return nil, errors.Internal("failed to encrypt data key", err)
	}

	project := &schemas.Project{
		Name:             name,
		Slug:             slug,
		Description:      description,
		EncryptedDataKey: encDataKey,
		DataKeyNonce:     nonce,
		RotationDays:     schemas.DefaultRotationDays,
	}

	err = service.orm.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(project).Error; err != nil {
			if stderrors.Is(err, gorm.ErrDuplicatedKey) {
				return errors.Conflict("project slug already exists")
			}
			return errors.Internal("failed to create project", err)
		}

		member := &schemas.ProjectMember{
			ProjectID: project.ID,
			UserID:    userID,
			Role:      "owner",
		}
		if err := tx.Create(member).Error; err != nil {
			return errors.Internal("failed to add owner", err)
		}

		defaultEnvs := []schemas.Environment{
			{ProjectID: project.ID, Name: "Development", Slug: "dev", Position: 0},
			{ProjectID: project.ID, Name: "Staging", Slug: "staging", Position: 1},
			{ProjectID: project.ID, Name: "Production", Slug: "prod", Position: 2},
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

	return &ProjectResponse{
		ID:           project.ID,
		Name:         project.Name,
		Slug:         project.Slug,
		Description:  project.Description,
		Role:         "owner",
		RotationDays: project.RotationDays,
		CreatedAt:    project.CreatedAt,
		UpdatedAt:    project.UpdatedAt,
	}, nil
}

func (service *Service) getProject(ctx context.Context, slug string, userID int64) (*ProjectResponse, error) {
	var result struct {
		schemas.Project
		Role string
	}

	err := service.orm.WithContext(ctx).
		Table("projects p").
		Select("p.*, pm.role").
		Joins("join project_members pm on pm.project_id = p.id").
		Where("p.slug = ? AND pm.user_id = ?", slug, userID).
		Scan(&result).Error
	if err != nil {
		return nil, errors.Internal("failed to get project", err)
	}
	if result.ID == 0 {
		return nil, errors.NotFound("project not found")
	}
	if err := ensureScopeAllowsProject(ctx, result.ID); err != nil {
		return nil, err
	}

	return &ProjectResponse{
		ID:           result.ID,
		Name:         result.Name,
		Slug:         result.Slug,
		Description:  result.Description,
		Role:         result.Role,
		RotationDays: result.RotationDays,
		CreatedAt:    result.CreatedAt,
		UpdatedAt:    result.UpdatedAt,
	}, nil
}

func (service *Service) updateProject(ctx context.Context, slug string, userID int64, name, description string, rotationDays *int) (*ProjectResponse, error) {
	project, role, err := service.getProjectWithRole(ctx, slug, userID)
	if err != nil {
		return nil, err
	}
	if role != "owner" && role != "admin" {
		return nil, errors.Forbidden("insufficient permissions")
	}

	project.Name = name
	project.Description = description
	if rotationDays != nil {
		project.RotationDays = *rotationDays
	}
	if err := service.orm.WithContext(ctx).Save(project).Error; err != nil {
		return nil, errors.Internal("failed to update project", err)
	}

	return &ProjectResponse{
		ID:           project.ID,
		Name:         project.Name,
		Slug:         project.Slug,
		Description:  project.Description,
		Role:         role,
		RotationDays: project.RotationDays,
		CreatedAt:    project.CreatedAt,
		UpdatedAt:    project.UpdatedAt,
	}, nil
}

func (service *Service) deleteProject(ctx context.Context, slug string, userID int64) error {
	_, role, err := service.getProjectWithRole(ctx, slug, userID)
	if err != nil {
		return err
	}
	if role != "owner" {
		return errors.Forbidden("only the owner can delete a project")
	}

	return service.orm.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var project schemas.Project
		if err := tx.Where("slug = ?", slug).First(&project).Error; err != nil {
			return errors.Internal("failed to find project for deletion", err)
		}

		var envIDs []int64
		tx.Model(&schemas.Environment{}).Where("project_id = ?", project.ID).Pluck("id", &envIDs)
		if len(envIDs) > 0 {
			if err := tx.Where("environment_id IN ?", envIDs).Delete(&schemas.Secret{}).Error; err != nil {
				return errors.Internal("failed to delete secrets", err)
			}
		}

		if err := tx.Where("project_id = ?", project.ID).Delete(&schemas.Environment{}).Error; err != nil {
			return errors.Internal("failed to delete environments", err)
		}

		if err := tx.Where("project_id = ?", project.ID).Delete(&schemas.ProjectMember{}).Error; err != nil {
			return errors.Internal("failed to delete members", err)
		}

		if err := tx.Delete(&project).Error; err != nil {
			return errors.Internal("failed to delete project", err)
		}

		return nil
	})
}

func (service *Service) listMembers(ctx context.Context, slug string, userID int64) ([]MemberResponse, error) {
	_, _, err := service.getProjectWithRole(ctx, slug, userID)
	if err != nil {
		return nil, err
	}

	var results []struct {
		ID    int64
		Email string
		Role  string
	}

	err = service.orm.WithContext(ctx).
		Table("project_members pm").
		Select("u.id, u.email, pm.role").
		Joins("join users u on u.id = pm.user_id").
		Joins("join projects p on p.id = pm.project_id").
		Where("p.slug = ?", slug).
		Order("pm.created_at asc").
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
	_, callerRole, err := service.getProjectWithRole(ctx, slug, callerID)
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

	var project schemas.Project
	if err := service.orm.WithContext(ctx).Where("slug = ?", slug).First(&project).Error; err != nil {
		return errors.Internal("failed to find project", err)
	}

	member := &schemas.ProjectMember{
		ProjectID: project.ID,
		UserID:    user.ID,
		Role:      role,
	}
	if err := service.orm.WithContext(ctx).Create(member).Error; err != nil {
		if stderrors.Is(err, gorm.ErrDuplicatedKey) {
			return errors.Conflict("user is already a member")
		}
		return errors.Internal("failed to add member", err)
	}
	return nil
}

func (service *Service) removeMember(ctx context.Context, slug string, callerID int64, targetUserID int64, callerEmail, clientIP string) error {
	_, callerRole, err := service.getProjectWithRole(ctx, slug, callerID)
	if err != nil {
		return err
	}
	if callerRole != "owner" && callerRole != "admin" {
		return errors.Forbidden("insufficient permissions")
	}

	var project schemas.Project
	if err := service.orm.WithContext(ctx).Where("slug = ?", slug).First(&project).Error; err != nil {
		return errors.Internal("failed to find project", err)
	}

	var target schemas.ProjectMember
	err = service.orm.WithContext(ctx).
		Where("project_id = ? AND user_id = ?", project.ID, targetUserID).
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
			Model(&schemas.ProjectMember{}).
			Where("project_id = ? AND role = ?", project.ID, "owner").
			Count(&ownerCount)
		if ownerCount <= 1 {
			return errors.Forbidden("cannot remove the last owner")
		}
	}

	if err := service.orm.WithContext(ctx).Delete(&target).Error; err != nil {
		return errors.Internal("failed to remove member", err)
	}

	if service.audit != nil {
		service.audit.Log(ctx, project.ID, strconv.FormatInt(callerID, 10), callerEmail, "member.removed", "user:"+strconv.FormatInt(targetUserID, 10), "", clientIP)
	}

	return nil
}

func (service *Service) createInvite(ctx context.Context, slug string, callerID int64, role string, maxUses int) (*InviteResponse, error) {
	_, callerRole, err := service.getProjectWithRole(ctx, slug, callerID)
	if err != nil {
		return nil, err
	}
	if callerRole != "owner" && callerRole != "admin" {
		return nil, errors.Forbidden("insufficient permissions")
	}

	var project schemas.Project
	if err := service.orm.WithContext(ctx).Where("slug = ?", slug).First(&project).Error; err != nil {
		return nil, errors.Internal("failed to find project", err)
	}

	tokenBytes := make([]byte, 24)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, errors.Internal("failed to generate invite token", err)
	}
	token := base64.RawURLEncoding.EncodeToString(tokenBytes)

	invite := &schemas.Invite{
		Token:     token,
		ProjectID: project.ID,
		Role:      role,
		MaxUses:   maxUses,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		CreatedBy: callerID,
	}
	if err := service.orm.WithContext(ctx).Create(invite).Error; err != nil {
		return nil, errors.Internal("failed to create invite", err)
	}

	return &InviteResponse{
		ID:        invite.ID,
		Token:     invite.Token,
		Role:      invite.Role,
		MaxUses:   invite.MaxUses,
		UseCount:  0,
		ExpiresAt: invite.ExpiresAt,
		CreatedAt: invite.CreatedAt,
	}, nil
}

func (service *Service) listInvites(ctx context.Context, slug string, callerID int64) ([]InviteResponse, error) {
	_, callerRole, err := service.getProjectWithRole(ctx, slug, callerID)
	if err != nil {
		return nil, err
	}
	if callerRole != "owner" && callerRole != "admin" {
		return nil, errors.Forbidden("insufficient permissions")
	}

	var project schemas.Project
	if err := service.orm.WithContext(ctx).Where("slug = ?", slug).First(&project).Error; err != nil {
		return nil, errors.Internal("failed to find project", err)
	}

	var invites []schemas.Invite
	err = service.orm.WithContext(ctx).
		Where("project_id = ? AND expires_at > ?", project.ID, time.Now()).
		Order("created_at desc").
		Find(&invites).Error
	if err != nil {
		return nil, errors.Internal("failed to list invites", err)
	}

	result := make([]InviteResponse, len(invites))
	for i, inv := range invites {
		result[i] = InviteResponse{
			ID:        inv.ID,
			Token:     inv.Token,
			Role:      inv.Role,
			MaxUses:   inv.MaxUses,
			UseCount:  inv.UseCount,
			ExpiresAt: inv.ExpiresAt,
			CreatedAt: inv.CreatedAt,
		}
	}
	return result, nil
}

func (service *Service) acceptInvite(ctx context.Context, token string, userID int64) error {
	var invite schemas.Invite
	err := service.orm.WithContext(ctx).Where("token = ?", token).First(&invite).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return errors.NotFound("invalid invite link")
	}
	if err != nil {
		return errors.Internal("failed to find invite", err)
	}

	if time.Now().After(invite.ExpiresAt) {
		return errors.Invalid("invite link has expired")
	}
	if invite.MaxUses > 0 && invite.UseCount >= invite.MaxUses {
		return errors.Invalid("invite link has reached max uses")
	}

	return service.orm.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		claim := tx.Model(&schemas.Invite{}).
			Where("id = ? AND (max_uses = 0 OR use_count < max_uses)", invite.ID).
			Update("use_count", gorm.Expr("use_count + 1"))
		if claim.Error != nil {
			return errors.Internal("failed to claim invite", claim.Error)
		}
		if claim.RowsAffected == 0 {
			return errors.Invalid("invite link has reached max uses")
		}

		member := &schemas.ProjectMember{
			ProjectID: invite.ProjectID,
			UserID:    userID,
			Role:      invite.Role,
		}
		if err := tx.Create(member).Error; err != nil {
			if stderrors.Is(err, gorm.ErrDuplicatedKey) {
				return errors.Conflict("you are already a member of this project")
			}
			return errors.Internal("failed to add member", err)
		}
		return nil
	})
}

func (service *Service) deleteInvite(ctx context.Context, slug string, callerID int64, inviteID int64) error {
	_, callerRole, err := service.getProjectWithRole(ctx, slug, callerID)
	if err != nil {
		return err
	}
	if callerRole != "owner" && callerRole != "admin" {
		return errors.Forbidden("insufficient permissions")
	}

	var project schemas.Project
	if err := service.orm.WithContext(ctx).Where("slug = ?", slug).First(&project).Error; err != nil {
		return errors.Internal("failed to find project", err)
	}

	result := service.orm.WithContext(ctx).
		Where("id = ? AND project_id = ?", inviteID, project.ID).
		Delete(&schemas.Invite{})
	if result.Error != nil {
		return errors.Internal("failed to delete invite", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.NotFound("invite not found")
	}
	return nil
}

func (service *Service) getProjectWithRole(ctx context.Context, slug string, userID int64) (*schemas.Project, string, error) {
	var project schemas.Project
	if err := service.orm.WithContext(ctx).Where("slug = ?", slug).First(&project).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", errors.NotFound("project not found")
		}
		return nil, "", errors.Internal("failed to find project", err)
	}

	if err := ensureScopeAllowsProject(ctx, project.ID); err != nil {
		return nil, "", err
	}

	var member schemas.ProjectMember
	err := service.orm.WithContext(ctx).
		Where("project_id = ? AND user_id = ?", project.ID, userID).
		First(&member).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return nil, "", errors.NotFound("project not found")
	}
	if err != nil {
		return nil, "", errors.Internal("failed to check membership", err)
	}

	return &project, member.Role, nil
}

func (service *Service) resolveProjectID(ctx context.Context, slug string, userID int64, requiredRoles ...string) (int64, error) {
	project, role, err := service.getProjectWithRole(ctx, slug, userID)
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

	return project.ID, nil
}

func (service *Service) ResolveProjectID(ctx context.Context, slug string, userID int64, requiredRoles ...string) (int64, error) {
	return service.resolveProjectID(ctx, slug, userID, requiredRoles...)
}

func (service *Service) GetProjectDataKey(ctx context.Context, projectID int64) ([]byte, error) {
	var project schemas.Project
	if err := service.orm.WithContext(ctx).
		Select("encrypted_data_key, data_key_nonce").
		Where("id = ?", projectID).
		First(&project).Error; err != nil {
		return nil, fmt.Errorf("failed to load project data key: %w", err)
	}

	if len(project.EncryptedDataKey) == 0 {
		return nil, fmt.Errorf("project %d has no data key", projectID)
	}

	return crypto.Decrypt(project.EncryptedDataKey, project.DataKeyNonce, service.keyring.Key())
}

func (service *Service) createApiToken(ctx context.Context, slug string, callerID int64, name, environments, permissions string, expiresInDays *int) (*ApiTokenResponse, error) {
	_, callerRole, err := service.getProjectWithRole(ctx, slug, callerID)
	if err != nil {
		return nil, err
	}
	if callerRole != "owner" && callerRole != "admin" {
		return nil, errors.Forbidden("insufficient permissions")
	}

	var project schemas.Project
	if err := service.orm.WithContext(ctx).Where("slug = ?", slug).First(&project).Error; err != nil {
		return nil, errors.Internal("failed to find project", err)
	}

	lifetime, err := tokenLifetime(expiresInDays)
	if err != nil {
		return nil, err
	}

	requested := parseRequestedEnvironments(environments)
	scoped, err := service.resolveScopeEnvironments(ctx, project.ID, requested)
	if err != nil {
		return nil, err
	}

	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, errors.Internal("failed to generate token", err)
	}
	rawToken := "casier_" + base64.RawURLEncoding.EncodeToString(tokenBytes)
	sum := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(sum[:])

	if permissions == "" {
		permissions = "read"
	}

	apiToken := &schemas.ApiToken{
		TokenHash:       tokenHash,
		Name:            name,
		ProjectID:       project.ID,
		AllEnvironments: len(requested) == 0,
		Permissions:     permissions,
		CreatedBy:       callerID,
		ExpiresAt:       time.Now().Add(lifetime),
	}

	err = service.orm.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(apiToken).Error; err != nil {
			return errors.Internal("failed to create API token", err)
		}
		for _, environment := range scoped {
			grant := &schemas.ApiTokenEnvironment{TokenID: apiToken.ID, EnvironmentID: environment.ID}
			if err := tx.Create(grant).Error; err != nil {
				return errors.Internal("failed to bind token scope", err)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &ApiTokenResponse{
		ID:           apiToken.ID,
		Token:        rawToken,
		Name:         apiToken.Name,
		Environments: strings.Join(environmentSlugs(scoped), ","),
		Permissions:  apiToken.Permissions,
		ExpiresAt:    apiToken.ExpiresAt,
		CreatedAt:    apiToken.CreatedAt,
	}, nil
}

// tokenLifetime refuses a token that never expires. A machine credential whose
// whole purpose is to sit in a CI runner is the one that most needs a horizon.
func tokenLifetime(expiresInDays *int) (time.Duration, error) {
	days := schemas.DefaultTokenDays
	if expiresInDays != nil {
		days = *expiresInDays
	}
	if days < 1 || days > schemas.MaxTokenDays {
		return 0, errors.Invalid(fmt.Sprintf("expires_in_days must be between 1 and %d", schemas.MaxTokenDays))
	}
	return time.Duration(days) * 24 * time.Hour, nil
}

// parseRequestedEnvironments reads the scope a caller asked for. Both an empty
// list and a lone "*" mean every environment — the dashboard has always offered
// "*" as the affordance for it, and refusing the string its own placeholder
// suggests would be a strange way to enforce a security boundary. Neither is an
// inference: whichever the caller sends, what gets stored is the explicit
// all_environments flag.
func parseRequestedEnvironments(raw string) []string {
	requested := authcontext.ParseScopeList(raw)
	if len(requested) == 1 && requested[0] == "*" {
		return nil
	}
	return requested
}

// resolveScopeEnvironments turns the requested slugs into environment rows, and
// refuses a slug the project does not have. Silently dropping it would mint a
// token narrower than the caller asked for, which reads as a broken deploy long
// after anyone remembers creating it.
func (service *Service) resolveScopeEnvironments(ctx context.Context, projectID int64, slugs []string) ([]schemas.Environment, error) {
	if len(slugs) == 0 {
		return nil, nil
	}

	var found []schemas.Environment
	if err := service.orm.WithContext(ctx).
		Where("project_id = ? AND slug IN ?", projectID, slugs).
		Find(&found).Error; err != nil {
		return nil, errors.Internal("failed to resolve environments", err)
	}
	if len(found) != len(slugs) {
		known := make(map[string]bool, len(found))
		for _, environment := range found {
			known[environment.Slug] = true
		}
		for _, slug := range slugs {
			if !known[slug] {
				return nil, errors.Invalid(fmt.Sprintf("unknown environment %q", slug))
			}
		}
	}
	return found, nil
}

func environmentSlugs(environments []schemas.Environment) []string {
	slugs := make([]string, len(environments))
	for i, environment := range environments {
		slugs[i] = environment.Slug
	}
	return slugs
}

func (service *Service) listApiTokens(ctx context.Context, slug string, callerID int64) ([]ApiTokenResponse, error) {
	_, callerRole, err := service.getProjectWithRole(ctx, slug, callerID)
	if err != nil {
		return nil, err
	}
	if callerRole != "owner" && callerRole != "admin" {
		return nil, errors.Forbidden("insufficient permissions")
	}

	var project schemas.Project
	if err := service.orm.WithContext(ctx).Where("slug = ?", slug).First(&project).Error; err != nil {
		return nil, errors.Internal("failed to find project", err)
	}

	var tokens []schemas.ApiToken
	err = service.orm.WithContext(ctx).
		Where("project_id = ?", project.ID).
		Order("created_at desc").
		Find(&tokens).Error
	if err != nil {
		return nil, errors.Internal("failed to list tokens", err)
	}

	var grants []struct {
		TokenID int64
		Slug    string
	}
	if err := service.orm.WithContext(ctx).
		Table("api_token_environments te").
		Select("te.token_id, e.slug").
		Joins("join environments e on e.id = te.environment_id").
		Joins("join api_tokens t on t.id = te.token_id").
		Where("t.project_id = ?", project.ID).
		Scan(&grants).Error; err != nil {
		return nil, errors.Internal("failed to read token scopes", err)
	}
	scopes := make(map[int64][]string, len(grants))
	for _, grant := range grants {
		scopes[grant.TokenID] = append(scopes[grant.TokenID], grant.Slug)
	}

	result := make([]ApiTokenResponse, len(tokens))
	for i, t := range tokens {
		result[i] = ApiTokenResponse{
			ID:           t.ID,
			Name:         t.Name,
			Environments: strings.Join(scopes[t.ID], ","),
			Permissions:  t.Permissions,
			ExpiresAt:    t.ExpiresAt,
			LastUsedAt:   t.LastUsedAt,
			RevokedAt:    t.RevokedAt,
			CreatedAt:    t.CreatedAt,
		}
	}
	return result, nil
}

// revokeApiToken revokes a token rather than deleting it: audit rows name the
// token that made a change, and deleting the row would turn every one of them
// into an anonymous event.
func (service *Service) revokeApiToken(ctx context.Context, slug string, callerID int64, tokenID int64) error {
	_, callerRole, err := service.getProjectWithRole(ctx, slug, callerID)
	if err != nil {
		return err
	}
	if callerRole != "owner" && callerRole != "admin" {
		return errors.Forbidden("insufficient permissions")
	}

	var project schemas.Project
	if err := service.orm.WithContext(ctx).Where("slug = ?", slug).First(&project).Error; err != nil {
		return errors.Internal("failed to find project", err)
	}

	result := service.orm.WithContext(ctx).
		Model(&schemas.ApiToken{}).
		Where("id = ? AND project_id = ? AND revoked_at IS NULL", tokenID, project.ID).
		Update("revoked_at", time.Now())
	if result.Error != nil {
		return errors.Internal("failed to revoke token", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.NotFound("token not found")
	}
	return nil
}

func (service *Service) createWebhook(ctx context.Context, slug string, callerID int64, url, events, secret string) (*WebhookResponse, error) {
	_, callerRole, err := service.getProjectWithRole(ctx, slug, callerID)
	if err != nil {
		return nil, err
	}
	if callerRole != "owner" && callerRole != "admin" {
		return nil, errors.Forbidden("insufficient permissions")
	}

	var project schemas.Project
	if err := service.orm.WithContext(ctx).Where("slug = ?", slug).First(&project).Error; err != nil {
		return nil, errors.Internal("failed to find project", err)
	}

	wh := &schemas.Webhook{
		ProjectID: project.ID,
		URL:       url,
		Events:    events,
		Secret:    secret,
		Active:    true,
		CreatedBy: callerID,
	}
	if err := service.orm.WithContext(ctx).Create(wh).Error; err != nil {
		return nil, errors.Internal("failed to create webhook", err)
	}

	return &WebhookResponse{
		ID:        wh.ID,
		URL:       wh.URL,
		Events:    wh.Events,
		Active:    wh.Active,
		CreatedAt: wh.CreatedAt,
	}, nil
}

func (service *Service) listWebhooks(ctx context.Context, slug string, callerID int64) ([]WebhookResponse, error) {
	_, callerRole, err := service.getProjectWithRole(ctx, slug, callerID)
	if err != nil {
		return nil, err
	}
	if callerRole != "owner" && callerRole != "admin" {
		return nil, errors.Forbidden("insufficient permissions")
	}

	var project schemas.Project
	if err := service.orm.WithContext(ctx).Where("slug = ?", slug).First(&project).Error; err != nil {
		return nil, errors.Internal("failed to find project", err)
	}

	var webhooks []schemas.Webhook
	err = service.orm.WithContext(ctx).
		Where("project_id = ?", project.ID).
		Order("created_at desc").
		Find(&webhooks).Error
	if err != nil {
		return nil, errors.Internal("failed to list webhooks", err)
	}

	result := make([]WebhookResponse, len(webhooks))
	for i, wh := range webhooks {
		result[i] = WebhookResponse{
			ID:        wh.ID,
			URL:       wh.URL,
			Events:    wh.Events,
			Active:    wh.Active,
			CreatedAt: wh.CreatedAt,
		}
	}
	return result, nil
}

func (service *Service) deleteWebhook(ctx context.Context, slug string, callerID int64, webhookID int64) error {
	_, callerRole, err := service.getProjectWithRole(ctx, slug, callerID)
	if err != nil {
		return err
	}
	if callerRole != "owner" && callerRole != "admin" {
		return errors.Forbidden("insufficient permissions")
	}

	var project schemas.Project
	if err := service.orm.WithContext(ctx).Where("slug = ?", slug).First(&project).Error; err != nil {
		return errors.Internal("failed to find project", err)
	}

	result := service.orm.WithContext(ctx).
		Where("id = ? AND project_id = ?", webhookID, project.ID).
		Delete(&schemas.Webhook{})
	if result.Error != nil {
		return errors.Internal("failed to delete webhook", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.NotFound("webhook not found")
	}
	return nil
}

func (service *Service) listAuditLogs(ctx context.Context, slug string, callerID int64, action string, limit int) ([]AuditLogEntry, error) {
	_, _, err := service.getProjectWithRole(ctx, slug, callerID)
	if err != nil {
		return nil, err
	}

	var project schemas.Project
	if err := service.orm.WithContext(ctx).Where("slug = ?", slug).First(&project).Error; err != nil {
		return nil, errors.Internal("failed to find project", err)
	}

	if limit <= 0 || limit > 100 {
		limit = 50
	}

	query := service.orm.WithContext(ctx).
		Where("project_id = ?", project.ID)
	if action != "" {
		query = query.Where("action = ?", action)
	}

	var logs []schemas.AuditLog
	err = query.
		Order("created_at desc").
		Limit(limit).
		Find(&logs).Error
	if err != nil {
		return nil, errors.Internal("failed to list audit logs", err)
	}

	result := make([]AuditLogEntry, len(logs))
	for i, l := range logs {
		result[i] = AuditLogEntry{
			ID:        l.ID,
			UserEmail: l.UserEmail,
			Action:    l.Action,
			Resource:  l.Resource,
			Detail:    l.Detail,
			ClientIP:  l.ClientIP,
			CreatedAt: l.CreatedAt,
		}
	}
	return result, nil
}

// RotateRootKey re-encrypts every project data key under newKey inside a single
// transaction, so a failure part way through cannot leave the instance with a
// mix of keys. The in-memory keyring is only swapped once the commit succeeds.
func (service *Service) RotateRootKey(ctx context.Context, newKey []byte, callerID, callerEmail, clientIP string) (int, error) {
	currentKey := service.keyring.Key()
	if bytes.Equal(newKey, currentKey) {
		return 0, errors.Invalid("new key is identical to the current key")
	}

	count := 0
	err := service.orm.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		count = 0

		var projects []schemas.Project
		if err := tx.Find(&projects).Error; err != nil {
			return errors.Internal("failed to list projects", err)
		}

		for _, project := range projects {
			if len(project.EncryptedDataKey) == 0 {
				continue
			}

			dataKey, err := crypto.Decrypt(project.EncryptedDataKey, project.DataKeyNonce, currentKey)
			if err != nil {
				return errors.Internal(fmt.Sprintf("failed to decrypt data key for project %d", project.ID), err)
			}

			newEncDataKey, newNonce, err := crypto.Encrypt(dataKey, newKey)
			if err != nil {
				return errors.Internal(fmt.Sprintf("failed to re-encrypt data key for project %d", project.ID), err)
			}

			if err := tx.Model(&schemas.Project{}).
				Where("id = ?", project.ID).
				Updates(map[string]any{
					"encrypted_data_key": newEncDataKey,
					"data_key_nonce":     newNonce,
				}).Error; err != nil {
				return errors.Internal(fmt.Sprintf("failed to save re-encrypted data key for project %d", project.ID), err)
			}
			count++
		}

		return nil
	})
	if err != nil {
		return 0, err
	}

	service.keyring.Replace(newKey)

	if service.audit != nil {
		service.audit.Log(ctx, 0, callerID, callerEmail, "instance.root_key_rotated", "encryption_key", fmt.Sprintf("%d project data keys re-encrypted", count), clientIP)
	}

	return count, nil
}

// ensureScopeAllowsProject stops an API token minted for one project from
// reaching another. It answers "not found" rather than "forbidden" so a token
// cannot use the endpoint to enumerate which projects exist.
func ensureScopeAllowsProject(ctx context.Context, projectID int64) error {
	if authcontext.ScopeFromContext(ctx).AllowsProject(projectID) {
		return nil
	}
	return errors.NotFound("project not found")
}

// isAdminEmail reports whether email is one of the instance admins, compared
// case-insensitively after trimming. An empty list denies everyone: instance
// wide operations fail closed when no admin has been configured.
func isAdminEmail(adminEmails []string, email string) bool {
	candidate := strings.ToLower(strings.TrimSpace(email))
	if candidate == "" {
		return false
	}

	for _, admin := range adminEmails {
		if strings.ToLower(strings.TrimSpace(admin)) == candidate {
			return true
		}
	}
	return false
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
