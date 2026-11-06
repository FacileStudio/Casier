# Casier — Architecture

How a secret is encrypted, who is allowed to read it, and how the API, dashboard, and CLI
fit into one container.

## Runtime topology

```
CLI (casier_… token) ──┐
Browser ───────────────┤
                       ▼
Internet ──▶ Traefik ──▶ Go binary (:4000) ──┬──▶ /api/*      handlers
                                              ├──▶ /health    liveness
                                              ├──▶ /ready     readiness (DB ping)
                                              ├──▶ /docs      module reference
                                              └──▶ everything SPA catch-all
                                                             │
                                                       Postgres 16
                                                  (values AES-256-GCM at rest)
```

One container. `apps/api/main.go` mounts every module inside `router.Route("/api", …)` and
puts `tronc/spa.Handler` on `/*` last, so a single Traefik router on
`Host(casier.facile.studio)` covers both halves — no `PathPrefix`, no strip middleware.
Postgres is an internal compose service with no published ports.

Two legacy redirects sit above the SPA: `GET /spaces` and `GET /spaces/*` answer `301` to
the matching `/projects` path, left over from the rename described below.

## Components

| Component | Path | Role |
|---|---|---|
| API | `apps/api` | Chi router, four modules, crypto, audit, mailer, webhooks |
| Client | `apps/client` | SvelteKit dashboard, `adapter-static`, served by the Go binary |
| CLI | `cli` | Rust binary that authenticates, fetches, and injects secrets |

Modules follow the same five-file shape: `types.go`, `service.go`, `controller.go`,
`router.go`, and `documentation.go`. The `documentation.go` files are collected at startup
into a single response served at `GET /docs`, so the API describes itself without a
generator or a checked-in spec.

## Encryption

Two levels of key, which is the whole point of the design.

The **root key** comes from `ENCRYPTION_KEY`, base64-encoded and exactly 32 bytes after
decoding — anything else exits 1 at startup. It is held in a `crypto.Keyring` shared by
every service in the process rather than copied into each, so a rotation is visible to all
of them at once instead of leaving one service decrypting with a key the database no longer
matches.

Each **project** gets its own AES-256-GCM data key, stored on the project row as
`encrypted_data_key` plus `data_key_nonce`, sealed with the root key. Secret values are
encrypted with the project's data key and their own 12-byte random nonce, stored as
`encrypted_value` and `nonce` on both `secrets` and `secret_versions`.

`POST /admin/rotate-key` re-wraps every project data key under a new root key. It is gated
on `CASIER_ADMIN_EMAILS`, refuses machine tokens outright, and fails closed when the list is
empty — with no admin emails configured, nobody can rotate.

## Authentication

Everything arrives as `Authorization: Bearer <token>`, and `middleware.RequireAuth` falls
back to the `casier_session` cookie when the header is absent, which is what lets the
same-origin dashboard work without touching `localStorage`.

**Sessions** are issued by the Registre OIDC callback and by the CLI exchange, stored SHA-256-hashed,
and carry both a 24-hour sliding window and a 7-day hard ceiling — an active session refreshes,
but never outlives a week. A cleanup goroutine prunes expired rows.

**There is no password sign-in.** An existing local account links to its Facile identity by
verified email on the first sign-in, and only when it holds no other subject of the same
issuer. API tokens (`casier_` prefix) authenticate on their own, with no session.

**Deletions.** Registre tells Casier when a person is erased (a signed push plus a poll,
state in `directory_deletion_state`). Their personal projects, tokens and account go; a
project other members use keeps its secrets, with authorship anonymised ("Deleted user") and
a new owner when the person was the last one.

**OIDC** is enabled by `OIDC_ISSUER` and federates to Registre.
Alongside the browser flow, `POST /api/auth/cli` exchanges an OIDC ID token for a session
token — and unlike the browser callback it does **not** auto-provision an unknown user.
User OIDC access and refresh tokens are themselves stored encrypted on the user row.

**Machine tokens** are `casier_<random>`, shown once at creation and stored only as a
SHA-256 hash. They have a lifetime: `DefaultTokenDays` is 90 and `MaxTokenDays` is 730. A
token that never expires is deliberately not offered.

## Authorization

Four layers, all of which are checked.

1. **Project role** — `owner`, `admin`, `member`, or `viewer` on `project_members`. Every
   project, environment, and secret path resolves through
   `projects.Service.resolveProjectID(ctx, slug, userID, requiredRoles...)`.
2. **Token scope** — carried on `authcontext.Identity.Scope`. A **nil scope means a user
   session** and is unrestricted by this layer; a non-nil scope is a machine token bounded by
   its project, its environments, and its permissions. The checks live at three choke points:
   `resolveProjectID`, `environments.Service.resolveEnvironmentID`, and
   `secrets.Controller.resolveEnvWithWrite`.
3. **Per-secret ACLs** — `secret_acls` narrows a project role for one secret. A secret with
   no ACL rows behaves as before, and `owner` always bypasses them.
4. **Protected environments** — `environments.protected` gates writes to `owner` or `admin`,
   enforced in the secrets write path through `roleMayWriteProtected`. Flipping the flag in
   *either* direction is itself restricted, because a member who could unprotect production
   and then write to it has bypassed the feature entirely, and deleting a protected
   environment is stricter still.

Two asymmetries are deliberate. Out-of-scope reads answer `not_found` rather than
`permission_denied`, so a token cannot enumerate what it cannot reach. And a token's
`all_environments` flag is stored explicitly rather than inferred from an empty environment
list: inferring it made "no environments left" and "every environment" the same state, so
cascading an environment delete would have *widened* a token instead of narrowing it.

Scoped tokens are refused outright on project administration — members, invites, webhooks,
minting further tokens, and root key rotation — by `requireUserSession`. A CI credential must
never be able to widen its own access.

Token scopes bind to environment **rows**, not slugs, through `api_token_environments`.
Deleting an environment now cascades the grant away; under the old slug-based storage the
grant lingered and recreating an environment with the same name silently handed old tokens
their access back. `schemas.bindTokenScopesToEnvironments` performs that one-way migration
and drops the old column.

## Data model

| Table | What it holds |
|---|---|
| `users` | Email, name, avatar URL, `locale` (from the OIDC `locale` claim) |
| `sessions` | Token hash as primary key, `user_id`, `expires_at` |
| `directory_deletion_state` | One row: the deletion poll's cursor and handled subjects |
| `projects` | Name, unique `slug`, `encrypted_data_key`, `data_key_nonce`, `rotation_days` |
| `project_members` | Role per user, unique on `(project_id, user_id)` |
| `environments` | Name, slug unique per project, `position`, `protected` (see the fourth authorization layer) |
| `secrets` | `key`, `encrypted_value`, `nonce`, `version`, `tags`, soft `deleted_at`, `deleted_by` |
| `secret_versions` | One row per previous value, with `changed_by` |
| `secret_acls` | `can_read` and `can_write` per user per secret |
| `api_tokens` | `token_hash`, `all_environments`, `permissions`, `expires_at`, `revoked_at` |
| `api_token_environments` | Token-to-environment grants, both sides cascading |
| `invites` | Token, role, `max_uses`, `use_count`, `expires_at` |
| `webhooks` | URL, `events`, signing `secret`, `active` |
| `audit_logs` | Action, resource, detail, actor email, client IP |

The unique index on `(environment_id, key)` deliberately ignores `deleted_at`, so a key
sitting in the trash stays reserved until it is restored or purged — otherwise recreating it
under the same name would leave two rows a restore could not tell apart. Trash retention is
30 days, enforced by a purge goroutine started at boot.

`schemas.Migrate` runs three hand-written steps around `AutoMigrate`:
`renameSpacesToProjects` for pre-rename databases, `renameResetTokenColumn` because the
column already held a digest and was merely named `token`, and
`bindTokenScopesToEnvironments` afterwards. All three are no-ops once applied.

Casier's container is a **Project**, not a Space — `space` is reserved suite-wide for the
team layer that Sablier, Nuage, Plume, Courrier, and Agenda all use. The rename happened on
2026-08-04 and the `/spaces` redirects above are its last visible trace.

## Rotation policy

`projects.rotation_days` defaults to 90, is capped at 3650, and `0` disables it. Each secret
is graded `ok`, `due_soon`, or `overdue`, where the "due soon" window is a quarter of the
period capped at 14 days so a short policy does not warn permanently.

Age is measured from `secret_versions`, not `secrets.updated_at` — a tag edit touches the row
without changing the value and must not reset the clock.

## Cross-app integration

Outbound `webhooks` fire through `internal/webhook` on secret changes, signed with the
webhook's own secret. Logging goes to Journal: `logger.New` takes a `Wrap` function that
installs `journal.NewHandler` when both `JOURNAL_URL` and `JOURNAL_TOKEN` are set, so
everything the app already logs is teed to the central instance. `JOURNAL_URL` must end in
`/api` or every line is silently dropped.
