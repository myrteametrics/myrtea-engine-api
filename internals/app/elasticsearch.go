package app

import (
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/myrteametrics/myrtea-sdk/v4/elasticsearchv6"
	"github.com/myrteametrics/myrtea-sdk/v4/elasticsearchv8"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func InitElasticsearch() {
	version := viper.GetInt("ELASTICSEARCH_VERSION")
	urls := viper.GetStringSlice("ELASTICSEARCH_URLS")

	switch version {
	case 6:
		elasticsearchv6.ReplaceGlobals(&elasticsearchv6.Credentials{
			URLs: urls,
		})
	case 7:
		fallthrough
	case 8:
		elasticsearchv8.ReplaceGlobals(elasticsearch.Config{
			Addresses: urls,
		})
	default:
		zap.L().Fatal("Unsupported Elasticsearch version", zap.Int("version", version))
	}
}
