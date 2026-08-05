package secrets

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/FacileStudio/Casier/apps/api/schemas"
)

func TestPurgeDue(t *testing.T) {
	now := time.Date(2026, 8, 5, 12, 0, 0, 0, time.UTC)
	retention := time.Duration(schemas.TrashRetentionDays) * 24 * time.Hour

	cases := []struct {
		name      string
		deletedAt time.Time
		want      bool
	}{
		{"deleted a moment ago stays", now.Add(-time.Minute), false},
		{"one day short of the window stays", now.Add(-retention + 24*time.Hour), false},
		{"one second short of the window stays", now.Add(-retention + time.Second), false},
		{"exactly at the window is destroyed", now.Add(-retention), true},
		{"one second past the window is destroyed", now.Add(-retention - time.Second), true},
		{"long past the window is destroyed", now.Add(-10 * retention), true},
		{"a live secret is never due", time.Time{}, false},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			if got := purgeDue(testCase.deletedAt, now); got != testCase.want {
				t.Fatalf("purgeDue(%s, %s) = %t, want %t", testCase.deletedAt, now, got, testCase.want)
			}
		})
	}
}

// TestTrashCutoffMatchesPurgeDue pins the loop's SQL predicate to the same
// decision the per-row helper makes. The sweep selects on deleted_at <= cutoff
// rather than on a computed deadline, and the two must not drift.
func TestTrashCutoffMatchesPurgeDue(t *testing.T) {
	now := time.Date(2026, 8, 5, 12, 0, 0, 0, time.UTC)
	cutoff := trashCutoff(now)

	for _, offset := range []time.Duration{-48 * time.Hour, -time.Second, 0, time.Second, 48 * time.Hour} {
		deletedAt := cutoff.Add(offset)
		selectedBySQL := !deletedAt.After(cutoff)
		if got := purgeDue(deletedAt, now); got != selectedBySQL {
			t.Fatalf("deleted_at %s: purgeDue = %t but the cutoff query would select = %t", deletedAt, got, selectedBySQL)
		}
	}
}

func TestPurgeAtIsTheAdvertisedDeadline(t *testing.T) {
	deletedAt := time.Date(2026, 8, 5, 12, 0, 0, 0, time.UTC)
	want := deletedAt.AddDate(0, 0, schemas.TrashRetentionDays)
	if got := purgeAt(deletedAt); !got.Equal(want) {
		t.Fatalf("purgeAt(%s) = %s, want %s", deletedAt, got, want)
	}
}

func TestRoleMayWriteProtected(t *testing.T) {
	cases := []struct {
		name      string
		role      string
		protected bool
		want      bool
	}{
		{"member writes an ordinary environment", "member", false, true},
		{"viewer is stopped by the project role, not by this", "viewer", false, true},
		{"owner writes a protected environment", roleOwner, true, true},
		{"admin writes a protected environment", roleAdmin, true, true},
		{"member cannot write a protected environment", "member", true, false},
		{"viewer cannot write a protected environment", "viewer", true, false},
		{"an unresolved role cannot write a protected environment", "", true, false},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			if got := roleMayWriteProtected(testCase.role, testCase.protected); got != testCase.want {
				t.Fatalf("roleMayWriteProtected(%q, %t) = %t, want %t", testCase.role, testCase.protected, got, testCase.want)
			}
		})
	}
}

// TestSecretResponseHidesValues is the contract the dashboard and the CLI both
// read: a hidden value is absent from the JSON, so no client can confuse it
// with a secret whose value is genuinely the empty string.
func TestSecretResponseHidesValues(t *testing.T) {
	empty := ""
	filled := "hunter2"

	cases := []struct {
		name  string
		value *string
		want  string
	}{
		{"omitted value has no field at all", nil, `{"id":1,"key":"API_KEY","version":2,"tags":null,"created_at":"0001-01-01T00:00:00Z","updated_at":"0001-01-01T00:00:00Z","rotated_at":"0001-01-01T00:00:00Z","rotation_status":""}`},
		{"an empty value is still sent", &empty, `{"id":1,"key":"API_KEY","value":"","version":2,"tags":null,"created_at":"0001-01-01T00:00:00Z","updated_at":"0001-01-01T00:00:00Z","rotated_at":"0001-01-01T00:00:00Z","rotation_status":""}`},
		{"a revealed value is sent", &filled, `{"id":1,"key":"API_KEY","value":"hunter2","version":2,"tags":null,"created_at":"0001-01-01T00:00:00Z","updated_at":"0001-01-01T00:00:00Z","rotated_at":"0001-01-01T00:00:00Z","rotation_status":""}`},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			encoded, err := json.Marshal(SecretResponse{ID: 1, Key: "API_KEY", Value: testCase.value, Version: 2})
			if err != nil {
				t.Fatalf("marshal failed: %v", err)
			}
			if string(encoded) != testCase.want {
				t.Fatalf("SecretResponse marshalled as %s, want %s", encoded, testCase.want)
			}
		})
	}
}

// TestTrashedSecretNeverCarriesAValue guards the one rule of the trash listing:
// it says what was deleted, never what was in it.
func TestTrashedSecretNeverCarriesAValue(t *testing.T) {
	encoded, err := json.Marshal(TrashedSecretResponse{ID: 1, Key: "API_KEY", DeletedBy: "7"})
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &fields); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if _, found := fields["value"]; found {
		t.Fatalf("trashed secret response carries a value field: %s", encoded)
	}
}

// TestBoolParam covers the edge of both new flags. Revealing plaintext and
// destroying a secret are opt-in, so anything that is not an explicit truth
// must read as false.
func TestBoolParam(t *testing.T) {
	cases := []struct {
		query string
		want  bool
	}{
		{"", false},
		{"?reveal=true", true},
		{"?reveal=1", true},
		{"?reveal=TRUE", true},
		{"?reveal=false", false},
		{"?reveal=0", false},
		{"?reveal=", false},
		{"?reveal=yes", false},
		{"?reveal=please", false},
		{"?tag=prod", false},
	}

	for _, testCase := range cases {
		t.Run("reveal"+testCase.query, func(t *testing.T) {
			request := httptest.NewRequest("GET", "/secrets"+testCase.query, nil)
			if got := boolParam(request, "reveal"); got != testCase.want {
				t.Fatalf("boolParam(%q) = %t, want %t", testCase.query, got, testCase.want)
			}
		})
	}
}

// TestReadDetail checks that the audit log records the size of an exposure in
// values, not in rows, whenever plaintext actually left the server.
func TestReadDetail(t *testing.T) {
	if got := readDetail(12, true); got != "12 values revealed" {
		t.Fatalf("readDetail(12, true) = %q", got)
	}
	if got := readDetail(12, false); got != "12 secrets" {
		t.Fatalf("readDetail(12, false) = %q", got)
	}
}

func TestRevealedValue(t *testing.T) {
	value := "hunter2"
	if got := revealedValue(SecretResponse{Value: &value}); got != value {
		t.Fatalf("revealedValue = %q, want %q", got, value)
	}
	if got := revealedValue(SecretResponse{}); got != "" {
		t.Fatalf("revealedValue of a hidden secret = %q, want empty", got)
	}
}
