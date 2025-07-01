-- +goose Up
-- +goose StatementBegin
ALTER TABLE calendar_v1
    ADD COLUMN timezone varchar(100) NOT NULL DEFAULT 'Europe/Paris';
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
ALTER TABLE calendar_v1
    DROP COLUMN timezone;
-- +goose StatementEnd

