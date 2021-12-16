-- drop table if exists users_v4;

-- create table users_v4 (
-- 	id uuid primary key not null,
-- 	login varchar(100) not null unique,
-- 	password varchar(100) not null,
-- 	created timestamptz not null,
-- 	last_name varchar(100) not null,
-- 	first_name varchar(100) not null,
-- 	email varchar(100) not null,
-- 	phone varchar(100)
-- );

insert into users_v4 (id, login, password, created, last_name, first_name, email, phone) VALUES ('5bb8d3b8-a6f4-481b-95a3-ce4b81aa95c3', 'nopermission', crypt('nopermission' ,gen_salt('md5')), current_timestamp, 'nopermission', 'nopermission', 'nopermission@admin.com', '0123456789');

-- drop table if exists roles_v4;

-- create table roles_v4 (
-- 		id uuid primary key,
-- 		name varchar(100) not null UNIQUE
-- 	);

insert into roles_v4 (id, name) VALUES ('81fa129e-5208-49b7-8085-b73ff355da9e', 'nopermission');

-- drop table if exists permissions_v4;

-- create table permissions_v4 (
-- 		id uuid primary key,
-- 		resource_type varchar(100) not null,
-- 		resource_id varchar(100) not null,
-- 		action varchar(100) not null
-- 	);

-- insert into permissions_v4 (id, resource_type, resource_id, action) VALUES ('b6a97654-5c96-4a05-b6dc-dc2088e1551e', '*', '*', '*');

-- drop table if exists users_roles_v4;

-- create table users_roles_v4 (
-- 		user_id uuid REFERENCES users_v4 (id),
-- 		role_id uuid REFERENCES roles_v4 (id)
-- 	);

insert into users_roles_v4 (user_id, role_id) VALUES ('5bb8d3b8-a6f4-481b-95a3-ce4b81aa95c3', '81fa129e-5208-49b7-8085-b73ff355da9e');

-- drop table if exists roles_permissions_v4;

-- create table roles_permissions_v4 (
-- 		role_id uuid REFERENCES roles_v4 (id),
-- 		permission_id uuid REFERENCES permissions_v4 (id)
-- 	);

-- insert into roles_permissions_v4 (role_id, permission_id) VALUES ('81fa129e-5208-49b7-8085-b73ff355da9e', 'b6a97654-5c96-4a05-b6dc-dc2088e1551e');