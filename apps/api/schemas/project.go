package schemas

import "time"

type Project struct {
	ID               int64     `gorm:"column:id;primaryKey"`
	Name             string    `gorm:"column:name"`
	Slug             string    `gorm:"column:slug;uniqueIndex"`
	Description      string    `gorm:"column:description"`
	EncryptedDataKey []byte    `gorm:"column:encrypted_data_key"`
	DataKeyNonce     []byte    `gorm:"column:data_key_nonce"`
	CreatedAt        time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt        time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (Project) TableName() string { return "projects" }
