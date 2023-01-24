package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/app"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/router"
	plugin "github.com/myrteametrics/myrtea-engine-api/v5/plugins"
	"github.com/myrteametrics/myrtea-engine-api/v5/plugins/assistant"
	"github.com/myrteametrics/myrtea-engine-api/v5/plugins/baseline"
	"github.com/myrteametrics/myrtea-sdk/v4/configuration"
	"github.com/myrteametrics/myrtea-sdk/v4/server"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var (
	// Version is the binary version (tag) + build number (CI pipeline)
	Version string
	// BuildDate is the date of build
	BuildDate string
)

// @version 1.0
// @description Myrtea Engine-API Swagger
// @termsOfService http://swagger.io/terms/

// @contact.name Myrtea Metrics
// @contact.url https://myrteametrics.ai/en/
// @contact.email contact@myrteametrics.com

// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization

func main() {

	app.InitConfiguration()
	zapConfig := configuration.InitLogger(viper.GetBool("LOGGER_PRODUCTION"))

	app.Init()
	defer app.Stop()
	zap.L().Info("Starting Engine-API", zap.String("version", Version), zap.String("build_date", BuildDate))

	plugins := make([]plugin.MyrteaPlugin, 0)

	if baselinePlugin := baseline.NewBaselinePlugin(); baselinePlugin != nil {
		defer baselinePlugin.Stop()
		if err := baselinePlugin.Start(); err != nil {
			zap.L().Error("Start baselinePlugin", zap.Error(err))
		} else {
			plugins = append(plugins, baselinePlugin)
		}
	}

	if assistantPlugin := assistant.NewAssistantPlugin(); assistantPlugin != nil {
		defer assistantPlugin.Stop()
		if err := assistantPlugin.Start(); err == nil {
			zap.L().Error("Start assistantPlugin", zap.Error(err))
		} else {
			plugins = append(plugins, assistantPlugin)
		}
	}

	serverPort := viper.GetInt("SERVER_PORT")
	serverEnableTLS := viper.GetBool("SERVER_ENABLE_TLS")
	serverTLSCert := viper.GetString("SERVER_TLS_FILE_CRT")
	serverTLSKey := viper.GetString("SERVER_TLS_FILE_KEY")

	routerConfig := router.Config{
		Production:         viper.GetBool("LOGGER_PRODUCTION"),
		CORS:               viper.GetBool("API_ENABLE_CORS"),
		Security:           viper.GetBool("API_ENABLE_SECURITY"),
		GatewayMode:        viper.GetBool("API_ENABLE_GATEWAY_MODE"),
		AuthenticationMode: viper.GetString("AUTHENTICATION_MODE"),
		LogLevel:           zapConfig.Level,
		Plugins:            plugins,
	}

	router := router.New(routerConfig)
	var srv *http.Server
	if serverEnableTLS {
		srv = server.NewSecuredServer(serverPort, serverTLSCert, serverTLSKey, router)
	} else {
		srv = server.NewUnsecuredServer(serverPort, router)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		var err error
		if serverEnableTLS {
			err = srv.ListenAndServeTLS(serverTLSCert, serverTLSKey)
		} else {
			err = srv.ListenAndServe()
		}
		if err != nil && err != http.ErrServerClosed {
			zap.L().Fatal("Server listen", zap.Error(err))
		}
	}()
	zap.L().Info("Server Started", zap.String("addr", srv.Addr))

	<-done

	ctxShutDown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		cancel()
	}()

	if err := srv.Shutdown(ctxShutDown); err != nil {
		zap.L().Fatal("Server shutdown failed", zap.Error(err))
	}
	zap.L().Info("Server shutdown")
}
