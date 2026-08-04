package schemas

import "time"

type Environment struct {
	ID        int64     `gorm:"column:id;primaryKey"`
	ProjectID int64     `gorm:"column:project_id;uniqueIndex:idx_project_env_slug"`
	Project   Project   `gorm:"constraint:OnDelete:CASCADE"`
	Name      string    `gorm:"column:name"`
	Slug      string    `gorm:"column:slug;uniqueIndex:idx_project_env_slug"`
	Position  int       `gorm:"column:position"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (Environment) TableName() string { return "environments" }
