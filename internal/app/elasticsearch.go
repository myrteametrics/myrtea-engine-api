package app

import (
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/config/esconfig"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/model"
	elasticsearchsdk "github.com/myrteametrics/myrtea-sdk/v5/elasticsearch"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func initElasticsearch() {
	config, exists, err := esconfig.R().GetDefault()
	replaceWithEnv := true

	if err != nil {
		zap.L().Error("Could not get default ElasticSearch config", zap.Error(err))
	} else if !exists {
		urls := viper.GetStringSlice("ELASTICSEARCH_URLS")
		zap.L().Warn("Default ElasticSearch config does not exists, creating one using ELASTICSEARCH_URLS", zap.Strings("urls", urls))
		config = model.ElasticSearchConfig{
			Name:    "default",
			URLs:    urls,
			Default: true,
		}
		id, err := esconfig.R().Create(config)
		if err != nil {
			zap.L().Error("Could not create default ElasticSearch config", zap.Error(err))
		} else {
			zap.L().Info("Default ElasticSearch config created", zap.Int64("id", id), zap.Strings("urls", urls))
		}
		replaceWithEnv = false
	}

	if replaceWithEnv && len(config.URLs) == 0 {
		config.URLs = viper.GetStringSlice("ELASTICSEARCH_URLS")
		zap.L().Warn("ElasticSearch default config does not contains any urls, using ELASTICSEARCH_URLS", zap.Strings("urls", config.URLs))
	}

	urls := viper.GetStringSlice("ELASTICSEARCH_URLS")
	err = elasticsearchsdk.ReplaceGlobals(elasticsearch.Config{
		Addresses: urls,
	})
	if err != nil {
		zap.L().Error("Could not init elasticsearch", zap.Error(err))

	}
}
