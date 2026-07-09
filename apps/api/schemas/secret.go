package schemas

import "time"

type Secret struct {
	ID             int64       `gorm:"column:id;primaryKey"`
	EnvironmentID  int64       `gorm:"column:environment_id;uniqueIndex:idx_env_key"`
	Environment    Environment `gorm:"constraint:OnDelete:CASCADE"`
	Key            string      `gorm:"column:key;uniqueIndex:idx_env_key"`
	EncryptedValue []byte      `gorm:"column:encrypted_value"`
	Nonce          []byte      `gorm:"column:nonce;size:12"`
	Version        int         `gorm:"column:version;default:1"`
	CreatedAt      time.Time   `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt      time.Time   `gorm:"column:updated_at;autoUpdateTime"`
}

func (Secret) TableName() string { return "secrets" }
