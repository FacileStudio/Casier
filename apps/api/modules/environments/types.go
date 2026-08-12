package environments

import "time"

// CreateRequest is the body of an environment creation.
type CreateRequest struct {
	Name      string `json:"name"`
	Protected bool   `json:"protected"`
}

// UpdateRequest is an environment update. Its Protected field is a pointer for
// the same reason Project.rotation_days is: an older client that PUTs only
// {"name": "..."} must leave protection alone rather than silently
// unprotecting production — absent means "leave it as it is".
type UpdateRequest struct {
	Name      string `json:"name"`
	Protected *bool  `json:"protected"`
}

// EnvironmentResponse describes one environment for the client.
type EnvironmentResponse struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	Position  int       `json:"position"`
	Protected bool      `json:"protected"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ListResponse is the list of every environment in a project, in order.
type ListResponse struct {
	Environments []EnvironmentResponse `json:"environments"`
}
