-- +goose Up
CREATE TABLE users (
    id UUID PRIMARY KEY,
    email TEXT NOT NULL,
    password_hash varchar(60) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    total_url_shortened INT NOT NULL DEFAULT 0
);

-- +goose Down
DROP TABLE users;