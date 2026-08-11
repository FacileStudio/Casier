package schemas

const porteSchema = `
CREATE TABLE IF NOT EXISTS porte_identities (
	user_id         bigint NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	provider        text NOT NULL,
	subject         text NOT NULL,
	password_hash   text NOT NULL DEFAULT '',
	access_token    text NOT NULL DEFAULT '',
	refresh_token   text NOT NULL DEFAULT '',
	token_expiry    timestamptz,
	roles           jsonb,
	roles_synced_at timestamptz,
	synced_at       timestamptz,
	created_at      timestamptz DEFAULT now(),
	PRIMARY KEY (provider, subject)
);

CREATE TABLE IF NOT EXISTS porte_sessions (
	id           bigserial PRIMARY KEY,
	token_hash   text NOT NULL UNIQUE,
	user_id      bigint NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	label        text NOT NULL DEFAULT '',
	created_at   timestamptz NOT NULL DEFAULT now(),
	last_used_at timestamptz NOT NULL DEFAULT now(),
	expires_at   timestamptz
);

CREATE TABLE IF NOT EXISTS porte_login_codes (
	code_hash   text PRIMARY KEY,
	user_id     bigint NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	expires_at  timestamptz NOT NULL,
	consumed_at timestamptz
);
`

const adoptExistingPasswords = `
INSERT INTO porte_identities (user_id, provider, subject, password_hash)
SELECT id, 'local', lower(btrim(email)), password_hash
  FROM users
 WHERE coalesce(password_hash, '') <> ''
ON CONFLICT (provider, subject) DO NOTHING;
`

const carryLegacySessionsOver = `
INSERT INTO porte_sessions (token_hash, user_id, created_at, last_used_at, expires_at)
SELECT s.token, s.user_id, s.created_at, now(), s.expires_at
  FROM sessions s JOIN users u ON u.id = s.user_id
 WHERE s.expires_at > now()
ON CONFLICT (token_hash) DO NOTHING;
`

const dropLegacySessions = `DROP TABLE IF EXISTS sessions;`

const dropOIDCColumns = `
ALTER TABLE users
  DROP COLUMN IF EXISTS oidc_access_token_enc,
  DROP COLUMN IF EXISTS oidc_access_token_nonce,
  DROP COLUMN IF EXISTS oidc_refresh_token_enc,
  DROP COLUMN IF EXISTS oidc_refresh_token_nonce,
  DROP COLUMN IF EXISTS oidc_token_expiry,
  DROP COLUMN IF EXISTS profile_synced_at;
`

const dropPasswordColumn = `ALTER TABLE users DROP COLUMN IF EXISTS password_hash;`
