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
	{Type: configuration.StringFlag, Name: "API_ENABLE_CORS", DefaultValue: "false", Description: "Run the API with CORS enabled"},
	{Type: configuration.StringFlag, Name: "API_ENABLE_SECURITY", DefaultValue: "true", Description: "Run the API in unsecured mode (without authentication)"},
	{Type: configuration.StringFlag, Name: "API_ENABLE_GATEWAY_MODE", DefaultValue: "false", Description: "Run the API without external Auth API (with gateway)"},
	{Type: configuration.StringFlag, Name: "API_ENABLE_VERBOSE_ERROR", DefaultValue: "false", Description: "Run the API with verbose error"},
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
	{Type: configuration.StringFlag, Name: "AUTHENTICATION_MODE", DefaultValue: "BASIC", Description: "Authentication mode"},
	{Type: configuration.StringFlag, Name: "AUTHENTICATION_SAML_ROOT_URL", DefaultValue: "http://localhost:9000/api/v4/", Description: "SAML Root URL"},
	{Type: configuration.StringFlag, Name: "AUTHENTICATION_SAML_ENTITYID", DefaultValue: "http://localhost:9000/", Description: "SAML EntityID"},
	{Type: configuration.StringFlag, Name: "AUTHENTICATION_SAML_KEY_FILE_PATH", DefaultValue: "resources/saml/certs/myservice.key", Description: "SAML SSL certificate private key"},
	{Type: configuration.StringFlag, Name: "AUTHENTICATION_SAML_CRT_FILE_PATH", DefaultValue: "resources/saml/certs/myservice.crt", Description: "SAML SSL certificate public key"},
	{Type: configuration.StringFlag, Name: "AUTHENTICATION_SAML_METADATA_MODE", DefaultValue: "FILE", Description: "SAML MetadataMode (FILE OR FETCH)"},
	{Type: configuration.StringFlag, Name: "AUTHENTICATION_SAML_METADATA_FILE_PATH", DefaultValue: "saml/idp-metadata.xml", Description: "SAML Metadata file path"},
	{Type: configuration.StringFlag, Name: "AUTHENTICATION_SAML_METADATA_FETCH_URL", DefaultValue: "https://samltest.id/saml/idp", Description: "SAML Metadata fetch url"},
	{Type: configuration.StringFlag, Name: "AUTHENTICATION_SAML_ENABLE_MEMBEROF_VALIDATION", DefaultValue: "true", Description: "SAML Enable memberOf validation"},
	{Type: configuration.StringFlag, Name: "AUTHENTICATION_SAML_ATTRIBUTE_USER_ID", DefaultValue: "uid", Description: "SAML Attribute userID"},
	{Type: configuration.StringFlag, Name: "AUTHENTICATION_SAML_ATTRIBUTE_USER_DISPLAYNAME", DefaultValue: "cn", Description: "SAML Attribute displayName"},
	{Type: configuration.StringFlag, Name: "AUTHENTICATION_SAML_ATTRIBUTE_USER_MEMBEROF", DefaultValue: "groups", Description: "SAML Attribute memberOf"},
	{Type: configuration.StringFlag, Name: "AUTHENTICATION_SAML_ADMIN_GROUP_NAME", DefaultValue: "administrator", Description: "SAML Admin group name"},
	{Type: configuration.StringFlag, Name: "AUTHENTICATION_SAML_COOKIE_MAX_AGE_DURATION", DefaultValue: "1h", Description: "SAML Cookie max age (time.Duration)"},
}

func initConfiguration() {
	configuration.InitializeConfig(AllowedConfigKey, ConfigName, ConfigPath, EnvPrefix)
}
