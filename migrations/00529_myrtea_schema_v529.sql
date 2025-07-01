-- +goose Up
-- +goose StatementBegin
ALTER TABLE api_keys RENAME TO api_keys_v1;
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
ALTER TABLE api_keys_v1 RENAME TO api_keys;
-- +goose StatementEnd