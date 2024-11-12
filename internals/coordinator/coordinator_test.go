package coordinator

import (
	"github.com/elastic/go-elasticsearch/v8"
	elasticsearchsdk "github.com/myrteametrics/myrtea-sdk/v5/elasticsearch"
	"sync"
	"testing"
	"time"

	"github.com/myrteametrics/myrtea-sdk/v5/helpers"
	"github.com/myrteametrics/myrtea-sdk/v5/modeler"
	"github.com/spf13/viper"
)

// ConfigPath is the toml configuration file path
var ConfigPath = "config"

// ConfigName is the toml configuration file name
var ConfigName = "engine-api"

// EnvPrefix is the standard environment variable prefix
var EnvPrefix = "MYRTEA"

// AllowedConfigKey list every allowed configuration key
var AllowedConfigKey = [][]helpers.ConfigKey{
	[]helpers.ConfigKey{
		{Type: helpers.StringFlag, Name: "DEBUG_MODE", DefaultValue: "false", Description: "Enable debug mode"},
		{Type: helpers.StringFlag, Name: "LOGGER_PRODUCTION", DefaultValue: "true", Description: "Enable or disable production log"},
		{Type: helpers.StringFlag, Name: "SERVER_PORT", DefaultValue: "9000", Description: "Server port"},
		{Type: helpers.StringFlag, Name: "SERVER_ENABLE_TLS", DefaultValue: "false", Description: "Run the server in unsecured mode (without SSL)"},
		{Type: helpers.StringFlag, Name: "SERVER_TLS_FILE_CRT", DefaultValue: "certs/server.rsa.crt", Description: "SSL certificate crt file location"},
		{Type: helpers.StringFlag, Name: "SERVER_TLS_FILE_KEY", DefaultValue: "certs/server.rsa.key", Description: "SSL certificate key file location"},
		{Type: helpers.StringFlag, Name: "API_ENABLE_CORS", DefaultValue: "false", Description: "Run the API with CORS enabled"},
		{Type: helpers.StringFlag, Name: "API_ENABLE_SECURITY", DefaultValue: "true", Description: "Run the API in unsecured mode (without authentication)"},
		{Type: helpers.StringFlag, Name: "API_ENABLE_GATEWAY_MODE", DefaultValue: "false", Description: "Run the API without external Auth API (with gateway)"},
		{Type: helpers.StringFlag, Name: "API_ENABLE_VERBOSE_ERROR", DefaultValue: "false", Description: "Run the API with verbose error"},
		{Type: helpers.StringFlag, Name: "INSTANCE_NAME", DefaultValue: "myrtea", Description: "Myrtea instance name"},
		{Type: helpers.StringFlag, Name: "SWAGGER_HOST", DefaultValue: "localhost:9000", Description: "Swagger UI target hostname"},
		{Type: helpers.StringFlag, Name: "SWAGGER_BASEPATH", DefaultValue: "/api/v5", Description: "Swagger UI target basepath"},
		{Type: helpers.StringSliceFlag, Name: "ELASTICSEARCH_URLS", DefaultValue: []string{"http://localhost:9200"}, Description: "Elasticsearch URLS"},
		{Type: helpers.StringFlag, Name: "POSTGRESQL_HOSTNAME", DefaultValue: "localhost", Description: "PostgreSQL hostname"},
		{Type: helpers.StringFlag, Name: "POSTGRESQL_PORT", DefaultValue: "5432", Description: "PostgreSQL port"},
		{Type: helpers.StringFlag, Name: "POSTGRESQL_DBNAME", DefaultValue: "postgres", Description: "PostgreSQL database name"},
		{Type: helpers.StringFlag, Name: "POSTGRESQL_USERNAME", DefaultValue: "postgres", Description: "PostgreSQL user"},
		{Type: helpers.StringFlag, Name: "POSTGRESQL_PASSWORD", DefaultValue: "postgres", Description: "PostgreSQL pasword"},
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
	},
}

var (
	liveIndicesCaches []string
	mu                sync.RWMutex
)

// func TestMaster(t *testing.T) {
// 	helpers.InitializeConfig(AllowedConfigKey, ConfigName, ConfigPath, EnvPrefix)
// 	helpers.InitLogger(viper.GetBool("LOGGER_PRODUCTION"))

// 	if err := GetInstance().InitInstance(viper.GetString("INSTANCE_NAME"), viper.GetStringSlice("ELASTICSEARCH_URLS"), map[int64]modeler.Model{
// 		1: {
// 			ID:   1,
// 			Name: "myindex",
// 			ElasticsearchOptions: modeler.ElasticsearchOptions{
// 				Rollmode:                  "none",
// 				Rollcron:                  "",
// 				EnablePurge:               true,
// 				PurgeMaxConcurrentIndices: 5,
// 				PatchAliasMaxIndices:      0,
// 			},
// 			Fields: []modeler.Field{
// 				&modeler.FieldLeaf{Name: "id", Ftype: modeler.String},
// 				&modeler.FieldLeaf{Name: "data", Ftype: modeler.String},
// 			},
// 		},
// 	}); err != nil {
// 		zap.L().Fatal("Intialisation of coordinator master", zap.Error(err))
// 	}

// 	ts := time.Date(2023, 1, 30, 12, 0, 0, 0, time.UTC)
// 	t.Log(GetInstance().Instances[viper.GetString("INSTANCE_NAME")])
// 	logicalIndex := GetInstance().LogicalIndex("myindex")
// 	t.Log(logicalIndex)
// 	err := logicalIndex.FetchIndices()
// 	if err != nil {
// 		t.Error(err)
// 	}

// 	indices := GetInstance().LogicalIndex("myindex").GetIndicesGte(ts, 20)
// 	t.Log(indices)
// 	t.Fail()

// }

func TestCoordinator(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping elasticsearch test in short mode")
	}
	helpers.InitializeConfig(AllowedConfigKey, ConfigName, ConfigPath, EnvPrefix)
	helpers.InitLogger(viper.GetBool("LOGGER_PRODUCTION"))

	elasticsearchsdk.ReplaceGlobals(elasticsearch.Config{Addresses: viper.GetStringSlice("ELASTICSEARCH_URLS")})

	err := InitInstance(
		viper.GetString("INSTANCE_NAME"),
		map[int64]modeler.Model{
			1: {
				ID:   1,
				Name: "myindex",
				ElasticsearchOptions: modeler.ElasticsearchOptions{
					Rollmode:                  modeler.RollmodeSettings{Type: modeler.RollmodeTimeBased},
					Rollcron:                  "0 * * * *",
					EnablePurge:               true,
					PurgeMaxConcurrentIndices: 30,
					PatchAliasMaxIndices:      0,
				},
			},
		},
	)
	if err != nil {
		t.Error(err)
	}
	t.Log(GetInstance().LogicalIndex("myindex").FindIndices(time.Now(), 26))
	t.Fail()
}

func TestPurge(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping elasticsearch test in short mode")
	}
	helpers.InitializeConfig(AllowedConfigKey, ConfigName, ConfigPath, EnvPrefix)
	helpers.InitLogger(viper.GetBool("LOGGER_PRODUCTION"))

	elasticsearchsdk.ReplaceGlobals(elasticsearch.Config{Addresses: viper.GetStringSlice("ELASTICSEARCH_URLS")})

	logicalIndex, err := NewLogicalIndexTimeBased("myrtea", modeler.Model{Name: "myindex", ElasticsearchOptions: modeler.ElasticsearchOptions{
		Rollmode:                  modeler.RollmodeSettings{Type: modeler.RollmodeTimeBased},
		Rollcron:                  "0 * * * *",
		EnablePurge:               true,
		PurgeMaxConcurrentIndices: 30,
		PatchAliasMaxIndices:      0,
	}})
	if err != nil {
		t.Error(err)
	}
	logicalIndex.purge()
	t.Fail()
}
