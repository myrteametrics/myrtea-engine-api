-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS fact_history_v5
(
    id                    serial PRIMARY KEY,
    fact_id               integer,
    situation_id          integer,
    situation_instance_id integer,
    ts                    timestamptz NOT NULL,
    result                jsonb
);

CREATE TABLE IF NOT EXISTS situation_history_v5
(
    id                    serial PRIMARY KEY,
    situation_id          integer,
    situation_instance_id integer,
    ts                    timestamptz NOT NULL,
    parameters            json,
    expression_facts      jsonb,
    metadatas             json
);

CREATE TABLE IF NOT EXISTS situation_fact_history_v5
(
    situation_history_id integer REFERENCES situation_history_v5 (id),
    fact_history_id      integer REFERENCES fact_history_v5 (id),
    fact_id              integer,
    PRIMARY KEY (situation_history_id, fact_history_id)
);

SELECT count(*)
FROM fact_history_v1;
INSERT INTO fact_history_v5 (fact_id, situation_id, situation_instance_id, ts, result)
SELECT id, situation_id, situation_instance_id, ts, result
FROM fact_history_v1;

SELECT count(*)
FROM situation_history_v1;
INSERT INTO situation_history_v5 (situation_id, situation_instance_id, ts, parameters, expression_facts, metadatas)
SELECT id, situation_instance_id, ts, parameters, expression_facts, metadatas
FROM situation_history_v1;

SELECT count(*)
FROM situation_history_v1;
INSERT INTO situation_fact_history_v5 (situation_history_id, fact_history_id)
SELECT s4.id, f4.id
FROM (SELECT id                                                                                      AS situation_id,
             situation_instance_id,
             ts                                                                                      AS situation_ts,
             js.key::integer                                                                         AS fact_id,
             to_timestamp(replace(js.value::text, '"', ''), 'YYYY-MM-DD HH24:MI:SS.US')::timestamptz AS fact_ts
      FROM situation_history_v1,
           json_each(situation_history_v1.facts_ids) AS js
      WHERE js.value::text <> 'null'
        AND js.value::text <> '"null"') sf
         INNER JOIN situation_history_v5 s4 ON
    sf.situation_id = s4.situation_id AND
    sf.situation_instance_id = s4.situation_instance_id AND
    sf.situation_ts = s4.ts
         INNER JOIN fact_history_v5 f4 ON
    sf.fact_id = f4.fact_id AND
    (sf.situation_id = f4.situation_id OR f4.situation_id = 0) AND
    (sf.situation_instance_id = f4.situation_instance_id OR f4.situation_instance_id = 0) AND
    sf.fact_ts = f4.ts;

UPDATE situation_fact_history_v5
SET fact_id = f.fact_id
FROM fact_history_v5 f
WHERE f.id = fact_history_id;

CREATE INDEX IF NOT EXISTS idx_fact_history_v5_combo ON fact_history_v5 (fact_id, ts DESC) include (id);
CREATE INDEX IF NOT EXISTS idx_situation_fact_history_v5_situation_history_id ON situation_fact_history_v5 (situation_history_id);
CREATE INDEX IF NOT EXISTS idx_situation_history_v5_combo ON situation_history_v5 (situation_id, situation_instance_id, ts DESC) include (id);
CREATE INDEX IF NOT EXISTS idx_connectors_executions_log_connector_id_name_ts ON connectors_executions_log_v1 (connector_id, name, ts DESC);

ALTER TABLE issues_v1
    ADD COLUMN IF NOT EXISTS situation_history_id integer REFERENCES situation_history_v5 (id);

UPDATE issues_v1 i
SET situation_history_id = sh.id
FROM situation_history_v5 sh
WHERE sh.situation_id = i.situation_id
  AND sh.situation_instance_id = i.situation_instance_id
  AND sh.ts = i.situation_date;
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
DROP INDEX if EXISTS idx_connectors_executions_log_connector_id_name_ts;

ALTER TABLE issues_v1
    DROP COLUMN IF EXISTS situation_history_id;

DROP TABLE IF EXISTS situation_fact_history_v5;
DROP TABLE IF EXISTS situation_history_v5;
DROP TABLE IF EXISTS fact_history_v5;
-- +goose StatementEnd

