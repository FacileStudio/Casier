package schemas

import "time"

type User struct {
	ID                        int64     `gorm:"column:id;primaryKey"`
	Email                     string    `gorm:"column:email;uniqueIndex"`
	Name                      string    `gorm:"column:name"`
	PasswordHash              string    `gorm:"column:password_hash"`
	OIDCAccessTokenEncrypted  []byte    `gorm:"column:oidc_access_token_enc"`
	OIDCAccessTokenNonce      []byte    `gorm:"column:oidc_access_token_nonce"`
	OIDCRefreshTokenEncrypted []byte    `gorm:"column:oidc_refresh_token_enc"`
	OIDCRefreshTokenNonce     []byte    `gorm:"column:oidc_refresh_token_nonce"`
	OIDCTokenExpiry           time.Time `gorm:"column:oidc_token_expiry"`
	ProfileSyncedAt           time.Time `gorm:"column:profile_synced_at"`
	CreatedAt                 time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt                 time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (User) TableName() string { return "users" }
