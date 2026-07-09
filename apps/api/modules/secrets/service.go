package secrets

import (
	"context"
	stderrors "errors"
	"log/slog"

	"clef/internal/crypto"
	"clef/internal/errors"
	"clef/schemas"

	"gorm.io/gorm"
)

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
	controller *Controller
}

func NewService(orm *gorm.DB, logger *slog.Logger, spaceRes SpaceResolver, envRes EnvironmentResolver, dataKeyRes DataKeyResolver) *Service {
	service := &Service{orm: orm, logger: logger, spaceRes: spaceRes, envRes: envRes, dataKeyRes: dataKeyRes}
	service.controller = newController(service)
	return service
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

func (service *Service) listSecrets(ctx context.Context, envID int64, userID, clientIP string) ([]SecretResponse, error) {
	var records []schemas.Secret
	err := service.orm.WithContext(ctx).
		Where("environment_id = ?", envID).
		Order("key asc").
		Find(&records).Error
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

	secrets := make([]SecretResponse, len(records))
	for i, r := range records {
		plaintext, err := crypto.Decrypt(r.EncryptedValue, r.Nonce, dataKey)
		if err != nil {
			return nil, errors.Internal("failed to decrypt secret", err)
		}
		secrets[i] = SecretResponse{
			ID:        r.ID,
			Key:       r.Key,
			Value:     string(plaintext),
			Version:   r.Version,
			CreatedAt: r.CreatedAt,
			UpdatedAt: r.UpdatedAt,
		}
	}
	return secrets, nil
}

func (service *Service) createSecret(ctx context.Context, envID int64, key, value, userID, clientIP string) (*SecretResponse, error) {
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
	}
	if err := service.orm.WithContext(ctx).Create(record).Error; err != nil {
		if stderrors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, errors.Conflict("secret key already exists in this environment")
		}
		return nil, errors.Internal("failed to create secret", err)
	}

	service.logger.Info("secrets.create", slog.String("user_id", userID), slog.String("client_ip", clientIP), slog.Int64("environment_id", envID), slog.String("key", key))

	return &SecretResponse{
		ID:        record.ID,
		Key:       record.Key,
		Value:     value,
		Version:   record.Version,
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

	if err := service.orm.WithContext(ctx).Save(&record).Error; err != nil {
		return nil, errors.Internal("failed to update secret", err)
	}

	service.logger.Info("secrets.update", slog.String("user_id", userID), slog.String("client_ip", clientIP), slog.Int64("environment_id", envID), slog.String("key", key), slog.Int("version", record.Version))

	return &SecretResponse{
		ID:        record.ID,
		Key:       record.Key,
		Value:     value,
		Version:   record.Version,
		CreatedAt: record.CreatedAt,
		UpdatedAt: record.UpdatedAt,
	}, nil
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
	return nil
}
