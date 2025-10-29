-- +goose Up
-- +goose StatementBegin
ALTER TABLE elasticsearch_config_v1
    ADD COLUMN auth     boolean NOT NULL DEFAULT FALSE,
    ADD COLUMN insecure boolean NOT NULL DEFAULT FALSE,
    ADD COLUMN username varchar(255),
    ADD COLUMN password text;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE elasticsearch_config_v1
    DROP COLUMN IF EXISTS auth,
    DROP COLUMN IF EXISTS insecure,
    DROP COLUMN IF EXISTS username,
    DROP COLUMN IF EXISTS password;
-- +goose StatementEnd

