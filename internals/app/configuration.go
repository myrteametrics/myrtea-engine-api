package app

import (
	"github.com/myrteametrics/myrtea-sdk/v4/helpers"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// ConfigPath is the toml configuration file path
var ConfigPath = "config"

// ConfigName is the toml configuration file name
var ConfigName = "engine-api"

// EnvPrefix is the standard environment variable prefix
var EnvPrefix = "MYRTEA"

// AllowedConfigKey list every allowed configuration key
var AllowedConfigKey = [][]helpers.ConfigKey{
	helpers.GetGeneralConfigKeys(),
	helpers.GetHTTPServerConfigKeys(),
	helpers.GetPostgresqlConfigKeys(),
	helpers.GetElasticsearchConfigKeys(),
	{
		{Type: helpers.StringFlag, Name: "HTTP_SERVER_API_ENABLE_VERBOSE_ERROR", DefaultValue: "false", Description: "Run the API with verbose error"},
		{Type: helpers.StringFlag, Name: "SWAGGER_HOST", DefaultValue: "localhost:9000", Description: "Swagger UI target hostname"},
		{Type: helpers.StringFlag, Name: "SWAGGER_BASEPATH", DefaultValue: "/api/v5", Description: "Swagger UI target basepath"},
		{Type: helpers.StringFlag, Name: "POSTGRESQL_CONN_POOL_MAX_OPEN", DefaultValue: "6", Description: "PostgreSQL connection pool max open"},
		{Type: helpers.StringFlag, Name: "POSTGRESQL_CONN_POOL_MAX_IDLE", DefaultValue: "3", Description: "PostgreSQL connection pool max idle"},
		{Type: helpers.StringFlag, Name: "POSTGRESQL_CONN_MAX_LIFETIME", DefaultValue: "0", Description: "PostgreSQL connection max lifetime"},
		{Type: helpers.StringFlag, Name: "ENABLE_CRONS_ON_START", DefaultValue: "true", Description: "Enable crons on startup"},
		{Type: helpers.StringFlag, Name: "AUTHENTICATION_MODE", DefaultValue: "BASIC", Description: "Authentication mode"},
		{Type: helpers.StringFlag, Name: "AUTHENTICATION_SAML_ROOT_URL", DefaultValue: "http://localhost:9000/api/v5/", Description: "SAML Root URL"},
		{Type: helpers.StringFlag, Name: "AUTHENTICATION_SAML_ENTITYID", DefaultValue: "http://localhost:9000/", Description: "SAML EntityID"},
		{Type: helpers.StringFlag, Name: "AUTHENTICATION_SAML_KEY_FILE_PATH", DefaultValue: "resources/saml/certs/myservice.key", Description: "SAML SSL certificate private key"},
		{Type: helpers.StringFlag, Name: "AUTHENTICATION_SAML_CRT_FILE_PATH", DefaultValue: "resources/saml/certs/myservice.crt", Description: "SAML SSL certificate public key"},
		{Type: helpers.StringFlag, Name: "AUTHENTICATION_SAML_METADATA_MODE", DefaultValue: "FILE", Description: "SAML MetadataMode (FILE OR FETCH)"},
		{Type: helpers.StringFlag, Name: "AUTHENTICATION_SAML_METADATA_FILE_PATH", DefaultValue: "saml/idp-metadata.xml", Description: "SAML Metadata file path"},
		{Type: helpers.StringFlag, Name: "AUTHENTICATION_SAML_METADATA_FETCH_URL", DefaultValue: "https://samltest.id/saml/idp", Description: "SAML Metadata fetch url"},
		{Type: helpers.StringFlag, Name: "AUTHENTICATION_SAML_ENABLE_MEMBEROF_VALIDATION", DefaultValue: "true", Description: "SAML Enable memberOf validation"},
		{Type: helpers.StringFlag, Name: "AUTHENTICATION_SAML_ATTRIBUTE_USER_ID", DefaultValue: "uid", Description: "SAML Attribute userID"},
		{Type: helpers.StringFlag, Name: "AUTHENTICATION_SAML_ATTRIBUTE_USER_DISPLAYNAME", DefaultValue: "cn", Description: "SAML Attribute displayName"},
		{Type: helpers.StringFlag, Name: "AUTHENTICATION_SAML_ATTRIBUTE_USER_MEMBEROF", DefaultValue: "groups", Description: "SAML Attribute memberOf"},
		{Type: helpers.StringFlag, Name: "AUTHENTICATION_SAML_ADMIN_GROUP_NAME", DefaultValue: "administrator", Description: "SAML Admin group name"},
		{Type: helpers.StringFlag, Name: "AUTHENTICATION_SAML_COOKIE_MAX_AGE_DURATION", DefaultValue: "1h", Description: "SAML Cookie max age (time.Duration)"},
		{Type: helpers.StringFlag, Name: "SMTP_USERNAME", DefaultValue: "smtp@example.com", Description: "SMTP Authentication Username"},
		{Type: helpers.StringFlag, Name: "SMTP_PASSWORD", DefaultValue: "", Description: "SMTP Authentication password"},
		{Type: helpers.StringFlag, Name: "SMTP_HOST", DefaultValue: "smtp.example.com", Description: "SMTP Authentication host"},
		{Type: helpers.StringFlag, Name: "SMTP_PORT", DefaultValue: "465", Description: "SMTP Authentication port"},
		{Type: helpers.StringFlag, Name: "AUTHENTICATION_OIDC_CLIENT_ID", DefaultValue: "", Description: "A unique identifier representing the client application seeking access to the server's resources."},
		{Type: helpers.StringFlag, Name: "AUTHENTICATION_OIDC_CLIENT_SECRET", DefaultValue: "", Description: "A shared secret between the client application and the authentication server to prove the client's identity."},
		{Type: helpers.StringFlag, Name: "AUTHENTICATION_OIDC_REDIRECT_URL", DefaultValue: "https://127.0.0.1:5556/auth/oidc/callback", Description: "The redirection URL to which the user will be redirected after successful authentication."},
		{Type: helpers.StringFlag, Name: "AUTHENTICATION_OIDC_ISSUER_URL", DefaultValue: "", Description: "The URL of the OIDC (OpenID Connect) server providing the authentication service."},
		{Type: helpers.StringFlag, Name: "AUTHENTICATION_OIDC_FRONT_END_URL", DefaultValue: "https://127.0.0.1:4200", Description: "The URL of the front-end application to which the user will be redirected after successful authentication."},
		{Type: helpers.StringFlag, Name: "AUTHENTICATION_OIDC_ENCRYPTION_KEY", DefaultValue: "hisis24characterslongs", Description: "The secret key used for state encryption/decryption in the OpenID Connect authentication process."},
		{Type: helpers.StringFlag, Name: "AUTHENTICATION_OIDC_SCOPES ", DefaultValue: "The scopes of access requested when authenticating with the OIDC server. (Only if AUTHENTICATION_MODE=`OIDC`)"},
	},
}

func InitConfiguration() {
	helpers.InitializeConfig(AllowedConfigKey, ConfigName, ConfigPath, EnvPrefix)

	// Custom plugins config
	v := viper.New()
	v.SetConfigName("plugins")
	v.AddConfigPath("config")
	err := v.ReadInConfig()
	if err != nil {
		zap.L().Warn("No plugins configuration found", zap.Error(err))
		return
	}
	err = viper.MergeConfigMap(v.AllSettings())
	if err != nil {
		zap.L().Warn("No plugins configuration found", zap.Error(err))
	}

}
