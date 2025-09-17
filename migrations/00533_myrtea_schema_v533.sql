-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS mail_templates_v1
(
    id          BIGINT PRIMARY KEY NOT NULL,
    name        VARCHAR(255) NOT NULL,
    description TEXT DEFAULT '',
    subject     VARCHAR(255) NOT NULL,
    body_html   TEXT NOT NULL
);
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS mail_templates_v1;
-- +goose StatementEnd