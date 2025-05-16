-- +goose Up
-- +goose StatementBegin

ALTER TABLE variables_config_v1
    ADD COLUMN scope varchar(32) NOT NULL DEFAULT 'global';
ALTER TABLE variables_config_v1
    DROP CONSTRAINT variables_config_v1_key_key;
ALTER TABLE variables_config_v1
    ADD CONSTRAINT variables_config_v1_scope_key_unique UNIQUE (scope, key);

-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin

DELETE FROM variables_config_v1
    WHERE scope <> 'global';
ALTER TABLE variables_config_v1
    DROP CONSTRAINT variables_config_v1_scope_key_unique;
ALTER TABLE variables_config_v1
    ADD CONSTRAINT variables_config_v1_key_key UNIQUE (key);
ALTER TABLE variables_config_v1
    DROP COLUMN scope;

-- +goose StatementEnd