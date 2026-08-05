package secrets

import "time"

const (
	RotationStatusOK      = "ok"
	RotationStatusDueSoon = "due_soon"
	RotationStatusOverdue = "overdue"
)

const maxRotationWarning = 14 * 24 * time.Hour

// rotationStatus grades a secret against its project's rotation policy.
// rotatedAt is the last time the *value* changed, not the last time the row was
// touched. A policy of zero days disables the warning entirely and yields "".
func rotationStatus(rotatedAt time.Time, rotationDays int, now time.Time) string {
	if rotationDays <= 0 || rotatedAt.IsZero() {
		return ""
	}

	period := time.Duration(rotationDays) * 24 * time.Hour
	due := rotatedAt.Add(period)
	if !now.Before(due) {
		return RotationStatusOverdue
	}

	warning := period / 4
	if warning > maxRotationWarning {
		warning = maxRotationWarning
	}
	if !now.Before(due.Add(-warning)) {
		return RotationStatusDueSoon
	}

	return RotationStatusOK
}
