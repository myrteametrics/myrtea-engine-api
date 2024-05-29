package app

import (
	"github.com/elastic/go-elasticsearch/v8"
	elasticsearchsdk "github.com/myrteametrics/myrtea-sdk/v4/elasticsearch"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func InitElasticsearch() {
	urls := viper.GetStringSlice("ELASTICSEARCH_URLS")
	err := elasticsearchsdk.ReplaceGlobals(elasticsearch.Config{
		Addresses: urls,
	})

	if err != nil {
		zap.L().Error("Failed to initialize elasticsearch", zap.Error(err))
	}
}
