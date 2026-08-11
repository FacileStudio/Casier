package auth

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"time"

	"github.com/FacileStudio/Casier/apps/api/internal/mailer"
	"github.com/FacileStudio/Casier/apps/api/schemas"
	"github.com/FacileStudio/porte"
	"github.com/FacileStudio/porte/local"
	"github.com/FacileStudio/tronc/errors"

	"gorm.io/gorm"
)

func (service *Service) requireMailer() error {
	if service.mailer == nil {
		return errors.Unavailable("password reset is unavailable, this deployment has no mail delivery configured")
	}
	return nil
}

func (service *Service) forgotPassword(ctx context.Context, email string) error {
	if err := service.requireMailer(); err != nil {
		return err
	}
	normalized := strings.TrimSpace(strings.ToLower(email))
	if normalized == "" || !strings.Contains(normalized, "@") {
		return nil
	}
	go service.deliverReset(context.WithoutCancel(ctx), normalized)
	return nil
}

func (service *Service) sendPasswordReset(ctx context.Context, email string) {
	var user schemas.User
	err := service.orm.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		return
	}

	_, err = service.identities.Find(ctx, porte.ProviderLocal, porte.LocalSubject(user.ID))
	if err != nil {
		return
	}

	token, err := porte.NewToken()
	if err != nil {
		service.logger.Error("failed to generate a password reset token", slog.Int64("user_id", user.ID), slog.Any("error", err))
		return
	}

	if err := service.orm.WithContext(ctx).Create(newPasswordReset(user.ID, porte.HashToken(token), time.Now())).Error; err != nil {
		service.logger.Error("failed to store a password reset token", slog.Int64("user_id", user.ID), slog.Any("error", err))
		return
	}

	if err := service.mailer.Send(ctx, passwordResetMessage(user.Email, resetLink(service.publicURL, token))); err != nil {
		service.logger.Error("failed to deliver a password reset email", slog.Int64("user_id", user.ID), slog.Any("error", err))
	}
}

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

	hashed := porte.HashToken(token)
	var reset schemas.PasswordReset
	err := service.orm.WithContext(ctx).Where("token_hash = ?", hashed).First(&reset).Error
	if err != nil {
		return errors.Invalid("invalid or expired reset token")
	}
	if !passwordResetUsable(reset, time.Now()) {
		return errors.Invalid("invalid or expired reset token")
	}

	hash, err := local.HashPassword(newPassword)
	if err != nil {
		return errors.Invalid("invalid password")
	}

	return service.orm.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		claim := tx.Model(&schemas.PasswordReset{}).
			Where("user_id = ? AND used = ?", reset.UserID, false).
			Update("used", true)
		if claim.Error != nil {
			return errors.Internal("failed to consume reset token", claim.Error)
		}
		if claim.RowsAffected == 0 {
			return errors.Invalid("invalid or expired reset token")
		}

		identity, err := service.identities.Find(ctx, porte.ProviderLocal, porte.LocalSubject(reset.UserID))
		if err != nil {
			return errors.Internal("failed to find identity", err)
		}
		identity.PasswordHash = hash
		if err := service.identities.Save(ctx, identity); err != nil {
			return errors.Internal("failed to update password", err)
		}

		if _, err := service.sessions.RevokeUser(ctx, reset.UserID); err != nil {
			return errors.Internal("failed to clear sessions", err)
		}
		return nil
	})
}
