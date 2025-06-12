-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS config_history_v1
(
    id        bigint PRIMARY KEY NOT NULL, -- timestamp in milliseconds
    commentary text DEFAULT '',
    type      varchar(100) NOT NULL,
    user      varchar(150) NOT NULL
);
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS config_history_v1;
-- +goose StatementEnd
