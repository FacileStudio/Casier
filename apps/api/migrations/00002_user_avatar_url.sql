-- The profile picture URL the identity provider hands over at sign-in. Casier
-- stores the URL, never the image: the provider serves the current picture at a
-- stable address, so a change shows on the next page load with nothing copied.

-- +goose Up
ALTER TABLE users ADD COLUMN avatar_url text NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE users DROP COLUMN avatar_url;
