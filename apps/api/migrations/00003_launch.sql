-- Launch floor: porte v0.12.0 binds CLI and mobile login codes to the identity
-- that vouched for them, Registre's locale reaches the app as users.locale, and
-- directory.Deletions keeps its cursor in one row. The password reset table
-- goes with the password sign-in it served.

-- +goose Up
ALTER TABLE porte_login_codes ADD COLUMN IF NOT EXISTS provider text NOT NULL DEFAULT '',
                              ADD COLUMN IF NOT EXISTS subject text NOT NULL DEFAULT '';
ALTER TABLE users ADD COLUMN IF NOT EXISTS locale text NOT NULL DEFAULT 'en';
CREATE TABLE IF NOT EXISTS directory_deletion_state (
    id integer PRIMARY KEY,
    state bytea NOT NULL,
    updated_at timestamp with time zone NOT NULL DEFAULT now()
);
DROP TABLE IF EXISTS password_resets;

-- +goose Down
CREATE TABLE IF NOT EXISTS password_resets (
    token_hash text NOT NULL PRIMARY KEY,
    user_id bigint,
    expires_at timestamp with time zone,
    used boolean DEFAULT false,
    created_at timestamp with time zone
);
CREATE INDEX IF NOT EXISTS idx_password_resets_user_id ON password_resets USING btree (user_id);
DROP TABLE IF EXISTS directory_deletion_state;
ALTER TABLE users DROP COLUMN IF EXISTS locale;
ALTER TABLE porte_login_codes DROP COLUMN IF EXISTS subject, DROP COLUMN IF EXISTS provider;
