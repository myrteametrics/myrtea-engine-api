-- +goose Up
-- +goose StatementBegin

CREATE TABLE functional_situation_v1 (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT DEFAULT ''::text,
    parent_id INTEGER REFERENCES functional_situation_v1 (id) ON DELETE SET NULL,
    color VARCHAR(7) DEFAULT '#0066CC',
    icon VARCHAR(50) DEFAULT 'folder',
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
    created_by VARCHAR(100) NOT NULL,
    metadata JSONB DEFAULT '{}'::jsonb
);

-- index on parent_id for quick lookups
CREATE INDEX idx_functional_situation_parent ON functional_situation_v1 (parent_id);

-- unique index to ensure name uniqueness at the same hierarchy level (NULL parent treated as 0)
CREATE UNIQUE INDEX unq_functional_situation_name_parent ON functional_situation_v1 ((COALESCE(parent_id, 0)), name);

-- pivot table linking functional situations to situation template instances
CREATE TABLE functional_situation_instances_v1 (
    functional_situation_id INTEGER NOT NULL REFERENCES functional_situation_v1 (id) ON DELETE CASCADE,
    template_instance_id INTEGER NOT NULL REFERENCES situation_template_instances_v1 (id) ON DELETE CASCADE,
    added_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
    added_by VARCHAR(100) NOT NULL,
    PRIMARY KEY (functional_situation_id, template_instance_id)
);

-- pivot table linking functional situations to (non-template) situations
CREATE TABLE functional_situation_situations_v1 (
    functional_situation_id INTEGER NOT NULL REFERENCES functional_situation_v1 (id) ON DELETE CASCADE,
    situation_id INTEGER NOT NULL REFERENCES situation_definition_v1 (id) ON DELETE CASCADE,
    added_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
    added_by VARCHAR(100) NOT NULL,
    PRIMARY KEY (functional_situation_id, situation_id)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Drop pivot tables first
DROP TABLE IF EXISTS functional_situation_situations_v1;
DROP TABLE IF EXISTS functional_situation_instances_v1;

-- Drop indexes (if any remain) and main table
DROP INDEX IF EXISTS unq_functional_situation_name_parent;
DROP INDEX IF EXISTS idx_functional_situation_parent;
DROP TABLE IF EXISTS functional_situation_v1;

-- +goose StatementEnd

