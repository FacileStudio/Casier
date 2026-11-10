# Changelog

## [Unreleased]

### Added

- feat: pending Registre space invitations in the user menu. `GET /api/me/invitations` is mounted through `porte/kit` when OIDC is configured, behind a user session only (an API token gets 401), and the sidebar and mobile nav show the count through muse.

### Fixed

- fix(api): erasure also anonymises an audit row that names the erased person as its subject (`member.removed`, resource `user:<id>`). The erasure test now runs the directory receiver's own two callbacks and checks every column that can name a person. Personal projects are still deleted and shared projects still kept, re-owned and anonymised.

### Changed

- build(deps): tronc v0.21.0, porte v0.18.0 (`directory` v0.4.0) and muse v0.18.0. Every entry between the old and new pins is additive.
- refactor(api): the session middleware, the route mount, the profile read and the picture backfill come from `porte/kit`. Routes and their unauthenticated status codes are unchanged across the bare, OIDC and OIDC with CLI configurations (170 lines walked, identical). A session whose account is gone still answers 401, now with the JSON error envelope instead of an empty body.
- The stored email is rewritten at sign-in only when Registre marks it verified, the suite rule `porte/kit` applies. Casier rewrote it on every sign-in.
- refactor(client): the fetch kernel, the 401 redirect and the theme store come from `@facile/client` v0.1.0. The `api` methods keep their signatures; a 401 now sends the browser to `/login?next=...` from any page, except the session probe, which is quiet. The theme keeps its `casier.theme` key, so stored choices survive. The client never stored a bearer token in `localStorage`; it stays on the cookie session.
- ci: filet v0.19.0 (action and binary) with `filet.yml` reduced to `preset: facile`; the finding list is identical to the explicit config's. `.filet-baseline.json` records the ratchet: no finding at `failOn: error`, so it is empty. douane 0.7.0 runs in CI with `-fail critical -scope prod`, from the release archive checked against its sha256. The pre-push hook adds `filet gate .`.
- build(api): `go 1.26.11`, go-jose v4.1.6, x/text v0.41.0 and chi v5.4.0, which clear every production-scope douane finding (critical stdlib advisories among them).
- Kept in Casier, not the kit: the sign-in write and its linking rule (Casier ignores a trailing slash on the provider, the kit does not), the API-token path before the session, the string user id and token scope on the request identity, `GET /auth/config` without OIDC, and the erasure.
- Licence: Casier is now licensed under `AGPL-3.0-only` (see `LICENSE`); before this, the repository carried no licence.

## [0.5.4] - 2026-11-07

### Changed

- build(client): bump muse to v0.16.1, replace the deprecated `outline` button variant with `secondary` and the `danger` icon buttons with `ghost-danger`.
- build(api): `GET /version` is baked at build time from `apps/client/package.json` by `scripts/version.sh`. The `APP_VERSION` build arg is gone from the Dockerfile and compose file, so a stale Dokploy env can no longer show the wrong version. The response is unchanged.

## [0.5.3] - 2026-11-07

### Changed

- build(client): bump muse to v0.14.4 (from v0.14.2). Switching the language in the user menu re-renders the whole UI at once, with no reload. A single 100% DonutChart slice draws as a full ring.

## [0.5.2] - 2026-11-07

### Security

- Registre tokens are sealed at rest. porte to v0.13.0: `porte_identities.access_token` and `refresh_token`
  are encrypted with AES-256-GCM under `PORTE_TOKEN_KEY`. Existing rows are sealed once, in the background,
  after boot, and the count is logged. An invalid `PORTE_TOKEN_KEY` stops the boot. With the key unset nothing changes.
- Losing `PORTE_TOKEN_KEY` makes the stored tokens unreadable: people sign in again, nothing else is lost.
  Do not roll back to a build older than porte v0.13.0 after the first reseal.

## [0.5.1] - 2026-11-07

### Changed

- build(client): bump muse to v0.14.2 (the user menu no longer overflows the viewport edge).
- feat(client): the user menu has a Subscription row that links to the Facile spaces page.

## [0.5.0] - 2026-11-07

### Changed

- feat(client): the sidebar user card opens the shared muse user menu (account, settings, language, theme, help, sign out). The cog link and the in-settings sign-out are gone.
- feat(client): settings use muse `SettingsLayout` with Profile, Preferences, API tokens and Advanced. Profile is read-only and links to Facile; language and theme live in Preferences. `/settings/appearance` and `/settings/api` redirect to `/settings/preferences` and `/settings/tokens`.
- feat(client): project tokens use muse `TokensSection`. The API calls, permissions, environments and expiry are unchanged; the default expiry in the create drawer is now 30 days.
- build(client): bump muse to v0.14.1.

## [0.4.1] - 2026-11-07

### Changed

- build(api): bump porte to v0.12.1. The GDPR erasure no longer fails on a schema without `porte_magic_links`.
- build(client): bump muse to v0.13.0. Form controls are borderless fills, `Select` is muse's own combobox and badges stay on one line.
- The loading screen and the add-member and create-token buttons use muse's `Spinner` without border classes.

## [0.4.0] - 2026-11-07
### Changed

- feat!: people sign in through Registre only. The login page is muse `AuthScreen`; the landing page uses `LandingHeader`, `LandingFooter` and `SignInWithFacile`
- build(api): bump porte to v0.12.0, add porte/directory v0.3.0, bump tronc to v0.17.1
- build(client): bump muse to v0.12.1; `provideLocale` follows the `locale` claim stored on the account
- feat(client): every screen is available in fr, en, es and de (es and de are machine translations and need native review); dates and numbers use `Intl` in the session locale
- feat(api): tronc's security headers are on (anti-framing enforced, the full CSP report-only), with the iconify script, the plan catalog and the icon API allowed
- feat(api): `OIDC_ISSUER` is required when `APP_ENV=production`
- OIDC account linking follows porte's rules: an existing account links only by verified email, and only when it holds no other identity of the same issuer
- ci: `ci.yml` runs gofmt, vet, build and tests against Postgres, the client check, message test and build, and filet (replaces `filet.yml`)
- fix(api): the audit log records the address tronc's RealIP resolved, not the first `X-Forwarded-For` entry a client could write; GDPR erasure also clears the stored address on the erased person's audit rows

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
