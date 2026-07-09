package schemas

import "time"

type Environment struct {
	ID        int64     `gorm:"column:id;primaryKey"`
	SpaceID   int64     `gorm:"column:space_id;uniqueIndex:idx_space_env_slug"`
	Space     Space     `gorm:"constraint:OnDelete:CASCADE"`
	Name      string    `gorm:"column:name"`
	Slug      string    `gorm:"column:slug;uniqueIndex:idx_space_env_slug"`
	Position  int       `gorm:"column:position"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (Environment) TableName() string { return "environments" }
