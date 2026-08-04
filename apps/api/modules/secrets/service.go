package secrets

import (
	"context"
	stderrors "errors"
	"log/slog"
	"regexp"
	"strings"
	"time"

	"casier/internal/audit"
	"casier/internal/crypto"
	"casier/internal/errors"
	"casier/internal/webhook"
	"casier/schemas"

	"gorm.io/gorm"
)

func joinTags(tags []string) string {
	if len(tags) == 0 {
		return ""
	}
	return strings.Join(tags, ",")
}

var refPattern = regexp.MustCompile(`\$\{ref:([^/]+)/([^}]+)\}`)

func (service *Service) resolveReferences(ctx context.Context, spaceID int64, value string, dataKey []byte, depth int) string {
	if depth > 3 || !strings.Contains(value, "${ref:") {
		return value
	}

	return refPattern.ReplaceAllStringFunc(value, func(match string) string {
		parts := refPattern.FindStringSubmatch(match)
		if len(parts) != 3 {
			return match
		}
		envSlug, secretKey := parts[1], parts[2]

		envID, err := service.envRes.ResolveEnvironmentID(ctx, spaceID, envSlug)
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

		return service.resolveReferences(ctx, spaceID, string(plaintext), dataKey, depth+1)
	})
}

func splitTags(s string) []string {
	if s == "" {
		return []string{}
	}
	return strings.Split(s, ",")
}

type SpaceResolver interface {
	ResolveSpaceID(ctx context.Context, slug string, userID int64, requiredRoles ...string) (int64, error)
}

type EnvironmentResolver interface {
	ResolveEnvironmentID(ctx context.Context, spaceID int64, envSlug string) (int64, error)
}

type DataKeyResolver interface {
	GetSpaceDataKey(ctx context.Context, spaceID int64) ([]byte, error)
}

type Service struct {
	orm        *gorm.DB
	logger     *slog.Logger
	spaceRes   SpaceResolver
	envRes     EnvironmentResolver
	dataKeyRes DataKeyResolver
	webhooks   *webhook.Dispatcher
	audit      *audit.Logger
	controller *Controller
}

func NewService(orm *gorm.DB, logger *slog.Logger, spaceRes SpaceResolver, envRes EnvironmentResolver, dataKeyRes DataKeyResolver, webhooks *webhook.Dispatcher, auditLog *audit.Logger) *Service {
	service := &Service{orm: orm, logger: logger, spaceRes: spaceRes, envRes: envRes, dataKeyRes: dataKeyRes, webhooks: webhooks, audit: auditLog}
	service.controller = newController(service)
	return service
}

func (service *Service) fireWebhook(ctx context.Context, eventType string, envID int64, key string, version int, userID, clientIP string) {
	spaceID, err := service.getSpaceIDFromEnv(ctx, envID)
	if err != nil {
		return
	}
	if service.audit != nil {
		service.audit.Log(ctx, spaceID, userID, "", eventType, "secret:"+key, "", clientIP)
	}
	if service.webhooks == nil {
		return
	}
	service.webhooks.Fire(ctx, webhook.Event{
		Type:        eventType,
		SpaceID:     spaceID,
		Key:         key,
		Version:     version,
		UserID:      userID,
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
	})
}

func (service *Service) getSpaceIDFromEnv(ctx context.Context, envID int64) (int64, error) {
	var env schemas.Environment
	if err := service.orm.WithContext(ctx).Where("id = ?", envID).First(&env).Error; err != nil {
		return 0, errors.Internal("failed to resolve environment", err)
	}
	return env.SpaceID, nil
}

func (service *Service) resolveDataKey(ctx context.Context, envID int64) ([]byte, error) {
	spaceID, err := service.getSpaceIDFromEnv(ctx, envID)
	if err != nil {
		return nil, err
	}
	dataKey, err := service.dataKeyRes.GetSpaceDataKey(ctx, spaceID)
	if err != nil {
		return nil, errors.Internal("failed to resolve data key", err)
	}
	return dataKey, nil
}

func (service *Service) listSecrets(ctx context.Context, envID int64, userID, clientIP, tagFilter string) ([]SecretResponse, error) {
	query := service.orm.WithContext(ctx).Where("environment_id = ?", envID)
	if tagFilter != "" {
		query = query.Where("tags LIKE ?", "%"+tagFilter+"%")
	}
	var records []schemas.Secret
	err := query.Order("key asc").Find(&records).Error
	if err != nil {
		return nil, errors.Internal("failed to list secrets", err)
	}

	service.logger.Info("secrets.list", slog.String("user_id", userID), slog.String("client_ip", clientIP), slog.Int64("environment_id", envID), slog.Int("count", len(records)))

	if len(records) == 0 {
		return []SecretResponse{}, nil
	}

	dataKey, err := service.resolveDataKey(ctx, envID)
	if err != nil {
		return nil, err
	}

	spaceID, _ := service.getSpaceIDFromEnv(ctx, envID)

	secrets := make([]SecretResponse, len(records))
	for i, r := range records {
		plaintext, err := crypto.Decrypt(r.EncryptedValue, r.Nonce, dataKey)
		if err != nil {
			return nil, errors.Internal("failed to decrypt secret", err)
		}
		resolved := service.resolveReferences(ctx, spaceID, string(plaintext), dataKey, 0)
		secrets[i] = SecretResponse{
			ID:        r.ID,
			Key:       r.Key,
			Value:     resolved,
			Version:   r.Version,
			Tags:      splitTags(r.Tags),
			CreatedAt: r.CreatedAt,
			UpdatedAt: r.UpdatedAt,
		}
	}
	return secrets, nil
}

func (service *Service) updateTags(ctx context.Context, envID int64, key string, tags []string) error {
	result := service.orm.WithContext(ctx).
		Model(&schemas.Secret{}).
		Where("environment_id = ? AND key = ?", envID, key).
		Update("tags", joinTags(tags))
	if result.Error != nil {
		return errors.Internal("failed to update tags", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.NotFound("secret not found")
	}
	return nil
}

func (service *Service) createSecret(ctx context.Context, envID int64, key, value, userID, clientIP string, tags []string) (*SecretResponse, error) {
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
				return errors.Conflict("secret key already exists in this environment")
			}
			return errors.Internal("failed to create secret", err)
		}
		return tx.Create(&schemas.SecretVersion{
			SecretID:       record.ID,
			Version:        1,
			EncryptedValue: ciphertext,
			Nonce:          nonce,
			ChangedBy:      userID,
		}).Error
	})
	if err != nil {
		return nil, err
	}

	service.logger.Info("secrets.create", slog.String("user_id", userID), slog.String("client_ip", clientIP), slog.Int64("environment_id", envID), slog.String("key", key))
	service.fireWebhook(ctx, "secret.created", envID, key, 1, userID, clientIP)

	return &SecretResponse{
		ID:        record.ID,
		Key:       record.Key,
		Value:     value,
		Version:   record.Version,
		Tags:      splitTags(record.Tags),
		CreatedAt: record.CreatedAt,
		UpdatedAt: record.UpdatedAt,
	}, nil
}

func (service *Service) updateSecret(ctx context.Context, envID int64, key, value, userID, clientIP string) (*SecretResponse, error) {
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
			ChangedBy:      userID,
		}).Error
	})
	if err != nil {
		return nil, err
	}

	service.logger.Info("secrets.update", slog.String("user_id", userID), slog.String("client_ip", clientIP), slog.Int64("environment_id", envID), slog.String("key", key), slog.Int("version", record.Version))
	service.fireWebhook(ctx, "secret.updated", envID, key, record.Version, userID, clientIP)

	return &SecretResponse{
		ID:        record.ID,
		Key:       record.Key,
		Value:     value,
		Version:   record.Version,
		Tags:      splitTags(record.Tags),
		CreatedAt: record.CreatedAt,
		UpdatedAt: record.UpdatedAt,
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

func (service *Service) rollbackSecret(ctx context.Context, envID int64, key string, targetVersion int, userID, clientIP string) (*SecretResponse, error) {
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
			ChangedBy:      userID,
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

	service.logger.Info("secrets.rollback", slog.String("user_id", userID), slog.String("client_ip", clientIP), slog.Int64("environment_id", envID), slog.String("key", key), slog.Int("to_version", targetVersion), slog.Int("new_version", secret.Version))

	return &SecretResponse{
		ID:        secret.ID,
		Key:       secret.Key,
		Value:     string(plaintext),
		Version:   secret.Version,
		Tags:      splitTags(secret.Tags),
		CreatedAt: secret.CreatedAt,
		UpdatedAt: secret.UpdatedAt,
	}, nil
}

func (service *Service) importEnv(ctx context.Context, envID int64, content, userID, clientIP string) (*ImportResponse, error) {
	dataKey, err := service.resolveDataKey(ctx, envID)
	if err != nil {
		return nil, err
	}

	result := &ImportResponse{}
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") && len(value) >= 2 {
			value = value[1 : len(value)-1]
		}
		if strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'") && len(value) >= 2 {
			value = value[1 : len(value)-1]
		}

		var existing schemas.Secret
		err := service.orm.WithContext(ctx).
			Where("environment_id = ? AND key = ?", envID, key).
			First(&existing).Error

		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			ciphertext, nonce, encErr := crypto.Encrypt([]byte(value), dataKey)
			if encErr != nil {
				return nil, errors.Internal("failed to encrypt", encErr)
			}
			record := &schemas.Secret{
				EnvironmentID:  envID,
				Key:            key,
				EncryptedValue: ciphertext,
				Nonce:          nonce,
				Version:        1,
			}
			if err := service.orm.WithContext(ctx).Create(record).Error; err != nil {
				result.Skipped++
				continue
			}
			service.orm.WithContext(ctx).Create(&schemas.SecretVersion{
				SecretID: record.ID, Version: 1,
				EncryptedValue: ciphertext, Nonce: nonce, ChangedBy: userID,
			})
			result.Created++
		} else if err == nil {
			ciphertext, nonce, encErr := crypto.Encrypt([]byte(value), dataKey)
			if encErr != nil {
				return nil, errors.Internal("failed to encrypt", encErr)
			}
			existing.EncryptedValue = ciphertext
			existing.Nonce = nonce
			existing.Version++
			if err := service.orm.WithContext(ctx).Save(&existing).Error; err != nil {
				result.Skipped++
				continue
			}
			service.orm.WithContext(ctx).Create(&schemas.SecretVersion{
				SecretID: existing.ID, Version: existing.Version,
				EncryptedValue: ciphertext, Nonce: nonce, ChangedBy: userID,
			})
			result.Updated++
		} else {
			result.Skipped++
		}
	}

	service.logger.Info("secrets.import", slog.String("user_id", userID), slog.String("client_ip", clientIP), slog.Int64("environment_id", envID), slog.Int("created", result.Created), slog.Int("updated", result.Updated))
	return result, nil
}

func (service *Service) exportEnv(ctx context.Context, envID int64, userID, clientIP string) (string, error) {
	secrets, err := service.listSecrets(ctx, envID, userID, clientIP, "")
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	for _, s := range secrets {
		sb.WriteString(s.Key)
		sb.WriteString("=")
		if strings.ContainsAny(s.Value, " \t\n\"'#") {
			sb.WriteString("\"")
			sb.WriteString(strings.ReplaceAll(s.Value, "\"", "\\\""))
			sb.WriteString("\"")
		} else {
			sb.WriteString(s.Value)
		}
		sb.WriteString("\n")
	}
	return sb.String(), nil
}

func (service *Service) diffEnvironments(ctx context.Context, envA, envB int64, userID, clientIP string) (*DiffResponse, error) {
	secretsA, err := service.listSecrets(ctx, envA, userID, clientIP, "")
	if err != nil {
		return nil, err
	}
	secretsB, err := service.listSecrets(ctx, envB, userID, clientIP, "")
	if err != nil {
		return nil, err
	}

	mapA := make(map[string]string, len(secretsA))
	for _, s := range secretsA {
		mapA[s.Key] = s.Value
	}
	mapB := make(map[string]string, len(secretsB))
	for _, s := range secretsB {
		mapB[s.Key] = s.Value
	}

	seen := make(map[string]bool)
	var entries []DiffEntry

	for _, s := range secretsA {
		seen[s.Key] = true
		valB, exists := mapB[s.Key]
		if !exists {
			entries = append(entries, DiffEntry{Key: s.Key, Status: "removed"})
		} else if s.Value != valB {
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

func (service *Service) deleteSecret(ctx context.Context, envID int64, key, userID, clientIP string) error {
	result := service.orm.WithContext(ctx).
		Where("environment_id = ? AND key = ?", envID, key).
		Delete(&schemas.Secret{})
	if result.Error != nil {
		return errors.Internal("failed to delete secret", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.NotFound("secret not found")
	}

	service.logger.Info("secrets.delete", slog.String("user_id", userID), slog.String("client_ip", clientIP), slog.Int64("environment_id", envID), slog.String("key", key))
	service.fireWebhook(ctx, "secret.deleted", envID, key, 0, userID, clientIP)
	return nil
}
