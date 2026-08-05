# Casier — Development

Local setup across three toolchains, the test layout, and the quality gate that runs before
every push.

## Prerequisites

| Tool | Version | Why |
|---|---|---|
| Go | 1.25 | `go 1.25.0` in `apps/api/go.mod`; the image builds on `golang:1.25-alpine` |
| Bun | 1.x | Client install, dev server, type-check, build |
| Rust | 2021 edition | The CLI in `cli/` |
| mise | recommended | Task runner; this repo uses mise's monorepo mode |
| Postgres | 16 | Either the repo-local instance or Docker |

Note the Go pin: Casier is on **1.25**, not the 1.24 most of the Go family uses. `mise.toml`
and both Dockerfile stages agree on it.

## Setup

```sh
mise trust && mise install
mise run bootstrap
```

`mise.toml` sets `experimental_monorepo_root` with `config_roots = ["apps/api",
"apps/client"]`, so each app carries its own `mise.toml` and the root tasks fan out to them
with `//apps/api:deps` style references. `bootstrap` is that fan-out for dependencies.

Enable the tracked git hooks once per clone:

```sh
mise run hooks
```

That is `git config core.hooksPath .githooks`. Without it the pre-push gate never runs.

## Running

```sh
mise run db-up
mise run dev
mise run db-down
```

`db-up` runs `scripts/postgres-up.sh`, which starts a **repo-local** Postgres cluster under
`.local/postgres` on port **5433** — deliberately not 5432, so it cannot collide with a
system Postgres or another suite app's container. `.local/` is gitignored.

`mise run dev` depends on `//apps/api:dev` and `//apps/client:dev`. The API task sources
`apps/api/.env` when present and otherwise supplies local defaults: `DATABASE_URL` pointing
at `127.0.0.1:5433`, `PORT=4000`, and a `CORS_ALLOWED_ORIGINS` list covering `3000` and
`5173` on both `localhost` and `127.0.0.1`. It prints every value it is about to use before
starting, which is the fastest way to catch a stale `.env`.

Individually:

```sh
cd apps/api && go run .
cd apps/client && bun run dev
cd cli && cargo run -- --help
```

Or the whole stack in Docker:

```sh
cp .env.example .env   # set ENCRYPTION_KEY
docker compose up --build
```

## Tests

```sh
cd apps/api && go test ./...
cd cli && cargo test
```

Go tests sit beside the code: `internal/crypto/crypto_test.go` covers the AES-256-GCM round
trip and rejects tampering with the ciphertext, the auth tag, the nonce, and the key;
`internal/authcontext` covers the token scope predicates, including the rule that a nil
scope means a user session and is unrestricted; and the module packages cover their own
services. `apps/api/authz_test.go` is the HTTP authorization matrix.

The authorization matrix needs a real Postgres and **skips loudly** without one. Point it at
a database with `TEST_DATABASE_URL`, or let `scripts/check.sh` do it for you — the gate
brings the repo-local cluster up, creates a `casier_test` database on port 5433, and exports
the URL itself. That machinery exists precisely so the gate cannot quietly stop covering
authorization.

Client type-check:

```sh
cd apps/client && bun run check
```

## Quality gate

```sh
mise run check        # the Go module, then the client, then the CLI
mise run check-go     # Go only
mise run format       # rewrite Go sources in place
```

All three call `scripts/check.sh`, which runs `gofmt -l`, `go vet ./...`, and `go test ./...`
in `apps/api`, then `bun run check` in `apps/client`, then `cargo fmt --check` and
`cargo test` in `cli`. It reports and never rewrites, except under `--format`. Each stage
skips with a message when its toolchain is missing rather than failing the run.

The CLI is in the gate on purpose: it speaks the same API as the client and breaks the same
way when the contract drifts, but it used to sit outside it entirely.

`.githooks/pre-push` execs the same script. Bypass a known-unrelated failure with
`git push --no-verify`.

Two details in the script are worth knowing before editing it. It is invoked as
`sh scripts/check.sh` rather than through `mise run`, because `mise run` resolves every tool
in the merged config first and an unrelated broken tool in a global config would take the
gate down with it. And it prefers `$GOROOT/bin/go` when `GOROOT` is set, because mise exports
`GOROOT` for the pinned version while leaving another `go` earlier on `PATH`, which fails
with `compile: version "X" does not match go tool version "Y"`.

## Conventions

Each `apps/api/modules/<name>/` carries `types.go`, `service.go`, `controller.go`,
`router.go`, and `documentation.go`. The last one feeds `GET /docs`, so adding a route means
describing it there in the same commit.

GORM models live in `apps/api/schemas/`, with `schemas/migrate.go` owning both `AutoMigrate`
and the hand-written DDL around it. Migrations run at boot.

The client is Svelte 5 runes only and a pure SPA — `adapter-static`, no load functions, data
fetched in `$effect`, every call through `src/lib/backend.ts`.

The design system is the suite house style: chroma-zero OKLCH tokens in
`apps/client/src/app.css`, Goga self-hosted, Iconify Solar icons. Never hardcode a hex,
never reach for a Tailwind color utility. Overlays use `src/lib/components/Drawer.svelte`,
not muse's `Modal` — that one is a native `<dialog>`, and Tailwind's preflight `margin: 0`
kills the `margin: auto` that would center it, so it renders in the top-left corner.
