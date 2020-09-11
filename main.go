package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/app"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/router"
	plugin "github.com/myrteametrics/myrtea-engine-api/v4/plugins"
	"github.com/myrteametrics/myrtea-engine-api/v4/plugins/assistant"
	"github.com/myrteametrics/myrtea-engine-api/v4/plugins/baseline"
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
	zapConfig := app.InitLogger()

	zap.L().Info("Starting Engine-API", zap.String("version", Version), zap.String("build_date", BuildDate))
	app.Init()
	defer app.Stop()

	plugins := make([]plugin.MyrteaPlugin, 0)

	baselinePlugin := baseline.NewBaselinePlugin()
	if baselinePlugin != nil {
		defer baselinePlugin.Stop()
		baselinePlugin.Start()
		baselinePlugin.Test()
		baselinePlugin.Test()
		plugins = append(plugins, baselinePlugin)
	}

	assistantPlugin := assistant.NewAssistantPlugin()
	if assistantPlugin != nil {
		defer assistantPlugin.Stop()
		assistantPlugin.Start()
		assistantPlugin.Test()
		assistantPlugin.Test()
		plugins = append(plugins, assistantPlugin)
	}

	serverPort := viper.GetInt("SERVER_PORT")
	serverEnableTLS := viper.GetBool("SERVER_ENABLE_TLS")
	serverTLSCert := viper.GetString("SERVER_TLS_FILE_CRT")
	serverTLSKey := viper.GetString("SERVER_TLS_FILE_KEY")

	apiEnableCORS := viper.GetBool("API_ENABLE_CORS")
	apiEnableSecurity := viper.GetBool("API_ENABLE_SECURITY")
	apiEnableGatewayMode := viper.GetBool("API_ENABLE_GATEWAY_MODE")

	if !apiEnableSecurity {
		zap.L().Info("Warning: API starting in unsecured mode, be sure to set API_UNSECURED=false in production")
	}
	if apiEnableGatewayMode {
		zap.L().Info("Server router will be started using API Gateway mode." +
			"Please ensure every request has been properly pre-verified by the auth-api")
	}

	router := router.NewChiRouter(apiEnableSecurity, apiEnableCORS, apiEnableGatewayMode, zapConfig.Level, plugins)

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
