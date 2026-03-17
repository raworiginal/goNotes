-- +goose Up
ALTER TABLE users ADD COLUMN updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;

-- +goose Down
ALTER TABLE users DROP COLUMN updated_at;
