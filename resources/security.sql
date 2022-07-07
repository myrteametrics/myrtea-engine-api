delete from users_roles_v4;
delete from roles_permissions_v4;
delete from users_v4;
delete from roles_v4;
delete from permissions_v4;
insert into users_v4 (id, login, password, created, last_name, first_name, email, phone) VALUES ('00000000-0000-0000-0000-000000000000', 'admin', crypt('myrtea' ,gen_salt('md5')), current_timestamp, 'admin', 'admin', 'admin@admin.com', '0123456789');
insert into roles_v4 (id, name) VALUES ('00000000-0000-0001-0000-000000000000', 'admin');
insert into permissions_v4 (id, resource_type, resource_id, action) VALUES ('00000000-0000-0002-0000-000000000000', '*', '*', '*');
insert into users_roles_v4 (user_id, role_id) VALUES ('00000000-0000-0000-0000-000000000000', '00000000-0000-0001-0000-000000000000');
insert into roles_permissions_v4 (role_id, permission_id) VALUES ('00000000-0000-0001-0000-000000000000', '00000000-0000-0002-0000-000000000000');

insert into users_v4 (id, login, password, created, last_name, first_name, email, phone) VALUES ('00000000-0000-0000-0001-000000000000', 'nopermission', crypt('nopermission' ,gen_salt('md5')), current_timestamp, 'nopermission', 'nopermission', 'nopermission@admin.com', '0123456789');
insert into roles_v4 (id, name) VALUES ('00000000-0000-0001-0001-000000000000', 'nopermission');

insert into users_v4 (id, login, password, created, last_name, first_name, email, phone) VALUES ('00000000-0000-0000-0002-000000000000', 'user1', crypt('user1' ,gen_salt('md5')), current_timestamp, 'user1', 'user1', 'user1@admin.com', '0123456789');
insert into roles_v4 (id, name) VALUES ('00000000-0000-0001-0002-000000000000', 'user1');
insert into permissions_v4 (id, resource_type, resource_id, action) VALUES ('00000000-0000-0002-0002-000000000000', '*', '*', 'list');
insert into permissions_v4 (id, resource_type, resource_id, action) VALUES ('00000000-0000-0002-0002-000000000001', 'model', '1', 'get');
insert into permissions_v4 (id, resource_type, resource_id, action) VALUES ('00000000-0000-0002-0002-000000000002', 'fact', '1', 'get');
insert into permissions_v4 (id, resource_type, resource_id, action) VALUES ('00000000-0000-0002-0002-000000000003', 'situation', '1', 'get');
insert into permissions_v4 (id, resource_type, resource_id, action) VALUES ('00000000-0000-0002-0002-000000000004', 'rule', '1', 'get');
insert into permissions_v4 (id, resource_type, resource_id, action) VALUES ('00000000-0000-0002-0002-000000000005', 'scheduler', '1', 'get');
insert into permissions_v4 (id, resource_type, resource_id, action) VALUES ('00000000-0000-0002-0002-000000000006', 'calendar', '1', 'get');
insert into permissions_v4 (id, resource_type, resource_id, action) VALUES ('00000000-0000-0002-0002-000000000007', 'frontend', 'view.user-settings', '*');
insert into permissions_v4 (id, resource_type, resource_id, action) VALUES ('00000000-0000-0002-0002-000000000008', 'frontend', 'view.home', '*');
insert into permissions_v4 (id, resource_type, resource_id, action) VALUES ('00000000-0000-0002-0002-000000000009', 'frontend', 'view.settings', '*');
insert into permissions_v4 (id, resource_type, resource_id, action) VALUES ('00000000-0000-0002-0002-000000000010', 'frontend', 'view.administration', '*');
insert into users_roles_v4 (user_id, role_id) VALUES ('00000000-0000-0000-0002-000000000000', '00000000-0000-0001-0002-000000000000');
insert into roles_permissions_v4 (role_id, permission_id) VALUES ('00000000-0000-0001-0002-000000000000', '00000000-0000-0002-0002-000000000000');
insert into roles_permissions_v4 (role_id, permission_id) VALUES ('00000000-0000-0001-0002-000000000000', '00000000-0000-0002-0002-000000000001');
insert into roles_permissions_v4 (role_id, permission_id) VALUES ('00000000-0000-0001-0002-000000000000', '00000000-0000-0002-0002-000000000002');
insert into roles_permissions_v4 (role_id, permission_id) VALUES ('00000000-0000-0001-0002-000000000000', '00000000-0000-0002-0002-000000000003');
insert into roles_permissions_v4 (role_id, permission_id) VALUES ('00000000-0000-0001-0002-000000000000', '00000000-0000-0002-0002-000000000004');
insert into roles_permissions_v4 (role_id, permission_id) VALUES ('00000000-0000-0001-0002-000000000000', '00000000-0000-0002-0002-000000000005');
insert into roles_permissions_v4 (role_id, permission_id) VALUES ('00000000-0000-0001-0002-000000000000', '00000000-0000-0002-0002-000000000006');
insert into roles_permissions_v4 (role_id, permission_id) VALUES ('00000000-0000-0001-0002-000000000000', '00000000-0000-0002-0002-000000000007');
insert into roles_permissions_v4 (role_id, permission_id) VALUES ('00000000-0000-0001-0002-000000000000', '00000000-0000-0002-0002-000000000008');
insert into roles_permissions_v4 (role_id, permission_id) VALUES ('00000000-0000-0001-0002-000000000000', '00000000-0000-0002-0002-000000000009');
insert into roles_permissions_v4 (role_id, permission_id) VALUES ('00000000-0000-0001-0002-000000000000', '00000000-0000-0002-0002-000000000010');


insert into users_v4 (id, login, password, created, last_name, first_name, email, phone) VALUES ('00000000-0000-0000-0003-000000000000', 'user2', crypt('user2' ,gen_salt('md5')), current_timestamp, 'user2', 'user2', 'user2@admin.com', '0123456789');
insert into roles_v4 (id, name) VALUES ('00000000-0000-0001-0003-000000000000', 'user2');
insert into permissions_v4 (id, resource_type, resource_id, action) VALUES ('00000000-0000-0002-0003-000000000000', '*', '*', 'list');
insert into permissions_v4 (id, resource_type, resource_id, action) VALUES ('00000000-0000-0002-0003-000000000001', '*', '*', 'get');
insert into permissions_v4 (id, resource_type, resource_id, action) VALUES ('00000000-0000-0002-0003-000000000002', 'fact', '*', 'list');
insert into permissions_v4 (id, resource_type, resource_id, action) VALUES ('00000000-0000-0002-0003-000000000003', 'fact', '*', 'get');
insert into permissions_v4 (id, resource_type, resource_id, action) VALUES ('00000000-0000-0002-0003-000000000004', 'frontend', 'view.user-settings', '*');
insert into permissions_v4 (id, resource_type, resource_id, action) VALUES ('00000000-0000-0002-0003-000000000005', 'frontend', 'view.settings', '*');
insert into users_roles_v4 (user_id, role_id) VALUES ('00000000-0000-0000-0003-000000000000', '00000000-0000-0001-0003-000000000000');
insert into roles_permissions_v4 (role_id, permission_id) VALUES ('00000000-0000-0001-0003-000000000000', '00000000-0000-0002-0003-000000000000');
insert into roles_permissions_v4 (role_id, permission_id) VALUES ('00000000-0000-0001-0003-000000000000', '00000000-0000-0002-0003-000000000001');
insert into roles_permissions_v4 (role_id, permission_id) VALUES ('00000000-0000-0001-0003-000000000000', '00000000-0000-0002-0003-000000000002');
insert into roles_permissions_v4 (role_id, permission_id) VALUES ('00000000-0000-0001-0003-000000000000', '00000000-0000-0002-0003-000000000003');
insert into roles_permissions_v4 (role_id, permission_id) VALUES ('00000000-0000-0001-0003-000000000000', '00000000-0000-0002-0003-000000000004');
insert into roles_permissions_v4 (role_id, permission_id) VALUES ('00000000-0000-0001-0003-000000000000', '00000000-0000-0002-0003-000000000005');

