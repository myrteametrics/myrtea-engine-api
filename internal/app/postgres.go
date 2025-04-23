package app

import (
	"github.com/myrteametrics/myrtea-engine-api/v5/migrations"
	"github.com/myrteametrics/myrtea-sdk/v5/postgres"
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

	zap.L().Info("Postgres connection initialized",
		zap.String("host", credentials.URL),
		zap.String("port", credentials.Port),
		zap.String("dbname", credentials.DbName),
		zap.String("user", credentials.User),
	)

	// Migrate the database
	if viper.GetBool("POSTGRESQL_MIGRATION_ON_STARTUP") {
		zap.L().Info("Migrating database")
		if err := migrations.Migrate(dbClient.DB); err != nil {
			zap.L().Fatal("Error migrating database", zap.Error(err))
		}
	} else {
		zap.L().Info("Skipping database migration")
	}

}
