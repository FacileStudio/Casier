# Changelog

## [Unreleased]

## [0.9.2] - 2026-11-14

### Security

- fix(mcp): `casier_reveal_secrets` is confirmed only on the `/pending` page in the person's own browser. A client that can elicit gets the confirm link like any other, and an in-client (MRTR) `accept` is refused, since a client binary could forge it. `casier_delete_secret` keeps the in-client path.
- fix(mcp): a reveal from a token with an empty `client_id` is refused (`permission_denied`), so two clients without an id never share a confirmation.
- fix(auth): porte v0.20.1 and porte/mcpauth v0.3.1 mark browser logins with `b~`. Under v0.20.0's `b-` mark, about one plain session token in 4096 (CLI, device, older sessions) passed as a browser login on the confirm page. A browser login from v0.9.0 or v0.9.1 still signs in but must sign in once more before its next confirm.

### Changed

- build(api): tronc v0.23.1 (request dedupe matches the OAuth client).
- test(api): the authz suite sends a real browser cookie (`porte.BrowserTokenPrefix`) to the confirm page routes: with `X-Facile-CSRF` they reach the handlers (404 on a missing action), without it they answer 403. A plain CLI token sent as the cookie answers 403.

## [0.9.1] - 2026-11-14

### Added

- feat(client): `app.html` loads Vision's `s.js` on `casier.facile.studio` only, with `data-mask` so ids in the path are recorded as `:id`. The API adds `https://vision.facile.studio` to the CSP `script-src`; the loader itself is allowed by the inline script hashes of the built client.

## [0.9.0] - 2026-11-14

### Security

- fix(mcp): the confirm page accepts only the session cookie a browser sign-in minted (`mcpauth.RequireOwnSession`, porte v0.20.0). A bearer, a CLI session, a `casier_` token, a delegated token or any request with an `Authorization` header answers 403. Each person signs in once more in the browser before their first confirm; the page says so and links to sign-in, in all four locales.
- fix(mcp): a reveal names the client that asked, so tronc v0.23.0's request dedupe never lets one client collect values the person confirmed for another.
- fix(mcp): `casier_reveal_secrets` stays within tronc's `ResultBudget` and lists the keys it did not send in `omitted`. A confirmation is spent only after a successful read, in one locked transaction, so a failed read can be retried and a second collection is refused.

### Fixed

- fix(mcp): `/api/mcp/pending` is mounted in production (`PendingPage.Mount`). It was only mounted in tests, so the confirm page could not load or confirm.
- fix(mcp): `casier_reveal_secrets` and `casier_delete_secret` are `Idempotent: false`: a tool that asks the person to confirm never is.

### Changed

- build(api): tronc v0.23.0, tronc/migrate v0.1.1, porte v0.20.0, porte/mcpauth v0.3.0, porte/forfait v0.1.3. Go 1.26.11 pinned in mise.
- refactor(mcp): `mcpkit.StateKey` (byte-identical), `KeepPruned` (actions resolved or expired more than 30 days ago, hourly), `outcome.Message`, `mcpauth.AccountExists` and the kit offset pager replace Casier's copies.
- docs(mcp): `docs/mcp.md` lists every tool and its hints, the reveal flow, the retry re-audit case and what is left out on purpose.

## [0.8.0] - 2026-11-14

### Added

- The landing shows a mise-en-scene demo: a secret is added to prod, then `casier run` starts an app with every value masked.
- A `/downloads` page lists every way to install the CLI (install script, `facile install casier`, Homebrew). The landing shows only the visitor's device option.

### Changed

- The landing copy is shorter: one hero line, four features, a closing call to action.
- build(client): muse v0.23.0 (Magnetik fonts) and mise-en-scene v0.2.1.

## [0.7.6] - 2026-11-13

### Changed

- build(client): muse v0.21.2.

### Fixed

- Icon-only controls show a muse tooltip; native `title` tooltips moved to muse `Tooltip` or the `IconButton` `tooltip` prop.

## [0.7.5] - 2026-11-13

### Changed

- chore(deploy): healthchecks run every 30 s with a 1 s start interval, so a deploy cutover stays short while steady `docker exec` load drops.

## [0.7.4] - 2026-11-13

### Changed

- refactor(auth): sign-in goes through porte/kit's user store, built with `kit.NewStore` instead of a `kit.Store` literal. The kit now owns the sign-in write Casier kept in GORM: the same linking rules (a verified address only, never an account that already holds a subject of the same issuer, a provider matching with or without one trailing slash, porte v0.18.1 and later) and the same profile refresh (the picture always, the email only when verified). `NewStore` fills `created_at` and `updated_at`, which have no database default, so a new account holds every column an older one does, and production stops logging "this Store was built as a literal" at boot. No schema change. A test pins `Unwritten()` as empty and checks a new account holds no NULL. `CASIER_ADMIN_EMAILS` still fails closed: the users table has no admin column, so signing in grants nothing.

## [0.7.3] - 2026-11-12

### Fixed

- fix(auth): an account re-created for a person Registre already erased (a bearer token minted before the erasure, an MCP first call) is erased again on the next deletion push or poll, with no restart. Sign-in, MCP provisioning and the profile sync refuse a subject Casier erased. porte v0.19.0 (`Deletions.Holds`, `oidc.Deps.Erased`), porte/directory v0.5.0.

## [0.7.2] - 2026-11-11

### Fixed

- fix(brand): favicons, touch and manifest icons, the link preview image and `site.webmanifest` come from muse v0.19.0's brand generator, and `app.html` carries its head block: the catalog title "Casier · A secrets vault for your team", its description, theme colours and Open Graph and Twitter tags. The landing page no longer overrides the title and description, so every locale shows the catalog's.

## [0.7.1] - 2026-11-11

### Fixed

- fix(deps): vite 7.3.8, @sveltejs/kit 2.70.4 and svelte 5.57.3, plus overrides for devalue 5.9.4, postcss 8.5.32, nanoid 3.3.20, source-map-js 1.2.2 and cookie 0.7.2. Clears every high and medium npm advisory douane reported in the client lockfile.

## [0.7.0] - 2026-11-11

### Added

- feat(api): a remote MCP server at `POST /mcp` on tronc/mcp v0.22.1 and porte/mcpauth v0.2.1, behind Registre tokens bound to `https://casier.facile.studio/mcp` (scopes `casier.read`, `casier.write`). Fourteen tools: `casier_list_projects`, `casier_list_secrets`, `casier_get_secret`, `casier_diff_environments`, `casier_list_trash`, `casier_list_audit_events`, `casier_create_project`, `casier_create_environment`, `casier_set_secret`, `casier_tag_secret`, `casier_rollback_secret`, `casier_restore_secret`, `casier_reveal_secrets` and `casier_delete_secret`. Each one runs the REST controller as the person, so roles, ACLs, protected environments and the audit trail apply. The first MCP call creates the account. MCP is free on every plan.
- Only `casier_reveal_secrets` returns values, and only after the person confirms reading those keys: in the client through MRTR when it can ask, otherwise on the new `/pending` page, after which the client that asked collects the values once within 15 minutes with `confirmation_id`. The page never shows a value. `casier_delete_secret` moves a secret to the trash after the same confirmation. Listings, the diff and every write answer metadata only.
- feat(client): the `/pending` page (muse, fr/en/es/de) lists what MCP clients asked for, with the keys and versions involved, to confirm or dismiss.
- Migration `00004_mcp_actions.sql` (additive) creates tronc/mcp's `mcp_actions` table. Erasing a person deletes their rows.

### Changed

- build(deps): tronc v0.22.1, porte v0.18.4, porte/mcpauth v0.2.1 and go-sdk v1.8.0; the Docker builder is `golang:1.26.11-alpine`.

## [0.6.0] - 2026-11-10

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
