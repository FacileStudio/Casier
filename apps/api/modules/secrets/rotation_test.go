package secrets

import (
	"testing"
	"time"
)

func TestRotationStatus(t *testing.T) {
	now := time.Date(2026, 8, 5, 12, 0, 0, 0, time.UTC)
	daysAgo := func(days int) time.Time { return now.AddDate(0, 0, -days) }

	cases := []struct {
		name      string
		rotatedAt time.Time
		days      int
		want      string
	}{
		{"disabled policy never warns", daysAgo(3650), 0, ""},
		{"negative policy never warns", daysAgo(3650), -1, ""},
		{"unknown rotation date never warns", time.Time{}, 90, ""},
		{"fresh secret is ok", daysAgo(1), 90, RotationStatusOK},
		{"one day before the warning window is ok", daysAgo(75), 90, RotationStatusOK},
		{"inside the warning window is due soon", daysAgo(77), 90, RotationStatusDueSoon},
		{"exactly at the deadline is overdue", daysAgo(90), 90, RotationStatusOverdue},
		{"past the deadline is overdue", daysAgo(200), 90, RotationStatusOverdue},
		{"short policies scale the window instead of always warning", daysAgo(1), 8, RotationStatusOK},
		{"short policies still warn near the deadline", daysAgo(7), 8, RotationStatusDueSoon},
		{"long policies cap the window at two weeks", daysAgo(700), 730, RotationStatusOK},
		{"long policies warn within the capped window", daysAgo(720), 730, RotationStatusDueSoon},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			got := rotationStatus(testCase.rotatedAt, testCase.days, now)
			if got != testCase.want {
				t.Fatalf("rotationStatus(%v, %d) = %q, want %q", testCase.rotatedAt, testCase.days, got, testCase.want)
			}
		})
	}
}
