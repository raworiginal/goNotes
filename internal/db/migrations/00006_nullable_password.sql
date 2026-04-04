-- +goose Up
ALTER TABLE users ALTER COLUMN password DROP NOT NULL;

-- +goose Down
ALTER TABLE users ALTER COLUMN password SET NOT NULL;
