# Changelog

## [Unreleased]

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
