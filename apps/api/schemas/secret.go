package schemas

import (
	"time"

	"gorm.io/gorm"
)

// TrashRetentionDays is how long a deleted secret stays restorable before the
// purge job destroys it and its versions for good.
const TrashRetentionDays = 30

// Secret is one stored value in an environment. Its unique index on
// (environment_id, key) deliberately ignores deleted_at, so a key in the trash
// stays reserved until it is restored or purged: recreating it under the same
// name would otherwise leave two rows a restore cannot tell apart.
type Secret struct {
	ID             int64          `gorm:"column:id;primaryKey"`
	EnvironmentID  int64          `gorm:"column:environment_id;uniqueIndex:idx_env_key"`
	Environment    Environment    `gorm:"constraint:OnDelete:CASCADE"`
	Key            string         `gorm:"column:key;uniqueIndex:idx_env_key"`
	EncryptedValue []byte         `gorm:"column:encrypted_value"`
	Nonce          []byte         `gorm:"column:nonce;size:12"`
	Version        int            `gorm:"column:version;default:1"`
	Tags           string         `gorm:"column:tags;default:''"`
	CreatedAt      time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt      time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt      gorm.DeletedAt `gorm:"column:deleted_at;index"`
	DeletedBy      string         `gorm:"column:deleted_by;default:''"`
}

func (Secret) TableName() string { return "secrets" }
