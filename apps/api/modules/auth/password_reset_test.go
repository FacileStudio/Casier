package auth

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/FacileStudio/Casier/apps/api/internal/mailer"
	"github.com/FacileStudio/Casier/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"
)

type fakeMailer struct {
	mu       sync.Mutex
	messages []mailer.Message
	err      error
}

func (f *fakeMailer) Send(_ context.Context, message mailer.Message) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.messages = append(f.messages, message)
	return f.err
}

func testService(mail mailer.Mailer) *Service {
	service := &Service{
		logger:    slog.New(slog.DiscardHandler),
		mailer:    mail,
		publicURL: "https://casier.facile.studio",
	}
	service.controller = newController(service)
	return service
}

func TestNewPasswordResetUsesTheSchemaTTL(t *testing.T) {
	now := time.Date(2026, 8, 5, 10, 0, 0, 0, time.UTC)

	reset := newPasswordReset(42, "digest", now)
	if reset.UserID != 42 {
		t.Errorf("UserID = %d, want 42", reset.UserID)
	}
	if reset.TokenHash != "digest" {
		t.Errorf("TokenHash = %q, want %q", reset.TokenHash, "digest")
	}
	if reset.Used {
		t.Error("a fresh reset must not be marked used")
	}
	if got := reset.ExpiresAt.Sub(now); got != schemas.PasswordResetTTL {
		t.Errorf("lifetime = %s, want %s", got, schemas.PasswordResetTTL)
	}
	if schemas.PasswordResetTTL > 30*time.Minute {
		t.Errorf("PasswordResetTTL = %s, OWASP guidance caps this flow at 30 minutes", schemas.PasswordResetTTL)
	}
}

func TestPasswordResetUsable(t *testing.T) {
	now := time.Date(2026, 8, 5, 10, 0, 0, 0, time.UTC)

	cases := []struct {
		name  string
		reset schemas.PasswordReset
		want  bool
	}{
		{
			name:  "fresh and unused",
			reset: schemas.PasswordReset{ExpiresAt: now.Add(time.Minute)},
			want:  true,
		},
		{
			name:  "issued a moment ago",
			reset: *newPasswordReset(1, "digest", now),
			want:  true,
		},
		{
			name:  "already consumed",
			reset: schemas.PasswordReset{ExpiresAt: now.Add(time.Minute), Used: true},
			want:  false,
		},
		{
			name:  "expired",
			reset: schemas.PasswordReset{ExpiresAt: now.Add(-time.Second)},
			want:  false,
		},
		{
			name:  "expiring exactly now",
			reset: schemas.PasswordReset{ExpiresAt: now},
			want:  false,
		},
		{
			name:  "consumed and expired",
			reset: schemas.PasswordReset{ExpiresAt: now.Add(-time.Hour), Used: true},
			want:  false,
		},
		{
			name:  "issued one tick past the TTL",
			reset: *newPasswordReset(1, "digest", now.Add(-schemas.PasswordResetTTL-time.Nanosecond)),
			want:  false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := passwordResetUsable(tc.reset, now); got != tc.want {
				t.Errorf("passwordResetUsable = %t, want %t", got, tc.want)
			}
		})
	}
}

func TestResetLink(t *testing.T) {
	cases := []struct {
		name      string
		publicURL string
		token     string
		want      string
	}{
		{
			name:      "plain origin",
			publicURL: "https://casier.facile.studio",
			token:     "abc123",
			want:      "https://casier.facile.studio/reset-password?token=abc123",
		},
		{
			name:      "trailing slash is not doubled",
			publicURL: "https://casier.facile.studio/",
			token:     "abc123",
			want:      "https://casier.facile.studio/reset-password?token=abc123",
		},
		{
			name:      "token characters are escaped into the query",
			publicURL: "https://casier.facile.studio",
			token:     "a+b/c=d&e",
			want:      "https://casier.facile.studio/reset-password?token=a%2Bb%2Fc%3Dd%26e",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := resetLink(tc.publicURL, tc.token); got != tc.want {
				t.Errorf("resetLink = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestPasswordResetMessageCarriesTheLinkAndNothingElse(t *testing.T) {
	link := resetLink("https://casier.facile.studio", "s3cret-token")
	message := passwordResetMessage("user@example.org", link)

	if message.To != "user@example.org" {
		t.Errorf("To = %q, want %q", message.To, "user@example.org")
	}
	if message.Subject == "" {
		t.Error("the message needs a subject")
	}
	if strings.Contains(message.Subject, "s3cret-token") {
		t.Error("the token must not travel in the subject line")
	}
	if !strings.Contains(message.Body, link) {
		t.Errorf("body does not carry the reset link:\n%s", message.Body)
	}
	if !strings.Contains(message.Body, "30 minutes") {
		t.Errorf("body does not state the lifetime:\n%s", message.Body)
	}
}

// The enumeration guarantee is structural: forgotPassword hands the address to
// the delivery path and returns, so there is no lookup on the request path to
// branch on. The service here has a nil database precisely to prove that — if
// anyone reintroduces an early return for unknown addresses, this panics.
func TestForgotPasswordAnswersIdenticallyForEveryAddress(t *testing.T) {
	cases := []struct {
		name  string
		email string
	}{
		{name: "registered address", email: "known@example.org"},
		{name: "unknown address", email: "nobody@example.org"},
		{name: "address that never had a password", email: "sso-only@example.org"},
		{name: "address of another tenant", email: "someone@example.net"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			handed := make(chan string, 1)
			service := testService(&fakeMailer{})
			service.deliverReset = func(_ context.Context, email string) { handed <- email }

			if err := service.forgotPassword(context.Background(), tc.email); err != nil {
				t.Fatalf("forgotPassword(%q) = %v, want a nil error for every address", tc.email, err)
			}

			select {
			case got := <-handed:
				if got != tc.email {
					t.Fatalf("delivery received %q, want %q", got, tc.email)
				}
			case <-time.After(2 * time.Second):
				t.Fatal("forgotPassword never handed the address to the delivery path")
			}
		})
	}
}

func TestForgotPasswordDetachesDeliveryFromTheRequest(t *testing.T) {
	handed := make(chan context.Context, 1)
	service := testService(&fakeMailer{})
	service.deliverReset = func(ctx context.Context, _ string) { handed <- ctx }

	ctx, cancel := context.WithCancel(context.Background())
	if err := service.forgotPassword(ctx, "user@example.org"); err != nil {
		t.Fatalf("forgotPassword: %v", err)
	}

	var delivery context.Context
	select {
	case delivery = <-handed:
	case <-time.After(2 * time.Second):
		t.Fatal("forgotPassword never handed the address to the delivery path")
	}

	cancel()
	if err := delivery.Err(); err != nil {
		t.Fatalf("delivery context died with the request: %v", err)
	}
}

func TestPasswordResetIsUnavailableWithoutAMailer(t *testing.T) {
	service := testService(nil)

	cases := []struct {
		name string
		call func() error
	}{
		{
			name: "forgot password",
			call: func() error { return service.forgotPassword(context.Background(), "user@example.org") },
		},
		{
			name: "reset password",
			call: func() error {
				return service.resetPassword(context.Background(), "some-token", "a-long-enough-password")
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.call()
			if err == nil {
				t.Fatal("the flow answered as if it worked with no mailer configured")
			}
			if got := errors.Status(err); got != http.StatusServiceUnavailable {
				t.Fatalf("status = %d, want %d", got, http.StatusServiceUnavailable)
			}
		})
	}
}

func TestForgotPasswordControllerValidatesTheAddress(t *testing.T) {
	cases := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{name: "plain address", email: "user@example.org"},
		{name: "address is lowercased and trimmed", email: "  User@Example.org  "},
		{name: "empty", email: "", wantErr: true},
		{name: "no at sign", email: "user.example.org", wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			handed := make(chan string, 1)
			service := testService(&fakeMailer{})
			service.deliverReset = func(_ context.Context, email string) { handed <- email }

			err := service.controller.forgotPassword(context.Background(), &ForgotPasswordRequest{Email: tc.email})
			if tc.wantErr {
				if err == nil {
					t.Fatal("controller accepted a malformed address")
				}
				return
			}
			if err != nil {
				t.Fatalf("controller.forgotPassword: %v", err)
			}
			select {
			case got := <-handed:
				if got != strings.ToLower(strings.TrimSpace(tc.email)) {
					t.Fatalf("delivery received %q, want the normalized address", got)
				}
			case <-time.After(2 * time.Second):
				t.Fatal("controller never handed the address to the delivery path")
			}
		})
	}
}
