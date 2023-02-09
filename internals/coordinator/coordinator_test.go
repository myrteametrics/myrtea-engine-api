package coordinator

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/myrteametrics/myrtea-sdk/v4/configuration"
	"github.com/myrteametrics/myrtea-sdk/v4/elasticsearch"
	"github.com/myrteametrics/myrtea-sdk/v4/modeler"
	"github.com/olivere/elastic"
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
var AllowedConfigKey = []configuration.ConfigKey{
	{Type: configuration.StringFlag, Name: "DEBUG_MODE", DefaultValue: "false", Description: "Enable debug mode"},
	{Type: configuration.StringFlag, Name: "LOGGER_PRODUCTION", DefaultValue: "true", Description: "Enable or disable production log"},
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
	{Type: configuration.StringFlag, Name: "SWAGGER_BASEPATH", DefaultValue: "/api/v5", Description: "Swagger UI target basepath"},
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
	{Type: configuration.StringFlag, Name: "AUTHENTICATION_SAML_ROOT_URL", DefaultValue: "http://localhost:9000/api/v5/", Description: "SAML Root URL"},
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

var (
	liveIndicesCaches []string
	mu                sync.RWMutex
)

// func TestMaster(t *testing.T) {
// 	configuration.InitializeConfig(AllowedConfigKey, ConfigName, ConfigPath, EnvPrefix)
// 	configuration.InitLogger(viper.GetBool("LOGGER_PRODUCTION"))

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

func TestGetIndices(t *testing.T) {
	configuration.InitializeConfig(AllowedConfigKey, ConfigName, ConfigPath, EnvPrefix)
	configuration.InitLogger(viper.GetBool("LOGGER_PRODUCTION"))

	client, err := elastic.NewClient(elastic.SetSniff(false),
		elastic.SetHealthcheckTimeoutStartup(60*time.Second),
		elastic.SetURL(viper.GetStringSlice("ELASTICSEARCH_URLS")...),
		// elastic.SetHttpClient(retryClient.StandardClient()),
	)
	if err != nil {
		zap.L().Error("Elasticsearch client initialization", zap.Error(err))
	} else {
		zap.L().Info("Initialize Elasticsearch client", zap.String("status", "done"))
	}

	indices, err := client.CatIndices().Index("myrtea-myindex-*").Columns("index").Do(context.Background())
	if err != nil {
		t.Error(err)
	}
	for _, index := range indices {
		t.Logf("%+v", index.Index)
	}
	t.Fail()
}

func TestCoordinator(t *testing.T) {

	configuration.InitializeConfig(AllowedConfigKey, ConfigName, ConfigPath, EnvPrefix)
	configuration.InitLogger(viper.GetBool("LOGGER_PRODUCTION"))

	err := InitInstance(
		viper.GetString("INSTANCE_NAME"),
		viper.GetStringSlice("ELASTICSEARCH_URLS"),
		map[int64]modeler.Model{
			1: {
				ID:   1,
				Name: "myindex",
				ElasticsearchOptions: modeler.ElasticsearchOptions{
					Rollmode:                  "timebased",
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

	configuration.InitializeConfig(AllowedConfigKey, ConfigName, ConfigPath, EnvPrefix)
	configuration.InitLogger(viper.GetBool("LOGGER_PRODUCTION"))

	client, err := elastic.NewClient(elastic.SetSniff(false),
		elastic.SetHealthcheckTimeoutStartup(60*time.Second),
		elastic.SetURL(viper.GetStringSlice("ELASTICSEARCH_URLS")...),
		// elastic.SetHttpClient(retryClient.StandardClient()),
	)
	if err != nil {
		zap.L().Error("Elasticsearch client initialization", zap.Error(err))
	} else {
		zap.L().Info("Initialize Elasticsearch client", zap.String("status", "done"))
	}

	logicalIndex, err := NewLogicalIndexTimeBased("myrtea", modeler.Model{Name: "myindex", ElasticsearchOptions: modeler.ElasticsearchOptions{
		Rollmode:                  "timebased",
		Rollcron:                  "0 * * * *",
		EnablePurge:               true,
		PurgeMaxConcurrentIndices: 30,
		PatchAliasMaxIndices:      0,
	}}, &elasticsearch.EsExecutor{Client: client})
	if err != nil {
		t.Error(err)
	}
	logicalIndex.purge()
	t.Fail()
}

func TestFindIndices(t *testing.T) {
	configuration.InitializeConfig(AllowedConfigKey, ConfigName, ConfigPath, EnvPrefix)
	configuration.InitLogger(viper.GetBool("LOGGER_PRODUCTION"))

	client, err := elastic.NewClient(elastic.SetSniff(false),
		elastic.SetHealthcheckTimeoutStartup(60*time.Second),
		elastic.SetURL(viper.GetStringSlice("ELASTICSEARCH_URLS")...),
		// elastic.SetHttpClient(retryClient.StandardClient()),
	)
	if err != nil {
		zap.L().Error("Elasticsearch client initialization", zap.Error(err))
	} else {
		zap.L().Info("Initialize Elasticsearch client", zap.String("status", "done"))
	}

	catIndicesResponse, err := client.CatIndices().Index("myrtea-myindex-*").Columns("index").Do(context.Background())
	if err != nil {
		t.Error(err)
	}

	indices := make([]string, 0)
	for _, index := range catIndicesResponse {
		indices = append(indices, index.Index)
	}
	t.Fail()
	// indices := []string{
	// 	"myrtea-myindex-2022-01-01",
	// 	"myrtea-myindex-2022-12-31",
	// 	"myrtea-myindex-2023-01-01",
	// 	"myrtea-myindex-2023-01-02",
	// 	"myrtea-myindex-2023-01-03",
	// 	"myrtea-myindex-2023-01-14",
	// 	"myrtea-myindex-2023-01-15",
	// 	"myrtea-myindex-2023-01-16",
	// 	"myrtea-myindex-2023-01-28",
	// 	"myrtea-myindex-2023-01-29",
	// 	"myrtea-myindex-2023-01-30",
	// 	"myrtea-myindex-2023-02-01",
	// 	"myrtea-myindex-2023-02-02",
	// 	"myrtea-myindex-2023-02-03",
	// }

	tsEnd := time.Date(2023, 1, 30, 12, 0, 0, 0, time.UTC)
	tsStart := tsEnd.Add(-20 * 24 * time.Hour)

	indexEnd := fmt.Sprintf("myrtea-%s-%s", "myindex", tsEnd.Format("2006-01-02"))
	indexStart := fmt.Sprintf("myrtea-%s-%s", "myindex", tsStart.Format("2006-01-02"))

	t.Log(indexEnd)
	t.Log(indexStart)

	subIndices := make([]string, 0)
	for _, index := range indices {
		if index < indexStart {
			// if index >= indexStart {
			subIndices = append(subIndices, index)
		}
	}
	t.Log(indices)
	t.Log(subIndices)
	t.Fail()

}
