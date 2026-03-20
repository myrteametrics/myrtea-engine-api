-- +goose Up
-- +goose StatementBegin

-- Create the folder table for external config hierarchy
CREATE TABLE external_config_folders_v1
(
    id         SERIAL PRIMARY KEY,
    name       VARCHAR(255) NOT NULL,
    parent_id  INTEGER      REFERENCES external_config_folders_v1 (id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Index on parent_id for quick hierarchy lookups
CREATE INDEX idx_external_config_folder_parent ON external_config_folders_v1 (parent_id);

-- Unique name within the same parent (NULL parent treated as root level)
CREATE UNIQUE INDEX unq_external_config_folder_name_parent
    ON external_config_folders_v1 ((COALESCE(parent_id, 0)), name);

-- Add folder_id column to the external config table
ALTER TABLE external_generic_config_v1
    ADD COLUMN folder_id INTEGER REFERENCES external_config_folders_v1 (id) ON DELETE SET NULL;

CREATE INDEX idx_external_generic_config_folder ON external_generic_config_v1 (folder_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Remove index and column from external config table
DROP INDEX IF EXISTS idx_external_generic_config_folder;
ALTER TABLE external_generic_config_v1
    DROP COLUMN IF EXISTS folder_id;

-- Drop folder table and its indexes
DROP INDEX IF EXISTS unq_external_config_folder_name_parent;
DROP INDEX IF EXISTS idx_external_config_folder_parent;
DROP TABLE IF EXISTS external_config_folders_v1;

-- +goose StatementEnd
