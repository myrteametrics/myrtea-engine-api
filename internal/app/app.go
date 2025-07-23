package app

import (
	"github.com/myrteametrics/myrtea-engine-api/v5/docs"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/security/users"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// Init initialize all the app configuration and components
func Init() {
	docs.SwaggerInfo.Host = viper.GetString("SWAGGER_HOST")
	docs.SwaggerInfo.BasePath = viper.GetString("SWAGGER_BASEPATH")

	initPostgres()
	initRepositories()
	initElasticsearch()
	initServices()

	if viper.GetBool("AUTHENTICATION_CREATE_SUPERUSER") {
		// we want to check if the superuser exists
		zap.L().Info("Trying to create superuser if not exists")
		if err := users.R().CreateSuperUserIfNotExists(); err != nil {
			zap.L().Error("Error creating superuser", zap.Error(err))
		}
	}
}

// Stop clean everything up before stopping the app
func Stop() {
	stopServices()
}
