-- +goose UP

ALTER TABLE users ADD COLUMN role VARCHAR(255) NOT NULL DEFAULT 'user' CHECK(role IN ('admin', 'user'));

-- +goose Down
ALTER TABLE users DROP COLUMN role;
