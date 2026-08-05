package projects

import "time"

type CreateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type UpdateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	// RotationDays is optional: a request that omits it leaves the existing
	// rotation policy alone rather than silently disabling it.
	RotationDays *int `json:"rotation_days"`
}

type AddMemberRequest struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

type ProjectResponse struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	Slug         string    `json:"slug"`
	Description  string    `json:"description"`
	Role         string    `json:"role"`
	RotationDays int       `json:"rotation_days"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type ListResponse struct {
	Projects []ProjectResponse `json:"projects"`
}

type MemberResponse struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

type MembersResponse struct {
	Members []MemberResponse `json:"members"`
}

type RemoveMemberResponse struct {
	Warning string `json:"warning"`
}

type CreateInviteRequest struct {
	Role    string `json:"role"`
	MaxUses int    `json:"max_uses"`
}

type InviteResponse struct {
	ID        int64     `json:"id"`
	Token     string    `json:"token"`
	Role      string    `json:"role"`
	MaxUses   int       `json:"max_uses"`
	UseCount  int       `json:"use_count"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

type InvitesResponse struct {
	Invites []InviteResponse `json:"invites"`
}

// ExpiresInDays is a pointer so that an older client omitting it still gets the
// default lifetime rather than a token that expires the instant it is minted.
type CreateApiTokenRequest struct {
	Name          string `json:"name"`
	Environments  string `json:"environments"`
	Permissions   string `json:"permissions"`
	ExpiresInDays *int   `json:"expires_in_days"`
}

type ApiTokenResponse struct {
	ID           int64      `json:"id"`
	Token        string     `json:"token,omitempty"`
	Name         string     `json:"name"`
	Environments string     `json:"environments"`
	Permissions  string     `json:"permissions"`
	ExpiresAt    time.Time  `json:"expires_at"`
	LastUsedAt   time.Time  `json:"last_used_at"`
	RevokedAt    *time.Time `json:"revoked_at"`
	CreatedAt    time.Time  `json:"created_at"`
}

type ApiTokensResponse struct {
	Tokens []ApiTokenResponse `json:"tokens"`
}

type CreateWebhookRequest struct {
	URL    string `json:"url"`
	Events string `json:"events"`
	Secret string `json:"secret"`
}

type WebhookResponse struct {
	ID        int64     `json:"id"`
	URL       string    `json:"url"`
	Events    string    `json:"events"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
}

type WebhooksResponse struct {
	Webhooks []WebhookResponse `json:"webhooks"`
}

type AuditLogEntry struct {
	ID        int64     `json:"id"`
	UserEmail string    `json:"user_email"`
	Action    string    `json:"action"`
	Resource  string    `json:"resource"`
	Detail    string    `json:"detail"`
	ClientIP  string    `json:"client_ip"`
	CreatedAt time.Time `json:"created_at"`
}

type AuditLogsResponse struct {
	Logs []AuditLogEntry `json:"logs"`
}

type RotateKeyRequest struct {
	NewKey string `json:"new_key"`
}

type RotateKeyResponse struct {
	Rotated int    `json:"rotated"`
	Warning string `json:"warning"`
}
