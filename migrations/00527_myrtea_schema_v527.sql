-- +goose Up
-- +goose StatementBegin

DO
$$
    DECLARE
        record_id          integer;
        current_definition jsonb;
        new_rollmode       jsonb;
    BEGIN
        FOR record_id, current_definition IN
            SELECT id, definition
            FROM model_v1
            LOOP
                -- Check if the rollmode field exists and is of type 'timebased'
                IF current_definition -> 'elasticsearchOptions' ->> 'rollmode' = 'timebased' THEN
                    -- Modify the rollmode field to be of type object with interval if it's 'timebased'
                    new_rollmode := jsonb_build_object(
                            'type', 'timebased',
                            'timebased', jsonb_build_object('interval', 'daily')
                                    );
                ELSE
                    -- Otherwise, set rollmode to an object with the type cron
                    new_rollmode := jsonb_build_object(
                            'type', 'cron'
                                    );
                END IF;

                -- Update the definition field with the new structure of rollmode
                UPDATE model_v1
                SET definition = jsonb_set(
                        current_definition,
                        '{elasticsearchOptions,rollmode}',
                        new_rollmode
                                 )
                WHERE id = record_id;
            END LOOP;
    END
$$;

-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin

DO
$$
    DECLARE
        record_id          integer;
        current_definition jsonb;
        original_rollmode  text;
    BEGIN
        FOR record_id, current_definition IN
            SELECT id, definition
            FROM model_v1
            LOOP
                -- Retrieve the current type of rollmode
                original_rollmode := current_definition -> 'elasticsearchOptions' -> 'rollmode' ->> 'type';

                -- Restore rollmode to its initial structure (string)
                UPDATE model_v1
                SET definition = jsonb_set(
                        current_definition,
                        '{elasticsearchOptions,rollmode}',
                        to_jsonb(original_rollmode)
                                 )
                WHERE id = record_id;
            END LOOP;
    END
$$;

-- +goose StatementEnd
