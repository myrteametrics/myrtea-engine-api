-- +goose Up
-- +goose StatementBegin
ALTER TABLE notifications_history_v1
    ADD COLUMN type varchar(100) DEFAULT NULL;

ALTER TABLE notifications_history_v1
    ADD COLUMN user_login varchar(100) NOT NULL DEFAULT '';
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
ALTER TABLE notifications_history_v1
    DROP COLUMN IF EXISTS user_login;

ALTER TABLE notifications_history_v1
    DROP COLUMN IF EXISTS type;
-- +goose StatementEnd
