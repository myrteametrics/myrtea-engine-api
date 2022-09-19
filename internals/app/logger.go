package app

import (
	"log"

	"go.uber.org/zap"
)

// InitLogger initialialize zap logging component
func InitLogger(production bool) zap.Config {
	var zapConfig zap.Config
	if production {
		zapConfig = zap.NewProductionConfig()
	} else {
		zapConfig = zap.NewDevelopmentConfig()
	}
	zapConfig.Level.SetLevel(zap.InfoLevel)
	logger, err := zapConfig.Build(zap.AddStacktrace(zap.ErrorLevel))
	if err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
	}
	defer logger.Sync()
	zap.ReplaceGlobals(logger)
	return zapConfig
}
