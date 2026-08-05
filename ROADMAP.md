# Casier Roadmap

> Formerly Clef. Renamed 2026-08-04 — grid of lockers, one compartment per
> project × environment. Registry-clean on crates.io and npm.

## v1 — Foundation

- [x] Go API with Chi + GORM + Postgres
- [x] Email + password auth (argon2id, session-based)
- [x] OIDC SSO via Authentik
- [x] Projects with per-project AES-256-GCM data keys (Infisical-style key hierarchy)
- [x] Environments (dev/staging/prod auto-created)
- [x] Secrets encrypted at rest with per-project data keys
- [x] OIDC tokens encrypted at rest
- [x] Audit logging (user, IP, key, operation)
- [x] Role-based access control (owner/admin/member/viewer)
- [x] SvelteKit dashboard (shadcn-svelte, dark theme)
- [x] Rust CLI with keyring token storage
- [x] `casier run` — inject secrets as env vars
- [x] Docker Compose deployment

## v2 — Hardening

- [x] Rate limiting on auth endpoints (login, register)
- [x] Password reset flow for non-OIDC users
- [x] Secret version history with point-in-time rollback
- [x] HttpOnly cookie auth as alternative to bearer tokens
- [x] OIDC callback via cookie instead of query param token
- [x] Session sliding window (extend on use, max 7 days)
- [x] Expired session cleanup (background purge)
- [x] CLI: `casier init` to scaffold `.casier.toml` per project

## v3 — Team Features

- [x] Root encryption key rotation (re-encrypt all project data keys) — unguarded and
      non-transactional until v6-1 fixed it, 2026-08-05
- [x] Invite links for project membership
- [x] Per-secret access control (beyond project-level roles) — stored but enforced by
      nothing until v6-3, 2026-08-05
- [x] Secret tagging and search
- [x] Bulk import/export (`.env` file ↔ Casier)
- [x] CLI: `casier diff` — compare secrets across environments
- [x] CLI: `casier sync` — push `.env` to Casier or pull from Casier
- [ ] Secret rotation automation (scheduled re-generation)
- [ ] External KMS support (AWS KMS, GCP KMS) for root key

## v4 — Platform

- [x] Webhook notifications on secret changes
- [x] API tokens (machine identity, scoped per project + environment) — the scopes were
      displayed but read by nothing until v6-2, 2026-08-05
- [x] Secret references (one secret pointing to another across environments)
- [x] Dashboard: activity feed with filterable audit log
- [x] Dashboard: secret diff view across environments
- [ ] Nook integration (emit secret change events to the Facile event bus)
- [ ] Go SDK (`@facile/casier-go`) for runtime secret fetching
- [ ] TypeScript SDK (`@facile/casier-ts`) for runtime secret fetching
- [ ] CI/CD integrations (GitHub Actions, Docker build args)

## v5 — Casier (current)

- [x] Rename Clef → Casier across repo, deployment, and infrastructure
- [x] CLI: offline cache — network-first with last-known-good fallback, making
      `casier run` safe when the server or the network is down
- [x] CLI: `casier check <file>` — CI gate, exit 1 on keys the store lacks
- [x] CLI: `casier push dokploy <composeId>` — push env straight to a Dokploy compose
- [x] API: `POST /auth/cli` — exchange an authentik ID token for an API credential
      (backend for device-code SSO login)
- [x] API: audit coverage for secret reads, `GET /ready` DB probe
- [x] Offboarding warning on member removal — removed members keep what they
      already fetched; rotate affected credentials
- [x] UI: production pass with the muse design system
- [ ] CLI: `casier login --sso` via OAuth device code flow (needs an authentik
      device-code flow created with Designation "Stage Configuration" — none
      ships by default)
- [ ] Runtime fetch pattern for deploys: static musl CLI baked into images,
      `ENTRYPOINT ["casier", "run", "--"]`, cache volume as restart fallback.
      Bootstrap tier (casier itself, postgres, authentik, traefik) stays
      Dokploy-injected forever

---

# Audit 2026-08-05

Source-level pass over `apps/api`, `apps/client/src/lib`, and the CLI surface. Every item
below was traced end to end in the code, not inferred from docs. Crypto (`internal/crypto`),
the key hierarchy, the schema, and the rotation grading are sound. **The authorization layer
is where the work is:** RBAC is enforced only inside the project/environment/secret
controllers, via `ResolveProjectID(..., requiredRoles...)`. Three things route around it.

The shared root cause of v6-1 through v6-3: `authcontext.Identity` carries `UserID`, `Email`,
`ClientIP` and nothing else. A caller's *capability* has nowhere to live, so scope enforcement
has no home to be written into. Widening that struct is the enabling change for the rest.

## v6 — Authorization (shipped 2026-08-05)

Ordered. 1 and 2 are the ones that change what Casier actually guarantees.

- [x] **v6-1 — Gate `POST /admin/rotate-key`.** `modules/projects/router.go:17` wraps the
      handler in `RequireAuth` and nothing else. Any authenticated user — a `viewer` on one
      project, or anyone self-registered while `SSO_ONLY=false` — can post a key of their
      choosing and re-encrypt every project data key in the instance. `RotateRootKey`
      (`projects/service.go:820`) then runs outside a transaction and mutates only
      `projects.Service.encKey`; `auth.Service` keeps the old key, and the next restart reads
      `ENCRYPTION_KEY` from env. A partial or unrecorded rotation is unrecoverable data loss
      for every project at once.
      Fix: instance-admin gate, wrap in a transaction, and hand the new key to every holder —
      or delete the route and rotate from the CLI with the process stopped, which is the
      honest shape for an operation that must not half-succeed.
- [x] **v6-2 — Enforce API token scope.** `authenticateApiToken` (`auth/service.go:241`)
      resolves a `casier_…` token to `api_tokens.created_by` and returns that user's identity.
      From there it is indistinguishable from a browser session: `ProjectID`, `Environments`
      and `Permissions` are stored, rendered in the token UI, and read by no code path. A
      token minted as "read-only, dev" can read prod, delete the project, add members, and
      mint more tokens.
      Fix: carry the token's scope on `Identity`, check it in the resolve helpers
      (`secrets/controller.go:237-263` is the choke point), and reject cross-project use at
      authentication time.
- [x] **v6-3 — Enforce or remove `secret_acls`.** `setACL` / `listACLs` / `deleteACL` write
      and read the table; no read or write path consults `can_read` / `can_write`. Enforcing
      it in `listSecrets` is the small change; the design question is whether per-secret ACLs
      earn their keep once v6-5 exists, or whether the table should go.
- [x] **v6-4 — Stop logging the password reset token.** `auth/service.go:325` logs
      `reset_token` verbatim at INFO, and Casier ships logs to Journal — log access is account
      takeover. There is no mailer, so that log line is the entire delivery mechanism. Either
      send a real email or disable the flow under `SSO_ONLY` and drop the log.
- [ ] **v6-5 — Environment-level roles.** "Any `member` can write prod" is the wrong default
      for a secrets manager. Prod should be gateable to admin without inventing a second
      project.
- [~] **v6-6 — An authorization test table.** There are no HTTP-level tests; coverage today is
      crypto, passwords, tokens, rotation grading, and OIDC parsing. One table-driven test over
      `(role, token scope, route) → expected status` would have caught v6-1 through v6-3, and
      is what keeps them fixed.

## v7 — Correctness (shipped 2026-08-05)

- [x] **Client swallows every API error.** The API writes `{"error":{"code","message"}}`;
      `apps/client/src/lib/backend.ts:31` reads `error.message`, which is therefore always
      undefined. Every toast and inline error in the dashboard reads "HTTP 403" instead of
      "insufficient permissions". One line, and it silently degraded every error path in the app.
- [x] **Webhooks and audit rows fire on a cancelled context.** `internal/webhook/webhook.go:47`
      spawns a goroutine and then queries with the *request* context, which dies the moment the
      handler returns — deliveries fail intermittently with `context canceled`. `internal/audit/
      audit.go:20` has the same shape and additionally discards the `Create` error, so audit
      loss is silent. For a secrets manager that is the one failure mode that must be loud.
      Detach with `context.WithoutCancel`, log the write error.
- [x] **`Event.Environment` is declared and never set** — every webhook payload ships an empty
      environment field.
- [x] **`${ref:env/KEY}` crosses environment boundaries.** `resolveReferences`
      (`secrets/service.go:30`) resolves against any environment in the project. Once v6-2 lands,
      that is the bypass: a token scoped to `dev` writes a reference and reads prod through it.
      Also N+1 queries per secret, on every list *and* every export.
- [x] **Export/import does not round-trip.** `exportEnv` escapes `"` but not backslashes or
      newlines, so multiline and escaped values come back mangled.
- [x] **`importEnv` is not transactional** — no per-key webhooks, no per-key audit, no key
      validation, and a partial failure leaves half an environment written.
- [x] **Tag filter is `LIKE '%' + raw + '%'`** — `%` and `_` unescaped, and it matches across
      tag boundaries.
- [x] **`acceptInvite` races on `use_count`** — concurrent accepts walk straight past `max_uses`.

## Shipped 2026-08-05 — what landed

Branch `hardening/authz-audit-2026-08`. The bullets above are kept as written because they are
the record of *why*; this is what actually changed. `gofmt`, `go build`, `go vet`, `go test`
and `svelte-check` are all clean.

**The enabling change.** `authcontext.Identity` gained `Scope *TokenScope`. A **nil scope means
a user session** (bounded by project roles alone); a non-nil scope is a `casier_…` machine token
carrying its project, environments and permissions. Every check is nil-safe, so session
behaviour is untouched. One deliberate asymmetry, documented on the type: an empty environment
list means *all* environments (that is how every pre-scoping token is stored), an empty
permission list means *no writes*. Fail open on reach, fail closed on write.

`internal/crypto.Keyring` now holds the root key for the whole process, so a rotation is
visible to `auth` and `projects` at once instead of leaving one service decrypting with a key
the database no longer matches.

**Three choke points now enforce token scope**, chosen because everything already funnels
through them: `projects.getProjectWithRole` (project), `environments.resolveEnvironmentID`
(environment), `secrets.resolveEnvWithRoles` (write permission). Out-of-scope reads answer
`not_found`, never `permission_denied`, so a token cannot enumerate what it cannot reach.
Scoped tokens are refused outright on member / invite / webhook / token routes.

**Rotation** is one transaction, `keyring.Replace` only after commit, a no-op key is rejected,
and the response carries a warning that `ENCRYPTION_KEY` must be updated in the environment too.

**ACLs are live.** A secret with at least one `secret_acls` row is hidden entirely from anyone
without a `can_read` row — dropped from list, export and diff rather than blanked, because a
blanked value silently corrupts both. `owner` bypasses. `setACL`/`deleteACL` tightened to
`owner|admin`, otherwise a member denied write could grant it back to themselves.

**Correctness:** `context.WithoutCancel` on webhook dispatch and audit writes (webhooks were
being dropped whenever delivery lost the race with the response); audit failures logged with
the full row and bounded by a semaphore; `.env` export/import round-trips byte-identically
through backslashes, quotes, newlines and tabs; import is one transaction with per-key audit
and webhooks; tag filter matches whole tags with proper `LIKE` escaping; `${ref:}` refuses to
resolve into an environment the caller cannot read; the webhook `environment` field is
populated; rollback finally emits an audit entry and a `secret.rolled_back` event.

**Client:** `ApiError` carries `status` and `code`, parsing survives empty and HTML bodies, and
the app shell logs out on a 401. Not global — `POST /auth/login` answers 401 for bad
credentials, and a global redirect would eat the login form's own error.

### Before deploying this to an instance with data

1. `CASIER_ADMIN_EMAILS` is unset everywhere today, so `POST /admin/rotate-key` will 403 for
   everyone until you set it. That is the intended posture; set it if you want rotation back.
2. `SELECT count(*) FROM api_tokens WHERE permissions = ''` — legacy rows with a blank
   permissions column become read-only. Fail-closed and correct, but it will look like a
   regression to whatever was using them.
3. `SELECT count(*) FROM secret_acls` — if this is non-zero, people currently seeing those
   secrets will stop seeing them, because until now the table enforced nothing.
4. Read-only and environment-scoped tokens start getting 403/404 on anything outside their
   scope. That is the headline fix, and it breaks exactly the automation it should.
5. The password reset routes no longer mount when `SSO_ONLY=true`, and the reset token is no
   longer written to the logs — which was its only delivery mechanism. Password reset is
   effectively disabled until there is a mailer (see v9).

### Follow-ups this batch created or uncovered

- **`api_tokens.environments` stores environment slugs with no foreign key.** Delete an
  environment and its slug stays in every token scope that named it; recreate a environment with
  the same slug later and those tokens silently regain access. Needs either an FK-backed join
  table or a cleanup on environment delete.
- **v6-6 is only half done.** The pure decision logic is now covered by table-driven tests in
  `authcontext`, `auth`, `projects`, `environments` and `secrets`, but there is still no
  HTTP-level `(role, token scope, route) → status` test, because there is no test-database
  harness. That harness is the missing piece.
- **A mailer.** v6-4 removed the reset token from the logs without replacing the delivery path.
- Per-list cost grew by one query for ACL grants and one for the caller's role.

## v8 — Refactors that pay rent

- [ ] **Collapse controller boilerplate.** ~30 handlers repeat identity-extract → parse → call.
      One helper removes the repetition and, more usefully, creates the single seam where v6-2's
      scope check belongs.
- [ ] **Stop re-querying the project.** Every method in `projects/service.go` calls
      `getProjectWithRole` (which loads the project) and then loads it again with
      `First(&project)`. Two to three round trips per handler for one row.
- [ ] **Drop the per-request goroutines.** Session sliding-expiry (`auth/service.go:224`) and
      `last_used_at` (`:264`) each spawn an unbounded goroutine and a DB write on every
      authenticated call. Batch, throttle, or accept coarser timestamps.
- [ ] **`resolveDataKey` resolves the project twice** — once directly, once inside
      `getProjectIDFromEnv`.

## v9 — Product

- [ ] **Trash for secrets.** `deleteSecret` cascades every version away with the row. Deleting
      `PROD_DATABASE_URL` is currently permanent and silent. 30-day soft delete, restorable.
- [ ] **Per-secret reveal endpoint.** `GET /secrets` returns every plaintext value and the
      dashboard masks them client-side with `••••••••` — the mask is theatre, the values are in
      the payload. It also means one audit row (`secrets.bulk_read`) covers a page load and a
      full production dump identically. A reveal-one endpoint fixes both the exposure and the
      audit trail.
- [ ] **Scoped tokens as a shipped feature** (the visible half of v6-2): settable TTL — today
      it is hard-coded to 365 days at `projects/service.go:597` — and revoke-rather-than-delete
      so a revoked token keeps its audit history.
- [ ] **Rotation warnings that leave the app.** The grading is built and nothing ever tells
      anyone. Weekly digest over webhook, or Courrier once it can send.

## Ideas under consideration

- E2E client-side encryption mode (age): zero-knowledge server, at the cost of
  dashboard search, cross-env diff, and server-side references. Only worth it
  if Casier is ever hosted for clients. Full design in the skatos repo
  (`SPEC.md`) — three-way merge protocol, recipient-in-ciphertext, escrow.
- Secret leases for short-lived credentials
- `casier history`/`undo` in the CLI on top of existing version history
