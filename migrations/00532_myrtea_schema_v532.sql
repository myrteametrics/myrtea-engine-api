-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS config_history_v1
(
    id        bigint PRIMARY KEY NOT NULL, -- timestamp in milliseconds
    commentary text DEFAULT '',
    update_type      varchar(100) NOT NULL,
    update_user      varchar(150) NOT NULL
);
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS config_history_v1;
-- +goose StatementEnd
