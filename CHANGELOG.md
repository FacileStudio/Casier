# Changelog

## [Unreleased]

### Changed

- feat!: people sign in through Registre only. The login page is muse `AuthScreen`; the landing page uses `LandingHeader`, `LandingFooter` and `SignInWithFacile`
- build(api): bump porte to v0.12.0, add porte/directory v0.3.0, bump tronc to v0.17.1
- build(client): bump muse to v0.12.1; `provideLocale` follows the `locale` claim stored on the account
- feat(client): every screen is available in fr, en, es and de (es and de are machine translations and need native review); dates and numbers use `Intl` in the session locale
- feat(api): tronc's security headers are on (anti-framing enforced, the full CSP report-only), with the iconify script, the plan catalog and the icon API allowed
- feat(api): `OIDC_ISSUER` is required when `APP_ENV=production`
- OIDC account linking follows porte's rules: an existing account links only by verified email, and only when it holds no other identity of the same issuer
- ci: `ci.yml` runs gofmt, vet, build and tests against Postgres, the client check, message test and build, and filet (replaces `filet.yml`)

### Removed

- `POST /api/auth/register`, `/login`, `/forgot-password` and `/reset-password`, the register, forgot-password and reset-password pages, and the password form on the login page
- `SSO_ONLY`, `PUBLIC_URL` and the `SMTP_*` settings, the mailer, and the `password_resets` table
- API tokens (`casier_` prefix) and the CLI sign-in are untouched

### Added

- feat(api): GDPR deletion. Registre's `user.deleted` notice (signed push on `/api/registre/deletions` plus a poll) erases the person's personal projects, tokens, invites and account; projects other members use keep their secrets, authorship becomes "Deleted user", and the longest-standing admin or member becomes owner when the person was the last owner
- feat(api): `GET /version`, set at build time with the `APP_VERSION` build arg
- feat(api): `users.locale`, returned by `GET /api/auth/me`
- migration `00003_launch.sql`: `porte_login_codes.provider` and `.subject`, `users.locale`, `directory_deletion_state`, and the drop of `password_resets`

## [0.3.0] - 2026-11-06

### Added

- feat(client): show the live plans on the landing page with muse's `PricingPlans` (a `LandingOffers` component below the existing content, hidden while the catalog is empty)

### Changed

- build(client): bump muse to v0.11.0; muse's built-in strings are now English by default (v0.10.0)

## [0.2.1] - 2026-11-03

### Changed

- Spaces use the share-circle icon
- build(client): bump muse to v0.8.1

## [0.2.0] - 2026-11-03

### Changed

- style(client): remove the decorative borders and dividers around the landing page, page headers, lists and the login tabs, to match muse v0.8.0
- build(client): bump muse to v0.8.0
- `casier login` now defaults to https://casier.facile.studio and signs in through the browser with no code to paste (casier-cli v0.3.0)

## [0.1.0] - 2026-11-03

### Added

- feat: show each person's Facile profile picture, with initials as the fallback, in the sidebar, settings, members, secret history, trash, audit log and API token owners
- `users.avatar_url` stores the `picture` URL from single sign-on on every sign-in, and a boot backfill fills it for users who have not signed in since
- `/auth/me`, members and ACLs carry `name` and `avatar_url`; versions, trash and audit entries carry an `actor`, and API tokens an `owner`

### Changed

- build(client): bump muse to v0.7.3

## [0.0.2] - 2026-11-02

### Fixed

- fix(client): handle 401 on auth check gracefully
