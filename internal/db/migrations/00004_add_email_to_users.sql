-- +goose Up
ALTER TABLE users ADD COLUMN email TEXT;
CREATE UNIQUE INDEX users_email_idx ON users (email) WHERE email IS NOT NULL;

-- +goose Down
DROP INDEX users_email_idx;
ALTER TABLE users DROP COLUMN email;
