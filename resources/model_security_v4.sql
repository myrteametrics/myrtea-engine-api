-- new tables
drop table if exists roles_permissions_v4;
drop table if exists users_roles_v4;
drop table if exists users_v4;
drop table if exists roles_v4;
drop table if exists permissions_v4;

create table users_v4 (
	id uuid primary key not null,
	login varchar(100) not null unique,
	password varchar(100) not null,
	created timestamptz not null,
	last_name varchar(100) not null,
	first_name varchar(100) not null,
	email varchar(100) not null,
	phone varchar(100)
);
create table roles_v4 (
	id uuid primary key,
	name varchar(100) not null UNIQUE
);
create table permissions_v4 (
	id uuid primary key,
	resource_type varchar(100) not null,
	resource_id varchar(100) not null,
	action varchar(100) not null
);
create table users_roles_v4 (
	user_id uuid REFERENCES users_v4 (id),
	role_id uuid REFERENCES roles_v4 (id)
);
create table roles_permissions_v4 (
	role_id uuid REFERENCES roles_v4 (id),
	permission_id uuid REFERENCES permissions_v4 (id)
);

insert into users_v4 (id, login, password, created, last_name, first_name, email, phone) values ('a86fb11b-0e01-4622-88d2-2eced9016cb4', 'admin', crypt('myrtea', gen_salt('md5')), now(), 'admin', 'admin', 'admin@gmail.com', '0123456789');
insert into roles_v4 (id, name) values ('ffa5fa16-cd5d-44f9-b785-5619c7f0bda8', 'admin');
insert into permissions_v4 (id, resource_type, resource_id, action) values ('10e62964-bbde-4c75-8b0a-5a1f9417daba', '*', '*', '*');
insert into users_roles_v4 (user_id, role_id) values ('a86fb11b-0e01-4622-88d2-2eced9016cb4', 'ffa5fa16-cd5d-44f9-b785-5619c7f0bda8');
insert into roles_permissions_v4 (role_id, permission_id) values ('ffa5fa16-cd5d-44f9-b785-5619c7f0bda8', '10e62964-bbde-4c75-8b0a-5a1f9417daba');

-- existing table alteration
alter table situation_definition_v1 drop column if exists groups;
update situation_definition_v1 set definition = definition::jsonb - 'groups';

alter table notifications_history_v1 drop column if exists groups;

truncate issue_detection_feedback_v3;
alter table issue_detection_feedback_v3 drop constraint if exists unq_issueid_userid;
alter table issue_detection_feedback_v3 drop column if exists user_id;
alter table issue_detection_feedback_v3 add column user_id uuid references users_v4(id);
alter table issue_detection_feedback_v3 add constraint unq_issueid_userid unique(issue_id,user_id);

 update rule_versions_v1 set data = replace(data::text, 'timeZone', 'timezone')::json;

