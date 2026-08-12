package schemas

import "time"

// DefaultRotationDays is the rotation policy a new project starts with: warn
// once a secret value has gone three months without being changed.
const DefaultRotationDays = 90

// MaxRotationDays keeps a policy inside a decade, and zero disables it.
const MaxRotationDays = 3650

// Project is one project, the container for environments and secrets.
type Project struct {
	ID               int64     `gorm:"column:id;primaryKey"`
	Name             string    `gorm:"column:name"`
	Slug             string    `gorm:"column:slug;uniqueIndex"`
	Description      string    `gorm:"column:description"`
	EncryptedDataKey []byte    `gorm:"column:encrypted_data_key"`
	DataKeyNonce     []byte    `gorm:"column:data_key_nonce"`
	RotationDays     int       `gorm:"column:rotation_days;default:90"`
	CreatedAt        time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt        time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (Project) TableName() string { return "projects" }
