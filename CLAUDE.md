# Casier

Self-hosted secrets and environment variable manager for teams. Go API + SvelteKit frontend + PostgreSQL.

## Tech Stack

| Layer    | Stack                                                        |
| -------- | ------------------------------------------------------------ |
| API      | Go 1.26, Chi router, GORM, PostgreSQL 16, tronc v0.17.1, porte v0.12.0 |
| Client   | SvelteKit 5 (Svelte 5 runes), Tailwind CSS 4, muse (`@facile/muse`) |
| Build    | `go build`, Bun                                              |
| Deploy   | Docker Compose, single distroless container serving API + client |
| Auth     | Sign in with Facile (Registre OIDC) only; scoped API tokens for machines |
| Crypto   | AES-256-GCM encryption for secret values at rest             |

## Project Structure

apps/
  api/                  Go backend
    main.go             Entrypoint
    authz_test.go       The authorization matrix (see below) — needs TEST_DATABASE_URL
    modules/            Domain modules (auth, projects, environments, secrets)
    internal/           Shared infra: audit, authcontext, authcrypto, crypto, database,
                        documentation, env, mailer, middleware, testdb, webhook.
                        Logging, HTTP errors and JSON come from tronc, not from here.
    schemas/            GORM models and migrations
  client/               SvelteKit frontend
    src/
      app.css           Design tokens (Facile Suite palette + muse token aliases)
      routes/           SvelteKit file-based routing
        +page.svelte    Public landing page
        login/          muse AuthScreen (the only sign-in page)
        (app)/          Authenticated shell: sidebar + mobile nav
          projects/     Project index, and per-project secrets/members/tokens/audit/settings
          settings/     Account, appearance, about
      lib/
        backend.ts      API client and domain types
        auth.svelte.ts  Auth state (Svelte 5 runes)
        theme.svelte.ts Light/dark/system preference
        components/     App components (shell chrome + project sections)
    static/fonts/       Goga typeface
docs/                   Reference docs: api, architecture, configuration, deployment, development
scripts/                check.sh (the quality gate) and the repo-local Postgres helpers
Dockerfile              Multi-stage: bun build -> golang:1.25-alpine -> distroless
docker-compose.yml      Full stack: db + app

The CLI is **not** here — it lives in its own repo, see below.

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

### The quality gate

    sh scripts/check.sh            gofmt -l, go vet, go test, then the client type-check
    sh scripts/check.sh --go-only  skip the client
    sh scripts/check.sh --format   rewrite Go sources in place (the only writing mode)
    mise install                   wire the git hooks; lefthook runs the gate before every push

It is deliberately not run through `mise`: `mise run` resolves every tool in the merged config
before running a task body, so an unrelated broken tool in a global config would take the gate
down with it. It also resolves `go` from `GOROOT` when set, because mise pins a version while
leaving another `go` earlier on PATH, and a mismatch fails with
`compile: version "X" does not match go tool version "Y"`.

When `pg_ctl` is on PATH the gate starts the repo-local Postgres and exports
`TEST_DATABASE_URL` itself, so the authorization matrix actually runs instead of skipping.

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
| OIDC_ISSUER            | In production | Registre issuer URL; sign-in is OIDC only      |
| OIDC_CLIENT_ID         | If OIDC  | OIDC client ID                                   |
| OIDC_CLIENT_SECRET     | If OIDC  | OIDC client secret                               |
| OIDC_REDIRECT_URL      | If OIDC  | Callback URL for OIDC flow                       |
| OIDC_SUCCESS_URL       | No       | Redirect after OIDC login (default: first CORS origin) |
| CASIER_ADMIN_EMAILS    | No       | Comma-separated emails allowed to call `POST /admin/rotate-key`. Empty means nobody — the endpoint fails closed. |
| TEST_DATABASE_URL      | No       | Test-only. Postgres for the HTTP authorization matrix; unset makes it skip loudly |

## CLI — a separate repo

The `casier` CLI lives in **[FacileStudio/casier-cli](https://github.com/FacileStudio/casier-cli)**
(checked out at `~/Code/Facile/casier-cli`), following the suite convention that every
CLI is its own repo — as opus-cli, sablier-cli, nuage-cli, capsule-cli and ardoise-cli already
are. It was split out of `cli/` on 2026-08-06 with `git subtree split`, so its history came
with it. **This repo no longer builds, tests or ships the CLI**; its `CLAUDE.md` covers the
client side.

What matters here is the API surface it depends on. **None of these may change without a
paired change in that repo**, and the two are versioned independently, so a mismatch is a real
deployment state rather than a hypothetical:

- `GET /auth/config` — how `casier login` discovers whether the API answers at the root (dev)
  or under `/api` (Traefik in production), and whether SSO is on.
- `GET /auth/oidc?flow=cli&port=<port>&cli_state=<nonce>` — the CLI opens a browser and listens on
  an ephemeral `127.0.0.1` port. After the OIDC round trip, porte redirects to
  `http://127.0.0.1:<port>/?code=<code>&state=<nonce>` where `code` is a **one-time login code**
  (≈60s TTL) and `state` echoes the caller's nonce. The port and nonce are validated
  server-side and the redirect host is hard-coded (only the number is taken from the
  request), so neither parameter can be turned into an open redirect; the CLI re-checks the
  nonce to reject code injection from other local processes.
- `POST /auth/oidc/exchange` — settles the flow. A JSON body `{"code": "<code>"}` returns
  `{"token": "…", "user_id": "…"}`, where `token` is a bearer session token. The code is
  single-use (a replay answers 401), stored hashed. This is porte's token endpoint in
  everything but name, so it answers under `Cache-Control: no-store`. Unlike the browser
  callback it does **not** auto-provision an unknown user on the register side (403 where the
  code is not valid).
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
   **An ACL gates every path that reads a secret, including the indirect ones.** Two were
   missing and were closed on 2026-08-08:
   - `${ref:env/KEY}` expansion checked only the token scope, so a member who could not list,
     reveal, export or diff a locked secret could still read it by writing a reference to it
     into a secret they *were* allowed to create, then revealing that. Expansion now runs the
     same `aclAllowsRead` gate and leaves a denied reference literal.
     `TestSecretReferenceRespectsACLs` pins it, and was confirmed to fail without the fix.
   - `GET .../secrets/{key}/versions` and `.../{key}/acls` answered any project member, leaking
     version history, who changed it, and the emails of every grantee for a secret the caller
     could not read. Both go through `service.readableSecret` now and answer `not_found`.
   Anything returning something *about* a secret must use `readableSecret`, not `findSecret`.
   Granting an ACL is administration, not a secret write: `PUT`/`DELETE .../{key}/acls` carry
   `middleware.RequireUserSession`, so no machine token can widen access however it is scoped.
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

**muse (`@facile/muse`) is the design system, and it owns the palette.** As of 2026-08-08
Casier pins `github:FacileStudio/muse#v0.12.1` and `app.css` is 71 lines: import Tailwind,
import `@facile/muse/styles`, `@source` the package so v4 scans it, and one `@theme inline`
block aliasing the suite's semantic names (`--color-background`, `--color-muted-foreground`, …)
**onto** muse's `fc-*` tokens. The aliases point that way round on purpose — muse is the source
of truth, and an app that redeclares the palette drifts from it. Both vocabularies therefore
work and resolve identically, which is what lets hand-written markup and muse primitives sit in
the same file.

- Pin the tag. Tracking the default branch means a library push silently changes this build.
- The house style is still chroma-zero OKLCH greyscale, Goga, radius `0.5rem` — muse now
  *matches* that rather than contradicting it. Never hardcode a hex or a Tailwind stock colour.
- **Colour exists only as a tone**: `fc-danger`, `fc-success`, `fc-warning`, `fc-info` and the
  role tones `fc-owner` / `fc-admin`, plus `fc-chart-1..6`. Everything else is greyscale.
- Fonts ship **inside the package**. There is no `static/fonts` and no preload — deleting the
  ten vendored Goga faces was part of the v0.5.0 adoption.
- **Dark mode needs both classes.** muse paints dark from `@media (prefers-color-scheme: dark)`
  scoped to `:root:not(.light)`, plus a `:root.dark` twin. Toggling only `dark` therefore cannot
  show light mode to a dark-preferring OS — the media block still wins. `theme.svelte.ts` and the
  pre-paint script in `app.html` both set `.light` *and* `.dark`.
- `vite.config.ts` needs `optimizeDeps.exclude: ['@facile/muse']` or `vite dev` dies. No gate
  catches this; the production build is unaffected, which is what makes it easy to reintroduce.
- **Overlays are muse's `Modal` / `ConfirmModal` / `Drawer`.** The old rule said the opposite —
  muse's `Modal` used to render in the top-left corner because Tailwind preflight's `margin: 0`
  kills the `<dialog>` UA centring. muse sets `m-auto` itself now, so the bug is gone and
  `bits-ui` has been dropped entirely. `lib/components/Drawer.svelte` and `ConfirmDialog.svelte`
  survive as thin adapters over muse so their call sites did not have to change.
- Containers are **borderless**: `Card`, `StatCard` and `Table` separate from the page by fill
  (`bg-fc-component` on `bg-fc-page`), not a 1px outline. 1px borders are still right *inside* a
  container, on form controls, and on floating surfaces.
- muse hides scrollbars globally, on purpose (layout stability). Never re-enable them; if
  content is unreachable, the container is the bug.
- Icons are Iconify **Solar** via the `<iconify-icon>` web component loaded in `app.html`:
  `-bold-duotone` for brand marks, `-linear` for UI chrome. Do not add `@lucide/svelte`.
  The brand mark is `solar:safe-square-bold-duotone`, and `static/favicon.svg` is the same glyph
  — solid plate, mark knocked out, inverted by `prefers-color-scheme` so it stays legible on a
  light or dark tab strip.
- **Settings is reached only from the sidebar user card** (CHARTE §14), never a nav row, and log
  out lives inside settings. The shell is muse's `SideBar` + `MobileNav`; only **one** `SideBar`
  may be mounted, because its collapse shortcut listener is document-level.
- `main` carries `min-w-0`. A flex item defaults to `min-width: auto`, so without it `main`
  refuses to shrink below its content and any inner horizontal scroller pushes the whole page
  sideways instead of scrolling itself.
- The auth screens are deliberately **raw elements, not muse** — they are the suite's canonical
  split-screen login unified byte-for-byte across 15 repos. Do not "finish the migration" there.
- Active navigation is inverted (`bg-foreground text-background`), not tinted.

## Conventions

- Go module pattern: each feature in modules/<name>/ has types.go, service.go, controller.go, router.go, documentation.go
- Cross-cutting HTTP plumbing comes from the **tronc** chassis (`httpx`, `httpjson`, `errors`,
  `logger`, `health`, `healthcheck`, `spa`, `env`) — do not reintroduce a local copy
- Migrations run on startup via `schemas.Migrate(db)`, and a failure is fatal. Raw SQL in a
  migration is only a string until Postgres parses it, so **run the built binary against a
  database with the pre-migration schema** before shipping one; unit tests cannot catch it
- Client uses adapter-static, pure SPA — no load functions, data is fetched in `$effect`. In
  production the **Go binary serves it** from one container via `tronc/spa`; there is no Nginx
- Svelte 5 runes enforced
- No inline comments in code
- Secrets encrypted at rest with AES-256-GCM (per-secret random nonce)
- Session tokens and API tokens are SHA-256 hashed before storage
- **Profile pictures are URLs, never bytes.** `users.avatar_url` holds Registre's stable
  `picture` claim (`<issuer>/avatars/<sub>`), written on every sign-in and backfilled at boot
  from `porte_identities` for the configured issuer. Every person in a response goes through
  `internal/people` (`Columns` for flat rows, `Resolve` + `Lookup` for `actor`/`owner`), and the
  client renders them with muse `Avatar` via `PersonLabel`, which falls back to initials on 404.
- A new route is not done until it has a case in the authorization matrix — the `chi.Walk`
  meta-test will fail the build until it does, which is the point
