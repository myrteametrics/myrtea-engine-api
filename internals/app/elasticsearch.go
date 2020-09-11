package app

import (
	"github.com/myrteametrics/myrtea-sdk/v4/elasticsearch"
	"github.com/spf13/viper"
)

func initElasticsearch() {
	urls := viper.GetStringSlice("ELASTICSEARCH_URLS")
	elasticsearch.ReplaceGlobals(&elasticsearch.Credentials{URLs: urls})
}
