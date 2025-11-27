package app

import (
	"crypto/tls"
	"net/http"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/config/esconfig"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/model"
	elasticsearchsdk "github.com/myrteametrics/myrtea-sdk/v5/elasticsearch"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func initElasticsearch() {
	config, exists, err := esconfig.R().GetDefaultForAuth() // Use ForAuth to get cleartext password
	replaceWithEnv := true

	if err != nil {
		zap.L().Error("Could not get default ElasticSearch config", zap.Error(err))
	} else if !exists {
		urls := viper.GetStringSlice("ELASTICSEARCH_URLS")
		auth := viper.GetBool("ELASTICSEARCH_AUTH")
		insecure := viper.GetBool("ELASTICSEARCH_INSECURE")
		username := viper.GetString("ELASTICSEARCH_USERNAME")
		password := viper.GetString("ELASTICSEARCH_PASSWORD")

		zap.L().Warn("Default ElasticSearch config does not exists, creating one using environment variables",
			zap.Strings("urls", urls), zap.Bool("auth", auth), zap.Bool("insecure", insecure))
		config = model.ElasticSearchConfig{
			Name:     "default",
			URLs:     urls,
			Default:  true,
			Auth:     auth,
			Insecure: insecure,
			Username: username,
			Password: password,
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

	// Build elasticsearch client config
	esClientConfig := elasticsearch.Config{
		Addresses: config.URLs,
	}

	// Apply authentication if enabled (password is now cleartext from GetDefaultForAuth)
	if config.Auth {
		esClientConfig.Username = config.Username
		esClientConfig.Password = config.Password // Now cleartext for authentication
		zap.L().Info("ElasticSearch authentication enabled", zap.String("username", config.Username))
	}

	// Apply insecure TLS if enabled
	if config.Insecure {
		esClientConfig.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		zap.L().Warn("ElasticSearch TLS verification disabled (insecure mode)")
	}

	zap.L().Info("Initializing ElasticSearch client", zap.Strings("urls", config.URLs),
		zap.Bool("auth", config.Auth), zap.Bool("insecure", config.Insecure))
	err = elasticsearchsdk.ReplaceGlobals(esClientConfig)
	if err != nil {
		zap.L().Error("Could not init elasticsearch", zap.Error(err), zap.Strings("urls", config.URLs))
	}
}
