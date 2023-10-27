package main

import (
	"context"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/metrics"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/app"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/router"
	plugin "github.com/myrteametrics/myrtea-engine-api/v5/plugins"
	"github.com/myrteametrics/myrtea-sdk/v4/helpers"
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

	hostname, _ := os.Hostname()
	metrics.InitMetricLabels(hostname)

	app.InitConfiguration()
	zapConfig := helpers.InitLogger(viper.GetBool("LOGGER_PRODUCTION"))

	app.Init()
	defer app.Stop()
	zap.L().Info("Starting Engine-API", zap.String("version", Version), zap.String("build_date", BuildDate))

	// Starting plugin core
	core := &plugin.Core{}
	core.RegisterPlugins()
	core.Start()
	defer core.Stop()

	serverPort := viper.GetInt("HTTP_SERVER_PORT")
	serverEnableTLS := viper.GetBool("HTTP_SERVER_ENABLE_TLS")
	serverTLSCert := viper.GetString("HTTP_SERVER_TLS_FILE_CRT")
	serverTLSKey := viper.GetString("HTTP_SERVER_TLS_FILE_KEY")

	routerConfig := router.Config{
		Production:         viper.GetBool("LOGGER_PRODUCTION"),
		CORS:               viper.GetBool("HTTP_SERVER_API_ENABLE_CORS"),
		Security:           viper.GetBool("HTTP_SERVER_API_ENABLE_SECURITY"),
		GatewayMode:        viper.GetBool("HTTP_SERVER_API_ENABLE_GATEWAY_MODE"),
		AuthenticationMode: viper.GetString("AUTHENTICATION_MODE"),
		LogLevel:           zapConfig.Level,
		PluginCore:         core,
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
