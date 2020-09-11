package app

import (
	"github.com/myrteametrics/myrtea-sdk/v4/postgres"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func initPostgres() {
	credentials := postgres.Credentials{
		URL:      viper.GetString("POSTGRESQL_HOSTNAME"),
		Port:     viper.GetString("POSTGRESQL_PORT"),
		DbName:   viper.GetString("POSTGRESQL_DBNAME"),
		User:     viper.GetString("POSTGRESQL_USERNAME"),
		Password: viper.GetString("POSTGRESQL_PASSWORD"),
	}
	dbClient, err := postgres.DbConnection(credentials)
	if err != nil {
		zap.L().Fatal("main.DbConnection:", zap.Error(err))
	}
	dbClient.SetMaxOpenConns(viper.GetInt("POSTGRESQL_CONN_POOL_MAX_OPEN"))
	dbClient.SetMaxIdleConns(viper.GetInt("POSTGRESQL_CONN_POOL_MAX_IDLE"))
	dbClient.SetConnMaxLifetime(viper.GetDuration("POSTGRESQL_CONN_MAX_LIFETIME"))
	postgres.ReplaceGlobals(dbClient)
}
