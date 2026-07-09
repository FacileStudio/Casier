package schemas

import "time"

type SecretVersion struct {
	ID             int64  `gorm:"column:id;primaryKey"`
	SecretID       int64  `gorm:"column:secret_id;index"`
	Secret         Secret `gorm:"constraint:OnDelete:CASCADE"`
	Version        int    `gorm:"column:version"`
	EncryptedValue []byte `gorm:"column:encrypted_value"`
	Nonce          []byte `gorm:"column:nonce;size:12"`
	ChangedBy      string `gorm:"column:changed_by"`
	CreatedAt      time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (SecretVersion) TableName() string { return "secret_versions" }
