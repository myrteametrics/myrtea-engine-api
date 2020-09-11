create extension pgcrypto;

drop table if exists issue_detection_feedback_v3;
drop table if exists job_schedules_v1;
drop table if exists user_memberships_v1;
drop table if exists situation_rules_v1;
drop table if exists notifications_history_v1;
drop table if exists issue_resolution_v1;
drop table if exists issue_resolution_draft_v1;
drop table if exists issues_v1;
drop table if exists ref_action_v1;
drop table if exists ref_rootcause_v1;
drop table if exists rule_versions_v1;
drop table if exists rules_v1;
drop table if exists fact_history_v1;
drop table if exists situation_history_v1;
drop table if exists situation_facts_v1;
drop table if exists fact_definition_v1;
drop table if exists situation_template_instances_v1;
drop table if exists situation_definition_v1;
drop table if exists calendar_union_v1;
drop table if exists calendar_v1;
drop table if exists user_groups_v1;
drop table if exists users_v1;
drop table if exists elasticsearch_indices_v1;
drop table if exists model_v1;


create table users_v1 (
	id serial primary key not null,
	login varchar(100) not null unique,
	password varchar(100) not null,
	role integer not null,
	created timestamptz not null,
	last_name varchar(100) not null,
	first_name varchar(100) not null,
	email varchar(100) not null,
	phone varchar(100)
);
insert into users_v1 (id, login, password, role, created, last_name, first_name, email, phone) VALUES (DEFAULT, 'admin', crypt('myrtea' ,gen_salt('md5')), 1, current_timestamp, 'admin', 'admin', 'admin@admin.com', '0123456789');

create table user_groups_v1 (
    id serial primary key not null,
    name varchar(100) not null UNIQUE
);
insert into user_groups_v1 (id, name) VALUES (DEFAULT, 'administrators');

create table user_memberships_v1 (
    user_id integer REFERENCES users_v1 (id),
	group_id integer REFERENCES user_groups_v1 (id),
	role integer,
	PRIMARY KEY(user_id, group_id)
);
insert into user_memberships_v1 (user_id, group_id, role ) VALUES (1,1,1);

create table if not exists fact_history_v1 (
	id integer not null,
	ts timestamptz not null,
	situation_id integer,
	situation_instance_id integer,
	result jsonb,
	success boolean,
	primary key (id, ts, situation_id, situation_instance_id)
);

create table situation_history_v1 (
	id integer,
	ts timestamptz not null,
	situation_instance_id integer,
	facts_ids json,
	parameters json,
	expression_facts jsonb,
	metadatas json,
	evaluated boolean,
	primary key (id, ts, situation_instance_id)
);


create table fact_definition_v1 (
	id serial primary key,
	name varchar(100) not null unique,
    definition json,
 	last_modified timestamptz not null
);

create table calendar_v1 (
	id serial primary key,
	name varchar(100) not null,
	description varchar(500) not null,
	period_data JSONB not null,
	enabled boolean not null,
	creation_date timestamptz not null,
	last_modified timestamptz not null
);

create table calendar_union_v1 (
	calendar_id integer REFERENCES calendar_v1 (id),
	sub_calendar_id integer REFERENCES calendar_v1 (id),
	priority integer,
	PRIMARY KEY(calendar_id, sub_calendar_id)
);

create table situation_definition_v1 (
	id serial primary key,
	groups integer[] not null,
	name varchar(100) not null unique,
	definition json,
	is_template boolean,
	is_object boolean,
	calendar_id integer REFERENCES calendar_v1 (id),
	last_modified timestamptz not null
);

create table situation_template_instances_v1 (
	id serial primary key,
	name varchar(100) not null unique,
	situation_id integer REFERENCES situation_definition_v1 (id),
	parameters json,
	calendar_id integer REFERENCES calendar_v1 (id),
	last_modified timestamptz not null
);

create table situation_facts_v1 (
	situation_id integer REFERENCES situation_definition_v1 (id),
	fact_id integer REFERENCES fact_definition_v1 (id),
	PRIMARY KEY(situation_id, fact_id)
);

create table notifications_history_v1 (
	id serial primary key,
	groups integer[] not null,
	data json,
	created_at timestamptz not null,
	isread boolean DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS rules_v1 (
	id serial PRIMARY KEY,
	name varchar(100) not null UNIQUE,
	enabled boolean not null,
	calendar_id integer REFERENCES calendar_v1 (id),
	last_modified timestamptz not null
);

CREATE TABLE IF NOT EXISTS rule_versions_v1 (
	rule_id INTEGER REFERENCES rules_v1 (id),
	version_number INTEGER not null,
	data json not null,
	creation_datetime timestamptz not null,
	PRIMARY KEY(rule_id, version_number)
);

create table issues_v1 (
	id serial primary key,
	key varchar(100) not null,
	name varchar(100) not null,
	level varchar(100) not null,
	situation_id integer REFERENCES situation_definition_v1 (id),
	situation_instance_id integer,
	situation_date timestamptz not null,
	expiration_date timestamptz not null,
	rule_data JSONB not null,
	state varchar(100) not null,
	last_modified timestamptz not null,
	created_at timestamptz not null,
	detection_rating_avg real,
	assigned_at timestamptz,
	assigned_to varchar(100),
	closed_at timestamptz,
	closed_by varchar(100)
);

CREATE TABLE job_schedules_v1 (
	id serial primary key,
	name varchar(100) not null,
	cronexpr varchar(100) not null,
	job_type varchar(100) not null,
	job_data json not null,
	last_modified timestamptz not null
);

create table ref_rootcause_v1 (
	id serial primary key,
	name varchar(100) not null,
	description varchar(500) not null,
	situation_id integer REFERENCES situation_definition_v1 (id),
	rule_id integer REFERENCES rules_v1 (id),
	CONSTRAINT unq_name_situationid UNIQUE(name,situation_id)
);

create table ref_action_v1 (
	id serial primary key,
	name varchar(100) not null,
	description varchar(500) not null,
	rootcause_id integer REFERENCES ref_rootcause_v1 (id),
	CONSTRAINT unq_name_rootcauseid UNIQUE(name,rootcause_id)
);

create table issue_resolution_v1 (
	feedback_date timestamptz not null,
	issue_id integer REFERENCES issues_v1 (id),
	rootcause_id integer REFERENCES ref_rootcause_v1 (id),
	action_id integer REFERENCES ref_action_v1 (id),
	CONSTRAINT unq_issue_rc_action UNIQUE(issue_id,rootcause_id,action_id)
);

create table issue_resolution_draft_v1 (
	issue_id integer REFERENCES issues_v1 (id) primary key,
	concurrency_uuid varchar(100) unique not null,
	last_modified timestamptz not null,
	data JSONB not null
);

create table IF NOT EXISTS model_v1 (
	id serial primary key,
	name varchar(100) unique not null,
	definition JSONB
);

create table IF NOT EXISTS elasticsearch_indices_v1 (
	id serial primary key,
	logical varchar(100) not null,
	technical varchar(100) not null,
	creation_date timestamptz not null
);

create table situation_rules_v1 (
	situation_id integer REFERENCES situation_definition_v1 (id),
	rule_id integer REFERENCES rules_v1 (id),
	execution_order integer,
	PRIMARY KEY(situation_id, rule_id)
);

create table issue_detection_feedback_v3 (
	id serial primary key,
	issue_id integer references issues_v1(id) not null,
	user_id integer references users_v1(id) not null,
	date timestamptz not null,
	rating integer,
	CONSTRAINT unq_issueid_userid UNIQUE(issue_id,user_id)
);

create table connectors_executions_log_v1 (
	id serial primary key not null,
	connector_id varchar(100) not null,
	name varchar(100) not null,
    ts timestamptz not null,
	success boolean
);