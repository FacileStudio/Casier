package mcp

import (
	stderrors "errors"
	"time"

	"github.com/FacileStudio/Casier/apps/api/internal/people"
	"github.com/FacileStudio/Casier/apps/api/modules/environments"
	"github.com/FacileStudio/Casier/apps/api/modules/projects"
	"github.com/FacileStudio/Casier/apps/api/modules/secrets"
	"github.com/FacileStudio/tronc/errors"
)

// stamp renders an instant for a model, in UTC.
func stamp(at time.Time) string {
	if at.IsZero() {
		return ""
	}
	return at.UTC().Format(time.RFC3339)
}

// projectOf renders a project with its environments.
func projectOf(p projects.ProjectResponse, envs []environments.EnvironmentResponse) projectOut {
	out := projectOut{
		Slug: p.Slug, Name: p.Name, Description: p.Description, Role: p.Role, RotationDays: p.RotationDays,
		Environments: make([]envOut, len(envs)),
	}
	for i, env := range envs {
		out.Environments[i] = envOf(env)
	}
	return out
}

// envOf renders an environment.
func envOf(e environments.EnvironmentResponse) envOut {
	return envOut{Slug: e.Slug, Name: e.Name, Protected: e.Protected}
}

// secretOf renders a secret's metadata and drops its value, whatever the response held.
func secretOf(s secrets.SecretResponse, detailed bool) secretOut {
	out := secretOut{
		Key: s.Key, Version: s.Version, Tags: s.Tags, UpdatedAt: stamp(s.UpdatedAt), RotationStatus: s.RotationStatus,
	}
	if detailed {
		out.CreatedAt, out.RotatedAt = stamp(s.CreatedAt), stamp(s.RotatedAt)
	}
	return out
}

// nameOf names a person for a model: their name, else their email, else the id stored.
func nameOf(person *people.Person, fallback string) string {
	switch {
	case person == nil:
		return fallback
	case person.Name != "":
		return person.Name
	case person.Email != "":
		return person.Email
	}
	return fallback
}

// isNotFound reports whether err is a tronc not_found.
func isNotFound(err error) bool {
	var known *errors.Error
	return stderrors.As(err, &known) && known.Code == "not_found"
}
