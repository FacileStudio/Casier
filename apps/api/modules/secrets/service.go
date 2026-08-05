package secrets

import (
	"context"
	stderrors "errors"
	"log/slog"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/FacileStudio/Casier/apps/api/internal/audit"
	"github.com/FacileStudio/Casier/apps/api/internal/authcontext"
	"github.com/FacileStudio/Casier/apps/api/internal/crypto"
	"github.com/FacileStudio/Casier/apps/api/internal/webhook"
	"github.com/FacileStudio/Casier/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"

	"gorm.io/gorm"
)

func joinTags(tags []string) string {
	if len(tags) == 0 {
		return ""
	}
	return strings.Join(tags, ",")
}

var refPattern = regexp.MustCompile(`\$\{ref:([^/]+)/([^}]+)\}`)

// resolveReferences expands ${ref:env/KEY} against the project's other
// environments. The caller's scope is honoured here: an environment-scoped API
// token must not be able to read a environment it was denied by writing a
// reference into one it was granted, so a denied reference is left literal,
// exactly like every other resolution failure.
func (service *Service) resolveReferences(ctx context.Context, projectID int64, value string, dataKey []byte, scope *authcontext.TokenScope, depth int) string {
	if depth > 3 || !strings.Contains(value, "${ref:") {
		return value
	}

	return refPattern.ReplaceAllStringFunc(value, func(match string) string {
		parts := refPattern.FindStringSubmatch(match)
		if len(parts) != 3 {
			return match
		}
		envSlug, secretKey := parts[1], parts[2]

		if !scope.AllowsEnvironment(envSlug) {
			return match
		}

		envID, err := service.envRes.ResolveEnvironmentID(ctx, projectID, envSlug)
		if err != nil {
			return match
		}

		var record schemas.Secret
		err = service.orm.WithContext(ctx).
			Where("environment_id = ? AND key = ?", envID, secretKey).
			First(&record).Error
		if err != nil {
			return match
		}

		plaintext, err := crypto.Decrypt(record.EncryptedValue, record.Nonce, dataKey)
		if err != nil {
			return match
		}

		return service.resolveReferences(ctx, projectID, string(plaintext), dataKey, scope, depth+1)
	})
}

func splitTags(s string) []string {
	if s == "" {
		return []string{}
	}
	return strings.Split(s, ",")
}

const (
	roleOwner = "owner"
	roleAdmin = "admin"
)

const maxSecretKeyLength = 256

const maxSecretValueSize = 64 * 1024

// The audit log distinguishes a page load from a data exfiltration. Listing
// metadata, revealing a page of values, revealing one key and exporting the
// whole environment used to collapse into a single "bulk_read" row, which made
// the log useless for the one question it exists to answer.
const (
	auditActionList       = "secrets.list"
	auditActionRevealList = "secrets.reveal_list"
	auditActionReveal     = "secrets.reveal"
	auditActionExport     = "secrets.export"
	auditActionDiff       = "secrets.diff"
	auditActionTrashList  = "secrets.trash_list"
	auditActionPurge      = "secrets.purge"
)

// roleMayWriteProtected answers whether a project role may write to an
// environment. A protected environment narrows writes to owner and admin;
// anywhere else the project role the caller already passed governs alone.
func roleMayWriteProtected(role string, protected bool) bool {
	if !protected {
		return true
	}
	return role == roleOwner || role == roleAdmin
}

// errDuplicateKey marks the unique-index violation on (environment_id, key) so
// the transaction can unwind before the collision is diagnosed. Postgres aborts
// a transaction on a failed statement, so the follow-up query that tells a live
// key from a trashed one has to run once the transaction is gone.
var errDuplicateKey = stderrors.New("duplicate secret key")

// caller is who a secret operation runs as: the resolved user, their role in
// the project, and the scope of the API token they authenticated with, if any.
type caller struct {
	userID   int64
	role     string
	clientIP string
	scope    *authcontext.TokenScope
}

func (person caller) logID() string {
	return strconv.FormatInt(person.userID, 10)
}

// secretGrant folds the secret_acls rows of one secret into a decision for one
// caller. A secret with no rows at all is unrestricted and governed by the
// project role alone.
type secretGrant struct {
	restricted bool
	canRead    bool
	canWrite   bool
}

// aclAllowsRead answers whether the caller may see a secret's value. An owner
// bypasses ACLs so a project can never be locked away from the person who owns
// it.
func aclAllowsRead(role string, grant secretGrant) bool {
	if role == roleOwner || !grant.restricted {
		return true
	}
	return grant.canRead
}

// aclAllowsWrite answers whether the caller may change a secret. The project
// role is checked before this, so this only narrows what a member may touch.
func aclAllowsWrite(role string, grant secretGrant) bool {
	if role == roleOwner || !grant.restricted {
		return true
	}
	return grant.canWrite
}

// secretGrants reads every ACL that applies to the given secrets in one query,
// folded server-side so the result holds one row per restricted secret rather
// than one row per grantee.
func (service *Service) secretGrants(ctx context.Context, orm *gorm.DB, secretIDs []int64, userID int64) (map[int64]secretGrant, error) {
	grants := make(map[int64]secretGrant)
	if len(secretIDs) == 0 {
		return grants, nil
	}

	var rows []struct {
		SecretID int64
		CanRead  bool
		CanWrite bool
	}
	err := orm.WithContext(ctx).
		Table("secret_acls").
		Select("secret_id, bool_or(user_id = ? and can_read) as can_read, bool_or(user_id = ? and can_write) as can_write", userID, userID).
		Where("secret_id IN ?", secretIDs).
		Group("secret_id").
		Scan(&rows).Error
	if err != nil {
		return nil, errors.Internal("failed to load secret ACLs", err)
	}

	for _, row := range rows {
		grants[row.SecretID] = secretGrant{restricted: true, canRead: row.CanRead, canWrite: row.CanWrite}
	}
	return grants, nil
}

func (service *Service) assertSecretWrite(ctx context.Context, secretID int64, person caller) error {
	grants, err := service.secretGrants(ctx, service.orm, []int64{secretID}, person.userID)
	if err != nil {
		return err
	}
	if !aclAllowsWrite(person.role, grants[secretID]) {
		return errors.Forbidden("insufficient permissions for this secret")
	}
	return nil
}

var likeEscaper = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)

// tagFilterPattern turns a user-supplied tag into a LIKE pattern that matches a
// whole tag and nothing else: wildcards in the filter are neutralised, and the
// comma padding stops "prod" from matching "not-production". It must be used
// with an explicit ESCAPE '\' clause.
func tagFilterPattern(tag string) string {
	return "%," + likeEscaper.Replace(tag) + ",%"
}

const tagFilterClause = `',' || tags || ',' LIKE ? ESCAPE '\'`

type ProjectResolver interface {
	ResolveProjectID(ctx context.Context, slug string, userID int64, requiredRoles ...string) (int64, error)
}

type EnvironmentResolver interface {
	ResolveEnvironmentID(ctx context.Context, projectID int64, envSlug string) (int64, error)
}

type DataKeyResolver interface {
	GetProjectDataKey(ctx context.Context, projectID int64) ([]byte, error)
}

type Service struct {
	orm        *gorm.DB
	logger     *slog.Logger
	projectRes ProjectResolver
	envRes     EnvironmentResolver
	dataKeyRes DataKeyResolver
	webhooks   *webhook.Dispatcher
	audit      *audit.Logger
	controller *Controller
}

func NewService(orm *gorm.DB, logger *slog.Logger, projectRes ProjectResolver, envRes EnvironmentResolver, dataKeyRes DataKeyResolver, webhooks *webhook.Dispatcher, auditLog *audit.Logger) *Service {
	service := &Service{orm: orm, logger: logger, projectRes: projectRes, envRes: envRes, dataKeyRes: dataKeyRes, webhooks: webhooks, audit: auditLog}
	service.controller = newController(service)
	return service
}

func (service *Service) fireWebhook(ctx context.Context, eventType string, envID int64, key string, version int, person caller) {
	projectID, envSlug, err := service.environmentRef(ctx, envID)
	if err != nil {
		return
	}
	service.emitSecretEvent(ctx, projectID, envSlug, eventType, key, version, person)
}

// emitSecretEvent records one secret mutation in the audit log and delivers it
// to the project's webhooks. It takes the project and environment already
// resolved so a bulk operation pays for that lookup once.
func (service *Service) emitSecretEvent(ctx context.Context, projectID int64, envSlug, eventType, key string, version int, person caller) {
	if service.audit != nil {
		service.audit.Log(ctx, projectID, person.logID(), "", eventType, "secret:"+key, "", person.clientIP)
	}
	if service.webhooks == nil {
		return
	}
	service.webhooks.Fire(ctx, webhook.Event{
		Type:        eventType,
		ProjectID:   projectID,
		Environment: envSlug,
		Key:         key,
		Version:     version,
		UserID:      person.logID(),
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
	})
}

// environmentRef resolves the project an environment belongs to and its slug,
// which audit entries and webhook payloads both name.
func (service *Service) environmentRef(ctx context.Context, envID int64) (int64, string, error) {
	var env schemas.Environment
	if err := service.orm.WithContext(ctx).Where("id = ?", envID).First(&env).Error; err != nil {
		return 0, "", errors.Internal("failed to resolve environment", err)
	}
	return env.ProjectID, env.Slug, nil
}

func (service *Service) getProjectIDFromEnv(ctx context.Context, envID int64) (int64, error) {
	projectID, _, err := service.environmentRef(ctx, envID)
	return projectID, err
}

// projectRole reads the caller's role in a project. The project resolver only
// checks roles, and ACL enforcement needs the role itself to let an owner
// through.
func (service *Service) projectRole(ctx context.Context, projectID, userID int64) string {
	var role string
	err := service.orm.WithContext(ctx).
		Model(&schemas.ProjectMember{}).
		Where("project_id = ? AND user_id = ?", projectID, userID).
		Pluck("role", &role).Error
	if err != nil {
		return ""
	}
	return role
}

func (service *Service) resolveDataKey(ctx context.Context, envID int64) ([]byte, error) {
	projectID, err := service.getProjectIDFromEnv(ctx, envID)
	if err != nil {
		return nil, err
	}
	dataKey, err := service.dataKeyRes.GetProjectDataKey(ctx, projectID)
	if err != nil {
		return nil, errors.Internal("failed to resolve data key", err)
	}
	return dataKey, nil
}

// secretRead describes one read of an environment: how it is filtered, whether
// plaintext leaves the server, and which audit action it is recorded under. The
// audit action is part of the request rather than inferred, because export and
// diff read exactly as much as a revealing list does and must not be logged as
// one.
type secretRead struct {
	tagFilter   string
	reveal      bool
	auditAction string
}

// listSecrets returns the secrets of an environment the caller is allowed to
// read. A secret carrying ACL rows is dropped entirely unless the caller holds
// a can_read grant on it, which is also what hides it from export and diff.
// Values are decrypted only when the read asks for them: a metadata listing
// never touches the data key, so a dashboard page load cannot leak plaintext it
// only intended to mask.
func (service *Service) listSecrets(ctx context.Context, envID int64, person caller, read secretRead) ([]SecretResponse, int, error) {
	query := service.orm.WithContext(ctx).Where("environment_id = ?", envID)
	if read.tagFilter != "" {
		query = query.Where(tagFilterClause, tagFilterPattern(read.tagFilter))
	}
	var records []schemas.Secret
	err := query.Order("key asc").Find(&records).Error
	if err != nil {
		return nil, 0, errors.Internal("failed to list secrets", err)
	}

	projectID, _ := service.getProjectIDFromEnv(ctx, envID)
	rotationDays := service.rotationDays(ctx, projectID)

	if len(records) == 0 {
		service.logger.Info(read.auditAction, slog.String("user_id", person.logID()), slog.String("client_ip", person.clientIP), slog.Int64("environment_id", envID), slog.Int("count", 0), slog.Bool("revealed", read.reveal))
		return []SecretResponse{}, rotationDays, nil
	}

	var dataKey []byte
	if read.reveal {
		dataKey, err = service.resolveDataKey(ctx, envID)
		if err != nil {
			return nil, 0, err
		}
	}

	ids := make([]int64, len(records))
	for i, r := range records {
		ids[i] = r.ID
	}
	grants, err := service.secretGrants(ctx, service.orm, ids, person.userID)
	if err != nil {
		return nil, 0, err
	}
	rotations := service.lastRotations(ctx, ids)
	now := time.Now().UTC()

	secrets := make([]SecretResponse, 0, len(records))
	for _, r := range records {
		if !aclAllowsRead(person.role, grants[r.ID]) {
			continue
		}
		rotatedAt, ok := rotations[r.ID]
		if !ok {
			rotatedAt = r.UpdatedAt
		}
		response := SecretResponse{
			ID:             r.ID,
			Key:            r.Key,
			Version:        r.Version,
			Tags:           splitTags(r.Tags),
			CreatedAt:      r.CreatedAt,
			UpdatedAt:      r.UpdatedAt,
			RotatedAt:      rotatedAt,
			RotationStatus: rotationStatus(rotatedAt, rotationDays, now),
		}
		if read.reveal {
			plaintext, err := crypto.Decrypt(r.EncryptedValue, r.Nonce, dataKey)
			if err != nil {
				return nil, 0, errors.Internal("failed to decrypt secret", err)
			}
			resolved := service.resolveReferences(ctx, projectID, string(plaintext), dataKey, person.scope, 0)
			response.Value = &resolved
		}
		secrets = append(secrets, response)
	}

	service.logger.Info(read.auditAction, slog.String("user_id", person.logID()), slog.String("client_ip", person.clientIP), slog.Int64("environment_id", envID), slog.Int("count", len(secrets)), slog.Bool("revealed", read.reveal))

	if service.audit != nil {
		service.audit.Log(ctx, projectID, person.logID(), "", read.auditAction, "environment:"+strconv.FormatInt(envID, 10), readDetail(len(secrets), read.reveal), person.clientIP)
	}

	return secrets, rotationDays, nil
}

// readDetail is the audit detail of a bulk read. A revealing read is spelled
// out in values rather than rows, because that number is the size of the
// exposure.
func readDetail(count int, reveal bool) string {
	if reveal {
		return strconv.Itoa(count) + " values revealed"
	}
	return strconv.Itoa(count) + " secrets"
}

// revealSecret returns the plaintext of one secret and records that in its own
// audit entry, so reading a single key is never indistinguishable from dumping
// an environment. A secret the caller may not read answers not_found, the same
// way it is simply absent from every listing: an ACL hides a secret, it does
// not announce it.
func (service *Service) revealSecret(ctx context.Context, envID int64, key string, person caller) (*SecretResponse, error) {
	record, err := service.findSecret(ctx, envID, key)
	if err != nil {
		return nil, err
	}

	grants, err := service.secretGrants(ctx, service.orm, []int64{record.ID}, person.userID)
	if err != nil {
		return nil, err
	}
	if !aclAllowsRead(person.role, grants[record.ID]) {
		return nil, errors.NotFound("secret not found")
	}

	dataKey, err := service.resolveDataKey(ctx, envID)
	if err != nil {
		return nil, err
	}
	plaintext, err := crypto.Decrypt(record.EncryptedValue, record.Nonce, dataKey)
	if err != nil {
		return nil, errors.Internal("failed to decrypt secret", err)
	}

	projectID, _ := service.getProjectIDFromEnv(ctx, envID)
	rotationDays := service.rotationDays(ctx, projectID)
	resolved := service.resolveReferences(ctx, projectID, string(plaintext), dataKey, person.scope, 0)

	rotatedAt, ok := service.lastRotations(ctx, []int64{record.ID})[record.ID]
	if !ok {
		rotatedAt = record.UpdatedAt
	}

	service.logger.Info(auditActionReveal, slog.String("user_id", person.logID()), slog.String("client_ip", person.clientIP), slog.Int64("environment_id", envID), slog.String("key", key))
	if service.audit != nil {
		service.audit.Log(ctx, projectID, person.logID(), "", auditActionReveal, "secret:"+key, "", person.clientIP)
	}

	return &SecretResponse{
		ID:             record.ID,
		Key:            record.Key,
		Value:          &resolved,
		Version:        record.Version,
		Tags:           splitTags(record.Tags),
		CreatedAt:      record.CreatedAt,
		UpdatedAt:      record.UpdatedAt,
		RotatedAt:      rotatedAt,
		RotationStatus: rotationStatus(rotatedAt, rotationDays, time.Now().UTC()),
	}, nil
}

// rotationSummary grades every secret in the project so the UI can flag a stale
// environment the user is not currently looking at. Values stay encrypted here:
// only timestamps are read.
func (service *Service) rotationSummary(ctx context.Context, projectID int64) (*RotationSummaryResponse, error) {
	rotationDays := service.rotationDays(ctx, projectID)

	var environments []schemas.Environment
	err := service.orm.WithContext(ctx).
		Where("project_id = ?", projectID).
		Order("position asc").
		Find(&environments).Error
	if err != nil {
		return nil, errors.Internal("failed to list environments", err)
	}

	summary := &RotationSummaryResponse{
		RotationDays: rotationDays,
		Environments: make([]EnvironmentRotation, len(environments)),
	}
	for i, env := range environments {
		summary.Environments[i] = EnvironmentRotation{Slug: env.Slug, Name: env.Name}
	}
	if rotationDays <= 0 || len(environments) == 0 {
		return summary, nil
	}

	var rows []struct {
		EnvironmentID int64
		RotatedAt     time.Time
	}
	err = service.orm.WithContext(ctx).
		Table("secrets s").
		Select("s.environment_id, coalesce(max(v.created_at), s.updated_at) as rotated_at").
		Joins("join environments e on e.id = s.environment_id").
		Joins("left join secret_versions v on v.secret_id = s.id").
		Where("e.project_id = ?", projectID).
		Group("s.id, s.environment_id, s.updated_at").
		Scan(&rows).Error
	if err != nil {
		return nil, errors.Internal("failed to summarise rotation", err)
	}

	index := make(map[int64]int, len(environments))
	for i, env := range environments {
		index[env.ID] = i
	}

	now := time.Now().UTC()
	for _, row := range rows {
		position, ok := index[row.EnvironmentID]
		if !ok {
			continue
		}
		switch rotationStatus(row.RotatedAt, rotationDays, now) {
		case RotationStatusOverdue:
			summary.Environments[position].Overdue++
		case RotationStatusDueSoon:
			summary.Environments[position].DueSoon++
		}
	}

	return summary, nil
}

// rotationDays reads the project's rotation policy. A project that predates the
// column, or one whose policy is unreadable, simply gets no warnings.
func (service *Service) rotationDays(ctx context.Context, projectID int64) int {
	if projectID == 0 {
		return 0
	}
	var days int
	err := service.orm.WithContext(ctx).
		Model(&schemas.Project{}).
		Where("id = ?", projectID).
		Pluck("rotation_days", &days).Error
	if err != nil {
		return 0
	}
	return days
}

// lastRotations maps each secret to the moment its value last changed. The
// version table is the only honest source for that: the secrets row is also
// touched by tag edits, which are not rotations.
func (service *Service) lastRotations(ctx context.Context, secretIDs []int64) map[int64]time.Time {
	var rows []struct {
		SecretID  int64
		RotatedAt time.Time
	}
	err := service.orm.WithContext(ctx).
		Model(&schemas.SecretVersion{}).
		Select("secret_id, max(created_at) as rotated_at").
		Where("secret_id IN ?", secretIDs).
		Group("secret_id").
		Scan(&rows).Error
	if err != nil {
		return map[int64]time.Time{}
	}

	rotations := make(map[int64]time.Time, len(rows))
	for _, row := range rows {
		rotations[row.SecretID] = row.RotatedAt
	}
	return rotations
}

func (service *Service) updateTags(ctx context.Context, envID int64, key string, tags []string, person caller) error {
	secret, err := service.findSecret(ctx, envID, key)
	if err != nil {
		return err
	}
	if err := service.assertSecretWrite(ctx, secret.ID, person); err != nil {
		return err
	}

	err = service.orm.WithContext(ctx).
		Model(&schemas.Secret{}).
		Where("id = ?", secret.ID).
		Update("tags", joinTags(tags)).Error
	if err != nil {
		return errors.Internal("failed to update tags", err)
	}
	return nil
}

func (service *Service) findSecret(ctx context.Context, envID int64, key string) (*schemas.Secret, error) {
	var record schemas.Secret
	err := service.orm.WithContext(ctx).
		Where("environment_id = ? AND key = ?", envID, key).
		First(&record).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.NotFound("secret not found")
	}
	if err != nil {
		return nil, errors.Internal("failed to find secret", err)
	}
	return &record, nil
}

func (service *Service) createSecret(ctx context.Context, envID int64, key, value string, person caller, tags []string) (*SecretResponse, error) {
	dataKey, err := service.resolveDataKey(ctx, envID)
	if err != nil {
		return nil, err
	}

	ciphertext, nonce, err := crypto.Encrypt([]byte(value), dataKey)
	if err != nil {
		return nil, errors.Internal("failed to encrypt secret", err)
	}

	record := &schemas.Secret{
		EnvironmentID:  envID,
		Key:            key,
		EncryptedValue: ciphertext,
		Nonce:          nonce,
		Version:        1,
		Tags:           joinTags(tags),
	}
	err = service.orm.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(record).Error; err != nil {
			if stderrors.Is(err, gorm.ErrDuplicatedKey) {
				return errDuplicateKey
			}
			return errors.Internal("failed to create secret", err)
		}
		return tx.Create(&schemas.SecretVersion{
			SecretID:       record.ID,
			Version:        1,
			EncryptedValue: ciphertext,
			Nonce:          nonce,
			ChangedBy:      person.logID(),
		}).Error
	})
	if stderrors.Is(err, errDuplicateKey) {
		return nil, service.duplicateKeyError(ctx, envID, key)
	}
	if err != nil {
		return nil, err
	}

	service.logger.Info("secrets.create", slog.String("user_id", person.logID()), slog.String("client_ip", person.clientIP), slog.Int64("environment_id", envID), slog.String("key", key))
	service.fireWebhook(ctx, "secret.created", envID, key, 1, person)

	return &SecretResponse{
		ID:             record.ID,
		Key:            record.Key,
		Value:          &value,
		Version:        record.Version,
		Tags:           splitTags(record.Tags),
		CreatedAt:      record.CreatedAt,
		UpdatedAt:      record.UpdatedAt,
		RotatedAt:      record.UpdatedAt,
		RotationStatus: RotationStatusOK,
	}, nil
}

func (service *Service) updateSecret(ctx context.Context, envID int64, key, value string, person caller) (*SecretResponse, error) {
	found, err := service.findSecret(ctx, envID, key)
	if err != nil {
		return nil, err
	}
	if err := service.assertSecretWrite(ctx, found.ID, person); err != nil {
		return nil, err
	}
	record := *found

	dataKey, err := service.resolveDataKey(ctx, envID)
	if err != nil {
		return nil, err
	}

	ciphertext, nonce, err := crypto.Encrypt([]byte(value), dataKey)
	if err != nil {
		return nil, errors.Internal("failed to encrypt secret", err)
	}

	record.EncryptedValue = ciphertext
	record.Nonce = nonce
	record.Version++

	err = service.orm.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&record).Error; err != nil {
			return errors.Internal("failed to update secret", err)
		}
		return tx.Create(&schemas.SecretVersion{
			SecretID:       record.ID,
			Version:        record.Version,
			EncryptedValue: ciphertext,
			Nonce:          nonce,
			ChangedBy:      person.logID(),
		}).Error
	})
	if err != nil {
		return nil, err
	}

	service.logger.Info("secrets.update", slog.String("user_id", person.logID()), slog.String("client_ip", person.clientIP), slog.Int64("environment_id", envID), slog.String("key", key), slog.Int("version", record.Version))
	service.fireWebhook(ctx, "secret.updated", envID, key, record.Version, person)

	return &SecretResponse{
		ID:             record.ID,
		Key:            record.Key,
		Value:          &value,
		Version:        record.Version,
		Tags:           splitTags(record.Tags),
		CreatedAt:      record.CreatedAt,
		UpdatedAt:      record.UpdatedAt,
		RotatedAt:      record.UpdatedAt,
		RotationStatus: RotationStatusOK,
	}, nil
}

func (service *Service) listVersions(ctx context.Context, envID int64, key string) ([]VersionResponse, error) {
	var secret schemas.Secret
	err := service.orm.WithContext(ctx).
		Where("environment_id = ? AND key = ?", envID, key).
		First(&secret).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.NotFound("secret not found")
	}
	if err != nil {
		return nil, errors.Internal("failed to find secret", err)
	}

	var versions []schemas.SecretVersion
	err = service.orm.WithContext(ctx).
		Where("secret_id = ?", secret.ID).
		Order("version desc").
		Find(&versions).Error
	if err != nil {
		return nil, errors.Internal("failed to list versions", err)
	}

	result := make([]VersionResponse, len(versions))
	for i, v := range versions {
		result[i] = VersionResponse{
			Version:   v.Version,
			ChangedBy: v.ChangedBy,
			CreatedAt: v.CreatedAt,
		}
	}
	return result, nil
}

func (service *Service) rollbackSecret(ctx context.Context, envID int64, key string, targetVersion int, person caller) (*SecretResponse, error) {
	found, err := service.findSecret(ctx, envID, key)
	if err != nil {
		return nil, err
	}
	if err := service.assertSecretWrite(ctx, found.ID, person); err != nil {
		return nil, err
	}
	secret := *found

	var version schemas.SecretVersion
	err = service.orm.WithContext(ctx).
		Where("secret_id = ? AND version = ?", secret.ID, targetVersion).
		First(&version).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.NotFound("version not found")
	}
	if err != nil {
		return nil, errors.Internal("failed to find version", err)
	}

	secret.EncryptedValue = version.EncryptedValue
	secret.Nonce = version.Nonce
	secret.Version++

	err = service.orm.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&secret).Error; err != nil {
			return errors.Internal("failed to rollback secret", err)
		}
		return tx.Create(&schemas.SecretVersion{
			SecretID:       secret.ID,
			Version:        secret.Version,
			EncryptedValue: version.EncryptedValue,
			Nonce:          version.Nonce,
			ChangedBy:      person.logID(),
		}).Error
	})
	if err != nil {
		return nil, err
	}

	dataKey, err := service.resolveDataKey(ctx, envID)
	if err != nil {
		return nil, err
	}
	plaintext, err := crypto.Decrypt(secret.EncryptedValue, secret.Nonce, dataKey)
	if err != nil {
		return nil, errors.Internal("failed to decrypt secret", err)
	}

	restored := string(plaintext)

	service.logger.Info("secrets.rollback", slog.String("user_id", person.logID()), slog.String("client_ip", person.clientIP), slog.Int64("environment_id", envID), slog.String("key", key), slog.Int("to_version", targetVersion), slog.Int("new_version", secret.Version))
	service.fireWebhook(ctx, "secret.rolled_back", envID, key, secret.Version, person)

	return &SecretResponse{
		ID:             secret.ID,
		Key:            secret.Key,
		Value:          &restored,
		Version:        secret.Version,
		Tags:           splitTags(secret.Tags),
		CreatedAt:      secret.CreatedAt,
		UpdatedAt:      secret.UpdatedAt,
		RotatedAt:      secret.UpdatedAt,
		RotationStatus: RotationStatusOK,
	}, nil
}

// importEnv applies a whole .env file in one transaction: either every parsed
// line lands or none does. Lines the caller may not write, and keys the schema
// cannot hold, are counted as skipped rather than failing the import.
func (service *Service) importEnv(ctx context.Context, envID int64, content string, person caller) (*ImportResponse, error) {
	dataKey, err := service.resolveDataKey(ctx, envID)
	if err != nil {
		return nil, err
	}

	entries, skipped := parseEnvFile(content)
	result := &ImportResponse{}
	var created, updated []importedSecret
	denied := 0
	conflictKey := ""

	err = service.orm.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result.Created, result.Updated = 0, 0
		created, updated, denied = nil, nil, 0
		conflictKey = ""

		var records []schemas.Secret
		if err := tx.Where("environment_id = ?", envID).Find(&records).Error; err != nil {
			return errors.Internal("failed to load environment secrets", err)
		}

		byKey := make(map[string]*schemas.Secret, len(records))
		ids := make([]int64, len(records))
		for i := range records {
			byKey[records[i].Key] = &records[i]
			ids[i] = records[i].ID
		}
		grants, err := service.secretGrants(ctx, tx, ids, person.userID)
		if err != nil {
			return err
		}

		for _, entry := range entries {
			ciphertext, nonce, encErr := crypto.Encrypt([]byte(entry.value), dataKey)
			if encErr != nil {
				return errors.Internal("failed to encrypt", encErr)
			}

			existing, found := byKey[entry.key]
			if found {
				if !aclAllowsWrite(person.role, grants[existing.ID]) {
					denied++
					continue
				}
				existing.EncryptedValue = ciphertext
				existing.Nonce = nonce
				existing.Version++
				if err := tx.Save(existing).Error; err != nil {
					return errors.Internal("failed to update secret", err)
				}
				if err := tx.Create(&schemas.SecretVersion{
					SecretID:       existing.ID,
					Version:        existing.Version,
					EncryptedValue: ciphertext,
					Nonce:          nonce,
					ChangedBy:      person.logID(),
				}).Error; err != nil {
					return errors.Internal("failed to record version", err)
				}
				result.Updated++
				updated = append(updated, importedSecret{key: entry.key, version: existing.Version})
				continue
			}

			record := &schemas.Secret{
				EnvironmentID:  envID,
				Key:            entry.key,
				EncryptedValue: ciphertext,
				Nonce:          nonce,
				Version:        1,
			}
			if err := tx.Create(record).Error; err != nil {
				if stderrors.Is(err, gorm.ErrDuplicatedKey) {
					conflictKey = entry.key
					return errDuplicateKey
				}
				return errors.Internal("failed to create secret", err)
			}
			if err := tx.Create(&schemas.SecretVersion{
				SecretID:       record.ID,
				Version:        1,
				EncryptedValue: ciphertext,
				Nonce:          nonce,
				ChangedBy:      person.logID(),
			}).Error; err != nil {
				return errors.Internal("failed to record version", err)
			}
			byKey[entry.key] = record
			result.Created++
			created = append(created, importedSecret{key: entry.key, version: 1})
		}
		return nil
	})
	if stderrors.Is(err, errDuplicateKey) {
		return nil, service.duplicateKeyError(ctx, envID, conflictKey)
	}
	if err != nil {
		return nil, err
	}
	result.Skipped = skipped + denied

	service.logger.Info("secrets.import", slog.String("user_id", person.logID()), slog.String("client_ip", person.clientIP), slog.Int64("environment_id", envID), slog.Int("created", result.Created), slog.Int("updated", result.Updated), slog.Int("skipped", result.Skipped))

	projectID, envSlug, err := service.environmentRef(ctx, envID)
	if err == nil {
		if service.audit != nil {
			detail := "created=" + strconv.Itoa(result.Created) + " updated=" + strconv.Itoa(result.Updated) + " skipped=" + strconv.Itoa(result.Skipped)
			service.audit.Log(ctx, projectID, person.logID(), "", "secrets.import", "environment:"+envSlug, detail, person.clientIP)
		}
		for _, secret := range created {
			service.emitSecretEvent(ctx, projectID, envSlug, "secret.created", secret.key, secret.version, person)
		}
		for _, secret := range updated {
			service.emitSecretEvent(ctx, projectID, envSlug, "secret.updated", secret.key, secret.version, person)
		}
	}

	return result, nil
}

type importedSecret struct {
	key     string
	version int
}

type envEntry struct {
	key   string
	value string
}

// parseEnvFile reads back the dialect exportEnv writes: one KEY=value per line,
// '#' comments, and quoted values whose escapes decodeEnvValue expands. Lines
// without a '=' and keys the secrets table cannot hold are counted as skipped
// instead of being inserted.
func parseEnvFile(content string) ([]envEntry, int) {
	var entries []envEntry
	skipped := 0

	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		rawKey, rawValue, ok := strings.Cut(line, "=")
		if !ok {
			skipped++
			continue
		}
		key := strings.TrimSpace(rawKey)
		if key == "" || len(key) > maxSecretKeyLength {
			skipped++
			continue
		}
		entries = append(entries, envEntry{key: key, value: decodeEnvValue(strings.TrimSpace(rawValue))})
	}

	return entries, skipped
}

var envValueEscaper = strings.NewReplacer(`\`, `\\`, `"`, `\"`, "\n", `\n`, "\r", `\r`, "\t", `\t`)

// encodeEnvValue renders a secret value as a .env literal. Anything a parser
// could misread - whitespace, quotes, a comment marker, a backslash - forces
// the double-quoted form, where backslashes are escaped before quotes and line
// breaks become two-character escapes so one secret stays on one line.
func encodeEnvValue(value string) string {
	if value == "" || !strings.ContainsAny(value, " \t\r\n\"'#\\") {
		return value
	}
	return `"` + envValueEscaper.Replace(value) + `"`
}

// decodeEnvValue is the exact inverse of encodeEnvValue. Single-quoted values
// are taken literally, as in every other .env dialect.
func decodeEnvValue(value string) string {
	if len(value) >= 2 && strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'") {
		return value[1 : len(value)-1]
	}
	if len(value) < 2 || !strings.HasPrefix(value, `"`) || !strings.HasSuffix(value, `"`) {
		return value
	}

	inner := value[1 : len(value)-1]
	if !strings.Contains(inner, `\`) {
		return inner
	}

	var sb strings.Builder
	sb.Grow(len(inner))
	for i := 0; i < len(inner); i++ {
		if inner[i] != '\\' || i+1 >= len(inner) {
			sb.WriteByte(inner[i])
			continue
		}
		i++
		switch inner[i] {
		case 'n':
			sb.WriteByte('\n')
		case 'r':
			sb.WriteByte('\r')
		case 't':
			sb.WriteByte('\t')
		case '\\':
			sb.WriteByte('\\')
		case '"':
			sb.WriteByte('"')
		default:
			sb.WriteByte('\\')
			sb.WriteByte(inner[i])
		}
	}
	return sb.String()
}

func (service *Service) exportEnv(ctx context.Context, envID int64, person caller) (string, error) {
	secrets, _, err := service.listSecrets(ctx, envID, person, secretRead{reveal: true, auditAction: auditActionExport})
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	for _, s := range secrets {
		sb.WriteString(s.Key)
		sb.WriteString("=")
		sb.WriteString(encodeEnvValue(revealedValue(s)))
		sb.WriteString("\n")
	}
	return sb.String(), nil
}

// revealedValue reads the plaintext off a response the service itself asked to
// have revealed. Nothing but a bug can leave it nil there, and an empty string
// is what a secret with no value would have produced anyway.
func revealedValue(secret SecretResponse) string {
	if secret.Value == nil {
		return ""
	}
	return *secret.Value
}

func (service *Service) diffEnvironments(ctx context.Context, envA, envB int64, person caller) (*DiffResponse, error) {
	secretsA, _, err := service.listSecrets(ctx, envA, person, secretRead{reveal: true, auditAction: auditActionDiff})
	if err != nil {
		return nil, err
	}
	secretsB, _, err := service.listSecrets(ctx, envB, person, secretRead{reveal: true, auditAction: auditActionDiff})
	if err != nil {
		return nil, err
	}

	mapA := make(map[string]string, len(secretsA))
	for _, s := range secretsA {
		mapA[s.Key] = revealedValue(s)
	}
	mapB := make(map[string]string, len(secretsB))
	for _, s := range secretsB {
		mapB[s.Key] = revealedValue(s)
	}

	seen := make(map[string]bool)
	var entries []DiffEntry

	for _, s := range secretsA {
		seen[s.Key] = true
		valB, exists := mapB[s.Key]
		if !exists {
			entries = append(entries, DiffEntry{Key: s.Key, Status: "removed"})
		} else if revealedValue(s) != valB {
			entries = append(entries, DiffEntry{Key: s.Key, Status: "changed"})
		}
	}
	for _, s := range secretsB {
		if !seen[s.Key] {
			entries = append(entries, DiffEntry{Key: s.Key, Status: "added"})
		}
	}

	if entries == nil {
		entries = []DiffEntry{}
	}

	return &DiffResponse{Entries: entries}, nil
}

func (service *Service) listACLs(ctx context.Context, envID int64, key string) ([]ACLResponse, error) {
	var secret schemas.Secret
	err := service.orm.WithContext(ctx).
		Where("environment_id = ? AND key = ?", envID, key).
		First(&secret).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.NotFound("secret not found")
	}
	if err != nil {
		return nil, errors.Internal("failed to find secret", err)
	}

	var results []struct {
		ID       int64
		Email    string
		CanRead  bool
		CanWrite bool
	}
	err = service.orm.WithContext(ctx).
		Table("secret_acls a").
		Select("a.id, u.email, a.can_read, a.can_write").
		Joins("join users u on u.id = a.user_id").
		Where("a.secret_id = ?", secret.ID).
		Scan(&results).Error
	if err != nil {
		return nil, errors.Internal("failed to list ACLs", err)
	}

	acls := make([]ACLResponse, len(results))
	for i, r := range results {
		acls[i] = ACLResponse{ID: r.ID, Email: r.Email, CanRead: r.CanRead, CanWrite: r.CanWrite}
	}
	return acls, nil
}

func (service *Service) setACL(ctx context.Context, envID int64, key, email string, canRead, canWrite bool) error {
	var secret schemas.Secret
	err := service.orm.WithContext(ctx).
		Where("environment_id = ? AND key = ?", envID, key).
		First(&secret).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return errors.NotFound("secret not found")
	}
	if err != nil {
		return errors.Internal("failed to find secret", err)
	}

	var user schemas.User
	if err := service.orm.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return errors.NotFound("user not found")
		}
		return errors.Internal("failed to find user", err)
	}

	var existing schemas.SecretACL
	err = service.orm.WithContext(ctx).
		Where("secret_id = ? AND user_id = ?", secret.ID, user.ID).
		First(&existing).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		acl := &schemas.SecretACL{
			SecretID: secret.ID,
			UserID:   user.ID,
			CanRead:  canRead,
			CanWrite: canWrite,
		}
		return service.orm.WithContext(ctx).Create(acl).Error
	}
	if err != nil {
		return errors.Internal("failed to check ACL", err)
	}

	existing.CanRead = canRead
	existing.CanWrite = canWrite
	return service.orm.WithContext(ctx).Save(&existing).Error
}

func (service *Service) deleteACL(ctx context.Context, envID int64, key string, aclID int64) error {
	var secret schemas.Secret
	err := service.orm.WithContext(ctx).
		Where("environment_id = ? AND key = ?", envID, key).
		First(&secret).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return errors.NotFound("secret not found")
	}
	if err != nil {
		return errors.Internal("failed to find secret", err)
	}

	result := service.orm.WithContext(ctx).
		Where("id = ? AND secret_id = ?", aclID, secret.ID).
		Delete(&schemas.SecretACL{})
	if result.Error != nil {
		return errors.Internal("failed to delete ACL", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.NotFound("ACL not found")
	}
	return nil
}

// deleteSecret moves a secret to the trash rather than destroying it: the row
// keeps its versions and its ACLs, disappears from every other route through
// GORM's soft-delete scope, and stays restorable for the retention window. The
// key remains reserved by the unique index while it sits there, so a restore
// can never collide with a replacement created in the meantime.
func (service *Service) deleteSecret(ctx context.Context, envID int64, key string, person caller) error {
	secret, err := service.findSecret(ctx, envID, key)
	if err != nil {
		return err
	}
	if err := service.assertSecretWrite(ctx, secret.ID, person); err != nil {
		return err
	}

	result := service.orm.WithContext(ctx).
		Model(&schemas.Secret{}).
		Where("id = ?", secret.ID).
		Updates(map[string]any{"deleted_at": time.Now().UTC(), "deleted_by": person.logID()})
	if result.Error != nil {
		return errors.Internal("failed to delete secret", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.NotFound("secret not found")
	}

	service.logger.Info("secrets.trash", slog.String("user_id", person.logID()), slog.String("client_ip", person.clientIP), slog.Int64("environment_id", envID), slog.String("key", key))
	service.fireWebhook(ctx, "secret.trashed", envID, key, secret.Version, person)
	return nil
}

// findTrashedSecret locates a secret that only the trash can see. Unscoped is
// what makes it reachable at all: every other query in this module is filtered
// by the soft-delete scope.
func (service *Service) findTrashedSecret(ctx context.Context, envID int64, key string) (*schemas.Secret, error) {
	var record schemas.Secret
	err := service.orm.WithContext(ctx).
		Unscoped().
		Where("environment_id = ? AND key = ? AND deleted_at IS NOT NULL", envID, key).
		First(&record).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.NotFound("secret not found in trash")
	}
	if err != nil {
		return nil, errors.Internal("failed to find trashed secret", err)
	}
	return &record, nil
}

// listTrash reports what can still be recovered, and until when. Values are
// never decrypted here: the trash answers "what did I delete", not "what was in
// it". Per-secret ACLs hide a trashed secret exactly as they hide a live one.
func (service *Service) listTrash(ctx context.Context, envID int64, person caller) (*TrashResponse, error) {
	var records []schemas.Secret
	err := service.orm.WithContext(ctx).
		Unscoped().
		Where("environment_id = ? AND deleted_at IS NOT NULL", envID).
		Order("deleted_at desc").
		Find(&records).Error
	if err != nil {
		return nil, errors.Internal("failed to list trashed secrets", err)
	}

	response := &TrashResponse{Secrets: []TrashedSecretResponse{}, RetentionDays: schemas.TrashRetentionDays}
	if len(records) == 0 {
		return response, nil
	}

	ids := make([]int64, len(records))
	for i, r := range records {
		ids[i] = r.ID
	}
	grants, err := service.secretGrants(ctx, service.orm, ids, person.userID)
	if err != nil {
		return nil, err
	}

	for _, r := range records {
		if !aclAllowsRead(person.role, grants[r.ID]) {
			continue
		}
		response.Secrets = append(response.Secrets, TrashedSecretResponse{
			ID:        r.ID,
			Key:       r.Key,
			Tags:      splitTags(r.Tags),
			DeletedAt: r.DeletedAt.Time,
			DeletedBy: r.DeletedBy,
			PurgeAt:   purgeAt(r.DeletedAt.Time),
		})
	}

	projectID, _ := service.getProjectIDFromEnv(ctx, envID)
	if service.audit != nil {
		service.audit.Log(ctx, projectID, person.logID(), "", auditActionTrashList, "environment:"+strconv.FormatInt(envID, 10), strconv.Itoa(len(response.Secrets))+" secrets", person.clientIP)
	}

	return response, nil
}

// restoreSecret lifts a secret back out of the trash. Its versions and ACLs
// were never touched, so clearing deleted_at is the whole of it.
func (service *Service) restoreSecret(ctx context.Context, envID int64, key string, person caller) error {
	secret, err := service.findTrashedSecret(ctx, envID, key)
	if err != nil {
		return err
	}
	if err := service.assertSecretWrite(ctx, secret.ID, person); err != nil {
		return err
	}

	result := service.orm.WithContext(ctx).
		Unscoped().
		Model(&schemas.Secret{}).
		Where("id = ?", secret.ID).
		Updates(map[string]any{"deleted_at": nil, "deleted_by": ""})
	if result.Error != nil {
		return errors.Internal("failed to restore secret", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.NotFound("secret not found in trash")
	}

	service.logger.Info("secrets.restore", slog.String("user_id", person.logID()), slog.String("client_ip", person.clientIP), slog.Int64("environment_id", envID), slog.String("key", key))
	service.fireWebhook(ctx, "secret.restored", envID, key, secret.Version, person)
	return nil
}

// purgeSecret destroys a secret for good, whether it is live or already in the
// trash. It is the only path that frees the key for reuse before the retention
// window elapses, which is why the controller restricts it to owner and admin.
func (service *Service) purgeSecret(ctx context.Context, envID int64, key string, person caller) error {
	var secret schemas.Secret
	err := service.orm.WithContext(ctx).
		Unscoped().
		Where("environment_id = ? AND key = ?", envID, key).
		First(&secret).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return errors.NotFound("secret not found")
	}
	if err != nil {
		return errors.Internal("failed to find secret", err)
	}
	if err := service.assertSecretWrite(ctx, secret.ID, person); err != nil {
		return err
	}

	if err := service.destroySecrets(ctx, []int64{secret.ID}); err != nil {
		return err
	}

	service.logger.Info(auditActionPurge, slog.String("user_id", person.logID()), slog.String("client_ip", person.clientIP), slog.Int64("environment_id", envID), slog.String("key", key))
	projectID, envSlug, refErr := service.environmentRef(ctx, envID)
	if refErr == nil {
		service.emitSecretEvent(ctx, projectID, envSlug, "secret.purged", key, secret.Version, person)
	}
	return nil
}

// destroySecrets removes secrets and everything that hangs off them in one
// transaction. The children go first so a failure can never leave versions or
// ACLs pointing at a secret that no longer exists.
func (service *Service) destroySecrets(ctx context.Context, secretIDs []int64) error {
	if len(secretIDs) == 0 {
		return nil
	}
	return service.orm.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("secret_id IN ?", secretIDs).Delete(&schemas.SecretVersion{}).Error; err != nil {
			return errors.Internal("failed to delete secret versions", err)
		}
		if err := tx.Where("secret_id IN ?", secretIDs).Delete(&schemas.SecretACL{}).Error; err != nil {
			return errors.Internal("failed to delete secret ACLs", err)
		}
		if err := tx.Unscoped().Where("id IN ?", secretIDs).Delete(&schemas.Secret{}).Error; err != nil {
			return errors.Internal("failed to delete secret", err)
		}
		return nil
	})
}

// StartTrashPurge destroys secrets that have outlived the recovery window. It
// runs for the life of the context the server is shut down with, the same way
// the session cleanup loop does.
func (service *Service) StartTrashPurge(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(trashPurgeInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				purged, err := service.purgeExpiredTrash(ctx, time.Now().UTC())
				if err != nil {
					service.logger.Error("secret trash purge failed", slog.Any("error", err))
				} else if purged > 0 {
					service.logger.Info("purged trashed secrets", slog.Int("count", purged))
				}
			}
		}
	}()
}

func (service *Service) purgeExpiredTrash(ctx context.Context, now time.Time) (int, error) {
	var ids []int64
	err := service.orm.WithContext(ctx).
		Unscoped().
		Model(&schemas.Secret{}).
		Where("deleted_at IS NOT NULL AND deleted_at <= ?", trashCutoff(now)).
		Pluck("id", &ids).Error
	if err != nil {
		return 0, errors.Internal("failed to select expired secrets", err)
	}
	if err := service.destroySecrets(ctx, ids); err != nil {
		return 0, err
	}
	return len(ids), nil
}

// duplicateKeyError turns a unique-index violation on (environment_id, key)
// into something the caller can act on. The index deliberately covers
// soft-deleted rows, so the collision may well be with a secret sitting in the
// trash, invisible to every route that would explain the failure.
func (service *Service) duplicateKeyError(ctx context.Context, envID int64, key string) error {
	var trashed int64
	err := service.orm.WithContext(ctx).
		Unscoped().
		Model(&schemas.Secret{}).
		Where("environment_id = ? AND key = ? AND deleted_at IS NOT NULL", envID, key).
		Count(&trashed).Error
	if err == nil && trashed > 0 {
		return errors.Conflict("secret key '" + key + "' is in the trash: restore it or purge it before creating it again")
	}
	return errors.Conflict("secret key already exists in this environment")
}

// environmentProtected reports whether an environment demands owner or admin
// for writes. It is read at the write choke point rather than carried on the
// environment resolver, which only ever hands back an ID.
func (service *Service) environmentProtected(ctx context.Context, envID int64) (bool, error) {
	var protected bool
	err := service.orm.WithContext(ctx).
		Model(&schemas.Environment{}).
		Where("id = ?", envID).
		Pluck("protected", &protected).Error
	if err != nil {
		return false, errors.Internal("failed to read environment protection", err)
	}
	return protected, nil
}
