package audit

import (
	"context"

	"casier/schemas"

	"gorm.io/gorm"
)

type Logger struct {
	orm *gorm.DB
}

func NewLogger(orm *gorm.DB) *Logger {
	return &Logger{orm: orm}
}

func (l *Logger) Log(ctx context.Context, projectID int64, userID, email, action, resource, detail, clientIP string) {
	go func() {
		l.orm.Create(&schemas.AuditLog{
			ProjectID: projectID,
			UserID:    userID,
			UserEmail: email,
			Action:    action,
			Resource:  resource,
			Detail:    detail,
			ClientIP:  clientIP,
		})
	}()
}
