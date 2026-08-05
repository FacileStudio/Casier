# Casier

Self-hosted secrets and environment variable manager for teams. Go API + SvelteKit frontend + PostgreSQL.

## Tech Stack

| Layer    | Stack                                                        |
| -------- | ------------------------------------------------------------ |
| API      | Go 1.25, Chi router, GORM, PostgreSQL 16, tronc v0.8.0      |
| Client   | SvelteKit 5 (Svelte 5 runes), Tailwind CSS 4, muse (`@facile/lib`) |
| Build    | `go build`, Bun                                              |
| Deploy   | Docker Compose, single distroless container serving API + client |
| Auth     | Email + password (session-based, argon2id), optional OIDC SSO |
| Crypto   | AES-256-GCM encryption for secret values at rest             |

## Project Structure

apps/
  api/                  Go backend
    main.go             Entrypoint
    modules/            Domain modules (auth, projects, environments, secrets)
    internal/           Shared infra (database, middleware, crypto, logger, etc.)
    schemas/            GORM models and migrations
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
Dockerfile              Multi-stage: bun build -> golang:1.25-alpine -> distroless
docker-compose.yml      Full stack: db + app

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
| DATABASE_URL           | Yes      | PostgreSQL connection string. The API exits 1 without it |
| APP_ENV                | No       | development/staging/production (default: development). Never gates security behaviour |
| PORT                   | No       | API port (default: 4000)                 |
| LOG_LEVEL              | No       | debug/info/warn/error (default: info)    |
| CORS_ALLOWED_ORIGINS   | No       | Comma-separated origins, read through `tronc/env`. `ALLOWED_ORIGINS`, `DOMAINS`, `DOMAIN`, `CORS_ORIGINS`, `TRUSTED_ORIGINS` and `CLIENT_ORIGIN` remain accepted fallbacks. Unset denies every cross-origin caller; `*` is refused at startup because the API sends credentials |
| ENCRYPTION_KEY         | Yes      | Base64-encoded 32-byte key for AES-256-GCM |
| JOURNAL_URL            | No       | Journal ingest URL — must end in `/api`, or every log line is silently dropped |
| JOURNAL_TOKEN          | No       | Per-app Journal API key                  |
| CLIENT_DIR             | No       | Built client directory served by the API (default: ./client) |
| OIDC_ISSUER            | No       | OIDC provider URL (enables SSO)                  |
| OIDC_CLIENT_ID         | If OIDC  | OIDC client ID                                   |
| OIDC_CLIENT_SECRET     | If OIDC  | OIDC client secret                               |
| OIDC_REDIRECT_URL      | If OIDC  | Callback URL for OIDC flow                       |
| OIDC_SUCCESS_URL       | No       | Redirect after OIDC login (default: first CORS origin) |
| SSO_ONLY               | No       | Hide password auth when true                     |
| CASIER_ADMIN_EMAILS    | No       | Comma-separated emails allowed to call `POST /admin/rotate-key`. Empty means nobody — the endpoint fails closed. |
| PUBLIC_URL             | No       | Public origin for links that leave the server (password reset). Defaults to the first CORS origin |
| SMTP_HOST              | No       | The switch for mail delivery. Without it `/auth/forgot-password` answers 503 rather than pretending |
| SMTP_PORT              | No       | Default 587 (STARTTLS)                           |
| SMTP_USERNAME          | No       | Optional — an internal relay often has none, but a username without a password is refused |
| SMTP_PASSWORD          | If user  | Required when SMTP_USERNAME is set               |
| SMTP_FROM              | If SMTP  | Envelope sender, e.g. `Casier <no-reply@facile.studio>` |
| TEST_DATABASE_URL      | No       | Test-only. Postgres for the HTTP authorization matrix; unset makes it skip loudly |

## CLI — a separate repo

The `casier` CLI lives in **[FacileStudio/casier-cli](https://github.com/FacileStudio/casier-cli)**
(checked out at `~/Projects/Facile/Code/casier-cli`), following the suite convention that every
CLI is its own repo — as opus-cli, sablier-cli, nuage-cli, capsule-cli and ardoise-cli already
are. It was split out of `cli/` on 2026-08-06 with `git subtree split`, so its history came
with it. **This repo no longer builds, tests or ships the CLI**; its `CLAUDE.md` covers the
client side.

What matters here is the API surface it depends on. **None of these may change without a
paired change in that repo**, and the two are versioned independently, so a mismatch is a real
deployment state rather than a hypothetical:

- `GET /auth/config` — how `casier login` discovers whether the API answers at the root (dev)
  or under `/api` (Traefik in production), and whether SSO is on.
- `GET /auth/oidc?cli_port=<port>&cli_state=<nonce>` — the CLI opens a browser and listens on
  an ephemeral `127.0.0.1` port. After the OIDC round trip the API mints a **separate** session
  token and redirects to `http://127.0.0.1:<port>/callback`. The port and nonce are validated
  server-side and the redirect host is hard-coded, so neither parameter can be turned into an
  open redirect; the CLI re-checks the nonce to reject token injection from other local
  processes.
- `POST /auth/cli` — exchanges an OIDC ID token for a session token. Unlike the browser
  callback it does **not** auto-provision an unknown user (403).
- `GET .../secrets` returns **metadata only**. Values come from `?reveal=true` or
  `GET .../secrets/{key}/reveal`, audited separately. This is the one change that already broke
  the CLI once: an older binary does not know to ask, so every value comes back absent and a
  naive client would inject an empty environment — the command runs, appears to work, and
  connects to nothing. The CLI now models a revealed secret as its own type and refuses a
  valueless answer outright. **Reinstall the CLI whenever this contract moves.**
- `casier_…` API tokens authenticate every non-interactive call via `CASIER_TOKEN`, bounded by
  the scope layer below.

## Authorization model

Four layers, and every one of them has to be checked — the 2026-08-05 audit found layers 2
and 3 were stored but never read.

1. **Project role** (`owner` / `admin` / `member` / `viewer`) via `ProjectMember`, enforced in
   `projects.Service.resolveProjectID(ctx, slug, userID, requiredRoles...)`. Every secret,
   environment and project path funnels through it.
2. **API token scope**, carried on `authcontext.Identity.Scope` as a `*TokenScope`. A **nil
   scope means a user session** and is unrestricted by this layer; a non-nil scope is a
   `casier_…` machine token bounded by its `project_id`, its `permissions` column, and its
   rows in **`api_token_environments`**. The three checks live at three choke points:
   `resolveProjectID` (project), `environments.Service.resolveEnvironmentID` (environment),
   and `secrets.Controller.resolveEnvWithWrite` (write permission). Scoped tokens are refused
   outright on the project administration routes — a CI credential never manages members.
   **A token cannot enumerate what it cannot reach**, because each layer answers
   out-of-scope and does-not-exist *identically*. The two layers pick different answers and
   that is fine as long as each is uniform: `environments.resolveEnvironmentID` says
   `not_found` for both, `secrets.resolveEnvFor` says `permission_denied` for both (its scope
   check runs before the lookup, so a slug that does not exist never gets as far as a 404).
   Verified end to end — probing a real out-of-scope environment and a made-up one through
   the secrets routes returns the same status. If you touch either helper, keep it uniform;
   the indistinguishability is the property, not the specific code.
   **Every scope check now fails closed.** "Every environment" is the stored
   `all_environments` flag, never an inference from an empty list — while empty meant
   everything, cascading away the last environment a token named *widened* it to the whole
   project, so the delete path was an escalation. Scopes reference environment rows with
   `ON DELETE CASCADE` and are resolved to slugs at authentication time, so a grant to a
   deleted environment is simply absent and recreating that slug does not resurrect it.
   Tokens expire (90 days by default, 730 at most, never infinite) and are **revoked, not
   deleted**, so audit rows naming a token still resolve.
3. **Per-secret ACLs** (`secret_acls`), which narrow a project role for one secret. A secret
   with no ACL rows behaves as before; `owner` always bypasses them.
4. **Protected environments** (`environments.protected`). Writing to a protected environment
   requires `owner` or `admin`, enforced in the secrets write path. Flipping the flag — in
   *either* direction — is itself `owner|admin` only, because a member who could unprotect
   production and then write to it has bypassed the whole feature. Deleting a protected
   environment requires `owner`, and no machine token may touch protection at all. This is
   v6-5, answered with one column rather than a per-environment role table: the need was
   "gate prod to admin without inventing a second project", and this is that.

`POST /admin/rotate-key` sits outside all of this: it is gated on `CASIER_ADMIN_EMAILS`,
refuses machine tokens, and fails closed when the list is empty.

**The matrix that keeps all four honest** is `apps/api/authz_test.go`: 58 routes × 14 caller
identities driven through the real Chi router against a real Postgres, asserting both status
*and* error code so the deliberate `not_found`-instead-of-`permission_denied` choice cannot be
helpfully "fixed" away. A companion meta-test walks the mounted routes with `chi.Walk` and
fails if any route has no case, so a new endpoint cannot ship unguarded. It needs
`TEST_DATABASE_URL`; `scripts/check.sh` provisions one when `pg_ctl` is on PATH, and the tests
skip *loudly* when it cannot.

## Secret lifecycle

`DELETE .../secrets/{key}` soft-deletes into a trash with a 30-day restore window
(`schemas.TrashRetentionDays`); `?purge=true` destroys immediately and is `owner|admin` only.
`GET .../secrets/trash` lists what is recoverable and when each entry expires,
`POST .../secrets/{key}/restore` brings one back, and `secrets.Service.StartTrashPurge`
sweeps hourly. The unique index on `(environment_id, key)` deliberately **covers** trashed
rows, so a deleted key stays reserved — creating it again returns a conflict naming the trash
rather than silently producing a second row a restore could not tell apart.

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
