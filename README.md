# Casier

Self-hosted secrets and environment variable manager for teams. Go API, SvelteKit dashboard,
Rust CLI.

Secret values are encrypted at rest with a per-project AES-256-GCM data key, wrapped by a
root key the process never writes down. Projects hold environments, environments hold
secrets, and every value change is versioned. Formerly named Clef.

Live at [casier.facile.studio](https://casier.facile.studio).

## What it does

- Encrypts every secret value with a per-project AES-256-GCM data key
- Groups secrets by project and environment, with slugs stable enough to script against
- Versions every value change, and rolls a secret back to any earlier version
- Soft-deletes to a trash with a 30-day restore window before the purge job runs
- Scoped machine tokens bound to a project, specific environments, and read or write
- Per-secret ACLs on top of project roles, plus an append-only audit log
- Flags secrets that have gone unrotated past the project's rotation policy
- Email and password accounts with password reset, or OIDC SSO with an `SSO_ONLY` mode
- Project invites, membership roles, and outbound webhooks on secret changes
- A Rust CLI that injects secrets into a process, diffs environments, and gates CI

## Stack

| Layer | Tech |
|---|---|
| API | Go 1.25, Chi v5, GORM, PostgreSQL 16, [tronc](https://github.com/FacileStudio/tronc) v0.8.0 |
| Client | SvelteKit 2, Svelte 5 (runes), Tailwind CSS 4, muse (`@facile/lib`), bits-ui |
| CLI | Rust 2021, clap, OS keychain token storage |
| Deploy | Docker Compose, single distroless container behind Traefik |

## Quick start

```sh
cp .env.example .env
# set ENCRYPTION_KEY: openssl rand -base64 32
docker compose up --build
```

The API serves the dashboard and its own routes from one container on port `4000`.

### Local development

```sh
mise trust && mise install
mise run bootstrap
mise run db-up
mise run dev
```

`mise run db-up` starts a repo-local Postgres on port `5433` under `.local/postgres`; no
Docker needed for the database. `mise run dev` runs the API and the client dev server
together. Stop the database again with `mise run db-down`.

### CLI

```sh
cd cli && cargo build --release
```

```sh
casier login
casier init
casier run -- npm start
```

## Configuration

| Variable | What it does |
|---|---|
| `DATABASE_URL` | Postgres connection string, required |
| `ENCRYPTION_KEY` | Base64-encoded **32 bytes** protecting every project data key; required |
| `CORS_ALLOWED_ORIGINS` | Allowed cross-origin callers; `*` is refused at startup |
| `PUBLIC_URL` | Public origin used to build password reset links |
| `CASIER_ADMIN_EMAILS` | Who may call `POST /admin/rotate-key`; empty means nobody |
| `SMTP_HOST` | Enables the mailer, and makes `SMTP_FROM` and a public origin required |
| `OIDC_ISSUER` | Enables OIDC; three companion variables become required with it |

Full reference: [docs/configuration.md](docs/configuration.md).

## Structure

```
apps/
  api/       Go backend — modules/ (auth, projects, environments, secrets),
             schemas/ (GORM models and migrations), internal/ (crypto, audit, mailer)
  client/    SvelteKit dashboard, built static and served by the Go binary
cli/         Rust CLI — commands/ (login, run, sync, diff, check, push)
scripts/     Repo-local Postgres helpers and the quality gate
docs/        Architecture, configuration, development, deployment, API
```

## Documentation

| Doc | What's in it |
|---|---|
| [Architecture](docs/architecture.md) | Request flow, data model, how the pieces fit |
| [Configuration](docs/configuration.md) | Every environment variable and default |
| [Development](docs/development.md) | Local setup, tests, the quality gate |
| [Deployment](docs/deployment.md) | Docker Compose, Dokploy, Traefik routing |
| [API](docs/api.md) | HTTP endpoints, payloads, and the CLI command reference |

---

Part of the [Facile Suite](https://facile.studio) — self-hosted tools for creative studios
and freelancers. One login, zero cloud dependency.
