package auth

import (
	"context"
	"log/slog"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/FacileStudio/Casier/apps/api/internal/mailer"
	"github.com/FacileStudio/Casier/apps/api/internal/testdb"
	"github.com/FacileStudio/Casier/apps/api/schemas"
	"github.com/FacileStudio/porte"
	"github.com/FacileStudio/porte/local"
	portepg "github.com/FacileStudio/porte/pg"
	"github.com/FacileStudio/porte/session"

	"gorm.io/gorm"
)

func resetTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	connection, ok := testdb.URL()
	if !ok {
		testdb.Announce()
		t.Skip(testdb.SkipReason)
	}
	db, err := testdb.Open(connection)
	if err != nil {
		t.Fatalf("testdb.Open: %v", err)
	}
	if err := testdb.Truncate(db); err != nil {
		t.Fatalf("testdb.Truncate: %v", err)
	}
	return db
}

func resetTestService(t *testing.T, db *gorm.DB) (*Service, *fakeMailer) {
	t.Helper()
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("db.DB: %v", err)
	}
	store := portepg.New(sqlDB)
	logger := slog.New(slog.DiscardHandler)
	sessions, err := session.New(porte.Config{}, session.Deps{Sessions: store.Sessions(), Logger: logger})
	if err != nil {
		t.Fatalf("session.New: %v", err)
	}
	passwords, err := local.New(local.Config{AllowRegistration: true, MinPasswordLength: 8}, local.Deps{
		Users:      NewUserStore(db),
		Identities: store.Identities(),
		Sessions:   sessions,
		Logger:     logger,
		Count:      NewUserStore(db).CountUsers,
	})
	if err != nil {
		t.Fatalf("local.New: %v", err)
	}
	fake := &fakeMailer{}
	service := NewService(db, nil, logger, fake, "https://casier.test", sessions, passwords, store.Identities())
	return service, fake
}

// tokenFromLastMessage pulls the reset token back out of the link that was
// mailed, which is the only place it ever exists in cleartext.
func tokenFromLastMessage(t *testing.T, fake *fakeMailer) string {
	t.Helper()
	fake.mu.Lock()
	defer fake.mu.Unlock()
	if len(fake.messages) == 0 {
		t.Fatal("no reset email was sent")
	}

	body := fake.messages[len(fake.messages)-1].Body
	start := strings.Index(body, "https://casier.test/reset-password?")
	if start < 0 {
		t.Fatalf("no reset link in the message body:\n%s", body)
	}
	end := start + len(strings.Fields(body[start:])[0])

	parsed, err := url.Parse(body[start:end])
	if err != nil {
		t.Fatalf("url.Parse(%q): %v", body[start:end], err)
	}
	token := parsed.Query().Get("token")
	if token == "" {
		t.Fatalf("reset link carries no token: %s", body[start:end])
	}
	return token
}

func countMessages(fake *fakeMailer) int {
	fake.mu.Lock()
	defer fake.mu.Unlock()
	return len(fake.messages)
}

func TestSendPasswordResetOnlyMailsAddressesThatCanUseTheLink(t *testing.T) {
	db := resetTestDB(t)
	service, fake := resetTestService(t, db)
	ctx := context.Background()

	if _, _, err := service.registerUser(ctx, "known@example.org", "correct horse battery"); err != nil {
		t.Fatalf("registerUser: %v", err)
	}
	if err := db.Create(&schemas.User{Email: "sso@example.org"}).Error; err != nil {
		t.Fatalf("create SSO user: %v", err)
	}

	cases := []struct {
		name      string
		email     string
		wantMail  bool
		wantToken bool
	}{
		{name: "registered password account", email: "known@example.org", wantMail: true, wantToken: true},
		{name: "unknown address", email: "nobody@example.org"},
		{name: "account with no password to reset", email: "sso@example.org"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			before := countMessages(fake)
			var resetsBefore int64
			db.Model(&schemas.PasswordReset{}).Count(&resetsBefore)

			service.sendPasswordReset(ctx, tc.email)

			sent := countMessages(fake) - before
			var resetsAfter int64
			db.Model(&schemas.PasswordReset{}).Count(&resetsAfter)

			if wantSent := boolToInt(tc.wantMail); sent != wantSent {
				t.Errorf("sent %d messages, want %d", sent, wantSent)
			}
			if got := resetsAfter - resetsBefore; got != int64(boolToInt(tc.wantToken)) {
				t.Errorf("created %d reset rows, want %d", got, boolToInt(tc.wantToken))
			}
		})
	}
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func TestPasswordResetTokenExpiresOnTheSchemaTTL(t *testing.T) {
	db := resetTestDB(t)
	service, _ := resetTestService(t, db)
	ctx := context.Background()

	if _, _, err := service.registerUser(ctx, "user@example.org", "correct horse battery"); err != nil {
		t.Fatalf("registerUser: %v", err)
	}

	issued := time.Now()
	service.sendPasswordReset(ctx, "user@example.org")

	var reset schemas.PasswordReset
	if err := db.First(&reset).Error; err != nil {
		t.Fatalf("read reset row: %v", err)
	}
	// issued is read before the row is written, so the stored lifetime is the
	// TTL plus however long the round trip took, never less.
	lifetime := reset.ExpiresAt.Sub(issued)
	if lifetime < schemas.PasswordResetTTL || lifetime > schemas.PasswordResetTTL+time.Minute {
		t.Fatalf("token lifetime = %s, want %s", lifetime, schemas.PasswordResetTTL)
	}
	if reset.Used {
		t.Error("a freshly issued token must not be marked used")
	}
}

func TestResetPasswordConsumesTheTokenAndEvictsEverySession(t *testing.T) {
	db := resetTestDB(t)
	service, fake := resetTestService(t, db)
	ctx := context.Background()

	userID, sessionToken, err := service.registerUser(ctx, "user@example.org", "correct horse battery")
	if err != nil {
		t.Fatalf("registerUser: %v", err)
	}
	if _, _, err := service.authenticateRequest(ctx, "Bearer "+sessionToken); err != nil {
		t.Fatalf("the session under test was not usable to begin with: %v", err)
	}

	service.sendPasswordReset(ctx, "user@example.org")
	token := tokenFromLastMessage(t, fake)

	if err := service.resetPassword(ctx, token, "a brand new password"); err != nil {
		t.Fatalf("resetPassword: %v", err)
	}

	if err := service.resetPassword(ctx, token, "a third password"); err == nil {
		t.Fatal("the reset token was accepted twice")
	}

	var sessions int64
	db.Raw("SELECT count(*) FROM porte_sessions WHERE user_id = ?", userID).Scan(&sessions)
	if sessions != 0 {
		t.Errorf("%d sessions survived the reset, want 0", sessions)
	}
	if _, _, err := service.authenticateRequest(ctx, "Bearer "+sessionToken); err == nil {
		t.Error("the pre-reset session token still authenticates")
	}

	if _, _, err := service.loginUser(ctx, "user@example.org", "a brand new password"); err != nil {
		t.Errorf("the new password does not work: %v", err)
	}
	if _, _, err := service.loginUser(ctx, "user@example.org", "correct horse battery"); err == nil {
		t.Error("the old password still works after a reset")
	}
}

func TestResetPasswordInvalidatesEveryOutstandingToken(t *testing.T) {
	db := resetTestDB(t)
	service, fake := resetTestService(t, db)
	ctx := context.Background()

	if _, _, err := service.registerUser(ctx, "user@example.org", "correct horse battery"); err != nil {
		t.Fatalf("registerUser: %v", err)
	}

	service.sendPasswordReset(ctx, "user@example.org")
	first := tokenFromLastMessage(t, fake)
	service.sendPasswordReset(ctx, "user@example.org")
	second := tokenFromLastMessage(t, fake)

	if first == second {
		t.Fatal("two reset requests produced the same token")
	}

	if err := service.resetPassword(ctx, second, "a brand new password"); err != nil {
		t.Fatalf("resetPassword: %v", err)
	}
	if err := service.resetPassword(ctx, first, "yet another password"); err == nil {
		t.Fatal("an older outstanding token survived a completed reset")
	}

	var outstanding int64
	db.Model(&schemas.PasswordReset{}).Where("used = ?", false).Count(&outstanding)
	if outstanding != 0 {
		t.Errorf("%d reset tokens are still outstanding, want 0", outstanding)
	}
}

func TestResetPasswordRejectsAnExpiredToken(t *testing.T) {
	db := resetTestDB(t)
	service, fake := resetTestService(t, db)
	ctx := context.Background()

	if _, _, err := service.registerUser(ctx, "user@example.org", "correct horse battery"); err != nil {
		t.Fatalf("registerUser: %v", err)
	}

	service.sendPasswordReset(ctx, "user@example.org")
	token := tokenFromLastMessage(t, fake)

	expired := time.Now().Add(-time.Second)
	if err := db.Model(&schemas.PasswordReset{}).Where("used = ?", false).Update("expires_at", expired).Error; err != nil {
		t.Fatalf("age the token: %v", err)
	}

	if err := service.resetPassword(ctx, token, "a brand new password"); err == nil {
		t.Fatal("an expired reset token was accepted")
	}
	if _, _, err := service.loginUser(ctx, "user@example.org", "correct horse battery"); err != nil {
		t.Errorf("the original password stopped working after a failed reset: %v", err)
	}
}

func TestSendPasswordResetStillWritesTheTokenWhenDeliveryFails(t *testing.T) {
	db := resetTestDB(t)
	fake := &fakeMailer{err: context.DeadlineExceeded}
	service, _ := resetTestService(t, db)
	service.mailer = fake
	service.deliverReset = service.sendPasswordReset
	ctx := context.Background()

	if _, _, err := service.registerUser(ctx, "user@example.org", "correct horse battery"); err != nil {
		t.Fatalf("registerUser: %v", err)
	}

	service.sendPasswordReset(ctx, "user@example.org")

	var count int64
	db.Model(&schemas.PasswordReset{}).Count(&count)
	if count != 1 {
		t.Fatalf("%d reset rows, want 1: a relay failure must not take the token with it", count)
	}
}

var _ mailer.Mailer = (*fakeMailer)(nil)
