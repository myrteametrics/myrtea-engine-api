-- +goose Up
-- +goose StatementBegin
ALTER TABLE public.situation_template_instances_v1
    ADD COLUMN enable_depends_on boolean DEFAULT FALSE;

ALTER TABLE public.situation_template_instances_v1
    ADD COLUMN depends_on_parameters json DEFAULT '{}'::json;
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
ALTER TABLE public.situation_template_instances_v1
    DROP COLUMN IF EXISTS enable_depends_on;

ALTER TABLE public.situation_template_instances_v1
    DROP COLUMN IF EXISTS depends_on_parameters;
-- +goose StatementEnd
