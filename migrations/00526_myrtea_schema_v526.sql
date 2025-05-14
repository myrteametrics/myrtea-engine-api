-- +goose Up
-- +goose StatementBegin

-- Step 1 : Create the external_generic_config_versions_v1 table
CREATE TABLE external_generic_config_versions_v1
(
    config_id       INTEGER   NOT NULL,
    created_at      TIMESTAMP NOT NULL DEFAULT NOW(),
    data            JSONB     NOT NULL,
    current_version BOOLEAN   NOT NULL DEFAULT FALSE,
    PRIMARY KEY (config_id, created_at)
);

-- Step 2 : Populate the external_generic_config_versions_v1 table
INSERT INTO external_generic_config_versions_v1 (config_id, data, current_version)
SELECT id, data, TRUE
FROM external_generic_config_v1;

-- Step 3 : Add a foreign key constraint between external_generic_config_versions_v1 and external_generic_config_v1
ALTER TABLE external_generic_config_versions_v1
    ADD CONSTRAINT fk_config_id
        FOREIGN KEY (config_id) REFERENCES external_generic_config_v1 (id)
            ON DELETE CASCADE;

-- Step 4 : Create a view to get the current version of the external_generic_config_v1 table
SELECT c.id              AS config_id,
       c.name            AS config_name,
       v.data            AS config_data,
       v.created_at      AS version_created_at,
       v.current_version AS current_version
FROM external_generic_config_v1 c
         JOIN
     external_generic_config_versions_v1 v
     ON
         c.id = v.config_id
WHERE v.current_version = TRUE;


-- Step 5 : Drop the 'data' column from the 'external_generic_config_v1' table
ALTER TABLE external_generic_config_v1
    DROP COLUMN data;



-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin


-- Step 1 : Add the 'data' column back to the 'external_generic_config_v1' table
ALTER TABLE external_generic_config_v1
    ADD COLUMN data JSONB;

-- Step 2 : Populate the 'data' column with the data from the 'external_generic_config_versions_v1' table
UPDATE external_generic_config_v1 c
SET data = v.data
FROM external_generic_config_versions_v1 v
WHERE c.id = v.config_id
  AND v.current_version = TRUE;

-- Step 3 : Drop the foreign key constraint between 'external_generic_config_versions_v1' and 'external_generic_config_v1'
ALTER TABLE external_generic_config_versions_v1
    DROP CONSTRAINT fk_config_id;

-- Step 4 : Drop the 'external_generic_config_versions_v1' table
DROP TABLE external_generic_config_versions_v1;

-- +goose StatementEnd
