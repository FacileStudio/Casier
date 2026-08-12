package audit

import (
	"context"
	"log/slog"
	"time"

	"github.com/FacileStudio/Casier/apps/api/schemas"

	"gorm.io/gorm"
)

const (
	writeTimeout   = 10 * time.Second
	maxConcurrency = 64
)

// Logger writes audit rows asynchronously, bounded by a writer pool.
type Logger struct {
	orm    *gorm.DB
	logger *slog.Logger
	slots  chan struct{}
}

// NewLogger wires an audit logger onto the database.
func NewLogger(orm *gorm.DB, logger *slog.Logger) *Logger {
	return &Logger{
		orm:    orm,
		logger: logger,
		slots:  make(chan struct{}, maxConcurrency),
	}
}

// Log writes an audit row in the background. The caller's context is detached
// with context.WithoutCancel so the write outlives the HTTP handler, and every
// failure - including a saturated writer pool - is logged at Error, because a
// silently lost audit row is the one failure a secrets manager cannot afford.
func (l *Logger) Log(ctx context.Context, projectID int64, userID, email, action, resource, detail, clientIP string) {
	entry := &schemas.AuditLog{
		ProjectID: projectID,
		UserID:    userID,
		UserEmail: email,
		Action:    action,
		Resource:  resource,
		Detail:    detail,
		ClientIP:  clientIP,
	}

	select {
	case l.slots <- struct{}{}:
	default:
		l.logger.Error("audit log dropped: writer pool saturated",
			slog.Int64("project_id", projectID),
			slog.String("user_id", userID),
			slog.String("action", action),
			slog.String("resource", resource),
			slog.String("detail", detail))
		return
	}

	writeCtx := context.WithoutCancel(ctx)
	go func() {
		defer func() { <-l.slots }()

		writeCtx, cancel := context.WithTimeout(writeCtx, writeTimeout)
		defer cancel()

		if err := l.orm.WithContext(writeCtx).Create(entry).Error; err != nil {
			l.logger.Error("failed to write audit log",
				slog.Int64("project_id", projectID),
				slog.String("user_id", userID),
				slog.String("action", action),
				slog.String("resource", resource),
				slog.String("detail", detail),
				slog.Any("error", err))
		}
	}()
}
