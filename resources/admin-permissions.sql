drop table if exists users_v4;

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

insert into users_v4 (id, login, password, created, last_name, first_name, email, phone) VALUES ('a14f7e12-b844-4fb1-9cdb-b2408f4bb3d6', 'admin', crypt('myrtea' ,gen_salt('md5')), current_timestamp, 'admin', 'admin', 'admin@admin.com', '0123456789');

drop table if exists roles_v4;

create table roles_v4 (
		id uuid primary key,
		name varchar(100) not null UNIQUE
	);

insert into roles_v4 (id, name) VALUES ('b7d0af68-8fa3-46af-997f-367b5136d4d9', 'admin');

drop table if exists permissions_v4;

create table permissions_v4 (
		id uuid primary key,
		resource_type varchar(100) not null,
		resource_id varchar(100) not null,
		action varchar(100) not null
	);

insert into permissions_v4 (id, resource_type, resource_id, action) VALUES ('d010d928-0bc2-4be5-8f20-0fda4d966a6a', '*', '*', '*');

drop table if exists users_roles_v4;

create table users_roles_v4 (
		user_id uuid REFERENCES users_v4 (id),
		role_id uuid REFERENCES roles_v4 (id)
	);

insert into users_roles_v4 (user_id, role_id) VALUES ('a14f7e12-b844-4fb1-9cdb-b2408f4bb3d6', 'b7d0af68-8fa3-46af-997f-367b5136d4d9');

drop table if exists roles_permissions_v4;

create table roles_permissions_v4 (
		role_id uuid REFERENCES roles_v4 (id),
		permission_id uuid REFERENCES permissions_v4 (id)
	);

insert into roles_permissions_v4 (role_id, permission_id) VALUES ('b7d0af68-8fa3-46af-997f-367b5136d4d9', 'd010d928-0bc2-4be5-8f20-0fda4d966a6a');