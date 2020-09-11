package app

import "github.com/myrteametrics/myrtea-sdk/v4/configuration"

// ConfigPath is the toml configuration file path
var ConfigPath = "config"

// ConfigName is the toml configuration file name
var ConfigName = "engine-api"

// EnvPrefix is the standard environment variable prefix
var EnvPrefix = "MYRTEA"

// AllowedConfigKey list every allowed configuration key
var AllowedConfigKey = []configuration.ConfigKey{
	{Type: configuration.StringFlag, Name: "DEBUG_MODE", DefaultValue: "false", Description: "Enable debug mode"},
	{Type: configuration.StringFlag, Name: "SERVER_PORT", DefaultValue: "9000", Description: "Server port"},
	{Type: configuration.StringFlag, Name: "SERVER_ENABLE_TLS", DefaultValue: "false", Description: "Run the server in unsecured mode (without SSL)"},
	{Type: configuration.StringFlag, Name: "SERVER_TLS_FILE_CRT", DefaultValue: "certs/server.rsa.crt", Description: "SSL certificate crt file location"},
	{Type: configuration.StringFlag, Name: "SERVER_TLS_FILE_KEY", DefaultValue: "certs/server.rsa.key", Description: "SSL certificate key file location"},
	{Type: configuration.StringFlag, Name: "API_ENABLE_CORS", DefaultValue: "false", Description: "Run the api with CORS enabled"},
	{Type: configuration.StringFlag, Name: "API_ENABLE_SECURITY", DefaultValue: "true", Description: "Run the api in unsecured mode (without authentication)"},
	{Type: configuration.StringFlag, Name: "API_ENABLE_GATEWAY_MODE", DefaultValue: "false", Description: "Run the api without external Auth API (with gateway)"},
	{Type: configuration.StringFlag, Name: "INSTANCE_NAME", DefaultValue: "myrtea", Description: "Myrtea instance name"},
	{Type: configuration.StringFlag, Name: "SWAGGER_HOST", DefaultValue: "localhost:9000", Description: "Swagger UI target hostname"},
	{Type: configuration.StringFlag, Name: "SWAGGER_BASEPATH", DefaultValue: "/api/v4", Description: "Swagger UI target basepath"},
	{Type: configuration.StringSliceFlag, Name: "ELASTICSEARCH_URLS", DefaultValue: []string{"http://localhost:9200"}, Description: "Elasticsearch URLS"},
	{Type: configuration.StringFlag, Name: "POSTGRESQL_HOSTNAME", DefaultValue: "localhost", Description: "PostgreSQL hostname"},
	{Type: configuration.StringFlag, Name: "POSTGRESQL_PORT", DefaultValue: "5432", Description: "PostgreSQL port"},
	{Type: configuration.StringFlag, Name: "POSTGRESQL_DBNAME", DefaultValue: "postgres", Description: "PostgreSQL database name"},
	{Type: configuration.StringFlag, Name: "POSTGRESQL_USERNAME", DefaultValue: "postgres", Description: "PostgreSQL user"},
	{Type: configuration.StringFlag, Name: "POSTGRESQL_PASSWORD", DefaultValue: "postgres", Description: "PostgreSQL pasword"},
	{Type: configuration.StringFlag, Name: "POSTGRESQL_CONN_POOL_MAX_OPEN", DefaultValue: "6", Description: "PostgreSQL connection pool max open"},
	{Type: configuration.StringFlag, Name: "POSTGRESQL_CONN_POOL_MAX_IDLE", DefaultValue: "3", Description: "PostgreSQL connection pool max idle"},
	{Type: configuration.StringFlag, Name: "POSTGRESQL_CONN_MAX_LIFETIME", DefaultValue: "0", Description: "PostgreSQL connection max lifetime"},
	{Type: configuration.StringFlag, Name: "ENABLE_CRONS_ON_START", DefaultValue: "true", Description: "Enable crons on startup"},
}

func initConfiguration() {
	configuration.InitializeConfig(AllowedConfigKey, ConfigName, ConfigPath, EnvPrefix)
}
