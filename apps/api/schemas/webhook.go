package schemas

import "time"

type Webhook struct {
	ID        int64     `gorm:"column:id;primaryKey"`
	ProjectID int64     `gorm:"column:project_id;index"`
	Project   Project   `gorm:"constraint:OnDelete:CASCADE"`
	URL       string    `gorm:"column:url"`
	Events    string    `gorm:"column:events"`
	Secret    string    `gorm:"column:secret"`
	Active    bool      `gorm:"column:active;default:true"`
	CreatedBy int64     `gorm:"column:created_by"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (Webhook) TableName() string { return "webhooks" }
