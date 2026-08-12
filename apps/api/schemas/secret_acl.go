package schemas

import "time"

// SecretACL is one user's override on a single secret.
type SecretACL struct {
	ID        int64     `gorm:"column:id;primaryKey"`
	SecretID  int64     `gorm:"column:secret_id;uniqueIndex:idx_secret_user"`
	Secret    Secret    `gorm:"constraint:OnDelete:CASCADE"`
	UserID    int64     `gorm:"column:user_id;uniqueIndex:idx_secret_user"`
	CanRead   bool      `gorm:"column:can_read;default:true"`
	CanWrite  bool      `gorm:"column:can_write;default:false"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (SecretACL) TableName() string { return "secret_acls" }
