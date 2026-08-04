# Casier

Self-hosted secrets and environment variable manager for teams. Go API + SvelteKit frontend + PostgreSQL.

## Tech Stack

| Layer    | Stack                                                        |
| -------- | ------------------------------------------------------------ |
| API      | Go 1.24, Chi router, GORM, PostgreSQL 16                    |
| Client   | SvelteKit 5 (Svelte 5 runes), Tailwind CSS 4, shadcn-svelte |
| Build    | `go build`, Bun                                              |
| Deploy   | Docker Compose (API via distroless, client via Nginx)        |
| Auth     | Email + password (session-based, argon2id), optional OIDC SSO |
| Crypto   | AES-256-GCM encryption for secret values at rest             |

## Project Structure

apps/
  api/                  Go backend
    main.go             Entrypoint
    modules/            Domain modules (auth, spaces, environments, secrets)
    internal/           Shared infra (database, middleware, crypto, logger, etc.)
    schemas/            GORM models and migrations
    Dockerfile          Multi-stage: golang:1.24-alpine -> distroless
  client/               SvelteKit frontend
    src/
      routes/           SvelteKit file-based routing
        (app)/          Authenticated layout (spaces, space detail)
        login/          Login page
        register/       Register page
      lib/
        backend.ts      API client
        auth.svelte.ts  Auth state (Svelte 5 runes)
        components/     shadcn-svelte UI components
    Dockerfile          Multi-stage: bun -> nginx:alpine
docker-compose.yml      Full stack: db + api + client

## Commands

### Full monorepo (from root)

mise trust && mise install
mise run bootstrap
mise run db-up
mise run dev
mise run db-down

### API only (apps/api/)

go run .
go test ./...
go build -o bin/api .

### Client only (apps/client/)

bun install
bun run dev
bun run build
bun run check

### Docker

cp .env.example .env
# Set ENCRYPTION_KEY (generate with: openssl rand -base64 32)
docker compose up --build

## Environment Variables

| Variable               | Required | Description                              |
|------------------------|----------|------------------------------------------|
| DATABASE_URL           | Yes      | PostgreSQL connection string             |
| PORT                   | No       | API port (default: 4000)                 |
| LOG_LEVEL              | No       | debug/info/warn/error (default: info)    |
| CORS_ALLOWED_ORIGINS   | No       | Comma-separated origins                  |
| ENCRYPTION_KEY         | Yes      | Base64-encoded 32-byte key for AES-256-GCM |
| VITE_API_BASE_URL      | No       | Client API URL (default: http://localhost:4000) |
| OIDC_ISSUER            | No       | OIDC provider URL (enables SSO)                  |
| OIDC_CLIENT_ID         | If OIDC  | OIDC client ID                                   |
| OIDC_CLIENT_SECRET     | If OIDC  | OIDC client secret                               |
| OIDC_REDIRECT_URL      | If OIDC  | Callback URL for OIDC flow                       |
| OIDC_SUCCESS_URL       | No       | Redirect after OIDC login (default: first CORS origin) |
| SSO_ONLY               | No       | Hide password auth when true                     |

## Conventions

- Go module pattern: each feature in modules/<name>/ has types.go, service.go, controller.go, router.go, documentation.go
- Migrations run on startup via schemas.Migrate(db)
- Client uses static adapter (Nginx in prod)
- shadcn-svelte UI components in src/lib/components/ui/
- Svelte 5 runes enforced
- Secrets encrypted at rest with AES-256-GCM (per-secret random nonce)
- Session tokens hashed with SHA-256 before storage
