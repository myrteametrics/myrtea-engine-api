-- +goose Up
-- +goose StatementBegin
ALTER TABLE external_generic_config_v1
    DROP CONSTRAINT external_generic_config_v1_pkey;

ALTER TABLE external_generic_config_v1
    ADD COLUMN id serial PRIMARY KEY;
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
ALTER TABLE external_generic_config_v1
    DROP COLUMN id;

ALTER TABLE external_generic_config_v1
    ADD PRIMARY KEY (name);
-- +goose StatementEnd