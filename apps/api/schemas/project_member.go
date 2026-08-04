package schemas

import "time"

type ProjectMember struct {
	ID        int64     `gorm:"column:id;primaryKey"`
	ProjectID int64     `gorm:"column:project_id;uniqueIndex:idx_project_user"`
	Project   Project   `gorm:"constraint:OnDelete:CASCADE"`
	UserID    int64     `gorm:"column:user_id;uniqueIndex:idx_project_user"`
	Role      string    `gorm:"column:role"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (ProjectMember) TableName() string { return "project_members" }
