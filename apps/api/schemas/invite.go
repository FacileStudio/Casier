package schemas

import "time"

type Invite struct {
	ID        int64     `gorm:"column:id;primaryKey"`
	Token     string    `gorm:"column:token;uniqueIndex"`
	ProjectID int64     `gorm:"column:project_id;index"`
	Project   Project   `gorm:"constraint:OnDelete:CASCADE"`
	Role      string    `gorm:"column:role"`
	MaxUses   int       `gorm:"column:max_uses;default:0"`
	UseCount  int       `gorm:"column:use_count;default:0"`
	ExpiresAt time.Time `gorm:"column:expires_at"`
	CreatedBy int64     `gorm:"column:created_by"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (Invite) TableName() string { return "invites" }
