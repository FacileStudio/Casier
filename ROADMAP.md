# Clef Roadmap

## v1 — Foundation (current)

- [x] Go API with Chi + GORM + Postgres
- [x] Email + password auth (argon2id, session-based)
- [x] OIDC SSO via Authentik
- [x] Spaces with per-space AES-256-GCM data keys (Infisical-style key hierarchy)
- [x] Environments (dev/staging/prod auto-created)
- [x] Secrets encrypted at rest with per-space data keys
- [x] OIDC tokens encrypted at rest
- [x] Audit logging (user, IP, key, operation)
- [x] Role-based access control (owner/admin/member/viewer)
- [x] SvelteKit dashboard (shadcn-svelte, dark theme)
- [x] Rust CLI with keyring token storage
- [x] `clef run` — inject secrets as env vars
- [x] Docker Compose deployment

## v2 — Hardening

- [ ] Rate limiting on auth endpoints (login, register)
- [ ] Password reset flow for non-OIDC users
- [ ] Secret version history with point-in-time rollback
- [ ] HttpOnly cookie auth as alternative to bearer tokens
- [ ] OIDC callback via POST/fragment instead of query param token
- [ ] Session sliding window (extend on use, max 7 days)
- [ ] Expired session cleanup (background purge)
- [ ] CLI: `clef init` to scaffold `.clef.toml` per project (space + env defaults)

## v3 — Team Features

- [ ] Secret rotation automation (scheduled re-generation)
- [ ] Root encryption key rotation (re-encrypt all space data keys)
- [ ] External KMS support (AWS KMS, GCP KMS) for root key
- [ ] Invite links for space membership
- [ ] Per-secret access control (beyond space-level roles)
- [ ] Secret tagging and search
- [ ] Bulk import/export (`.env` file ↔ Clef)
- [ ] CLI: `clef diff` — compare secrets across environments
- [ ] CLI: `clef sync` — push `.env` to Clef or pull from Clef

## v4 — Platform

- [ ] Webhook notifications on secret changes
- [ ] Nook integration (emit secret change events to the Facile event bus)
- [ ] Go SDK (`@facile/clef-go`) for runtime secret fetching
- [ ] TypeScript SDK (`@facile/clef-ts`) for runtime secret fetching
- [ ] API tokens (machine identity, scoped per space + environment)
- [ ] CI/CD integrations (GitHub Actions, Docker build args)
- [ ] Secret references (one secret pointing to another across environments)
- [ ] Dashboard: activity feed with filterable audit log
- [ ] Dashboard: secret diff view across environments
