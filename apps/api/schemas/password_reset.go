package schemas

import "time"

type PasswordReset struct {
	Token     string    `gorm:"column:token;primaryKey"`
	UserID    int64     `gorm:"column:user_id;index"`
	ExpiresAt time.Time `gorm:"column:expires_at"`
	Used      bool      `gorm:"column:used;default:false"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (PasswordReset) TableName() string { return "password_resets" }
