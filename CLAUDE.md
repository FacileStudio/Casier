# Casier

Self-hosted secrets and environment variable manager for teams. Go API + SvelteKit frontend + PostgreSQL.

## Tech Stack

| Layer    | Stack                                                        |
| -------- | ------------------------------------------------------------ |
| API      | Go 1.24, Chi router, GORM, PostgreSQL 16                    |
| Client   | SvelteKit 5 (Svelte 5 runes), Tailwind CSS 4, muse (`@facile/lib`) |
| Build    | `go build`, Bun                                              |
| Deploy   | Docker Compose (API via distroless, client via Nginx)        |
| Auth     | Email + password (session-based, argon2id), optional OIDC SSO |
| Crypto   | AES-256-GCM encryption for secret values at rest             |

## Project Structure

apps/
  api/                  Go backend
    main.go             Entrypoint
    modules/            Domain modules (auth, projects, environments, secrets)
    internal/           Shared infra (database, middleware, crypto, logger, etc.)
    schemas/            GORM models and migrations
    Dockerfile          Multi-stage: golang:1.24-alpine -> distroless
  client/               SvelteKit frontend
    src/
      app.css           Design tokens (Facile Suite palette + muse token aliases)
      routes/           SvelteKit file-based routing
        +page.svelte    Public landing page
        (auth)/         Split-screen auth shell (login, register, password reset)
        (app)/          Authenticated shell: sidebar + mobile nav
          projects/     Project index, and per-project secrets/members/tokens/audit/settings
          settings/     Account, appearance, about
      lib/
        backend.ts      API client and domain types
        auth.svelte.ts  Auth state (Svelte 5 runes)
        theme.svelte.ts Light/dark/system preference
        components/     App components (shell chrome + project sections)
    static/fonts/       Goga typeface
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

## CLI (cli/)

    casier login
    casier projects
    casier run -p <project> -e <env> -- <command>

`login` resolves the server from `--server`, then `CASIER_SERVER_URL`, then the saved config,
then a prompt defaulting to `https://casier.facile.studio/api`, and saves the winner to
`<config_dir>/casier/config.toml`. Because Traefik serves the API under `/api` in production
while dev serves it at the root, login probes `<url>/auth/config` and falls back to
`<url>/api/auth/config`, saving whichever answers — so both `https://casier.facile.studio` and
`.../api` work. Against a local API, override it: `casier login --server http://localhost:4000`.

Against an SSO server the CLI opens a browser and receives the session token on an ephemeral
`127.0.0.1` listener: it calls `/auth/oidc?cli_port=<port>&cli_state=<nonce>`, and after the
OIDC round trip the API mints a **separate** session token and redirects to
`http://127.0.0.1:<port>/callback`. The port and nonce are validated server-side and the
redirect host is hard-coded, so the parameters cannot be turned into an open redirect. The
nonce is checked by the CLI to reject token injection from other local processes.

Tokens live in the OS keychain, keyed by server URL. `CASIER_TOKEN` overrides it for CI and
containers, where no keychain exists.

`keyring` must keep its per-target platform features in `cli/Cargo.toml` (`apple-native`,
`windows-native`, `sync-secret-service`). With none of them enabled the crate silently falls
back to an in-memory mock store: login prints "Logged in as …", writes the config, and the
token disappears when the process exits, so every later command reports "Not logged in".
`store_token` reads the token back after writing to make that failure loud if it ever returns.

## Domain Vocabulary

Casier follows the Facile Suite vocabulary: **`space` is reserved suite-wide for the team
layer** (Sablier, Nuage, Plume, Courrier, Agenda all use `Space` + `SpaceMember` for teams).
Casier's own container is a **Project** — it groups environments, which group secrets:

    Project (slug, per-project AES-256-GCM data key)
      └── Environment (dev / staging / prod)
            └── Secret (encrypted value + versions)

Membership, invites, API tokens, webhooks and audit logs all hang off a Project. If a team
layer is added later, follow Sablier's model: a `Space` above `Project`, never a rename back.

## Rotation policy

`Project.rotation_days` (default 90, `0` disables, max 3650) is how long a secret value may
sit unchanged before Casier flags it. `modules/secrets/rotation.go` grades each secret as
`ok` / `due_soon` / `overdue`; the "due soon" window is a quarter of the period, capped at
14 days, so a short policy does not warn permanently.

Age is measured from **`secret_versions`**, not `secrets.updated_at` — a tag edit touches the
row without changing the value and must not reset the clock. `GET .../secrets` returns
`rotated_at` + `rotation_status` per secret plus the project's `rotation_days`;
`GET /projects/{slug}/secrets/rotation` counts stale secrets per environment so the UI can
flag an environment the user is not currently looking at.

`PUT /projects/{slug}` treats an absent `rotation_days` as "leave it alone" — the field is a
`*int` precisely so an older client that PUTs only name and description cannot silently
disable a project's warnings.

Renamed from `space` on 2026-08-04. `schemas.renameSpacesToProjects` migrates pre-rename
databases in place on startup and is a no-op afterwards.

## Design system

Casier follows the Facile Suite house style: **black and white, no colour**. The palette in
`apps/client/src/app.css` is chroma-zero OKLCH (`oklch(L 0 0)`) and mirrors Sablier, Nuage,
Plume and Vision token for token — including a greyscale `--destructive`. Typeface is Goga,
self-hosted from `static/fonts`. Radius is `0.5rem`.

- Suite tokens (`bg-background`, `text-muted-foreground`, `border-border`, `bg-primary`, …)
  are the vocabulary. Never hardcode a hex, never use a Tailwind colour utility.
- muse's `fc-*` tokens are aliased onto the same values in `app.css`, so `@facile/lib`
  primitives (Table, Badge, Field, Input, …) render in suite colours automatically.
- Icons are Iconify **Solar** via the `<iconify-icon>` web component loaded in `app.html`:
  `-bold-duotone` for brand marks, `-linear` for UI chrome. Do not add `@lucide/svelte`.
  The brand mark is `solar:safe-square-bold-duotone` (`Logo.svelte`), and `static/favicon.svg`
  is the same glyph — solid plate, mark knocked out, inverted by `prefers-color-scheme` so it
  stays legible on a light or dark tab strip.
- **Overlays are `src/lib/components/Drawer.svelte`, never muse's `Modal`.** That `Modal` is a
  native `<dialog>`, and Tailwind's preflight `margin: 0` kills the `margin: auto` that centres
  one — it renders in the top-left corner. The drawer is a right-anchored panel built on the
  bits-ui `Dialog` (already a dependency; it is what shadcn-svelte builds on too), full-width
  below `sm`. Its slide/fade keyframes live at the bottom of `app.css` because bits-ui renders
  the overlay and panel itself, out of reach of scoped styles.
- Dark mode is a `.dark` class on `<html>`, resolved before first paint by an inline script
  in `app.html` and driven by `src/lib/theme.svelte.ts` (light / dark / system).
- Active navigation is inverted (`bg-foreground text-background`), not tinted. Separation is
  done with 1px borders, not shadows.

## Conventions

- Go module pattern: each feature in modules/<name>/ has types.go, service.go, controller.go, router.go, documentation.go
- Migrations run on startup via schemas.Migrate(db)
- Client uses static adapter (Nginx in prod), pure SPA — no load functions, data is fetched in `$effect`
- Svelte 5 runes enforced
- No inline comments in code
- Secrets encrypted at rest with AES-256-GCM (per-secret random nonce)
- Session tokens hashed with SHA-256 before storage
