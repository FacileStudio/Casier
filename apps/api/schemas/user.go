package schemas

import "time"

// User is the local identity row.
type User struct {
	ID        int64     `gorm:"column:id;primaryKey"`
	Email     string    `gorm:"column:email;uniqueIndex"`
	Name      string    `gorm:"column:name"`
	AvatarURL string    `gorm:"column:avatar_url"`
	Locale    string    `gorm:"column:locale;default:en"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (User) TableName() string { return "users" }
