-- +goose Up
-- +goose StatementBegin
CREATE SCHEMA IF NOT EXISTS url_shortener;

CREATE TABLE IF NOT EXISTS url_shortener.links 
(
    short_key  TEXT NOT NULL PRIMARY KEY,
    long_url   TEXT NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS url_shortener.links;

-- +goose StatementEnd