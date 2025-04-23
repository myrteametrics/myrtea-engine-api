-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS variables_config_v1
(
    id    serial PRIMARY KEY,
    key   varchar(100) UNIQUE NOT NULL,
    value varchar(100)        NOT NULL
);
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS variables_config_v1;
-- +goose StatementEnd
