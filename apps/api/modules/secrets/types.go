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

type SecretResponse struct {
	ID        int64     `json:"id"`
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	Version   int       `json:"version"`
	Tags      []string  `json:"tags"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ListResponse struct {
	Secrets []SecretResponse `json:"secrets"`
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
