package schemas

import "time"

type SpaceMember struct {
	ID        int64     `gorm:"column:id;primaryKey"`
	SpaceID   int64     `gorm:"column:space_id;uniqueIndex:idx_space_user"`
	Space     Space     `gorm:"constraint:OnDelete:CASCADE"`
	UserID    int64     `gorm:"column:user_id;uniqueIndex:idx_space_user"`
	Role      string    `gorm:"column:role"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (SpaceMember) TableName() string { return "space_members" }
