-- +goose Up
-- +goose StatementBegin
ALTER TABLE roles_v4
    ADD COLUMN home_page varchar(255);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE roles_v4
    DROP COLUMN IF EXISTS home_page;
-- +goose StatementEnd

