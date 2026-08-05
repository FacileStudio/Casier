package secrets

import "time"

type CreateRequest struct {
	Key   string   `json:"key"`
	Value string   `json:"value"`
	Tags  []string `json:"tags"`
}

type UpdateRequest struct {
	Value string `json:"value"`
}

type UpdateTagsRequest struct {
	Tags []string `json:"tags"`
}

// SecretResponse carries a value only when the caller asked for one and was
// allowed to have it. Value is a pointer so an omitted value is absent from the
// JSON entirely: a client can never mistake "hidden" for "the secret is empty".
type SecretResponse struct {
	ID             int64     `json:"id"`
	Key            string    `json:"key"`
	Value          *string   `json:"value,omitempty"`
	Version        int       `json:"version"`
	Tags           []string  `json:"tags"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	RotatedAt      time.Time `json:"rotated_at"`
	RotationStatus string    `json:"rotation_status"`
}

type ListResponse struct {
	Secrets      []SecretResponse `json:"secrets"`
	RotationDays int              `json:"rotation_days"`
}

// TrashedSecretResponse never carries a value. The trash is a recovery listing,
// not a read path: restoring a secret is what puts it back within reach of the
// audited reveal routes.
type TrashedSecretResponse struct {
	ID        int64     `json:"id"`
	Key       string    `json:"key"`
	Tags      []string  `json:"tags"`
	DeletedAt time.Time `json:"deleted_at"`
	DeletedBy string    `json:"deleted_by"`
	PurgeAt   time.Time `json:"purge_at"`
}

type TrashResponse struct {
	Secrets       []TrashedSecretResponse `json:"secrets"`
	RetentionDays int                     `json:"retention_days"`
}

type EnvironmentRotation struct {
	Slug    string `json:"slug"`
	Name    string `json:"name"`
	Overdue int    `json:"overdue"`
	DueSoon int    `json:"due_soon"`
}

type RotationSummaryResponse struct {
	RotationDays int                   `json:"rotation_days"`
	Environments []EnvironmentRotation `json:"environments"`
}

type VersionResponse struct {
	Version   int       `json:"version"`
	ChangedBy string    `json:"changed_by"`
	CreatedAt time.Time `json:"created_at"`
}

type VersionsResponse struct {
	Versions []VersionResponse `json:"versions"`
}

type RollbackRequest struct {
	Version int `json:"version"`
}

type ImportRequest struct {
	Content string `json:"content"`
}

type ImportResponse struct {
	Created int `json:"created"`
	Updated int `json:"updated"`
	Skipped int `json:"skipped"`
}

type ExportResponse struct {
	Content string `json:"content"`
}

type DiffEntry struct {
	Key    string `json:"key"`
	Status string `json:"status"`
	Left   string `json:"left,omitempty"`
	Right  string `json:"right,omitempty"`
}

type DiffResponse struct {
	From    string      `json:"from"`
	To      string      `json:"to"`
	Entries []DiffEntry `json:"entries"`
}

type SetACLRequest struct {
	Email    string `json:"email"`
	CanRead  bool   `json:"can_read"`
	CanWrite bool   `json:"can_write"`
}

type ACLResponse struct {
	ID       int64  `json:"id"`
	Email    string `json:"email"`
	CanRead  bool   `json:"can_read"`
	CanWrite bool   `json:"can_write"`
}

type ACLsResponse struct {
	ACLs []ACLResponse `json:"acls"`
}
