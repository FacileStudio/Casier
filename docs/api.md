# Casier — API

Every HTTP route the Go binary registers, plus the CLI command reference, taken from
`apps/api/main.go` and each module's `router.go`.

Application routes live under `/api`. `/health`, `/ready`, `/docs` and `/docs/openapi.json`
sit at the root, and `/health` and `/ready` also answer under `/api`. Anything else falls
through to the
dashboard's index document. There is no strip-prefix middleware in front, so the paths below
are what the server sees and what a public URL contains.

Credentials are `Authorization: Bearer <token>`, falling back to the `casier_session` cookie
when the header is absent.

| Auth column | Meaning |
|---|---|
| public | No credential |
| session | User session, or a machine token where the scope allows it |
| user-only | User session exclusively; machine tokens are refused with `403` |
| admin | User session whose email is in `CASIER_ADMIN_EMAILS` |

## Health and self-description

| Method | Path | Auth |
|---|---|---|
| GET | `/health`, `/api/health` | public |
| GET | `/ready`, `/api/ready` | public |
| GET | `/docs` | public |
| GET | `/docs/openapi.json` | public |

`GET /docs` is the API reference — a Scalar UI over `GET /docs/openapi.json`, an OpenAPI 3.1
document generated at startup from each module's `documentation.go`. Both are served by
`tronc/apiref`, so every Facile backend answers on the same two paths. There is no
checked-in spec file: the document is the registry, rendered.

`apps/api/apiref_test.go` walks the mounted routes and fails if one is missing from the
registry, so the reference cannot drift from the code the way a hand-maintained spec does.

`GET /spaces` and `GET /spaces/*` answer `301` to the matching `/projects` path, left from
the 2026-08-04 rename.

## Auth

| Method | Path | Auth |
|---|---|---|
| GET | `/api/auth/config` | public |
| POST | `/api/auth/register` | public |
| POST | `/api/auth/login` | public |
| POST | `/api/auth/forgot-password` | public |
| POST | `/api/auth/reset-password` | public |
| POST | `/api/auth/logout` | session |
| GET | `/api/auth/me` | session |
| GET | `/api/auth/oidc` | public |
| GET | `/api/auth/oidc/callback` | public |
| POST | `/api/auth/cli` | public |
| POST | `/api/auth/sync-profile` | session |

`GET /api/auth/config` returns `{"sso_only","oidc_enabled","password_reset_enabled"}`.
`password_reset_enabled` is true only when passwords are on *and* a mailer is configured, so
the login screen never offers a reset that cannot be delivered.

The five password routes are **not registered at all** when `SSO_ONLY=true` — they answer
`404` rather than a policy error. The four OIDC routes are registered only when
`OIDC_ISSUER` is set and provider discovery succeeded at startup; a discovery failure is
logged and leaves them unregistered.

Register and login set the `casier_session` cookie alongside returning the token, so the
dashboard never touches `localStorage`. Logout clears it and answers `204`.

`POST /api/auth/cli` exchanges an OIDC ID token for a session token. Unlike the browser
callback it does **not** auto-provision an unknown user.

Register, login, forgot-password, reset-password, and the CLI exchange are rate limited to 5
requests per minute per IP by a token-bucket limiter.

## Projects

| Method | Path | Auth |
|---|---|---|
| GET | `/api/projects` | session |
| POST | `/api/projects` | session |
| GET | `/api/projects/{slug}` | session |
| PUT | `/api/projects/{slug}` | session |
| DELETE | `/api/projects/{slug}` | session |
| GET | `/api/projects/{slug}/audit` | session |
| GET | `/api/projects/{slug}/members` | user-only |
| POST | `/api/projects/{slug}/members` | user-only |
| DELETE | `/api/projects/{slug}/members/{userId}` | user-only |
| GET | `/api/projects/{slug}/invites` | user-only |
| POST | `/api/projects/{slug}/invites` | user-only |
| DELETE | `/api/projects/{slug}/invites/{inviteId}` | user-only |
| POST | `/api/invites/{token}/accept` | user-only |
| GET | `/api/projects/{slug}/tokens` | user-only |
| POST | `/api/projects/{slug}/tokens` | user-only |
| DELETE | `/api/projects/{slug}/tokens/{tokenId}` | user-only |
| GET | `/api/projects/{slug}/webhooks` | user-only |
| POST | `/api/projects/{slug}/webhooks` | user-only |
| DELETE | `/api/projects/{slug}/webhooks/{webhookId}` | user-only |
| POST | `/api/admin/rotate-key` | admin |

Everything marked user-only is refused for machine tokens by `requireUserSession`. Project
administration is a people operation: a CI credential must never be able to mint another
token, add a member, or widen its own access.

`PUT /api/projects/{slug}` treats an absent `rotation_days` as "leave it alone" — the field
is a pointer precisely so an older client sending only name and description cannot silently
disable a project's rotation warnings.

`POST /api/projects/{slug}/tokens` returns the raw `casier_<random>` token **once**; only its
SHA-256 hash is stored. A token carries a project, either an explicit environment list or
`all_environments`, a permission level, and an expiry — default 90 days, maximum 730.

`DELETE .../members/{userId}` answers `200` with a warning rather than a bare `204`: removed
members keep any secrets they already fetched, and the affected credentials need rotating.

`GET /api/projects/{slug}/audit` accepts `action` and `limit`, defaulting to 50 entries.

`POST /api/admin/rotate-key` takes the new base64 root key, re-wraps every project data key,
and returns the count plus a warning. It is gated on `CASIER_ADMIN_EMAILS`, refuses machine
tokens, and fails closed when the list is empty.

## Environments

| Method | Path | Auth |
|---|---|---|
| GET | `/api/projects/{slug}/environments` | session |
| POST | `/api/projects/{slug}/environments` | session |
| PUT | `/api/projects/{slug}/environments/{envSlug}` | session |
| DELETE | `/api/projects/{slug}/environments/{envSlug}` | session |

An environment has a name, a slug unique within its project, an ordering `position`, and a
`protected` flag. Deleting one cascades its secrets and, through
`api_token_environments`, any token grants that pointed at it.

`protected` gates writes to the environment's secrets to `owner` or `admin`. Changing the
flag in either direction is itself restricted, and deleting a protected environment is
stricter still — otherwise unprotecting production would be a one-request way around it.

## Secrets

| Method | Path | Auth |
|---|---|---|
| GET | `/api/projects/{slug}/environments/{envSlug}/secrets` | session |
| POST | `/api/projects/{slug}/environments/{envSlug}/secrets` | session |
| PUT | `/api/projects/{slug}/environments/{envSlug}/secrets/{key}` | session |
| DELETE | `/api/projects/{slug}/environments/{envSlug}/secrets/{key}` | session |
| GET | `/api/projects/{slug}/environments/{envSlug}/secrets/{key}/reveal` | session |
| PUT | `/api/projects/{slug}/environments/{envSlug}/secrets/{key}/tags` | session |
| GET | `/api/projects/{slug}/environments/{envSlug}/secrets/{key}/versions` | session |
| POST | `/api/projects/{slug}/environments/{envSlug}/secrets/{key}/rollback` | session |
| GET | `/api/projects/{slug}/environments/{envSlug}/secrets/{key}/acls` | session |
| PUT | `/api/projects/{slug}/environments/{envSlug}/secrets/{key}/acls` | session |
| DELETE | `/api/projects/{slug}/environments/{envSlug}/secrets/{key}/acls/{aclId}` | session |
| GET | `/api/projects/{slug}/environments/{envSlug}/secrets/trash` | session |
| POST | `/api/projects/{slug}/environments/{envSlug}/secrets/{key}/restore` | session |
| POST | `/api/projects/{slug}/environments/{envSlug}/secrets/import` | session |
| GET | `/api/projects/{slug}/environments/{envSlug}/secrets/export` | session |
| GET | `/api/projects/{slug}/secrets/diff` | session |
| GET | `/api/projects/{slug}/secrets/rotation` | session |

**Listing returns metadata, not values.** Plaintext comes only from `?reveal=true` on the
list route or from the dedicated `/{key}/reveal` route, and those are audited separately. A
client that expects a value and receives none should fail rather than inject an empty
variable — which is exactly what the CLI does.

Two query flags widen what a request does: `reveal` on the list route and `purge` on delete,
which destroys a secret immediately instead of trashing it. Both are parsed with
`strconv.ParseBool`, and **an unparseable value is false** — a typo must never read as
consent.

`GET .../secrets` also accepts `tag` to filter, and returns `rotated_at` and
`rotation_status` per secret alongside the project's `rotation_days`.

`GET /api/projects/{slug}/secrets/diff` requires both `from` and `to` environment slugs;
either one missing is `400`. `GET /api/projects/{slug}/secrets/rotation` counts stale
secrets per environment, so the dashboard can flag an environment the user is not currently
looking at.

Deleted secrets sit in `/secrets/trash` for 30 days and are restorable by key. The unique
index on `(environment_id, key)` ignores `deleted_at`, so the key stays reserved while it is
in the trash — recreating it under the same name would otherwise leave two rows a restore
could not tell apart.

## Errors and rate limits

Errors come from `tronc/errors` and are written by `tronc/httpjson`, so every failure is
JSON with the status its type maps to: `400` invalid, `401` unauthorized, `403` forbidden,
`404` not found, `409` conflict, `429` rate limited, `500` internal.

Out-of-scope reads deliberately answer `not_found` rather than `permission_denied`, so a
machine token cannot enumerate what it cannot reach.

Rate limiting is 5 requests per minute per IP on the credential routes, refilling
continuously rather than resetting on a fixed window.

## CLI

The `casier` binary, which lives in [FacileStudio/casier-cli](https://github.com/FacileStudio/casier-cli).
It resolves its server from `CASIER_SERVER_URL`, then
`~/.config/casier/config.toml`, then the built-in default
`https://casier.facile.studio/api`. Tokens live in the OS keychain under service `casier`,
keyed by server URL, unless `CASIER_TOKEN` is set — which is the CI and container path,
since a container has no keychain.

Most commands take `--project` and `--env`. Both fall back to a `casier.yml` in the
working directory with a `[project]` table carrying `slug` and `environment`, and
`environment` finally falls back to `dev`.

| Command | What it does |
|---|---|
| `casier login [--server URL] [--no-browser]` | Opens a browser for the OIDC round trip and stores the token. `--no-browser` prints the URL instead |
| `casier logout` | Clears the stored credential |
| `casier init` | Writes a `casier.yml` for the current directory |
| `casier projects` | Lists the projects you belong to |
| `casier secrets list -p <slug> [-e <env>] [--show]` | Lists keys; `--show` reveals values |
| `casier secrets get -p <slug> [-e <env>] <key>` | Prints one value |
| `casier secrets set -p <slug> [-e <env>] <key> <value>` | Creates or updates a secret |
| `casier secrets delete -p <slug> [-e <env>] <key>` | Deletes a secret |
| `casier diff -p <slug> --from <env> --to <env>` | Compares two environments |
| `casier sync push -p <slug> [-e <env>] [-f .env]` | Pushes a `.env` file into Casier |
| `casier sync pull -p <slug> [-e <env>] [-f .env]` | Writes secrets out to a `.env` file |
| `casier run [-p <slug>] [-e <env>] [--offline] -- <cmd>` | Injects secrets as environment variables and execs the command |
| `casier check [file] [-p <slug>] [-e <env>]` | Compares a `.env` file against the remote set and exits 1 when keys are missing |
| `casier push dokploy <composeId> [-p] [-e]` | Pushes secrets into a Dokploy compose service |

Realistic uses:

```sh
casier login --server https://casier.facile.studio/api
casier init
casier run -- npm start
```

```sh
casier check .env.example -p my-app -e prod
```

```sh
DOKPLOY_URL=https://gare.facile.studio DOKPLOY_API_KEY=... \
  casier push dokploy my-compose-id -p my-app -e prod
```

`casier run --offline` uses the cache under `~/.config/casier/cache/<project>-<env>.json`
without contacting the server. Only secrets that actually carry a value are cached: a
valueless entry would poison an offline run with an empty variable long after the reason it
was missing had been fixed.
