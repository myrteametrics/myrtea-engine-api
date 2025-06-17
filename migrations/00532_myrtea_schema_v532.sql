-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS config_history_v1
(
    id               BIGINT PRIMARY KEY NOT NULL, -- timestamp in milliseconds
    commentary       TEXT DEFAULT '',
    update_type      VARCHAR(100) NOT NULL,
    update_user      VARCHAR(150) NOT NULL,
    config           TEXT DEFAULT ''
);
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS config_history_v1;
-- +goose StatementEnd
