-- +goose Up
-- +goose StatementBegin
ALTER TABLE issues_v1
    ADD COLUMN IF NOT EXISTS comment TEXT;
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
ALTER TABLE issues_v1
    DROP COLUMN IF EXISTS comment;
-- +goose StatementEnd
