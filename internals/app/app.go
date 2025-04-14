package app

import (
	"github.com/myrteametrics/myrtea-engine-api/v5/docs"
	"github.com/spf13/viper"
)

// Init initialize all the app configuration and components
func Init() {
	docs.SwaggerInfo.Host = viper.GetString("SWAGGER_HOST")
	docs.SwaggerInfo.BasePath = viper.GetString("SWAGGER_BASEPATH")

	initPostgres()
	initRepositories()
	//initElasticsearch()
	initServices()
}

// Stop cleanup everything before stopping the app
func Stop() {
	stopServices()
}
