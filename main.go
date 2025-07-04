package main

import (
	"context"
	"errors"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/export"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/handler"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/metrics"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/service"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/plugins"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/app"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/router"
	"github.com/myrteametrics/myrtea-sdk/v5/helpers"
	"github.com/myrteametrics/myrtea-sdk/v5/server"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var (
	// Version is the binary version (tag) + build number (CI pipeline)
	Version string
	// BuildDate is the date of build
	BuildDate string
)

//	@version		1.0
//	@title			Myrtea Engine-API
//	@description	Myrtea Engine-API Swagger
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	Myrtea Metrics
//	@contact.url	https://myrteametrics.ai/en/
//	@contact.email	contact@myrteametrics.com

//	@securityDefinitions.apikey	Bearer
//	@in							header
//	@name						Authorization

// @securityDefinitions.apikey	ApiKeyAuth
// @in							header
// @name						X-API-Key
// @description				Authentication using the X-API-Key header
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
	}

	// Exports
	directDownload := viper.GetBool("EXPORT_DIRECT_DOWNLOAD")
	indirectDownloadUrl := viper.GetString("EXPORT_INDIRECT_DOWNLOAD_URL")

	exportWrapper := export.NewWrapper(
		viper.GetString("EXPORT_BASE_PATH"),        // basePath
		viper.GetInt("EXPORT_WORKERS_COUNT"),       // workersCount
		viper.GetInt("EXPORT_DISK_RETENTION_DAYS"), // diskRetentionDays
		viper.GetInt("EXPORT_QUEUE_MAX_SIZE"),      // queueMaxSize
	)
	exportWrapper.Init(context.Background())

	// Init services
	serviceManager := service.NewManager()

	if err := serviceManager.LoadConnectors(); err != nil {
		zap.L().Error("Error loading service connectors", zap.Error(err))
	}

	if err := serviceManager.LoadPlugins(core); err != nil {
		zap.L().Error("Error loading service plugins", zap.Error(err))
	}

	// Init router services struct (used to inject services into the router)
	routerServices := router.Services{
		PluginCore:       core,
		ProcessorHandler: handler.NewProcessorHandler(),
		ExportHandler:    handler.NewExportHandler(exportWrapper, directDownload, indirectDownloadUrl),
		ServiceHandler:   handler.NewServiceHandler(serviceManager),
		ApiKeyHandler:    handler.NewApiKeyHandler(viper.GetDuration("API_KEY_CACHE_DURATION")),
	}

	mux := router.New(routerConfig, routerServices)
	var srv *http.Server
	if serverEnableTLS {
		srv = server.NewSecuredServer(serverPort, serverTLSCert, serverTLSKey, mux)
	} else {
		srv = server.NewUnsecuredServer(serverPort, mux)
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
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
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
