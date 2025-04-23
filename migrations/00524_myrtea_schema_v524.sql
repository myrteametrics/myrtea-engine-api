-- +goose Up
-- +goose StatementBegin
CREATE TABLE elasticsearch_config_v1
(
    id               serial PRIMARY KEY,
    name             varchar(255) NOT NULL,
    urls             text,
    "default"        boolean      NOT NULL,
    export_activated boolean      NOT NULL DEFAULT TRUE
);

CREATE UNIQUE index unique_default_true ON elasticsearch_config_v1 ("default")
    WHERE
        "default" = TRUE;

CREATE OR REPLACE FUNCTION prevent_delete_default_row() returns trigger AS
$$
begin
    if OLD.name = 'default' then
        raise exception 'Cannot delete the row with name "default"';
    end if;
    return OLD;
end;
$$ language plpgsql;

CREATE TRIGGER prevent_delete_default_row_trigger
    before delete
    ON elasticsearch_config_v1
    FOR each ROW
EXECUTE function prevent_delete_default_row();

CREATE OR REPLACE FUNCTION ensure_one_default_row() returns trigger AS
$$
begin
    -- If the row being deleted is the current default
    if OLD."default" = true then
        -- Update the 'default' row to be the new default
        update elasticsearch_config_v1
        set "default" = TRUE
        where name = 'default';
    end if;
    return OLD;
end;
$$ language plpgsql;

CREATE TRIGGER ensure_one_default_row_trigger
    AFTER delete
    ON elasticsearch_config_v1
    FOR each ROW
EXECUTE function ensure_one_default_row();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER ensure_one_default_row_trigger ON elasticsearch_config_v1;
DROP FUNCTION ensure_one_default_row ();

DROP TRIGGER prevent_delete_default_row_trigger ON elasticsearch_config_v1;
DROP FUNCTION prevent_delete_default_row ();

DROP INDEX unique_default_true;
DROP TABLE elasticsearch_config_v1;
-- +goose StatementEnd
