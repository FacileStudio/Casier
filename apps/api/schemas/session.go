package schemas

import "time"

const (
	SessionSlidingWindow = 24 * time.Hour
	SessionMaxLifetime   = 7 * 24 * time.Hour
)

type Session struct {
	Token     string    `gorm:"column:token;primaryKey"`
	UserID    int64     `gorm:"column:user_id;index"`
	ExpiresAt time.Time `gorm:"column:expires_at"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (Session) TableName() string { return "sessions" }
