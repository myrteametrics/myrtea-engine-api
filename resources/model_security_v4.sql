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

drop table if exists roles_v4;
create table roles_v4 (
	id uuid primary key,
	name varchar(100) not null UNIQUE
);

drop table if exists permissions_v4;
create table permissions_v4 (
	id uuid primary key,
	resource_type varchar(100) not null,
	resource_id varchar(100) not null,
	action varchar(100) not null
);

drop table if exists users_roles_v4;
create table users_roles_v4 (
	user_id uuid REFERENCES users_v4 (id),
	role_id uuid REFERENCES roles_v4 (id)
);

drop table if exists roles_permissions_v4;
create table roles_permissions_v4 (
	role_id uuid REFERENCES roles_v4 (id),
	permission_id uuid REFERENCES permissions_v4 (id)
);