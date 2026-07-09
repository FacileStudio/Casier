package schemas

import "time"

type AuditLog struct {
	ID        int64     `gorm:"column:id;primaryKey"`
	SpaceID   int64     `gorm:"column:space_id;index"`
	UserID    string    `gorm:"column:user_id"`
	UserEmail string    `gorm:"column:user_email"`
	Action    string    `gorm:"column:action"`
	Resource  string    `gorm:"column:resource"`
	Detail    string    `gorm:"column:detail"`
	ClientIP  string    `gorm:"column:client_ip"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime;index"`
}

func (AuditLog) TableName() string { return "audit_logs" }
