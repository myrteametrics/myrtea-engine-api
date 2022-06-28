create table issue_detection_feedback_v3 (
	id serial primary key,
	issue_id integer references issues_v1(id) not null,
	-- user_id integer references users_v1(id) not null,
	date timestamptz not null,
	rating integer,
	CONSTRAINT unq_issueid_userid UNIQUE(issue_id,user_id)
);

TRUNCATE issue_detection_feedback_v3;
ALTER TABLE issue_detection_feedback_v3 ALTER COLUMN user_id TYPE uuid;
ALTER TABLE situation_definition_v1 DROP COLUMN groups;
-- ALTER TABLE notifications_history_v1 ALTER COLUMN groups TYPE []uuid;

create table situation_definition_v1 (
	id serial primary key,	
    -- groups integer[] not null,
	name varchar(100) not null unique,
	definition json,
	is_template boolean,
	is_object boolean,
	calendar_id integer REFERENCES calendar_v1 (id),
	last_modified timestamptz not null
);

create table notifications_history_v1 (
	id serial primary key,
	-- groups integer[] not null,
	data json,
	created_at timestamptz not null,
	isread boolean DEFAULT FALSE
);