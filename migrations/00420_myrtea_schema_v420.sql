-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users_v4
(
    id         uuid PRIMARY KEY NOT NULL,
    login      varchar(100)     NOT NULL UNIQUE,
    password   varchar(100)     NOT NULL,
    created    timestamptz      NOT NULL,
    last_name  varchar(100)     NOT NULL,
    first_name varchar(100)     NOT NULL,
    email      varchar(100)     NOT NULL,
    phone      varchar(100)
);

CREATE TABLE IF NOT EXISTS roles_v4
(
    id   uuid PRIMARY KEY,
    name varchar(100) NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS permissions_v4
(
    id            uuid PRIMARY KEY,
    resource_type varchar(100) NOT NULL,
    resource_id   varchar(100) NOT NULL,
    action        varchar(100) NOT NULL
);

CREATE TABLE IF NOT EXISTS users_roles_v4
(
    user_id uuid REFERENCES users_v4 (id),
    role_id uuid REFERENCES roles_v4 (id)
);

CREATE TABLE IF NOT EXISTS roles_permissions_v4
(
    role_id       uuid REFERENCES roles_v4 (id),
    permission_id uuid REFERENCES permissions_v4 (id)
);

-- existing table alteration
ALTER TABLE situation_definition_v1
    DROP COLUMN IF EXISTS groups;

UPDATE situation_definition_v1
SET definition = definition::jsonb - 'groups';

ALTER TABLE notifications_history_v1
    DROP COLUMN IF EXISTS groups;

TRUNCATE issue_detection_feedback_v3;

ALTER TABLE issue_detection_feedback_v3
    DROP CONSTRAINT if EXISTS unq_issueid_userid;

ALTER TABLE issue_detection_feedback_v3
    DROP COLUMN IF EXISTS user_id;

ALTER TABLE issue_detection_feedback_v3
    ADD COLUMN user_id uuid REFERENCES users_v4 (id);

ALTER TABLE issue_detection_feedback_v3
    ADD CONSTRAINT unq_issueid_userid UNIQUE (issue_id, user_id);
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin

ALTER TABLE situation_definition_v1
    ADD COLUMN IF NOT EXISTS groups INTEGER[] NOT NULL;

ALTER TABLE notifications_history_v1
    ADD COLUMN IF NOT EXISTS groups INTEGER[] NOT NULL;

ALTER TABLE issue_detection_feedback_v3
    DROP CONSTRAINT if EXISTS unq_issueid_userid;

ALTER TABLE issue_detection_feedback_v3
    DROP COLUMN IF EXISTS user_id;

ALTER TABLE issue_detection_feedback_v3
    ADD COLUMN user_id integer REFERENCES users_v1 (id) NOT NULL;

ALTER TABLE issue_detection_feedback_v3
    ADD CONSTRAINT unq_issueid_userid UNIQUE (issue_id, user_id);

DROP TABLE IF EXISTS roles_permissions_v4;
DROP TABLE IF EXISTS users_roles_v4;
DROP TABLE IF EXISTS users_v4;
DROP TABLE IF EXISTS roles_v4;
DROP TABLE IF EXISTS permissions_v4;
-- +goose StatementEnd

