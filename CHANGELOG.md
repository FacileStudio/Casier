# Changelog

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
