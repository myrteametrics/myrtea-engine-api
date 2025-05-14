-- +goose Up
-- +goose StatementBegin
ALTER TABLE job_schedules_v1
    ADD COLUMN IF NOT EXISTS enabled boolean;

UPDATE job_schedules_v1
SET enabled = TRUE;
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
ALTER TABLE job_schedules_v1
    DROP COLUMN IF EXISTS enabled;
-- +goose StatementEnd
