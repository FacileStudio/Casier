package schemas

import "time"

// PasswordResetTTL follows the OWASP forgot-password guidance: short-lived,
// single use, and stored only as a SHA-256 digest so a database read cannot be
// replayed into an account takeover.
const PasswordResetTTL = 30 * time.Minute

type PasswordReset struct {
	TokenHash string    `gorm:"column:token_hash;primaryKey"`
	UserID    int64     `gorm:"column:user_id;index"`
	ExpiresAt time.Time `gorm:"column:expires_at"`
	Used      bool      `gorm:"column:used;default:false"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (PasswordReset) TableName() string { return "password_resets" }
