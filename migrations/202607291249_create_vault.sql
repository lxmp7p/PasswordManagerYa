-- +goose Up

CREATE TABLE vault_items (
    id BIGSERIAL PRIMARY KEY,

    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    type VARCHAR(50) NOT NULL,

    title TEXT NOT NULL,

    secret_data BYTEA NOT NULL,

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE vault_metadata (
    id BIGSERIAL PRIMARY KEY,

    item_id BIGINT NOT NULL REFERENCES vault_items(id)
        ON DELETE CASCADE,

    key TEXT NOT NULL,
    value TEXT NOT NULL
);

CREATE TABLE vault_files (
    id BIGSERIAL PRIMARY KEY,
    item_id BIGINT REFERENCES vault_items(id),
    filename TEXT,
    content BYTEA
);