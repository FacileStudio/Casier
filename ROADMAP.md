# Casier Roadmap

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

- [x] Root encryption key rotation (re-encrypt all space data keys)
- [x] Invite links for space membership
- [x] Per-secret access control (beyond space-level roles)
- [x] Secret tagging and search
- [x] Bulk import/export (`.env` file ↔ Casier)
- [x] CLI: `casier diff` — compare secrets across environments
- [x] CLI: `casier sync` — push `.env` to Casier or pull from Casier
- [ ] Secret rotation automation (scheduled re-generation)
- [ ] External KMS support (AWS KMS, GCP KMS) for root key

## v4 — Platform

- [x] Webhook notifications on secret changes
- [x] API tokens (machine identity, scoped per space + environment)
- [x] Secret references (one secret pointing to another across environments)
- [x] Dashboard: activity feed with filterable audit log
- [x] Dashboard: secret diff view across environments
- [ ] Nook integration (emit secret change events to the Facile event bus)
- [ ] Go SDK (`@facile/casier-go`) for runtime secret fetching
- [ ] TypeScript SDK (`@facile/casier-ts`) for runtime secret fetching
- [ ] CI/CD integrations (GitHub Actions, Docker build args)
