package export

import (
	"testing"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/coordinator"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/fact"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/history"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/modeler"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/rule"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/situation"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/tests"
	"github.com/myrteametrics/myrtea-sdk/v4/elasticsearchv8"
	"github.com/myrteametrics/myrtea-sdk/v4/helpers"
	"github.com/myrteametrics/myrtea-sdk/v4/postgres"
	"github.com/spf13/viper"

	"github.com/myrteametrics/myrtea-sdk/v4/engine"
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
		{Type: helpers.StringFlag, Name: "SMTP_PASSWORD", DefaultValue: "password", Description: "SMTP Authentication password"},
		{Type: helpers.StringFlag, Name: "SMTP_HOST", DefaultValue: "smtp.example.com", Description: "SMTP Authentication host"},
		{Type: helpers.StringFlag, Name: "SMTP_PORT", DefaultValue: "465", Description: "SMTP Authentication port"},
	},
}

func TestExportFactHits(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping elasticsearch test in short mode")
	}

	helpers.InitializeConfig(AllowedConfigKey, ConfigName, "../../config", EnvPrefix)
	helpers.InitLogger(viper.GetBool("LOGGER_PRODUCTION"))
	urls := viper.GetStringSlice("ELASTICSEARCH_URLS")
	elasticsearchv8.ReplaceGlobals(elasticsearch.Config{
		Addresses: urls,
	})

	db := tests.DBClient(t)
	postgres.ReplaceGlobals(db)
	history.ReplaceGlobals(history.New(db))
	fact.ReplaceGlobals(fact.NewPostgresRepository(db))
	situation.ReplaceGlobals(situation.NewPostgresRepository(db))
	rule.ReplaceGlobals(rule.NewPostgresRepository(db))
	modeler.ReplaceGlobals(modeler.NewPostgresRepository(db))
	instanceName := viper.GetString("INSTANCE_NAME")
	models, err := modeler.R().GetAll()
	if err != nil {
		t.Error(err)
	}

	if err := coordinator.InitInstance(instanceName, models); err != nil {
		t.Error(err)
	}

	f := engine.Fact{
		ID:               1,
		Name:             "test",
		IsObject:         false,
		IsTemplate:       false,
		Model:            "test", // Don't forget to change to an existing model (model, term & field)
		CalculationDepth: 90,
		Intent:           &engine.IntentFragment{Name: "doc_count", Operator: engine.Count, Term: "test"},
		Condition: &engine.BooleanFragment{Operator: engine.And, Fragments: []engine.ConditionFragment{
			&engine.LeafConditionFragment{Operator: engine.For, Field: "my_bool", Value: true},
		},
		},
	}
	hits, err := ExportFactHitsFull(f)
	if err != nil {
		t.Error(err)
	}
	t.Log(len(hits))
}
