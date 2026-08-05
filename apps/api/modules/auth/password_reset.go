package auth

import (
	"context"
	stderrors "errors"
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"time"

	"github.com/FacileStudio/Casier/apps/api/internal/authcrypto"
	"github.com/FacileStudio/Casier/apps/api/internal/mailer"
	"github.com/FacileStudio/Casier/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"

	"gorm.io/gorm"
)

// requireMailer keeps the flow honest when no relay is configured. Answering
// 200 without a way to deliver the link is what made this endpoint a silent
// no-op in the first place, so both halves refuse instead.
func (service *Service) requireMailer() error {
	if service.mailer == nil {
		return errors.Unavailable("password reset is unavailable, this deployment has no mail delivery configured")
	}
	return nil
}

// forgotPassword answers before it knows whether the address exists. The
// lookup, the token write and the SMTP round trip all run on a detached
// context, so a caller cannot separate a registered address from an unknown one
// by status, body or clock — the request path never sees the answer either.
// That is OWASP's forgot-password guidance, and it is why nothing here branches
// on the email.
func (service *Service) forgotPassword(ctx context.Context, email string) error {
	if err := service.requireMailer(); err != nil {
		return err
	}
	go service.deliverReset(context.WithoutCancel(ctx), email)
	return nil
}

// sendPasswordReset runs after the response has already gone out, so every
// failure below is logged rather than returned: an operator needs to see a
// bounced relay, and the caller must not be told whether there was anything to
// deliver.
func (service *Service) sendPasswordReset(ctx context.Context, email string) {
	var user schemas.User
	err := service.orm.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return
	}
	if err != nil {
		service.logger.Error("password reset lookup failed", slog.Any("error", err))
		return
	}
	if user.PasswordHash == "" {
		return
	}

	token, err := authcrypto.NewToken()
	if err != nil {
		service.logger.Error("failed to generate a password reset token", slog.Int64("user_id", user.ID), slog.Any("error", err))
		return
	}

	if err := service.orm.WithContext(ctx).Create(newPasswordReset(user.ID, hashToken(token), time.Now())).Error; err != nil {
		service.logger.Error("failed to store a password reset token", slog.Int64("user_id", user.ID), slog.Any("error", err))
		return
	}

	if err := service.mailer.Send(ctx, passwordResetMessage(user.Email, resetLink(service.publicURL, token))); err != nil {
		service.logger.Error("failed to deliver a password reset email", slog.Int64("user_id", user.ID), slog.Any("error", err))
	}
}

// newPasswordReset stamps the expiry from schemas.PasswordResetTTL rather than
// an ad-hoc window, so the lifetime lives next to the table that stores it.
func newPasswordReset(userID int64, tokenHash string, now time.Time) *schemas.PasswordReset {
	return &schemas.PasswordReset{
		TokenHash: tokenHash,
		UserID:    userID,
		ExpiresAt: now.Add(schemas.PasswordResetTTL),
	}
}

func passwordResetUsable(reset schemas.PasswordReset, now time.Time) bool {
	return !reset.Used && now.Before(reset.ExpiresAt)
}

// resetLink points at the client's /reset-password route, which reads the token
// out of the query string.
func resetLink(publicURL string, token string) string {
	return strings.TrimSuffix(publicURL, "/") + "/reset-password?token=" + url.QueryEscape(token)
}

func passwordResetMessage(email string, link string) mailer.Message {
	return mailer.Message{
		To:      email,
		Subject: "Reset your Casier password",
		Body: fmt.Sprintf(`Someone asked to reset the password on your Casier account.

Open this link to choose a new one:

%s

The link works once and expires in %d minutes. If this was not you, ignore this
email: nothing has changed, and no one can reset your password without the link.
`, link, int(schemas.PasswordResetTTL.Minutes())),
	}
}

func (service *Service) resetPassword(ctx context.Context, token string, newPassword string) error {
	if err := service.requireMailer(); err != nil {
		return err
	}

	hashed := hashToken(token)
	var reset schemas.PasswordReset
	err := service.orm.WithContext(ctx).Where("token_hash = ?", hashed).First(&reset).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return errors.Invalid("invalid or expired reset token")
	}
	if err != nil {
		return errors.Internal("failed to validate reset token", err)
	}
	if !passwordResetUsable(reset, time.Now()) {
		return errors.Invalid("invalid or expired reset token")
	}

	hash, err := authcrypto.HashPassword(newPassword)
	if err != nil {
		return errors.Invalid("invalid password")
	}

	return service.orm.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// One statement burns the token that was presented and every other
		// outstanding one for that user: a reset invalidates the whole queue, and
		// the affected-row count is what makes the claim atomic, so two requests
		// racing on the same token cannot both win.
		claim := tx.Model(&schemas.PasswordReset{}).
			Where("user_id = ? AND used = ?", reset.UserID, false).
			Update("used", true)
		if claim.Error != nil {
			return errors.Internal("failed to consume reset token", claim.Error)
		}
		if claim.RowsAffected == 0 {
			return errors.Invalid("invalid or expired reset token")
		}
		if err := tx.Model(&schemas.User{}).Where("id = ?", reset.UserID).Update("password_hash", hash).Error; err != nil {
			return errors.Internal("failed to update password", err)
		}
		// Every session dies with the old password: a reset is the one moment a
		// user has to be able to evict whoever else was holding the account.
		if err := tx.Where("user_id = ?", reset.UserID).Delete(&schemas.Session{}).Error; err != nil {
			return errors.Internal("failed to clear sessions", err)
		}
		return nil
	})
}
