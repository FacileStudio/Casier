# Casier — Configuration

Every environment variable the code actually reads, and what happens when one is wrong.

`apps/api/internal/env` builds on `tronc/env.LoadCore`, so Casier inherits the suite's core
variables and adds its own. Configuration is validated at startup: a bad value logs and
exits 1 before the listener binds, so a container restart loop is the first symptom.

## Core, from `tronc/env`

| Variable | Required | Default | What it does |
|---|---|---|---|
| `DATABASE_URL` | yes | — | Postgres connection string. Absent or blank exits 1 |
| `PORT` | no | `8080` | HTTP listen port; rejected unless it parses to 1–65535. Compose sets `4000` |
| `LOG_LEVEL` | no | `info` | `debug`, `info`, `warn`, or `error`. Anything else means `info` |
| `APP_ENV` | no | `development` | `development`, `staging`, or `production`. Never gates security behavior |
| `CORS_ALLOWED_ORIGINS` | no | — | Comma-separated origins. Unset denies every cross-origin caller |
| `JOURNAL_URL` | no | — | Journal API base URL for log shipping |
| `JOURNAL_TOKEN` | no | — | Per-app Journal ingest key |

`tronc` reads the CORS list from the first name that is set, in order:
`CORS_ALLOWED_ORIGINS`, `ALLOWED_ORIGINS`, `DOMAINS`, `DOMAIN`, `CORS_ORIGINS`,
`TRUSTED_ORIGINS`, `CLIENT_ORIGIN`. The first is canonical; the rest exist so a repo can
adopt `tronc` without changing its deployment config.

**`*` is a startup panic here, not a permissive setting.** Casier sets
`AllowCredentials: true`, and `tronc/middleware.CORS` refuses the combination of a wildcard
origin and credentials outright.

Log shipping turns on only when both `JOURNAL_URL` and `JOURNAL_TOKEN` are non-empty. The
SDK posts to `<JOURNAL_URL>/ingest` and Journal serves ingest at `/api/ingest`, so the value
must end in `/api`. Journal's SPA catch-all answers unmatched paths with `200`, so a base
URL missing the suffix looks like a successful ingest while every line is discarded.

## Encryption

| Variable | Required | Default | What it does |
|---|---|---|---|
| `ENCRYPTION_KEY` | **yes** | — | Root key wrapping every project data key |
| `CASIER_ADMIN_EMAILS` | no | — | Comma-separated emails allowed to call `POST /admin/rotate-key` |

`ENCRYPTION_KEY` must be **base64-encoded and exactly 32 bytes once decoded**. Both halves
of that matter: a value that is not valid base64 and a value that decodes to 31 or 33 bytes
fail with different messages, and both exit 1. Generate one with `openssl rand -base64 32` —
which produces 44 characters ending in `=`. A 32-character random string is a common
mistake: it is base64-decodable but decodes to 24 bytes, so it is rejected.

`CASIER_ADMIN_EMAILS` fails closed. An empty list means nobody can rotate the root key, not
that everybody can.

## Auth

| Variable | Required | Default | What it does |
|---|---|---|---|
| `OIDC_ISSUER` | in production | — | Registre's issuer URL; makes the next three required. A production boot without it exits 1 |
| `OIDC_CLIENT_ID` | with issuer | — | Client identifier |
| `OIDC_CLIENT_SECRET` | with issuer | — | Client secret |
| `OIDC_REDIRECT_URL` | with issuer | — | Callback URL, e.g. `https://casier.facile.studio/api/auth/oidc/callback` |
| `OIDC_SUCCESS_URL` | no | first CORS origin | Where to send the browser after login |
| `MCP_RESOURCE` | no | `https://casier.facile.studio/mcp` in compose | Canonical URI of the remote MCP server, the RFC 8707 resource Registre binds its tokens to. Mounts `POST /mcp` only while OIDC is on; empty leaves it off |

Setting `OIDC_ISSUER` while leaving client id, secret, or redirect empty exits 1 with an
explicit message. `OIDC_SUCCESS_URL` falls back to the first CORS origin.

The same client credentials sign the deletion poll against Registre. `SSO_ONLY`, `PUBLIC_URL`
and the `SMTP_*` variables are no longer read; delete them from the environment.

## Static client

| Variable | Required | Default | What it does |
|---|---|---|---|
| `CLIENT_DIR` | no | `./client` | Directory holding the built dashboard, read by `tronc/spa.DirFromEnv` |

The SPA route is skipped entirely when the directory has no `index.html`, so an API-only
build still serves its routes. The image sets `CLIENT_DIR=/client` explicitly, because the
distroless `:nonroot` base carries its own working directory and a relative `./client` would
resolve somewhere else.

`PORT` is read a second time by `tronc/healthcheck`, which is what the compose healthcheck
`["CMD", "/api", "healthcheck"]` invokes.

## Compose substitutions

Interpolated by `docker-compose.yml`; they never reach the Go process directly.

| Variable | Default | What it does |
|---|---|---|
| `POSTGRES_USER` | `casier` | Internal Postgres user, also spliced into `DATABASE_URL` |
| `POSTGRES_PASSWORD` | `casier-internal-db` | Internal Postgres password |
| `POSTGRES_DB` | `casier` | Internal database name |

Compose also defaults `CORS_ALLOWED_ORIGINS` to `https://casier.facile.studio`, which in
turn seeds `OIDC_SUCCESS_URL` when it is left empty.

## CLI

The Rust CLI reads three variables of its own.

| Variable | What it does |
|---|---|
| `CASIER_SERVER_URL` | Overrides the saved server URL. Default `https://casier.facile.studio/api` |
| `CASIER_TOKEN` | Bearer token, checked before the OS keychain. The CI and container path |
| `DOKPLOY_URL`, `DOKPLOY_API_KEY` | Required by `casier push dokploy` |

Without `CASIER_TOKEN`, the CLI reads the token from the OS keychain under service `casier`,
keyed by the server URL — so one machine can hold credentials for several instances.
Per-project defaults live in a `casier.yml` at the repo root with a `[project]` table
carrying `slug` and `environment`; `environment` defaults to `dev`.

## Test database

| Variable | What it does |
|---|---|
| `TEST_DATABASE_URL` | Postgres for the authorization matrix tests; without it they skip loudly |

## Traps

- **`ENCRYPTION_KEY` must decode to 32 bytes.** Base64 of anything else is rejected. This is
  the single most common first-boot failure.
- **`*` in the CORS list panics** rather than allowing everything, because the API sends
  credentials.
- **No `CASIER_ADMIN_EMAILS` means no root key rotation at all.**
- **A production boot without `OIDC_ISSUER` exits 1**: with no password sign-in left, such an instance would have no way in.
- **`.env.example` is a compose template** and drifts. This page is derived from the code
  that reads the variables.
