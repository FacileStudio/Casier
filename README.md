# Casier

Self-hosted secrets and environment variable manager for teams — Go API, SvelteKit dashboard, Rust CLI.

Encrypts secrets at rest with per-project AES-256-GCM data keys, supports email/password or OIDC SSO auth, and ships a CLI for injecting secrets straight into your dev or CI environment. Formerly named Clef.

Live at [casier.facile.studio](https://casier.facile.studio).

## Stack

| Layer  | Tech |
|--------|------|
| API    | Go 1.24, Chi, GORM, PostgreSQL 16 |
| Client | SvelteKit 5 (runes), Tailwind CSS 4, muse (`@facile/lib`) |
| CLI    | Rust, OS keychain token storage |
| Deploy | Docker Compose (API distroless, client Nginx) |

## Structure

```
apps/
  api/      Go backend — modules/ (auth, projects, environments, secrets), schemas/ (GORM + migrations)
  client/   SvelteKit dashboard
cli/        Rust CLI (casier login / projects / run / check / push / sync / diff)
scripts/    Local Postgres helpers
```

Data model: `Project` (own AES-256-GCM key) → `Environment` (dev/staging/prod) → `Secret` (encrypted value + version history).

## Quickstart

```sh
mise trust && mise install
mise run bootstrap
mise run db-up
mise run dev
```

Or with Docker:

```sh
cp .env.example .env
# set ENCRYPTION_KEY, e.g. openssl rand -base64 32
docker compose up --build
```

Individually: `go run .` (API, in `apps/api/`) or `bun run dev` (client, in `apps/client/`).

## CLI

```sh
casier login
casier projects
casier run -p <project> -e <env> -- <command>
casier check <file>          # CI gate: exit 1 on missing keys
casier push dokploy <composeId>
casier sync
casier diff
```

Tokens are stored in the OS keychain, keyed by server URL. Set `CASIER_TOKEN` for CI/containers where no keychain exists.

## Environment variables

See `.env.example`. Required: `DATABASE_URL`, `ENCRYPTION_KEY`. Optional: `PORT`, `LOG_LEVEL`, `CORS_ALLOWED_ORIGINS`, `VITE_API_BASE_URL`, and `OIDC_*` / `SSO_ONLY` for SSO.

## Docs

See `CLAUDE.md` for architecture details, auth flow internals, domain vocabulary, and design system conventions. Roadmap in `ROADMAP.md`.
