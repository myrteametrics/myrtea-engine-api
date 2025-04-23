-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION if NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS users_v1
(
    id         serial PRIMARY KEY NOT NULL,
    login      varchar(100)       NOT NULL UNIQUE,
    password   varchar(100)       NOT NULL,
    role       integer            NOT NULL,
    created    timestamptz        NOT NULL,
    last_name  varchar(100)       NOT NULL,
    first_name varchar(100)       NOT NULL,
    email      varchar(100)       NOT NULL,
    phone      varchar(100)
);

CREATE TABLE IF NOT EXISTS user_groups_v1
(
    id   serial PRIMARY KEY NOT NULL,
    name varchar(100)       NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS user_memberships_v1
(
    user_id  integer REFERENCES users_v1 (id),
    group_id integer REFERENCES user_groups_v1 (id),
    role     integer,
    PRIMARY KEY (user_id, group_id)
);

CREATE TABLE IF NOT EXISTS fact_history_v1
(
    id                    integer     NOT NULL,
    ts                    timestamptz NOT NULL,
    situation_id          integer,
    situation_instance_id integer,
    result                jsonb,
    success               boolean,
    PRIMARY KEY (id, ts, situation_id, situation_instance_id)
);

CREATE TABLE IF NOT EXISTS situation_history_v1
(
    id                    integer,
    ts                    timestamptz NOT NULL,
    situation_instance_id integer,
    facts_ids             json,
    parameters            json,
    expression_facts      jsonb,
    metadatas             json,
    evaluated             boolean,
    PRIMARY KEY (id, ts, situation_instance_id)
);

CREATE TABLE IF NOT EXISTS fact_definition_v1
(
    id            serial PRIMARY KEY,
    name          varchar(100) NOT NULL UNIQUE,
    definition    json,
    last_modified timestamptz  NOT NULL
);

CREATE TABLE IF NOT EXISTS calendar_v1
(
    id            serial PRIMARY KEY,
    name          varchar(100) NOT NULL,
    description   varchar(500) NOT NULL,
    period_data   jsonb        NOT NULL,
    enabled       boolean      NOT NULL,
    creation_date timestamptz  NOT NULL,
    last_modified timestamptz  NOT NULL
);

CREATE TABLE IF NOT EXISTS calendar_union_v1
(
    calendar_id     integer REFERENCES calendar_v1 (id),
    sub_calendar_id integer REFERENCES calendar_v1 (id),
    priority        integer,
    PRIMARY KEY (calendar_id, sub_calendar_id)
);

CREATE TABLE IF NOT EXISTS situation_definition_v1
(
    id            serial PRIMARY KEY,
    groups        integer[]    NOT NULL,
    name          varchar(100) NOT NULL UNIQUE,
    definition    json,
    is_template   boolean,
    is_object     boolean,
    calendar_id   integer REFERENCES calendar_v1 (id),
    last_modified timestamptz  NOT NULL
);

CREATE TABLE IF NOT EXISTS situation_template_instances_v1
(
    id            serial PRIMARY KEY,
    name          varchar(100) NOT NULL,
    situation_id  integer REFERENCES situation_definition_v1 (id),
    parameters    json,
    calendar_id   integer REFERENCES calendar_v1 (id),
    last_modified timestamptz  NOT NULL,
    CONSTRAINT unq_situation_template_instances_v1_situationid_name UNIQUE (situation_id, name)
);

CREATE TABLE IF NOT EXISTS situation_facts_v1
(
    situation_id integer REFERENCES situation_definition_v1 (id),
    fact_id      integer REFERENCES fact_definition_v1 (id),
    PRIMARY KEY (situation_id, fact_id)
);

CREATE TABLE IF NOT EXISTS notifications_history_v1
(
    id         serial PRIMARY KEY,
    groups     integer[]   NOT NULL,
    data       json,
    created_at timestamptz NOT NULL,
    isread     boolean DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS rules_v1
(
    id            serial PRIMARY KEY,
    name          varchar(100) NOT NULL UNIQUE,
    enabled       boolean      NOT NULL,
    calendar_id   integer REFERENCES calendar_v1 (id),
    last_modified timestamptz  NOT NULL
);

CREATE TABLE IF NOT EXISTS rule_versions_v1
(
    rule_id           integer REFERENCES rules_v1 (id),
    version_number    integer     NOT NULL,
    data              json        NOT NULL,
    creation_datetime timestamptz NOT NULL,
    PRIMARY KEY (rule_id, version_number)
);

CREATE TABLE IF NOT EXISTS issues_v1
(
    id                    serial PRIMARY KEY,
    key                   varchar(100) NOT NULL,
    name                  varchar(100) NOT NULL,
    level                 varchar(100) NOT NULL,
    situation_id          integer REFERENCES situation_definition_v1 (id),
    situation_instance_id integer,
    situation_date        timestamptz  NOT NULL,
    expiration_date       timestamptz  NOT NULL,
    rule_data             jsonb        NOT NULL,
    state                 varchar(100) NOT NULL,
    last_modified         timestamptz  NOT NULL,
    created_at            timestamptz  NOT NULL,
    detection_rating_avg  real,
    assigned_at           timestamptz,
    assigned_to           varchar(100),
    closed_at             timestamptz,
    closed_by             varchar(100)
);

CREATE TABLE IF NOT EXISTS job_schedules_v1
(
    id            serial PRIMARY KEY,
    name          varchar(100) NOT NULL,
    cronexpr      varchar(100) NOT NULL,
    job_type      varchar(100) NOT NULL,
    job_data      json         NOT NULL,
    last_modified timestamptz  NOT NULL
);

CREATE TABLE IF NOT EXISTS ref_rootcause_v1
(
    id           serial PRIMARY KEY,
    name         varchar(100) NOT NULL,
    description  varchar(500) NOT NULL,
    situation_id integer REFERENCES situation_definition_v1 (id),
    rule_id      integer REFERENCES rules_v1 (id),
    CONSTRAINT unq_name_situationid_ruleid UNIQUE (name, situation_id, rule_id)
);

CREATE TABLE IF NOT EXISTS ref_action_v1
(
    id           serial PRIMARY KEY,
    name         varchar(100) NOT NULL,
    description  varchar(500) NOT NULL,
    rootcause_id integer REFERENCES ref_rootcause_v1 (id),
    CONSTRAINT unq_name_rootcauseid UNIQUE (name, rootcause_id)
);

CREATE TABLE IF NOT EXISTS issue_resolution_v1
(
    feedback_date timestamptz NOT NULL,
    issue_id      integer REFERENCES issues_v1 (id),
    rootcause_id  integer REFERENCES ref_rootcause_v1 (id),
    action_id     integer REFERENCES ref_action_v1 (id),
    CONSTRAINT unq_issue_rc_action UNIQUE (issue_id, rootcause_id, action_id)
);

CREATE TABLE IF NOT EXISTS issue_resolution_draft_v1
(
    issue_id         integer REFERENCES issues_v1 (id) PRIMARY KEY,
    concurrency_uuid varchar(100) UNIQUE NOT NULL,
    last_modified    timestamptz         NOT NULL,
    data             jsonb               NOT NULL
);

CREATE TABLE IF NOT EXISTS model_v1
(
    id         serial PRIMARY KEY,
    name       varchar(100) UNIQUE NOT NULL,
    definition jsonb
);

CREATE TABLE IF NOT EXISTS elasticsearch_indices_v1
(
    id            serial PRIMARY KEY,
    logical       varchar(100) NOT NULL,
    technical     varchar(100) NOT NULL,
    creation_date timestamptz  NOT NULL
);

CREATE TABLE IF NOT EXISTS situation_rules_v1
(
    situation_id    integer REFERENCES situation_definition_v1 (id),
    rule_id         integer REFERENCES rules_v1 (id),
    execution_order integer,
    PRIMARY KEY (situation_id, rule_id)
);

CREATE TABLE IF NOT EXISTS issue_detection_feedback_v3
(
    id       serial PRIMARY KEY,
    issue_id integer REFERENCES issues_v1 (id) NOT NULL,
    user_id  integer REFERENCES users_v1 (id)  NOT NULL,
    date     timestamptz                       NOT NULL,
    rating   integer,
    CONSTRAINT unq_issueid_userid UNIQUE (issue_id, user_id)
);

CREATE TABLE IF NOT EXISTS connectors_config_v1
(
    id            serial PRIMARY KEY NOT NULL,
    connector_id  varchar(100)       NOT NULL UNIQUE,
    name          varchar(100)       NOT NULL,
    current       text               NOT NULL,
    previous      text,
    last_modified timestamptz        NOT NULL
);

CREATE TABLE IF NOT EXISTS connectors_executions_log_v1
(
    id           serial PRIMARY KEY NOT NULL,
    connector_id varchar(100)       NOT NULL,
    name         varchar(100)       NOT NULL,
    ts           timestamptz        NOT NULL,
    success      boolean
);
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
drop table if exists connectors_executions_log_v1;
drop table if exists connectors_config_v1;
drop table if exists issue_detection_feedback_v3;
drop table if exists situation_rules_v1;
drop table if exists elasticsearch_indices_v1;
drop table if exists model_v1;
drop table if exists issue_resolution_draft_v1;
drop table if exists issue_resolution_v1;
drop table if exists ref_action_v1;
drop table if exists ref_rootcause_v1;
drop table if exists job_schedules_v1;
drop table if exists issues_v1;
drop table if exists rule_versions_v1;
drop table if exists rules_v1;
drop table if exists notifications_history_v1;
drop table if exists situation_facts_v1;
drop table if exists situation_template_instances_v1;
drop table if exists situation_definition_v1;
drop table if exists calendar_union_v1;
drop table if exists calendar_v1;
drop table if exists fact_definition_v1;
drop table if exists situation_history_v1;
drop table if exists fact_history_v1;
drop table if exists user_memberships_v1;
drop table if exists user_groups_v1;
drop table if exists users_v1;

drop extension if exists pgcrypto;
-- +goose StatementEnd

