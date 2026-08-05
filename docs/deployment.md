# Casier — Deployment

How the image is built, what Compose starts, and how it is routed on la ruche.

## Image

One root `Dockerfile` with the repo root as its context, in three stages:

1. `oven/bun:1` installs `apps/client` with `--frozen-lockfile` and runs `bun run build`.
2. `golang:1.25-alpine` downloads the API module's dependencies and builds a static binary
   with `CGO_ENABLED=0` and `-trimpath -ldflags="-s -w"`.
3. `gcr.io/distroless/static-debian12:nonroot` receives the binary at `/api` and the built
   client at `/client`, sets `CLIENT_DIR=/client`, exposes `4000`, and runs as
   `nonroot:nonroot`.

`CLIENT_DIR` is set explicitly because the distroless `:nonroot` variant carries its own
working directory; a relative `./client` would resolve there and the dashboard would
silently not be served. The `:nonroot` variant works because the container mounts no
writable volume — every byte of state is in Postgres.

The CLI is not part of the image, and no longer lives in this repo at all — it is a
developer and CI tool in [FacileStudio/casier-cli](https://github.com/FacileStudio/casier-cli), built
separately with `cargo build --release`.

## Compose topology

```
dokploy-network ──▶ casier-api (:4000, Traefik-routed, no host ports)
                          │
default network ──────────┴──▶ casier-db (postgres:16-alpine, expose only)
```

| Service | Notes |
|---|---|
| `casier-db` | `postgres:16-alpine`, named volume `casier_db_data`, `pg_isready` healthcheck every 5s |
| `casier-api` | Built from the root `Dockerfile`, joins both `default` and the external `dokploy-network`, waits on `casier-db` being healthy |

Postgres publishes nothing — it only `expose`s `5432` on the compose network, and
`DATABASE_URL` is assembled inside the compose file from `POSTGRES_USER`,
`POSTGRES_PASSWORD`, and `POSTGRES_DB`. There is no dev overlay in this repo; local work
uses the repo-local Postgres on port 5433 instead. See
[development.md](development.md).

The API healthcheck is `["CMD", "/api", "healthcheck"]` — the binary's own subcommand,
handled by `tronc/healthcheck`, which reads `PORT`. It must track the binary's path inside
the image if that ever moves.

## Traefik

One hostname, one service, two routers:

```yaml
traefik.enable: "true"
traefik.docker.network: dokploy-network
traefik.http.routers.casier-web.rule: "Host(`casier.facile.studio`)"
traefik.http.routers.casier-web.entrypoints: web
traefik.http.routers.casier-web.middlewares: redirect-to-https@file
traefik.http.routers.casier-web.service: casier-svc
traefik.http.routers.casier-secure.rule: "Host(`casier.facile.studio`)"
traefik.http.routers.casier-secure.entrypoints: websecure
traefik.http.routers.casier-secure.tls.certresolver: letsencrypt
traefik.http.routers.casier-secure.service: casier-svc
traefik.http.services.casier-svc.loadbalancer.server.port: "4000"
```

This is the suite's one-container, one-router, one-hostname rule. There is no `PathPrefix`
and no strip-prefix middleware: `/api` is part of the Go router itself, so `/api/projects`
arrives as `/api/projects`.

That matters to the CLI. `casier login` probes `GET /auth/config` to discover whether the
API answers at the root — the local dev shape — or under `/api`, the production shape. The
default server URL it ships with, `https://casier.facile.studio/api`, already carries the
prefix.

## Deploying on la ruche

Deployments are managed through Dokploy at `gare.facile.studio`, which owns the environment
file and triggers the Compose build. Prefer the `dokploy` CLI over SSH plus `docker`.

Set `ENCRYPTION_KEY` — base64, 32 bytes decoded, from `openssl rand -base64 32`. Set
`POSTGRES_USER` and `POSTGRES_PASSWORD` rather than keeping the built-in defaults. Set
`APP_ENV=production`, and `CASIER_ADMIN_EMAILS` if anyone should be able to rotate the root
key. `CORS_ALLOWED_ORIGINS` already defaults to the public host in the compose file, which
also seeds `PUBLIC_URL` and `OIDC_SUCCESS_URL`.

For SSO, set `OIDC_ISSUER`, `OIDC_CLIENT_ID`, `OIDC_CLIENT_SECRET`, and `OIDC_REDIRECT_URL`
together, pointing the redirect at
`https://casier.facile.studio/api/auth/oidc/callback`. For password reset, set `SMTP_HOST`
and `SMTP_FROM`. For log shipping, set `JOURNAL_URL` ending in `/api` and a `JOURNAL_TOKEN`
minted on Journal's Keys page.

### Rotating the root key

`ENCRYPTION_KEY` is not something to change in the environment file alone. Every project
data key is wrapped under the current root key, so replacing the variable without
re-wrapping makes every secret in the instance undecryptable.

The supported path is `POST /api/admin/rotate-key` with the new key in the body. It re-wraps
every project data key in place, returns how many it rotated, and returns a warning with it.
Only an email listed in `CASIER_ADMIN_EMAILS` may call it, machine tokens are refused, and
the endpoint fails closed when the list is empty. Update the environment variable to the new
key **after** the rotation reports success, then redeploy.

## Migrations

There is no migration tool and no separate step. `schemas.Migrate` runs at boot, before the
listener binds. In order: `renameSpacesToProjects` for databases predating the 2026-08-04
rename, `renameResetTokenColumn` because that column already held a digest and was merely
named `token`, `AutoMigrate` over the fourteen models, and then
`bindTokenScopesToEnvironments`, which moves token scopes off a comma-separated slug column
onto rows referencing environments by id and drops the old column.

Every step is guarded and idempotent, so a redeploy is safe. A failure exits 1 rather than
serving against a half-migrated schema.

One-way detail worth knowing before rolling back: `bindTokenScopesToEnvironments` drops
`api_tokens.environments`. A token stored with an empty list meant "every environment" under
the old semantics and becomes `all_environments`; slugs that no longer resolve are dropped,
which is the point — they were grants to an environment that does not exist.

## Health and readiness

`/health` and `/ready` answer at both the root and under `/api`, so the same probe works
whether the edge forwards everything or only `/api/*`. `/health` returns `{"status":"ok"}`
as soon as the process is serving and touches no dependency; `/ready` pings Postgres with a
2-second timeout and returns `503 {"status":"not_ready"}` when it fails. Use `/health` for
restarts and `/ready` for traffic.

A green `/api/health` says nothing about the dashboard — it only proves the Go process is
up, and the SPA route is skipped entirely when `CLIENT_DIR` has no `index.html`. Verify a
deploy by loading the site root and confirming the dashboard renders.
