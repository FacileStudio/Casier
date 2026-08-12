package projects

import "time"

// CreateRequest is the body of a project creation.
type CreateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// UpdateRequest is the body of a project update. Its RotationDays is optional:
// a request that omits it leaves the existing rotation policy alone rather
// than silently disabling it.
type UpdateRequest struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	RotationDays *int   `json:"rotation_days"`
}

// AddMemberRequest adds a member to a project with a role.
type AddMemberRequest struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

// ProjectResponse describes one project for the client.
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

// ListResponse is the list of every project the caller can see.
type ListResponse struct {
	Projects []ProjectResponse `json:"projects"`
}

// MemberResponse describes one project member.
type MemberResponse struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

// MembersResponse is the list of a project's members.
type MembersResponse struct {
	Members []MemberResponse `json:"members"`
}

// RemoveMemberResponse explains what removing a member cost, if anything.
type RemoveMemberResponse struct {
	Warning string `json:"warning"`
}

// CreateInviteRequest is the body of an invite creation.
type CreateInviteRequest struct {
	Role    string `json:"role"`
	MaxUses int    `json:"max_uses"`
}

// InviteResponse describes one invite for the client.
type InviteResponse struct {
	ID        int64     `json:"id"`
	Token     string    `json:"token"`
	Role      string    `json:"role"`
	MaxUses   int       `json:"max_uses"`
	UseCount  int       `json:"use_count"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// InvitesResponse is the list of a project's invites.
type InvitesResponse struct {
	Invites []InviteResponse `json:"invites"`
}

// CreateApiTokenRequest mints a machine token. Its ExpiresInDays is a pointer
// so that an older client omitting it still gets the default lifetime rather
// than a token that expires the instant it is minted.
type CreateApiTokenRequest struct {
	Name          string `json:"name"`
	Environments  string `json:"environments"`
	Permissions   string `json:"permissions"`
	ExpiresInDays *int   `json:"expires_in_days"`
}

// ApiTokenResponse describes one machine token.
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

// ApiTokensResponse is the list of a project's tokens.
type ApiTokensResponse struct {
	Tokens []ApiTokenResponse `json:"tokens"`
}

// CreateWebhookRequest is the body of a webhook registration.
type CreateWebhookRequest struct {
	URL    string `json:"url"`
	Events string `json:"events"`
	Secret string `json:"secret"`
}

// WebhookResponse describes one registered webhook.
type WebhookResponse struct {
	ID        int64     `json:"id"`
	URL       string    `json:"url"`
	Events    string    `json:"events"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
}

// WebhooksResponse is the list of a project's webhooks.
type WebhooksResponse struct {
	Webhooks []WebhookResponse `json:"webhooks"`
}

// AuditLogEntry is one audited action.
type AuditLogEntry struct {
	ID        int64     `json:"id"`
	UserEmail string    `json:"user_email"`
	Action    string    `json:"action"`
	Resource  string    `json:"resource"`
	Detail    string    `json:"detail"`
	ClientIP  string    `json:"client_ip"`
	CreatedAt time.Time `json:"created_at"`
}

// AuditLogsResponse is a page of a project's audit trail.
type AuditLogsResponse struct {
	Logs []AuditLogEntry `json:"logs"`
}

// RotateKeyRequest carries the replacement root key.
type RotateKeyRequest struct {
	NewKey string `json:"new_key"`
}

// RotateKeyResponse reports how many keys were re-wrapped.
type RotateKeyResponse struct {
	Rotated int    `json:"rotated"`
	Warning string `json:"warning"`
}
