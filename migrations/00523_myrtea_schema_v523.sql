-- +goose Up
-- +goose StatementBegin
ALTER TABLE variables_config_v1
    ALTER COLUMN value type text;

-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
ALTER TABLE variables_config_v1
    ALTER COLUMN value type varchar(100);
-- +goose StatementEnd