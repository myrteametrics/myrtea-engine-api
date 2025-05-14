-- +goose Up
-- +goose StatementBegin

ALTER TABLE external_generic_config_v1
    ALTER COLUMN name TYPE character varying(255);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE external_generic_config_v1
    ALTER COLUMN name TYPE character varying(100);

-- +goose StatementEnd
