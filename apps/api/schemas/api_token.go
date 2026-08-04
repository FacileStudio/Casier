package schemas

import "time"

type ApiToken struct {
	ID           int64     `gorm:"column:id;primaryKey"`
	TokenHash    string    `gorm:"column:token_hash;uniqueIndex"`
	Name         string    `gorm:"column:name"`
	ProjectID    int64     `gorm:"column:project_id;index"`
	Project      Project   `gorm:"constraint:OnDelete:CASCADE"`
	Environments string    `gorm:"column:environments"`
	Permissions  string    `gorm:"column:permissions;default:'read'"`
	CreatedBy    int64     `gorm:"column:created_by"`
	ExpiresAt    time.Time `gorm:"column:expires_at"`
	LastUsedAt   time.Time `gorm:"column:last_used_at"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (ApiToken) TableName() string { return "api_tokens" }
