package tests

const (

	// UsersDropTableV1 SQL statement for table drop
	UsersDropTableV1 string = `DROP TABLE IF EXISTS users_v1;`
	// UsersTableV1 SQL statement for the users table
	UsersTableV1 string = `create table IF NOT EXISTS users_v1 (
		id serial primary key not null,
		login varchar(100) not null unique,
		password varchar(100) not null,
		role integer not null,
		created timestamptz not null,
		last_name varchar(100) not null,
		first_name varchar(100) not null,
		email varchar(100) not null,
		phone varchar(100)
	);`

	// GroupsDropTableV1 SQL statement for table drop
	GroupsDropTableV1 string = `DROP TABLE IF EXISTS user_groups_v1;`
	// GroupsTableV1 SQL statement for the groups table
	GroupsTableV1 string = `create table IF NOT EXISTS user_groups_v1 (
		id serial primary key not null,
		name varchar(100) not null UNIQUE
	);`

	//UserMembershipsDropTableV1 SQL statement for table drop
	UserMembershipsDropTableV1 string = `DROP TABLE IF EXISTS user_memberships_v1;`
	// UserMembershipsTableV1 SQL statement for the user memberships table
	UserMembershipsTableV1 string = `create table IF NOT EXISTS user_memberships_v1 (
		user_id integer REFERENCES users_v1 (id),
		group_id integer REFERENCES user_groups_v1 (id),
		role integer,
		PRIMARY KEY(user_id, group_id)
	);`

	// RulesDropTableV1 SQL statement for table drop
	RulesDropTableV1 string = `DROP TABLE IF EXISTS rules_v1;`
	// RulesTableV1 SQL statement for the rules table
	RulesTableV1 string = `CREATE TABLE IF NOT EXISTS rules_v1 (
		id serial PRIMARY KEY,
		name varchar(100) not null UNIQUE,
		enabled boolean not null,
		calendar_id integer REFERENCES calendar_v1 (id),
		last_modified timestamptz not null
	);`

	// RuleVersionsDropTableV1 SQL statement for table drop
	RuleVersionsDropTableV1 string = `DROP TABLE IF EXISTS rule_versions_v1;`
	// RuleVersionsTableV1 SQL statement for the rules table
	RuleVersionsTableV1 string = `CREATE TABLE IF NOT EXISTS rule_versions_v1 (
		rule_id INTEGER REFERENCES rules_v1 (id),
		version_number INTEGER not null,
		data json not null,
		creation_datetime timestamptz not null,
		PRIMARY KEY(rule_id, version_number)
	);`

	//SituationHistoryDropTableV1 SQL statement for table drop
	SituationHistoryDropTableV1 string = `DROP TABLE IF EXISTS situation_history_v1;`
	// SituationHistoryTableV1 SQL statement for the situation history
	SituationHistoryTableV1 string = `create table situation_history_v1 (
		id integer,
		ts timestamptz not null,
		situation_instance_id integer,
		facts_ids json,
		parameters json,
		expression_facts jsonb,
		metadatas json,
		evaluated boolean,
		primary key (id, ts, situation_instance_id)
	);`

	//FactHistoryDropTableV1 SQL statement for table drop
	FactHistoryDropTableV1 string = `DROP TABLE IF EXISTS fact_history_v1;`
	// FactHistoryTableV1 SQL statement for the fact history
	FactHistoryTableV1 string = `create table if not exists fact_history_v1 (
		id integer not null,
		ts timestamptz not null,
		situation_id integer,
		situation_instance_id integer,
		result jsonb,
		success boolean,
		primary key (id, ts, situation_id, situation_instance_id)
	);`

	//FactDefinitionDropTableV1 SQL statement for table drop
	FactDefinitionDropTableV1 string = `DROP TABLE IF EXISTS fact_definition_v1;`
	// FactDefinitionTableV1 SQL statement for the fact definition table
	FactDefinitionTableV1 string = `create table fact_definition_v1 (
		id serial primary key,
		name varchar(100) not null unique,
		definition json,
		 last_modified timestamptz not null
	);`

	// SituationDefinitionDropTableV1 SQL statement for table drop
	SituationDefinitionDropTableV1 string = `DROP TABLE IF EXISTS situation_definition_v1;`
	// SituationDefinitionTableV1 SQL statement for the situation definition table
	SituationDefinitionTableV1 string = `create table situation_definition_v1 (
		id serial primary key,
		groups integer[] not null,
		name varchar(100) not null unique,
		definition json,
		is_template boolean,
		is_object boolean,
		calendar_id integer REFERENCES calendar_v1 (id),
		last_modified timestamptz not null
	);`

	// SituationTemplateInstancesDropTableV1 SQL statement to drop table situation_template_instances_v1
	SituationTemplateInstancesDropTableV1 string = `DROP TABLE IF EXISTS situation_template_instances_v1;`
	// SituationTemplateInstancesTableV1 SQL statement to create table situation_template_instances_v1
	SituationTemplateInstancesTableV1 string = `create table situation_template_instances_v1 (
		id serial primary key,
		name varchar(100) not null unique,
		situation_id integer REFERENCES situation_definition_v1 (id),
		parameters json,
		calendar_id integer REFERENCES calendar_v1 (id),
		last_modified timestamptz not null
	);`

	// SituationFactsDropTableV1 SQL statement for table drop
	SituationFactsDropTableV1 string = `DROP TABLE IF EXISTS situation_facts_v1;`
	// SituationFactsTableV1 SQL statement for the situation facts table
	SituationFactsTableV1 string = `create table situation_facts_v1 (
		situation_id integer REFERENCES situation_definition_v1 (id),
		fact_id integer REFERENCES fact_definition_v1 (id),
		PRIMARY KEY(situation_id, fact_id)
	);`

	// NotificationHistoryDropTableV1 SQL statement for table drop
	NotificationHistoryDropTableV1 string = `DROP TABLE IF EXISTS notifications_history_v1;`
	// NotificationHistoryTableV1 SQL statement for the notification history
	NotificationHistoryTableV1 string = `create table notifications_history_v1 (
		id serial primary key not null,
		groups integer[] not null,
		data json,
		created_at timestamptz not null,
		isread boolean default FALSE
	);`

	// JobSchedulesDropTableV1 SQL statement for table drop job schedules
	JobSchedulesDropTableV1 string = `DROP TABLE IF EXISTS job_schedules_v1;`
	// JobSchedulesTableV1 SQL statement for the job schedules
	JobSchedulesTableV1 string = `CREATE TABLE job_schedules_v1 (
		id serial primary key,
		name varchar(100) not null,
		cronexpr varchar(100) not null,
		job_type varchar(100) not null,
		job_data json not null,
		last_modified timestamptz not null
	);`

	// IssuesDropTableV1 SQL statement for table drop
	IssuesDropTableV1 string = `DROP TABLE IF EXISTS issues_v1;`
	// IssuesTableV1 SQL statement for the issues table
	IssuesTableV1 string = `create table issues_v1 (
		id serial primary key,
		key varchar(100) not null,
		name varchar(100) not null,
		level varchar(100) not null,
		situation_id integer REFERENCES situation_definition_v1 (id),
		situation_instance_id integer,
		situation_date timestamptz not null,
		expiration_date timestamptz not null,
		rule_data JSONB,
		state varchar(100) not null,
		last_modified timestamptz not null,
		created_at timestamptz not null,
		detection_rating_avg real,
		assigned_at timestamptz,
		assigned_to varchar(100),
		closed_at timestamptz,
		closed_by varchar(100),
		comment text
	);`

	// RefRootCauseDropTableV1 SQL statement for table drop
	RefRootCauseDropTableV1 string = `DROP TABLE IF EXISTS ref_rootcause_v1;`
	// RefRootCauseTableV1 SQL statement for the rootcause references
	RefRootCauseTableV1 string = `create table ref_rootcause_v1 (
		id serial primary key,
		name varchar(100) not null,
		description varchar(500) not null,
		situation_id integer REFERENCES situation_definition_v1 (id),
		rule_id integer REFERENCES rules_v1 (id),
		CONSTRAINT unq_name_situationid UNIQUE(name,situation_id)
	);`

	// RefActionDropTableV1 SQL statement for table drop
	RefActionDropTableV1 string = `DROP TABLE IF EXISTS ref_action_v1;`
	// RefActionTableV1 SQL statement for the action references
	RefActionTableV1 string = `create table ref_action_v1 (
		id serial primary key,
		name varchar(100) not null,
		description varchar(500) not null,
		rootcause_id integer REFERENCES ref_rootcause_v1 (id),
		CONSTRAINT unq_name_rootcauseid UNIQUE(name,rootcause_id)
	);`

	// IssueResolutionDropTableV1 SQL statement for table drop
	IssueResolutionDropTableV1 string = `DROP TABLE IF EXISTS issue_resolution_v1;`
	// IssueResolutionTableV1 SQL statement for the issue resolution statistics
	IssueResolutionTableV1 string = `create table issue_resolution_v1 (
		feedback_date timestamptz not null,
		issue_id integer REFERENCES issues_v1 (id),
		rootcause_id integer REFERENCES ref_rootcause_v1 (id),
		action_id integer REFERENCES ref_action_v1 (id),
		CONSTRAINT unq_issue_rc_action UNIQUE(issue_id,rootcause_id,action_id)
	);`

	// IssueResolutionDraftDropTableV1 SQL statement for table drop
	IssueResolutionDraftDropTableV1 string = `DROP TABLE IF EXISTS issue_resolution_draft_v1;`
	// IssueResolutionDraftTableV1 SQL statement for the issue resolution statistics
	IssueResolutionDraftTableV1 string = `create table issue_resolution_draft_v1 (
		issue_id integer REFERENCES issues_v1 (id) primary key,
		concurrency_uuid varchar(100) unique not null,
		last_modified timestamptz not null,
		data JSONB not null
	);`

	// SituationRulesDropTableV1 SQL statement to drop table situation_rules_v1
	SituationRulesDropTableV1 string = `DROP TABLE IF EXISTS situation_rules_v1;`
	// SituationRulesTableV1 SQL statement to create table situation_rules_v1
	SituationRulesTableV1 string = `create table situation_rules_v1 (
		situation_id integer REFERENCES situation_definition_v1 (id),
		rule_id integer REFERENCES rules_v1 (id),
		execution_order integer,
		PRIMARY KEY(situation_id, rule_id)
	);`

	// ModelDropTableV1 SQL statement to drop table model_v1
	ModelDropTableV1 string = `DROP TABLE IF EXISTS model_v1;`
	// ModelTableV1 SQL statement to create table model_v1
	ModelTableV1 string = `create table IF NOT EXISTS model_v1 (
		id serial primary key,
		name varchar(100) unique not null,
		definition JSONB
	);`

	// EsIndicesDropTableV1 SQL statement to drop table elasticsearch_indices_v1
	EsIndicesDropTableV1 string = `DROP TABLE IF EXISTS elasticsearch_indices_v1;`
	// EsIndicesTableV1 SQL statement to create table elasticsearch_indices_v1
	EsIndicesTableV1 string = `create table IF NOT EXISTS elasticsearch_indices_v1 (
		id serial primary key,
		logical varchar(100) not null,
		technical varchar(100) not null,
		creation_date timestamptz not null
	);`

	// IssueFeedbackDropTableV3 SQL statement to drop table issue_detection_feedback_v3
	IssueFeedbackDropTableV3 string = `DROP TABLE IF EXISTS issue_detection_feedback_v3;`
	// IssueFeedbackTableV3 SQL statement to create table issue_detection_feedback_v3
	IssueFeedbackTableV3 string = `	create table issue_detection_feedback_v3 (
		id serial primary key,
		issue_id integer references issues_v1(id) not null,
		user_id integer references users_v1(id) not null,
		date timestamptz not null,
		rating integer,
		CONSTRAINT unq_issueid_userid UNIQUE(issue_id,user_id)
	);`

	// CalendarDropTableV3 SQL statement to drop table calendar_v1
	CalendarDropTableV3 string = `DROP TABLE IF EXISTS calendar_V1;`
	// CalendarTableV3 SQL statement to create table calendar_v1
	CalendarTableV3 string = `create table calendar_v1 (
		id serial primary key,
		name varchar(100) not null,
		description varchar(500) not null,
		period_data JSONB not null,
		enabled boolean not null,
		creation_date timestamptz not null,
		last_modified timestamptz not null
	);`

	// CalendarUnionDropTableV3 SQL statement to drop table calendar_union_v1
	CalendarUnionDropTableV3 string = `DROP TABLE IF EXISTS calendar_union_v1;`
	// CalendarUnionTableV3 SQL statement to create table calendar_union_v1
	CalendarUnionTableV3 string = `create table calendar_union_v1 (
		calendar_id integer REFERENCES calendar_v1 (id),
		sub_calendar_id integer REFERENCES calendar_v1 (id),
		priority integer,
		PRIMARY KEY(calendar_id, sub_calendar_id)
	);`

	//ConnectorExecutionsLogsDropTableV1 SQL statement to drop table connectors_executions_log_v1
	ConnectorExecutionsLogsDropTableV1 string = `DROP TABLE IF EXISTS connectors_executions_log_v1;`
	//ConnectorExecutionsLogsTableV1 SQL statement to create table connectors_executions_log_v1
	ConnectorExecutionsLogsTableV1 string = `create table connectors_executions_log_v1 (
		id serial primary key not null,
		connector_id varchar(100) not null,
		name varchar(100) not null,
		ts timestamptz not null,
		success boolean
	);`
)
