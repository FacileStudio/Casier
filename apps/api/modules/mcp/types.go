package mcp

// envRef names one environment of one project.
type envRef struct {
	Project     string `json:"project" jsonschema:"the project slug, from casier_list_projects"`
	Environment string `json:"environment" jsonschema:"the environment slug, such as dev, staging or prod, from casier_list_projects"`
}

// secretRef names one secret of one environment.
type secretRef struct {
	envRef
	Key string `json:"key" jsonschema:"the secret's key, such as DATABASE_URL, from casier_list_secrets"`
}

// projectsInput is casier_list_projects's input.
type projectsInput struct {
	Query string `json:"query,omitempty" jsonschema:"keep projects whose name or slug contains this, case ignored"`
}

// secretsInput is casier_list_secrets's input.
type secretsInput struct {
	envRef
	Query          string `json:"query,omitempty" jsonschema:"keep keys that contain this, case ignored"`
	Tag            string `json:"tag,omitempty" jsonschema:"keep secrets carrying this tag"`
	Cursor         string `json:"cursor,omitempty" jsonschema:"next_cursor from the previous page"`
	ResponseFormat string `json:"response_format,omitempty" jsonschema:"concise (default) or detailed, which adds created_at and rotated_at"`
}

// diffInput is casier_diff_environments's input.
type diffInput struct {
	Project string `json:"project" jsonschema:"the project slug, from casier_list_projects"`
	From    string `json:"from" jsonschema:"the first environment slug"`
	To      string `json:"to" jsonschema:"the second environment slug"`
}

// auditInput is casier_list_audit_events's input.
type auditInput struct {
	Project string `json:"project" jsonschema:"the project slug, from casier_list_projects"`
	Action  string `json:"action,omitempty" jsonschema:"keep one action, such as secret.reveal or secret.update"`
	Limit   int    `json:"limit,omitempty" jsonschema:"how many events, newest first, 1 to 100. Default 50"`
}

// createProjectInput is casier_create_project's input.
type createProjectInput struct {
	Name        string `json:"name" jsonschema:"the project's name; its slug is derived from it"`
	Description string `json:"description,omitempty" jsonschema:"what the project holds secrets for"`
}

// createEnvironmentInput is casier_create_environment's input.
type createEnvironmentInput struct {
	Project   string `json:"project" jsonschema:"the project slug, from casier_list_projects"`
	Name      string `json:"name" jsonschema:"the environment's name, such as staging; its slug is derived from it"`
	Protected bool   `json:"protected,omitempty" jsonschema:"only owners and admins may then write to it"`
}

// setInput is casier_set_secret's input.
type setInput struct {
	secretRef
	Value string   `json:"value" jsonschema:"the plaintext value, at most 64 KB; Casier encrypts it and never returns it"`
	Tags  []string `json:"tags,omitempty" jsonschema:"tags for a new secret; an existing secret keeps its tags (use casier_tag_secret)"`
}

// tagInput is casier_tag_secret's input.
type tagInput struct {
	secretRef
	Tags []string `json:"tags" jsonschema:"the secret's whole tag list; an empty list clears it"`
}

// rollbackInput is casier_rollback_secret's input.
type rollbackInput struct {
	secretRef
	Version int `json:"version" jsonschema:"the version to bring back, from casier_get_secret"`
}

// revealInput is casier_reveal_secrets's input.
type revealInput struct {
	envRef
	Keys           []string `json:"keys,omitempty" jsonschema:"the keys to read, 1 to 50, from casier_list_secrets"`
	ConfirmationID string   `json:"confirmation_id,omitempty" jsonschema:"the confirmation_id of an earlier call the user has since confirmed in Casier; the other arguments are then ignored"`
}

// projectOut is one project the caller is a member of.
type projectOut struct {
	Slug         string   `json:"slug"`
	Name         string   `json:"name"`
	Description  string   `json:"description,omitempty"`
	Role         string   `json:"role"`
	RotationDays int      `json:"rotation_days"`
	Environments []envOut `json:"environments"`
}

// envOut is one environment.
type envOut struct {
	Slug      string `json:"slug"`
	Name      string `json:"name"`
	Protected bool   `json:"protected,omitempty"`
}

// projectsOutput answers casier_list_projects.
type projectsOutput struct {
	Projects []projectOut `json:"projects"`
}

// secretOut is a secret's metadata. It never carries the value.
type secretOut struct {
	Key            string   `json:"key"`
	Version        int      `json:"version"`
	Tags           []string `json:"tags,omitempty"`
	UpdatedAt      string   `json:"updated_at"`
	RotationStatus string   `json:"rotation_status,omitempty"`
	CreatedAt      string   `json:"created_at,omitempty"`
	RotatedAt      string   `json:"rotated_at,omitempty"`
}

// secretsOutput answers casier_list_secrets.
type secretsOutput struct {
	Project      string      `json:"project"`
	Environment  string      `json:"environment"`
	RotationDays int         `json:"rotation_days"`
	Total        int         `json:"total"`
	Secrets      []secretOut `json:"secrets"`
	NextCursor   string      `json:"next_cursor,omitempty"`
}

// versionOut is one stored version of a secret, without its value.
type versionOut struct {
	Version   int    `json:"version"`
	ChangedBy string `json:"changed_by"`
	CreatedAt string `json:"created_at"`
}

// secretOutput answers casier_get_secret.
type secretOutput struct {
	Secret   secretOut    `json:"secret"`
	Versions []versionOut `json:"versions"`
}

// diffOut is one key compared across two environments, without values.
type diffOut struct {
	Key    string `json:"key"`
	Status string `json:"status"`
}

// diffOutput answers casier_diff_environments.
type diffOutput struct {
	From    string    `json:"from"`
	To      string    `json:"to"`
	Entries []diffOut `json:"entries"`
}

// trashOut is one secret in the trash.
type trashOut struct {
	Key       string   `json:"key"`
	Tags      []string `json:"tags,omitempty"`
	DeletedAt string   `json:"deleted_at"`
	DeletedBy string   `json:"deleted_by"`
	PurgeAt   string   `json:"purge_at"`
}

// trashOutput answers casier_list_trash.
type trashOutput struct {
	RetentionDays int        `json:"retention_days"`
	Secrets       []trashOut `json:"secrets"`
}

// auditOut is one audit event.
type auditOut struct {
	Action    string `json:"action"`
	Resource  string `json:"resource"`
	Detail    string `json:"detail,omitempty"`
	Person    string `json:"person"`
	CreatedAt string `json:"created_at"`
}

// auditOutput answers casier_list_audit_events.
type auditOutput struct {
	Events []auditOut `json:"events"`
}

// writeOutput answers the tools that change one thing.
type writeOutput struct {
	Status      string      `json:"status"`
	Project     *projectOut `json:"project,omitempty"`
	Environment *envOut     `json:"environment,omitempty"`
	Secret      *secretOut  `json:"secret,omitempty"`
	Message     string      `json:"message,omitempty"`
}

// revealedOut is one secret's value, read once.
type revealedOut struct {
	Key     string `json:"key"`
	Version int    `json:"version"`
	Value   string `json:"value"`
}

// revealOutput answers casier_reveal_secrets.
type revealOutput struct {
	Status         string        `json:"status"`
	Project        string        `json:"project,omitempty"`
	Environment    string        `json:"environment,omitempty"`
	Secrets        []revealedOut `json:"secrets,omitempty"`
	Omitted        []string      `json:"omitted,omitempty"`
	ConfirmURL     string        `json:"confirm_url,omitempty"`
	ConfirmationID string        `json:"confirmation_id,omitempty"`
	Message        string        `json:"message,omitempty"`
}

// deleteOutput answers casier_delete_secret.
type deleteOutput struct {
	Status     string `json:"status"`
	Key        string `json:"key,omitempty"`
	ConfirmURL string `json:"confirm_url,omitempty"`
	Message    string `json:"message,omitempty"`
}
