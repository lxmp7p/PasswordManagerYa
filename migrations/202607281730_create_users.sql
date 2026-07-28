-- +goose Up

CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    login VARCHAR(255) NOT NULL UNIQUE,
    password_hash TEXT NOT NULL
);