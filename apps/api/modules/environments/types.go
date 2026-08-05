package environments

import "time"

type CreateRequest struct {
	Name      string `json:"name"`
	Protected bool   `json:"protected"`
}

// Protected is a pointer for the same reason Project.rotation_days is: an older
// client that PUTs only {"name": "..."} must leave protection alone rather than
// silently unprotecting production. Absent means "leave it as it is".
type UpdateRequest struct {
	Name      string `json:"name"`
	Protected *bool  `json:"protected"`
}

type EnvironmentResponse struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	Position  int       `json:"position"`
	Protected bool      `json:"protected"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ListResponse struct {
	Environments []EnvironmentResponse `json:"environments"`
}
