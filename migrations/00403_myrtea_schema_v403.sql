-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS external_generic_config_v1
(
    name varchar(100) PRIMARY KEY,
    data jsonb NOT NULL
);
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS external_generic_config_v1;
-- +goose StatementEnd
