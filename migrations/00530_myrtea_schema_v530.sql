-- +goose Up
-- +goose StatementBegin
CREATE TABLE tags_v1
(
    id          serial PRIMARY KEY,
    name        varchar(64) NOT NULL,
    description text        NOT NULL DEFAULT '',
    color       varchar(7)  NOT NULL DEFAULT '#000000',
    created_at  timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_name_color UNIQUE (name, color)
);

CREATE TABLE tags_situations_v1
(
    tag_id       bigint NOT NULL,
    situation_id bigint NOT NULL,
    PRIMARY KEY (tag_id, situation_id),
    FOREIGN KEY (tag_id) REFERENCES tags_v1 (id) ON DELETE CASCADE,
    FOREIGN KEY (situation_id) REFERENCES situation_definition_v1 (id) ON DELETE CASCADE
);

CREATE TABLE tags_situation_template_instances_v1
(
    tag_id                         bigint NOT NULL,
    situation_template_instance_id bigint NOT NULL,
    PRIMARY KEY (tag_id, situation_template_instance_id),
    FOREIGN KEY (tag_id) REFERENCES tags_v1 (id) ON DELETE CASCADE,
    FOREIGN KEY (situation_template_instance_id) REFERENCES situation_template_instances_v1 (id) ON DELETE CASCADE
);
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS tags_situation_template_instances_v1;
DROP TABLE IF EXISTS tags_situations_v1;
DROP TABLE IF EXISTS tags_v1;
-- +goose StatementEnd