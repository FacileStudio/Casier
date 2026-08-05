package schemas

import "time"

// DefaultTokenDays is how long a machine token lives when the caller does not
// pick a lifetime, and MaxTokenDays is the ceiling. A token that never expires
// is not offered: the whole point of a machine credential is that it lapses.
const (
	DefaultTokenDays = 90
	MaxTokenDays     = 730
)

// AllEnvironments is stored explicitly rather than inferred from an empty
// environment list. Inferring it made "no environments left" and "every
// environment" the same state, so cascading an environment delete would have
// widened a token's reach instead of narrowing it.
type ApiToken struct {
	ID              int64      `gorm:"column:id;primaryKey"`
	TokenHash       string     `gorm:"column:token_hash;uniqueIndex"`
	Name            string     `gorm:"column:name"`
	ProjectID       int64      `gorm:"column:project_id;index"`
	Project         Project    `gorm:"constraint:OnDelete:CASCADE"`
	AllEnvironments bool       `gorm:"column:all_environments;default:false"`
	Permissions     string     `gorm:"column:permissions;default:'read'"`
	CreatedBy       int64      `gorm:"column:created_by"`
	ExpiresAt       time.Time  `gorm:"column:expires_at"`
	LastUsedAt      time.Time  `gorm:"column:last_used_at"`
	RevokedAt       *time.Time `gorm:"column:revoked_at;index"`
	CreatedAt       time.Time  `gorm:"column:created_at;autoCreateTime"`
}

func (ApiToken) TableName() string { return "api_tokens" }

// ApiTokenEnvironment binds a token's scope to environment rows instead of to
// their slugs. Deleting an environment now cascades the grant away; before, the
// slug lingered and recreating an environment with the same name silently
// handed the old tokens their access back.
type ApiTokenEnvironment struct {
	TokenID       int64       `gorm:"column:token_id;primaryKey"`
	Token         ApiToken    `gorm:"constraint:OnDelete:CASCADE"`
	EnvironmentID int64       `gorm:"column:environment_id;primaryKey"`
	Environment   Environment `gorm:"constraint:OnDelete:CASCADE"`
}

func (ApiTokenEnvironment) TableName() string { return "api_token_environments" }
