--
-- PostgreSQL database dump
--

-- Dumped from database version 16.2
-- Dumped by pg_dump version 16.2

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: pgcrypto; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS pgcrypto WITH SCHEMA public;


--
-- Name: EXTENSION pgcrypto; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION pgcrypto IS 'cryptographic functions';


--
-- Name: ensure_one_default_row(); Type: FUNCTION; Schema: public;d
--

CREATE FUNCTION public.ensure_one_default_row() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
begin
    -- If the row being deleted is the current default
    if OLD."default" = true then
        -- Update the 'default' row to be the new default
        update elasticsearch_config_v1
        set "default" = TRUE
        where name = 'default';
    end if;
    return OLD;
end;
$$;


ALTER FUNCTION

--
-- Name: prevent_delete_default_row(); Type: FUNCTION; Schema: public;d
--

CREATE FUNCTION public.prevent_delete_default_row() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
begin
    if OLD.name = 'default' then
        raise exception 'Cannot delete the row with name "default"';
    end if;
    return OLD;
end;
$$;


ALTER FUNCTION

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: calendar_union_v1; Type: TABLE; Schema: public;d
--

CREATE TABLE public.calendar_union_v1 (
    calendar_id integer NOT NULL,
    sub_calendar_id integer NOT NULL,
    priority integer
);




--
-- Name: calendar_v1; Type: TABLE; Schema: public;d
--

CREATE TABLE public.calendar_v1 (
    id integer NOT NULL,
    name character varying(100) NOT NULL,
    description character varying(500) NOT NULL,
    period_data jsonb NOT NULL,
    enabled boolean NOT NULL,
    creation_date timestamp with time zone NOT NULL,
    last_modified timestamp with time zone NOT NULL,
    timezone character varying(100) DEFAULT 'Europe/Paris'::character varying NOT NULL
);




--
-- Name: calendar_v1_id_seq; Type: SEQUENCE; Schema: public;d
--

CREATE SEQUENCE public.calendar_v1_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;




--
-- Name: calendar_v1_id_seq; Type: SEQUENCE OWNED BY; Schema: public;d
--

ALTER SEQUENCE public.calendar_v1_id_seq OWNED BY public.calendar_v1.id;


--
-- Name: connectors_config_v1; Type: TABLE; Schema: public;d
--

CREATE TABLE public.connectors_config_v1 (
    id integer NOT NULL,
    connector_id character varying(100) NOT NULL,
    name character varying(100) NOT NULL,
    current text NOT NULL,
    previous text,
    last_modified timestamp with time zone NOT NULL
);




--
-- Name: connectors_config_v1_id_seq; Type: SEQUENCE; Schema: public;d
--

CREATE SEQUENCE public.connectors_config_v1_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;




--
-- Name: connectors_config_v1_id_seq; Type: SEQUENCE OWNED BY; Schema: public;d
--

ALTER SEQUENCE public.connectors_config_v1_id_seq OWNED BY public.connectors_config_v1.id;


--
-- Name: connectors_executions_log_v1; Type: TABLE; Schema: public;d
--

CREATE TABLE public.connectors_executions_log_v1 (
    id integer NOT NULL,
    connector_id character varying(100) NOT NULL,
    name character varying(100) NOT NULL,
    ts timestamp with time zone NOT NULL,
    success boolean
);




--
-- Name: connectors_executions_log_v1_id_seq; Type: SEQUENCE; Schema: public;d
--

CREATE SEQUENCE public.connectors_executions_log_v1_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;




--
-- Name: connectors_executions_log_v1_id_seq; Type: SEQUENCE OWNED BY; Schema: public;d
--

ALTER SEQUENCE public.connectors_executions_log_v1_id_seq OWNED BY public.connectors_executions_log_v1.id;


--
-- Name: elasticsearch_config_v1; Type: TABLE; Schema: public;d
--

CREATE TABLE public.elasticsearch_config_v1 (
    id integer NOT NULL,
    name character varying(255) NOT NULL,
    urls text,
    "default" boolean NOT NULL,
    export_activated boolean DEFAULT true NOT NULL
);




--
-- Name: elasticsearch_config_v1_id_seq; Type: SEQUENCE; Schema: public;d
--

CREATE SEQUENCE public.elasticsearch_config_v1_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;




--
-- Name: elasticsearch_config_v1_id_seq; Type: SEQUENCE OWNED BY; Schema: public;d
--

ALTER SEQUENCE public.elasticsearch_config_v1_id_seq OWNED BY public.elasticsearch_config_v1.id;


--
-- Name: elasticsearch_indices_v1; Type: TABLE; Schema: public;d
--

CREATE TABLE public.elasticsearch_indices_v1 (
    id integer NOT NULL,
    logical character varying(100) NOT NULL,
    technical character varying(100) NOT NULL,
    creation_date timestamp with time zone NOT NULL
);




--
-- Name: elasticsearch_indices_v1_id_seq; Type: SEQUENCE; Schema: public;d
--

CREATE SEQUENCE public.elasticsearch_indices_v1_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;




--
-- Name: elasticsearch_indices_v1_id_seq; Type: SEQUENCE OWNED BY; Schema: public;d
--

ALTER SEQUENCE public.elasticsearch_indices_v1_id_seq OWNED BY public.elasticsearch_indices_v1.id;


--
-- Name: external_generic_config_v1; Type: TABLE; Schema: public;d
--

CREATE TABLE public.external_generic_config_v1 (
    name character varying(255) NOT NULL,
    id integer NOT NULL
);




--
-- Name: external_generic_config_v1_id_seq; Type: SEQUENCE; Schema: public;d
--

CREATE SEQUENCE public.external_generic_config_v1_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;




--
-- Name: external_generic_config_v1_id_seq; Type: SEQUENCE OWNED BY; Schema: public;d
--

ALTER SEQUENCE public.external_generic_config_v1_id_seq OWNED BY public.external_generic_config_v1.id;


--
-- Name: external_generic_config_versions_v1; Type: TABLE; Schema: public;d
--

CREATE TABLE public.external_generic_config_versions_v1 (
    config_id integer NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    data jsonb NOT NULL,
    current_version boolean DEFAULT false NOT NULL
);




--
-- Name: fact_definition_v1; Type: TABLE; Schema: public;d
--

CREATE TABLE public.fact_definition_v1 (
    id integer NOT NULL,
    name character varying(100) NOT NULL,
    definition json,
    last_modified timestamp with time zone NOT NULL
);




--
-- Name: fact_definition_v1_id_seq; Type: SEQUENCE; Schema: public;d
--

CREATE SEQUENCE public.fact_definition_v1_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;




--
-- Name: fact_definition_v1_id_seq; Type: SEQUENCE OWNED BY; Schema: public;d
--

ALTER SEQUENCE public.fact_definition_v1_id_seq OWNED BY public.fact_definition_v1.id;


--
-- Name: fact_history_v1; Type: TABLE; Schema: public;d
--

CREATE TABLE public.fact_history_v1 (
    id integer NOT NULL,
    ts timestamp with time zone NOT NULL,
    situation_id integer NOT NULL,
    situation_instance_id integer NOT NULL,
    result jsonb,
    success boolean
);




--
-- Name: fact_history_v5; Type: TABLE; Schema: public;d
--

CREATE TABLE public.fact_history_v5 (
    id integer NOT NULL,
    fact_id integer,
    situation_id integer,
    situation_instance_id integer,
    ts timestamp with time zone NOT NULL,
    result jsonb
);




--
-- Name: fact_history_v5_id_seq; Type: SEQUENCE; Schema: public;d
--

CREATE SEQUENCE public.fact_history_v5_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;




--
-- Name: fact_history_v5_id_seq; Type: SEQUENCE OWNED BY; Schema: public;d
--

ALTER SEQUENCE public.fact_history_v5_id_seq OWNED BY public.fact_history_v5.id;


--
-- Name: goose_db_version; Type: TABLE; Schema: public;d
--

CREATE TABLE public.goose_db_version (
    id integer NOT NULL,
    version_id bigint NOT NULL,
    is_applied boolean NOT NULL,
    tstamp timestamp without time zone DEFAULT now()
);




--
-- Name: goose_db_version_id_seq; Type: SEQUENCE; Schema: public;d
--

CREATE SEQUENCE public.goose_db_version_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;




--
-- Name: goose_db_version_id_seq; Type: SEQUENCE OWNED BY; Schema: public;d
--

ALTER SEQUENCE public.goose_db_version_id_seq OWNED BY public.goose_db_version.id;


--
-- Name: issue_detection_feedback_v3; Type: TABLE; Schema: public;d
--

CREATE TABLE public.issue_detection_feedback_v3 (
    id integer NOT NULL,
    issue_id integer NOT NULL,
    date timestamp with time zone NOT NULL,
    rating integer,
    user_id uuid
);



--
-- Name: issue_detection_feedback_v3_id_seq; Type: SEQUENCE; Schema: public;d
--

CREATE SEQUENCE public.issue_detection_feedback_v3_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;




--
-- Name: issue_detection_feedback_v3_id_seq; Type: SEQUENCE OWNED BY; Schema: public;d
--

ALTER SEQUENCE public.issue_detection_feedback_v3_id_seq OWNED BY public.issue_detection_feedback_v3.id;


--
-- Name: issue_resolution_draft_v1; Type: TABLE; Schema: public;d
--

CREATE TABLE public.issue_resolution_draft_v1 (
    issue_id integer NOT NULL,
    concurrency_uuid character varying(100) NOT NULL,
    last_modified timestamp with time zone NOT NULL,
    data jsonb NOT NULL
);




--
-- Name: issue_resolution_v1; Type: TABLE; Schema: public;d
--

CREATE TABLE public.issue_resolution_v1 (
    feedback_date timestamp with time zone NOT NULL,
    issue_id integer,
    rootcause_id integer,
    action_id integer
);




--
-- Name: issues_v1; Type: TABLE; Schema: public;d
--

CREATE TABLE public.issues_v1 (
    id integer NOT NULL,
    key character varying(100) NOT NULL,
    name character varying(100) NOT NULL,
    level character varying(100) NOT NULL,
    situation_id integer,
    situation_instance_id integer,
    situation_date timestamp with time zone NOT NULL,
    expiration_date timestamp with time zone NOT NULL,
    rule_data jsonb NOT NULL,
    state character varying(100) NOT NULL,
    last_modified timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    detection_rating_avg real,
    assigned_at timestamp with time zone,
    assigned_to character varying(100),
    closed_at timestamp with time zone,
    closed_by character varying(100),
    comment text,
    situation_history_id integer
);




--
-- Name: issues_v1_id_seq; Type: SEQUENCE; Schema: public;d
--

CREATE SEQUENCE public.issues_v1_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;




--
-- Name: issues_v1_id_seq; Type: SEQUENCE OWNED BY; Schema: public;d
--

ALTER SEQUENCE public.issues_v1_id_seq OWNED BY public.issues_v1.id;


--
-- Name: job_schedules_v1; Type: TABLE; Schema: public;d
--

CREATE TABLE public.job_schedules_v1 (
    id integer NOT NULL,
    name character varying(100) NOT NULL,
    cronexpr character varying(100) NOT NULL,
    job_type character varying(100) NOT NULL,
    job_data json NOT NULL,
    last_modified timestamp with time zone NOT NULL,
    enabled boolean
);




--
-- Name: job_schedules_v1_id_seq; Type: SEQUENCE; Schema: public;d
--

CREATE SEQUENCE public.job_schedules_v1_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;




--
-- Name: job_schedules_v1_id_seq; Type: SEQUENCE OWNED BY; Schema: public;d
--

ALTER SEQUENCE public.job_schedules_v1_id_seq OWNED BY public.job_schedules_v1.id;


--
-- Name: model_v1; Type: TABLE; Schema: public;d
--

CREATE TABLE public.model_v1 (
    id integer NOT NULL,
    name character varying(100) NOT NULL,
    definition jsonb
);




--
-- Name: model_v1_id_seq; Type: SEQUENCE; Schema: public;d
--

CREATE SEQUENCE public.model_v1_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;




--
-- Name: model_v1_id_seq; Type: SEQUENCE OWNED BY; Schema: public;d
--

ALTER SEQUENCE public.model_v1_id_seq OWNED BY public.model_v1.id;


--
-- Name: notifications_history_v1; Type: TABLE; Schema: public;d
--

CREATE TABLE public.notifications_history_v1 (
    id integer NOT NULL,
    data json,
    created_at timestamp with time zone NOT NULL,
    isread boolean DEFAULT false,
    type character varying(100) DEFAULT NULL::character varying,
    user_login character varying(100) DEFAULT ''::character varying NOT NULL
);




--
-- Name: notifications_history_v1_id_seq; Type: SEQUENCE; Schema: public;d
--

CREATE SEQUENCE public.notifications_history_v1_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;




--
-- Name: notifications_history_v1_id_seq; Type: SEQUENCE OWNED BY; Schema: public;d
--

ALTER SEQUENCE public.notifications_history_v1_id_seq OWNED BY public.notifications_history_v1.id;


--
-- Name: permissions_v4; Type: TABLE; Schema: public;d
--

CREATE TABLE public.permissions_v4 (
    id uuid NOT NULL,
    resource_type character varying(100) NOT NULL,
    resource_id character varying(100) NOT NULL,
    action character varying(100) NOT NULL
);




--
-- Name: ref_action_v1; Type: TABLE; Schema: public;d
--

CREATE TABLE public.ref_action_v1 (
    id integer NOT NULL,
    name character varying(100) NOT NULL,
    description character varying(500) NOT NULL,
    rootcause_id integer
);




--
-- Name: ref_action_v1_id_seq; Type: SEQUENCE; Schema: public;d
--

CREATE SEQUENCE public.ref_action_v1_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;




--
-- Name: ref_action_v1_id_seq; Type: SEQUENCE OWNED BY; Schema: public;d
--

ALTER SEQUENCE public.ref_action_v1_id_seq OWNED BY public.ref_action_v1.id;


--
-- Name: ref_rootcause_v1; Type: TABLE; Schema: public;d
--

CREATE TABLE public.ref_rootcause_v1 (
    id integer NOT NULL,
    name character varying(100) NOT NULL,
    description character varying(500) NOT NULL,
    situation_id integer,
    rule_id integer
);




--
-- Name: ref_rootcause_v1_id_seq; Type: SEQUENCE; Schema: public;d
--

CREATE SEQUENCE public.ref_rootcause_v1_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;




--
-- Name: ref_rootcause_v1_id_seq; Type: SEQUENCE OWNED BY; Schema: public;d
--

ALTER SEQUENCE public.ref_rootcause_v1_id_seq OWNED BY public.ref_rootcause_v1.id;


--
-- Name: roles_permissions_v4; Type: TABLE; Schema: public;d
--

CREATE TABLE public.roles_permissions_v4 (
    role_id uuid,
    permission_id uuid
);




--
-- Name: roles_v4; Type: TABLE; Schema: public;d
--

CREATE TABLE public.roles_v4 (
    id uuid NOT NULL,
    name character varying(100) NOT NULL
);




--
-- Name: rule_versions_v1; Type: TABLE; Schema: public;d
--

CREATE TABLE public.rule_versions_v1 (
    rule_id integer NOT NULL,
    version_number integer NOT NULL,
    data json NOT NULL,
    creation_datetime timestamp with time zone NOT NULL
);




--
-- Name: rules_v1; Type: TABLE; Schema: public;d
--

CREATE TABLE public.rules_v1 (
    id integer NOT NULL,
    name character varying(100) NOT NULL,
    enabled boolean NOT NULL,
    calendar_id integer,
    last_modified timestamp with time zone NOT NULL
);




--
-- Name: rules_v1_id_seq; Type: SEQUENCE; Schema: public;d
--

CREATE SEQUENCE public.rules_v1_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;




--
-- Name: rules_v1_id_seq; Type: SEQUENCE OWNED BY; Schema: public;d
--

ALTER SEQUENCE public.rules_v1_id_seq OWNED BY public.rules_v1.id;


--
-- Name: situation_definition_v1; Type: TABLE; Schema: public;d
--

CREATE TABLE public.situation_definition_v1 (
    id integer NOT NULL,
    name character varying(100) NOT NULL,
    definition json,
    is_template boolean,
    is_object boolean,
    calendar_id integer,
    last_modified timestamp with time zone NOT NULL
);




--
-- Name: situation_definition_v1_id_seq; Type: SEQUENCE; Schema: public;d
--

CREATE SEQUENCE public.situation_definition_v1_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;




--
-- Name: situation_definition_v1_id_seq; Type: SEQUENCE OWNED BY; Schema: public;d
--

ALTER SEQUENCE public.situation_definition_v1_id_seq OWNED BY public.situation_definition_v1.id;


--
-- Name: situation_fact_history_v5; Type: TABLE; Schema: public;d
--

CREATE TABLE public.situation_fact_history_v5 (
    situation_history_id integer NOT NULL,
    fact_history_id integer NOT NULL,
    fact_id integer
);




--
-- Name: situation_facts_v1; Type: TABLE; Schema: public;d
--

CREATE TABLE public.situation_facts_v1 (
    situation_id integer NOT NULL,
    fact_id integer NOT NULL
);




--
-- Name: situation_history_v1; Type: TABLE; Schema: public;d
--

CREATE TABLE public.situation_history_v1 (
    id integer NOT NULL,
    ts timestamp with time zone NOT NULL,
    situation_instance_id integer NOT NULL,
    facts_ids json,
    parameters json,
    expression_facts jsonb,
    metadatas json,
    evaluated boolean
);




--
-- Name: situation_history_v5; Type: TABLE; Schema: public;d
--

CREATE TABLE public.situation_history_v5 (
    id integer NOT NULL,
    situation_id integer,
    situation_instance_id integer,
    ts timestamp with time zone NOT NULL,
    parameters json,
    expression_facts jsonb,
    metadatas json
);




--
-- Name: situation_history_v5_id_seq; Type: SEQUENCE; Schema: public;d
--

CREATE SEQUENCE public.situation_history_v5_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;




--
-- Name: situation_history_v5_id_seq; Type: SEQUENCE OWNED BY; Schema: public;d
--

ALTER SEQUENCE public.situation_history_v5_id_seq OWNED BY public.situation_history_v5.id;


--
-- Name: situation_rules_v1; Type: TABLE; Schema: public;d
--

CREATE TABLE public.situation_rules_v1 (
    situation_id integer NOT NULL,
    rule_id integer NOT NULL,
    execution_order integer
);




--
-- Name: situation_template_instances_v1; Type: TABLE; Schema: public;d
--

CREATE TABLE public.situation_template_instances_v1 (
    id integer NOT NULL,
    name character varying(100) NOT NULL,
    situation_id integer,
    parameters json,
    calendar_id integer,
    last_modified timestamp with time zone NOT NULL,
    enable_depends_on boolean DEFAULT false,
    depends_on_parameters json DEFAULT '{}'::json
);




--
-- Name: situation_template_instances_v1_id_seq; Type: SEQUENCE; Schema: public;d
--

CREATE SEQUENCE public.situation_template_instances_v1_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;




--
-- Name: situation_template_instances_v1_id_seq; Type: SEQUENCE OWNED BY; Schema: public;d
--

ALTER SEQUENCE public.situation_template_instances_v1_id_seq OWNED BY public.situation_template_instances_v1.id;


--
-- Name: user_groups_v1; Type: TABLE; Schema: public;d
--

CREATE TABLE public.user_groups_v1 (
    id integer NOT NULL,
    name character varying(100) NOT NULL
);




--
-- Name: user_groups_v1_id_seq; Type: SEQUENCE; Schema: public;d
--

CREATE SEQUENCE public.user_groups_v1_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;




--
-- Name: user_groups_v1_id_seq; Type: SEQUENCE OWNED BY; Schema: public;d
--

ALTER SEQUENCE public.user_groups_v1_id_seq OWNED BY public.user_groups_v1.id;


--
-- Name: user_memberships_v1; Type: TABLE; Schema: public;d
--

CREATE TABLE public.user_memberships_v1 (
    user_id integer NOT NULL,
    group_id integer NOT NULL,
    role integer
);




--
-- Name: users_roles_v4; Type: TABLE; Schema: public;d
--

CREATE TABLE public.users_roles_v4 (
    user_id uuid,
    role_id uuid
);




--
-- Name: users_v1; Type: TABLE; Schema: public;d
--

CREATE TABLE public.users_v1 (
    id integer NOT NULL,
    login character varying(100) NOT NULL,
    password character varying(100) NOT NULL,
    role integer NOT NULL,
    created timestamp with time zone NOT NULL,
    last_name character varying(100) NOT NULL,
    first_name character varying(100) NOT NULL,
    email character varying(100) NOT NULL,
    phone character varying(100)
);




--
-- Name: users_v1_id_seq; Type: SEQUENCE; Schema: public;d
--

CREATE SEQUENCE public.users_v1_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;




--
-- Name: users_v1_id_seq; Type: SEQUENCE OWNED BY; Schema: public;d
--

ALTER SEQUENCE public.users_v1_id_seq OWNED BY public.users_v1.id;


--
-- Name: users_v4; Type: TABLE; Schema: public;d
--

CREATE TABLE public.users_v4 (
    id uuid NOT NULL,
    login character varying(100) NOT NULL,
    password character varying(100) NOT NULL,
    created timestamp with time zone NOT NULL,
    last_name character varying(100) NOT NULL,
    first_name character varying(100) NOT NULL,
    email character varying(100) NOT NULL,
    phone character varying(100)
);




--
-- Name: variables_config_v1; Type: TABLE; Schema: public;d
--

CREATE TABLE public.variables_config_v1 (
    id integer NOT NULL,
    key character varying(100) NOT NULL,
    value text NOT NULL
);




--
-- Name: variables_config_v1_id_seq; Type: SEQUENCE; Schema: public;d
--

CREATE SEQUENCE public.variables_config_v1_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;




--
-- Name: variables_config_v1_id_seq; Type: SEQUENCE OWNED BY; Schema: public;d
--

ALTER SEQUENCE public.variables_config_v1_id_seq OWNED BY public.variables_config_v1.id;


--
-- Name: calendar_v1 id; Type: DEFAULT; Schema: public;d
--

ALTER TABLE ONLY public.calendar_v1 ALTER COLUMN id SET DEFAULT nextval('public.calendar_v1_id_seq'::regclass);


--
-- Name: connectors_config_v1 id; Type: DEFAULT; Schema: public;d
--

ALTER TABLE ONLY public.connectors_config_v1 ALTER COLUMN id SET DEFAULT nextval('public.connectors_config_v1_id_seq'::regclass);


--
-- Name: connectors_executions_log_v1 id; Type: DEFAULT; Schema: public;d
--

ALTER TABLE ONLY public.connectors_executions_log_v1 ALTER COLUMN id SET DEFAULT nextval('public.connectors_executions_log_v1_id_seq'::regclass);


--
-- Name: elasticsearch_config_v1 id; Type: DEFAULT; Schema: public;d
--

ALTER TABLE ONLY public.elasticsearch_config_v1 ALTER COLUMN id SET DEFAULT nextval('public.elasticsearch_config_v1_id_seq'::regclass);


--
-- Name: elasticsearch_indices_v1 id; Type: DEFAULT; Schema: public;d
--

ALTER TABLE ONLY public.elasticsearch_indices_v1 ALTER COLUMN id SET DEFAULT nextval('public.elasticsearch_indices_v1_id_seq'::regclass);


--
-- Name: external_generic_config_v1 id; Type: DEFAULT; Schema: public;d
--

ALTER TABLE ONLY public.external_generic_config_v1 ALTER COLUMN id SET DEFAULT nextval('public.external_generic_config_v1_id_seq'::regclass);


--
-- Name: fact_definition_v1 id; Type: DEFAULT; Schema: public;d
--

ALTER TABLE ONLY public.fact_definition_v1 ALTER COLUMN id SET DEFAULT nextval('public.fact_definition_v1_id_seq'::regclass);


--
-- Name: fact_history_v5 id; Type: DEFAULT; Schema: public;d
--

ALTER TABLE ONLY public.fact_history_v5 ALTER COLUMN id SET DEFAULT nextval('public.fact_history_v5_id_seq'::regclass);


--
-- Name: goose_db_version id; Type: DEFAULT; Schema: public;d
--

ALTER TABLE ONLY public.goose_db_version ALTER COLUMN id SET DEFAULT nextval('public.goose_db_version_id_seq'::regclass);


--
-- Name: issue_detection_feedback_v3 id; Type: DEFAULT; Schema: public;d
--

ALTER TABLE ONLY public.issue_detection_feedback_v3 ALTER COLUMN id SET DEFAULT nextval('public.issue_detection_feedback_v3_id_seq'::regclass);


--
-- Name: issues_v1 id; Type: DEFAULT; Schema: public;d
--

ALTER TABLE ONLY public.issues_v1 ALTER COLUMN id SET DEFAULT nextval('public.issues_v1_id_seq'::regclass);


--
-- Name: job_schedules_v1 id; Type: DEFAULT; Schema: public;d
--

ALTER TABLE ONLY public.job_schedules_v1 ALTER COLUMN id SET DEFAULT nextval('public.job_schedules_v1_id_seq'::regclass);


--
-- Name: model_v1 id; Type: DEFAULT; Schema: public;d
--

ALTER TABLE ONLY public.model_v1 ALTER COLUMN id SET DEFAULT nextval('public.model_v1_id_seq'::regclass);


--
-- Name: notifications_history_v1 id; Type: DEFAULT; Schema: public;d
--

ALTER TABLE ONLY public.notifications_history_v1 ALTER COLUMN id SET DEFAULT nextval('public.notifications_history_v1_id_seq'::regclass);


--
-- Name: ref_action_v1 id; Type: DEFAULT; Schema: public;d
--

ALTER TABLE ONLY public.ref_action_v1 ALTER COLUMN id SET DEFAULT nextval('public.ref_action_v1_id_seq'::regclass);


--
-- Name: ref_rootcause_v1 id; Type: DEFAULT; Schema: public;d
--

ALTER TABLE ONLY public.ref_rootcause_v1 ALTER COLUMN id SET DEFAULT nextval('public.ref_rootcause_v1_id_seq'::regclass);


--
-- Name: rules_v1 id; Type: DEFAULT; Schema: public;d
--

ALTER TABLE ONLY public.rules_v1 ALTER COLUMN id SET DEFAULT nextval('public.rules_v1_id_seq'::regclass);


--
-- Name: situation_definition_v1 id; Type: DEFAULT; Schema: public;d
--

ALTER TABLE ONLY public.situation_definition_v1 ALTER COLUMN id SET DEFAULT nextval('public.situation_definition_v1_id_seq'::regclass);


--
-- Name: situation_history_v5 id; Type: DEFAULT; Schema: public;d
--

ALTER TABLE ONLY public.situation_history_v5 ALTER COLUMN id SET DEFAULT nextval('public.situation_history_v5_id_seq'::regclass);


--
-- Name: situation_template_instances_v1 id; Type: DEFAULT; Schema: public;d
--

ALTER TABLE ONLY public.situation_template_instances_v1 ALTER COLUMN id SET DEFAULT nextval('public.situation_template_instances_v1_id_seq'::regclass);


--
-- Name: user_groups_v1 id; Type: DEFAULT; Schema: public;d
--

ALTER TABLE ONLY public.user_groups_v1 ALTER COLUMN id SET DEFAULT nextval('public.user_groups_v1_id_seq'::regclass);


--
-- Name: users_v1 id; Type: DEFAULT; Schema: public;d
--

ALTER TABLE ONLY public.users_v1 ALTER COLUMN id SET DEFAULT nextval('public.users_v1_id_seq'::regclass);


--
-- Name: variables_config_v1 id; Type: DEFAULT; Schema: public;d
--

ALTER TABLE ONLY public.variables_config_v1 ALTER COLUMN id SET DEFAULT nextval('public.variables_config_v1_id_seq'::regclass);


--
-- Name: calendar_union_v1 calendar_union_v1_pkey; Type: CONSTRAINT; Schema: public;d
--

ALTER TABLE ONLY public.calendar_union_v1
    ADD CONSTRAINT calendar_union_v1_pkey PRIMARY KEY (calendar_id, sub_calendar_id);


--
-- Name: calendar_v1 calendar_v1_pkey; Type: CONSTRAINT; Schema: public;d
--

ALTER TABLE ONLY public.calendar_v1
    ADD CONSTRAINT calendar_v1_pkey PRIMARY KEY (id);


--
-- Name: connectors_config_v1 connectors_config_v1_connector_id_key; Type: CONSTRAINT; Schema: public;d
--

ALTER TABLE ONLY public.connectors_config_v1
    ADD CONSTRAINT connectors_config_v1_connector_id_key UNIQUE (connector_id);


--
-- Name: connectors_config_v1 connectors_config_v1_pkey; Type: CONSTRAINT; Schema: public;d
--

ALTER TABLE ONLY public.connectors_config_v1
    ADD CONSTRAINT connectors_config_v1_pkey PRIMARY KEY (id);


--
-- Name: connectors_executions_log_v1 connectors_executions_log_v1_pkey; Type: CONSTRAINT; Schema: public;d
--

ALTER TABLE ONLY public.connectors_executions_log_v1
    ADD CONSTRAINT connectors_executions_log_v1_pkey PRIMARY KEY (id);


--
-- Name: elasticsearch_config_v1 elasticsearch_config_v1_pkey; Type: CONSTRAINT; Schema: public;d
--

ALTER TABLE ONLY public.elasticsearch_config_v1
    ADD CONSTRAINT elasticsearch_config_v1_pkey PRIMARY KEY (id);


--
-- Name: elasticsearch_indices_v1 elasticsearch_indices_v1_pkey; Type: CONSTRAINT; Schema: public;d
--

ALTER TABLE ONLY public.elasticsearch_indices_v1
    ADD CONSTRAINT elasticsearch_indices_v1_pkey PRIMARY KEY (id);


--
-- Name: external_generic_config_v1 external_generic_config_v1_pkey; Type: CONSTRAINT; Schema: public;d
--

ALTER TABLE ONLY public.external_generic_config_v1
    ADD CONSTRAINT external_generic_config_v1_pkey PRIMARY KEY (id);


--
-- Name: external_generic_config_versions_v1 external_generic_config_versions_v1_pkey; Type: CONSTRAINT; Schema: public;d
--

ALTER TABLE ONLY public.external_generic_config_versions_v1
    ADD CONSTRAINT external_generic_config_versions_v1_pkey PRIMARY KEY (config_id, created_at);


--
-- Name: fact_definition_v1 fact_definition_v1_name_key; Type: CONSTRAINT; Schema: public;d
--

ALTER TABLE ONLY public.fact_definition_v1
    ADD CONSTRAINT fact_definition_v1_name_key UNIQUE (name);


--
-- Name: fact_definition_v1 fact_definition_v1_pkey; Type: CONSTRAINT; Schema: public;d
--

ALTER TABLE ONLY public.fact_definition_v1
    ADD CONSTRAINT fact_definition_v1_pkey PRIMARY KEY (id);


--
-- Name: fact_history_v1 fact_history_v1_pkey; Type: CONSTRAINT; Schema: public;d
--

ALTER TABLE ONLY public.fact_history_v1
    ADD CONSTRAINT fact_history_v1_pkey PRIMARY KEY (id, ts, situation_id, situation_instance_id);


--
-- Name: fact_history_v5 fact_history_v5_pkey; Type: CONSTRAINT; Schema: public;d
--

ALTER TABLE ONLY public.fact_history_v5
    ADD CONSTRAINT fact_history_v5_pkey PRIMARY KEY (id);


--
-- Name: goose_db_version goose_db_version_pkey; Type: CONSTRAINT; Schema: public;d
--

ALTER TABLE ONLY public.goose_db_version
    ADD CONSTRAINT goose_db_version_pkey PRIMARY KEY (id);


--
-- Name: issue_detection_feedback_v3 issue_detection_feedback_v3_pkey; Type: CONSTRAINT; Schema: public;d
--

ALTER TABLE ONLY public.issue_detection_feedback_v3
    ADD CONSTRAINT issue_detection_feedback_v3_pkey PRIMARY KEY (id);


--
-- Name: issue_resolution_draft_v1 issue_resolution_draft_v1_concurrency_uuid_key; Type: CONSTRAINT; Schema: public;d
--

ALTER TABLE ONLY public.issue_resolution_draft_v1
    ADD CONSTRAINT issue_resolution_draft_v1_concurrency_uuid_key UNIQUE (concurrency_uuid);


--
-- Name: issue_resolution_draft_v1 issue_resolution_draft_v1_pkey; Type: CONSTRAINT; Schema: public;d
--

ALTER TABLE ONLY public.issue_resolution_draft_v1
    ADD CONSTRAINT issue_resolution_draft_v1_pkey PRIMARY KEY (issue_id);


--
-- Name: issues_v1 issues_v1_pkey; Type: CONSTRAINT; Schema: public;d
--

ALTER TABLE ONLY public.issues_v1
    ADD CONSTRAINT issues_v1_pkey PRIMARY KEY (id);


--
-- Name: job_schedules_v1 job_schedules_v1_pkey; Type: CONSTRAINT; Schema: public;d
--

ALTER TABLE ONLY public.job_schedules_v1
    ADD CONSTRAINT job_schedules_v1_pkey PRIMARY KEY (id);


--
-- Name: model_v1 model_v1_name_key; Type: CONSTRAINT; Schema: public;d
--

ALTER TABLE ONLY public.model_v1
    ADD CONSTRAINT model_v1_name_key UNIQUE (name);


--
-- Name: model_v1 model_v1_pkey; Type: CONSTRAINT; Schema: public;d
--

ALTER TABLE ONLY public.model_v1
    ADD CONSTRAINT model_v1_pkey PRIMARY KEY (id);


--
-- Name: notifications_history_v1 notifications_history_v1_pkey; Type: CONSTRAINT; Schema: public;d
--

ALTER TABLE ONLY public.notifications_history_v1
    ADD CONSTRAINT notifications_history_v1_pkey PRIMARY KEY (id);


--
-- Name: permissions_v4 permissions_v4_pkey; Type: CONSTRAINT; Schema: public;d
--

ALTER TABLE ONLY public.permissions_v4
    ADD CONSTRAINT permissions_v4_pkey PRIMARY KEY (id);


--
-- Name: ref_action_v1 ref_action_v1_pkey; Type: CONSTRAINT; Schema: public;d
--

ALTER TABLE ONLY public.ref_action_v1
    ADD CONSTRAINT ref_action_v1_pkey PRIMARY KEY (id);


--
-- Name: ref_rootcause_v1 ref_rootcause_v1_pkey; Type: CONSTRAINT; Schema: public;d
--

ALTER TABLE ONLY public.ref_rootcause_v1
    ADD CONSTRAINT ref_rootcause_v1_pkey PRIMARY KEY (id);


--
-- Name: roles_v4 roles_v4_name_key; Type: CONSTRAINT; Schema: public;d
--

ALTER TABLE ONLY public.roles_v4
    ADD CONSTRAINT roles_v4_name_key UNIQUE (name);


--
-- Name: roles_v4 roles_v4_pkey; Type: CONSTRAINT; Schema: public;d
--

ALTER TABLE ONLY public.roles_v4
    ADD CONSTRAINT roles_v4_pkey PRIMARY KEY (id);


--
-- Name: rule_versions_v1 rule_versions_v1_pkey; Type: CONSTRAINT; Schema: public;d
--

ALTER TABLE ONLY public.rule_versions_v1
    ADD CONSTRAINT rule_versions_v1_pkey PRIMARY KEY (rule_id, version_number);


--
-- Name: rules_v1 rules_v1_name_key; Type: CONSTRAINT; Schema: public;d
--

ALTER TABLE ONLY public.rules_v1
    ADD CONSTRAINT rules_v1_name_key UNIQUE (name);


--
-- Name: rules_v1 rules_v1_pkey; Type: CONSTRAINT; Schema: public;d
--

ALTER TABLE ONLY public.rules_v1
    ADD CONSTRAINT rules_v1_pkey PRIMARY KEY (id);


--
-- Name: situation_definition_v1 situation_definition_v1_name_key; Type: CONSTRAINT; Schema: public;d
--

ALTER TABLE ONLY public.situation_definition_v1
    ADD CONSTRAINT situation_definition_v1_name_key UNIQUE (name);


--
-- Name: situation_definition_v1 situation_definition_v1_pkey; Type: CONSTRAINT; Schema: public;d
--

ALTER TABLE ONLY public.situation_definition_v1
    ADD CONSTRAINT situation_definition_v1_pkey PRIMARY KEY (id);


--
-- Name: situation_fact_history_v5 situation_fact_history_v5_pkey; Type: CONSTRAINT; Schema: public;d
--

ALTER TABLE ONLY public.situation_fact_history_v5
    ADD CONSTRAINT situation_fact_history_v5_pkey PRIMARY KEY (situation_history_id, fact_history_id);


--
-- Name: situation_facts_v1 situation_facts_v1_pkey; Type: CONSTRAINT; Schema: public;d
--

ALTER TABLE ONLY public.situation_facts_v1
    ADD CONSTRAINT situation_facts_v1_pkey PRIMARY KEY (situation_id, fact_id);


--
-- Name: situation_history_v1 situation_history_v1_pkey; Type: CONSTRAINT; Schema: public;d
--

ALTER TABLE ONLY public.situation_history_v1
    ADD CONSTRAINT situation_history_v1_pkey PRIMARY KEY (id, ts, situation_instance_id);


--
-- Name: situation_history_v5 situation_history_v5_pkey; Type: CONSTRAINT; Schema: public;d
--

ALTER TABLE ONLY public.situation_history_v5
    ADD CONSTRAINT situation_history_v5_pkey PRIMARY KEY (id);


--
-- Name: situation_rules_v1 situation_rules_v1_pkey; Type: CONSTRAINT; Schema: public;d
--

ALTER TABLE ONLY public.situation_rules_v1
    ADD CONSTRAINT situation_rules_v1_pkey PRIMARY KEY (situation_id, rule_id);


--
-- Name: situation_template_instances_v1 situation_template_instances_v1_pkey; Type: CONSTRAINT; Schema: public;d
--

ALTER TABLE ONLY public.situation_template_instances_v1
    ADD CONSTRAINT situation_template_instances_v1_pkey PRIMARY KEY (id);


--
-- Name: issue_resolution_v1 unq_issue_rc_action; Type: CONSTRAINT; Schema: public;d
--

ALTER TABLE ONLY public.issue_resolution_v1
    ADD CONSTRAINT unq_issue_rc_action UNIQUE (issue_id, rootcause_id, action_id);


--
-- Name: issue_detection_feedback_v3 unq_issueid_userid; Type: CONSTRAINT; Schema: public;d
--

ALTER TABLE ONLY public.issue_detection_feedback_v3
    ADD CONSTRAINT unq_issueid_userid UNIQUE (issue_id, user_id);


--
-- Name: ref_action_v1 unq_name_rootcauseid; Type: CONSTRAINT; Schema: public;d
--

ALTER TABLE ONLY public.ref_action_v1
    ADD CONSTRAINT unq_name_rootcauseid UNIQUE (name, rootcause_id);


--
-- Name: ref_rootcause_v1 unq_name_situationid_ruleid; Type: CONSTRAINT; Schema: public;d
--

ALTER TABLE ONLY public.ref_rootcause_v1
    ADD CONSTRAINT unq_name_situationid_ruleid UNIQUE (name, situation_id, rule_id);


--
-- Name: situation_template_instances_v1 unq_situation_template_instances_v1_situationid_name; Type: CONSTRAINT; Schema: public;d
--

ALTER TABLE ONLY public.situation_template_instances_v1
    ADD CONSTRAINT unq_situation_template_instances_v1_situationid_name UNIQUE (situation_id, name);


--
-- Name: user_groups_v1 user_groups_v1_name_key; Type: CONSTRAINT; Schema: public;d
--

ALTER TABLE ONLY public.user_groups_v1
    ADD CONSTRAINT user_groups_v1_name_key UNIQUE (name);


--
-- Name: user_groups_v1 user_groups_v1_pkey; Type: CONSTRAINT; Schema: public;d
--

ALTER TABLE ONLY public.user_groups_v1
    ADD CONSTRAINT user_groups_v1_pkey PRIMARY KEY (id);


--
-- Name: user_memberships_v1 user_memberships_v1_pkey; Type: CONSTRAINT; Schema: public;d
--

ALTER TABLE ONLY public.user_memberships_v1
    ADD CONSTRAINT user_memberships_v1_pkey PRIMARY KEY (user_id, group_id);


--
-- Name: users_v1 users_v1_login_key; Type: CONSTRAINT; Schema: public;d
--

ALTER TABLE ONLY public.users_v1
    ADD CONSTRAINT users_v1_login_key UNIQUE (login);


--
-- Name: users_v1 users_v1_pkey; Type: CONSTRAINT; Schema: public;d
--

ALTER TABLE ONLY public.users_v1
    ADD CONSTRAINT users_v1_pkey PRIMARY KEY (id);


--
-- Name: users_v4 users_v4_login_key; Type: CONSTRAINT; Schema: public;d
--

ALTER TABLE ONLY public.users_v4
    ADD CONSTRAINT users_v4_login_key UNIQUE (login);


--
-- Name: users_v4 users_v4_pkey; Type: CONSTRAINT; Schema: public;d
--

ALTER TABLE ONLY public.users_v4
    ADD CONSTRAINT users_v4_pkey PRIMARY KEY (id);


--
-- Name: variables_config_v1 variables_config_v1_key_key; Type: CONSTRAINT; Schema: public;d
--

ALTER TABLE ONLY public.variables_config_v1
    ADD CONSTRAINT variables_config_v1_key_key UNIQUE (key);


--
-- Name: variables_config_v1 variables_config_v1_pkey; Type: CONSTRAINT; Schema: public;d
--

ALTER TABLE ONLY public.variables_config_v1
    ADD CONSTRAINT variables_config_v1_pkey PRIMARY KEY (id);


--
-- Name: idx_connectors_executions_log_connector_id_name_ts; Type: INDEX; Schema: public;d
--

CREATE INDEX idx_connectors_executions_log_connector_id_name_ts ON public.connectors_executions_log_v1 USING btree (connector_id, name, ts DESC);


--
-- Name: idx_fact_history_v5_combo; Type: INDEX; Schema: public;d
--

CREATE INDEX idx_fact_history_v5_combo ON public.fact_history_v5 USING btree (fact_id, ts DESC) INCLUDE (id);


--
-- Name: idx_situation_fact_history_v5_situation_history_id; Type: INDEX; Schema: public;d
--

CREATE INDEX idx_situation_fact_history_v5_situation_history_id ON public.situation_fact_history_v5 USING btree (situation_history_id);


--
-- Name: idx_situation_history_v5_combo; Type: INDEX; Schema: public;d
--

CREATE INDEX idx_situation_history_v5_combo ON public.situation_history_v5 USING btree (situation_id, situation_instance_id, ts DESC) INCLUDE (id);


--
-- Name: unique_default_true; Type: INDEX; Schema: public;d
--

CREATE UNIQUE INDEX unique_default_true ON public.elasticsearch_config_v1 USING btree ("default") WHERE ("default" = true);


--
-- Name: elasticsearch_config_v1 ensure_one_default_row_trigger; Type: TRIGGER; Schema: public;d
--

CREATE TRIGGER ensure_one_default_row_trigger AFTER DELETE ON public.elasticsearch_config_v1 FOR EACH ROW EXECUTE FUNCTION public.ensure_one_default_row();


--
-- Name: elasticsearch_config_v1 prevent_delete_default_row_trigger; Type: TRIGGER; Schema: public;d
--

CREATE TRIGGER prevent_delete_default_row_trigger BEFORE DELETE ON public.elasticsearch_config_v1 FOR EACH ROW EXECUTE FUNCTION public.prevent_delete_default_row();


--
-- Name: calendar_union_v1 calendar_union_v1_calendar_id_fkey; Type: FK CONSTRAINT; Schema: public;d
--

ALTER TABLE ONLY public.calendar_union_v1
    ADD CONSTRAINT calendar_union_v1_calendar_id_fkey FOREIGN KEY (calendar_id) REFERENCES public.calendar_v1(id);


--
-- Name: calendar_union_v1 calendar_union_v1_sub_calendar_id_fkey; Type: FK CONSTRAINT; Schema: public;d
--

ALTER TABLE ONLY public.calendar_union_v1
    ADD CONSTRAINT calendar_union_v1_sub_calendar_id_fkey FOREIGN KEY (sub_calendar_id) REFERENCES public.calendar_v1(id);


--
-- Name: external_generic_config_versions_v1 fk_config_id; Type: FK CONSTRAINT; Schema: public;d
--

ALTER TABLE ONLY public.external_generic_config_versions_v1
    ADD CONSTRAINT fk_config_id FOREIGN KEY (config_id) REFERENCES public.external_generic_config_v1(id) ON DELETE CASCADE;


--
-- Name: issue_detection_feedback_v3 issue_detection_feedback_v3_issue_id_fkey; Type: FK CONSTRAINT; Schema: public;d
--

ALTER TABLE ONLY public.issue_detection_feedback_v3
    ADD CONSTRAINT issue_detection_feedback_v3_issue_id_fkey FOREIGN KEY (issue_id) REFERENCES public.issues_v1(id);


--
-- Name: issue_detection_feedback_v3 issue_detection_feedback_v3_user_id_fkey; Type: FK CONSTRAINT; Schema: public;d
--

ALTER TABLE ONLY public.issue_detection_feedback_v3
    ADD CONSTRAINT issue_detection_feedback_v3_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users_v4(id);


--
-- Name: issue_resolution_draft_v1 issue_resolution_draft_v1_issue_id_fkey; Type: FK CONSTRAINT; Schema: public;d
--

ALTER TABLE ONLY public.issue_resolution_draft_v1
    ADD CONSTRAINT issue_resolution_draft_v1_issue_id_fkey FOREIGN KEY (issue_id) REFERENCES public.issues_v1(id);


--
-- Name: issue_resolution_v1 issue_resolution_v1_action_id_fkey; Type: FK CONSTRAINT; Schema: public;d
--

ALTER TABLE ONLY public.issue_resolution_v1
    ADD CONSTRAINT issue_resolution_v1_action_id_fkey FOREIGN KEY (action_id) REFERENCES public.ref_action_v1(id);


--
-- Name: issue_resolution_v1 issue_resolution_v1_issue_id_fkey; Type: FK CONSTRAINT; Schema: public;d
--

ALTER TABLE ONLY public.issue_resolution_v1
    ADD CONSTRAINT issue_resolution_v1_issue_id_fkey FOREIGN KEY (issue_id) REFERENCES public.issues_v1(id);


--
-- Name: issue_resolution_v1 issue_resolution_v1_rootcause_id_fkey; Type: FK CONSTRAINT; Schema: public;d
--

ALTER TABLE ONLY public.issue_resolution_v1
    ADD CONSTRAINT issue_resolution_v1_rootcause_id_fkey FOREIGN KEY (rootcause_id) REFERENCES public.ref_rootcause_v1(id);


--
-- Name: issues_v1 issues_v1_situation_history_id_fkey; Type: FK CONSTRAINT; Schema: public;d
--

ALTER TABLE ONLY public.issues_v1
    ADD CONSTRAINT issues_v1_situation_history_id_fkey FOREIGN KEY (situation_history_id) REFERENCES public.situation_history_v5(id);


--
-- Name: issues_v1 issues_v1_situation_id_fkey; Type: FK CONSTRAINT; Schema: public;d
--

ALTER TABLE ONLY public.issues_v1
    ADD CONSTRAINT issues_v1_situation_id_fkey FOREIGN KEY (situation_id) REFERENCES public.situation_definition_v1(id);


--
-- Name: ref_action_v1 ref_action_v1_rootcause_id_fkey; Type: FK CONSTRAINT; Schema: public;d
--

ALTER TABLE ONLY public.ref_action_v1
    ADD CONSTRAINT ref_action_v1_rootcause_id_fkey FOREIGN KEY (rootcause_id) REFERENCES public.ref_rootcause_v1(id);


--
-- Name: ref_rootcause_v1 ref_rootcause_v1_rule_id_fkey; Type: FK CONSTRAINT; Schema: public;d
--

ALTER TABLE ONLY public.ref_rootcause_v1
    ADD CONSTRAINT ref_rootcause_v1_rule_id_fkey FOREIGN KEY (rule_id) REFERENCES public.rules_v1(id);


--
-- Name: ref_rootcause_v1 ref_rootcause_v1_situation_id_fkey; Type: FK CONSTRAINT; Schema: public;d
--

ALTER TABLE ONLY public.ref_rootcause_v1
    ADD CONSTRAINT ref_rootcause_v1_situation_id_fkey FOREIGN KEY (situation_id) REFERENCES public.situation_definition_v1(id);


--
-- Name: roles_permissions_v4 roles_permissions_v4_permission_id_fkey; Type: FK CONSTRAINT; Schema: public;d
--

ALTER TABLE ONLY public.roles_permissions_v4
    ADD CONSTRAINT roles_permissions_v4_permission_id_fkey FOREIGN KEY (permission_id) REFERENCES public.permissions_v4(id);


--
-- Name: roles_permissions_v4 roles_permissions_v4_role_id_fkey; Type: FK CONSTRAINT; Schema: public;d
--

ALTER TABLE ONLY public.roles_permissions_v4
    ADD CONSTRAINT roles_permissions_v4_role_id_fkey FOREIGN KEY (role_id) REFERENCES public.roles_v4(id);


--
-- Name: rule_versions_v1 rule_versions_v1_rule_id_fkey; Type: FK CONSTRAINT; Schema: public;d
--

ALTER TABLE ONLY public.rule_versions_v1
    ADD CONSTRAINT rule_versions_v1_rule_id_fkey FOREIGN KEY (rule_id) REFERENCES public.rules_v1(id);


--
-- Name: rules_v1 rules_v1_calendar_id_fkey; Type: FK CONSTRAINT; Schema: public;d
--

ALTER TABLE ONLY public.rules_v1
    ADD CONSTRAINT rules_v1_calendar_id_fkey FOREIGN KEY (calendar_id) REFERENCES public.calendar_v1(id);


--
-- Name: situation_definition_v1 situation_definition_v1_calendar_id_fkey; Type: FK CONSTRAINT; Schema: public;d
--

ALTER TABLE ONLY public.situation_definition_v1
    ADD CONSTRAINT situation_definition_v1_calendar_id_fkey FOREIGN KEY (calendar_id) REFERENCES public.calendar_v1(id);


--
-- Name: situation_fact_history_v5 situation_fact_history_v5_fact_history_id_fkey; Type: FK CONSTRAINT; Schema: public;d
--

ALTER TABLE ONLY public.situation_fact_history_v5
    ADD CONSTRAINT situation_fact_history_v5_fact_history_id_fkey FOREIGN KEY (fact_history_id) REFERENCES public.fact_history_v5(id);


--
-- Name: situation_fact_history_v5 situation_fact_history_v5_situation_history_id_fkey; Type: FK CONSTRAINT; Schema: public;d
--

ALTER TABLE ONLY public.situation_fact_history_v5
    ADD CONSTRAINT situation_fact_history_v5_situation_history_id_fkey FOREIGN KEY (situation_history_id) REFERENCES public.situation_history_v5(id);


--
-- Name: situation_facts_v1 situation_facts_v1_fact_id_fkey; Type: FK CONSTRAINT; Schema: public;d
--

ALTER TABLE ONLY public.situation_facts_v1
    ADD CONSTRAINT situation_facts_v1_fact_id_fkey FOREIGN KEY (fact_id) REFERENCES public.fact_definition_v1(id);


--
-- Name: situation_facts_v1 situation_facts_v1_situation_id_fkey; Type: FK CONSTRAINT; Schema: public;d
--

ALTER TABLE ONLY public.situation_facts_v1
    ADD CONSTRAINT situation_facts_v1_situation_id_fkey FOREIGN KEY (situation_id) REFERENCES public.situation_definition_v1(id);


--
-- Name: situation_rules_v1 situation_rules_v1_rule_id_fkey; Type: FK CONSTRAINT; Schema: public;d
--

ALTER TABLE ONLY public.situation_rules_v1
    ADD CONSTRAINT situation_rules_v1_rule_id_fkey FOREIGN KEY (rule_id) REFERENCES public.rules_v1(id);


--
-- Name: situation_rules_v1 situation_rules_v1_situation_id_fkey; Type: FK CONSTRAINT; Schema: public;d
--

ALTER TABLE ONLY public.situation_rules_v1
    ADD CONSTRAINT situation_rules_v1_situation_id_fkey FOREIGN KEY (situation_id) REFERENCES public.situation_definition_v1(id);


--
-- Name: situation_template_instances_v1 situation_template_instances_v1_calendar_id_fkey; Type: FK CONSTRAINT; Schema: public;d
--

ALTER TABLE ONLY public.situation_template_instances_v1
    ADD CONSTRAINT situation_template_instances_v1_calendar_id_fkey FOREIGN KEY (calendar_id) REFERENCES public.calendar_v1(id);


--
-- Name: situation_template_instances_v1 situation_template_instances_v1_situation_id_fkey; Type: FK CONSTRAINT; Schema: public;d
--

ALTER TABLE ONLY public.situation_template_instances_v1
    ADD CONSTRAINT situation_template_instances_v1_situation_id_fkey FOREIGN KEY (situation_id) REFERENCES public.situation_definition_v1(id);


--
-- Name: user_memberships_v1 user_memberships_v1_group_id_fkey; Type: FK CONSTRAINT; Schema: public;d
--

ALTER TABLE ONLY public.user_memberships_v1
    ADD CONSTRAINT user_memberships_v1_group_id_fkey FOREIGN KEY (group_id) REFERENCES public.user_groups_v1(id);


--
-- Name: user_memberships_v1 user_memberships_v1_user_id_fkey; Type: FK CONSTRAINT; Schema: public;d
--

ALTER TABLE ONLY public.user_memberships_v1
    ADD CONSTRAINT user_memberships_v1_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users_v1(id);


--
-- Name: users_roles_v4 users_roles_v4_role_id_fkey; Type: FK CONSTRAINT; Schema: public;d
--

ALTER TABLE ONLY public.users_roles_v4
    ADD CONSTRAINT users_roles_v4_role_id_fkey FOREIGN KEY (role_id) REFERENCES public.roles_v4(id);


--
-- Name: users_roles_v4 users_roles_v4_user_id_fkey; Type: FK CONSTRAINT; Schema: public;d
--

ALTER TABLE ONLY public.users_roles_v4
    ADD CONSTRAINT users_roles_v4_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users_v4(id);


--
-- PostgreSQL database dump complete
--

