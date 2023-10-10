package router

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/jwtauth/v5"
	"github.com/lestrrat-go/jwx/jwa"
	_ "github.com/myrteametrics/myrtea-engine-api/v5/docs" // docs is generated by Swag CL
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/handlers"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/handlers/render"
	oidcAuth "github.com/myrteametrics/myrtea-engine-api/v5/internals/router/oidc"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/security"
	plugin "github.com/myrteametrics/myrtea-engine-api/v5/plugins"
	"github.com/myrteametrics/myrtea-sdk/v4/postgres"
	"github.com/spf13/viper"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"
)

// Config wraps common configuration parameters
type Config struct {
	Production         bool
	Security           bool
	CORS               bool
	GatewayMode        bool
	VerboseError       bool
	AuthenticationMode string
	LogLevel           zap.AtomicLevel
	PluginCore         *plugin.Core
}

// Check clean up the configuration and logs comments if required
func (config *Config) Check() {
	if !config.Security {
		zap.L().Warn("API starting in unsecured mode, be sure to set HTTP_SERVER_API_ENABLE_SECURITY=true in production")
	}
	if config.VerboseError {
		zap.L().Warn("API starting in verbose error mode, be sure to set HTTP_SERVER_API_ENABLE_VERBOSE_MODE=false in production")
	}
	if config.GatewayMode {
		zap.L().Warn("Server router will be started using API Gateway mode. " +
			"Please ensure every request has been properly pre-verified by the auth-api")
		if !config.Security {
			zap.L().Warn("Gateway mode has no use if API security is not enabled (HTTP_SERVER_API_ENABLE_SECURITY=false)")
			config.GatewayMode = false
		}
	}
	if config.Security && config.GatewayMode && config.AuthenticationMode == "SAML" {
		zap.L().Warn("SAML Authentication mode is not compatible with HTTP_SERVER_API_ENABLE_GATEWAY_MODE=true")
		config.GatewayMode = false
	}
	if config.Security && config.GatewayMode && config.AuthenticationMode == "OIDC" {
		zap.L().Warn("OIDC Authentication mode is not compatible with HTTP_SERVER_API_ENABLE_GATEWAY_MODE=true")
		config.GatewayMode = false
	}
	if config.AuthenticationMode != "BASIC" && config.AuthenticationMode != "SAML" && config.AuthenticationMode != "OIDC" {
		zap.L().Warn("Authentication mode not supported. Back to default value 'BASIC'", zap.String("AuthenticationMode", config.AuthenticationMode))
		config.AuthenticationMode = "BASIC"
	}
}

// New returns a new fully configured instance of chi.Mux
// It instanciates all middlewares including the security ones, all routes and route groups
func New(config Config) *chi.Mux {

	config.Check()

	r := chi.NewRouter()
	// Global middleware stack
	// TODO: Add CORS middleware
	if config.CORS {
		cors := cors.New(cors.Options{
			AllowedOrigins:   []string{"*"},
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
			ExposedHeaders:   []string{"Link", "Authenticate-To", "Content-Disposition"},
			AllowCredentials: true,
			MaxAge:           300, // Maximum value not ignored by any of major browsers
		})
		r.Use(cors.Handler)
	}

	r.Use(chimiddleware.SetHeader("Strict-Transport-Security", "max-age=63072000; includeSubDomains"))
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.StripSlashes)
	r.Use(chimiddleware.RedirectSlashes)
	if config.Production {
		r.Use(CustomZapLogger)
	} else {
		r.Use(CustomLogger)
	}
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.Timeout(60 * time.Second))

	var routes func(r chi.Router)
	var err error

	switch config.AuthenticationMode {
	case "BASIC":
		routes, err = buildRoutesV3Basic(config)
	case "SAML":
		routes, err = buildRoutesV3SAML(config)
	case "OIDC":
		routes, err = buildRoutesV3OIDC(config)
	default:
		zap.L().Panic("Authentication mode not supported", zap.String("AuthenticationMode", config.AuthenticationMode))
		return nil
	}
	if err != nil {
		zap.L().Panic("Cannot initialize API routes", zap.String("AuthenticationMode", config.AuthenticationMode), zap.Error(err))
	}

	r.Route("/api/v4", routes)
	r.Route("/api/v5", routes)

	return r
}

func buildRoutesV3Basic(config Config) (func(r chi.Router), error) {
	signingKey := []byte(security.RandString(128))
	securityMiddleware := security.NewMiddlewareJWT(signingKey, security.NewDatabaseAuth(postgres.DB()))

	return func(r chi.Router) {

		// Public routes
		r.Group(func(rg chi.Router) {
			rg.Get("/isalive", handlers.IsAlive)
			rg.Get("/swagger/*", httpSwagger.WrapHandler)

			rg.Post("/login", securityMiddleware.GetToken())
			r.Get("/authmode", handlers.GetAuthenticationMode)
		})

		// Protected routes
		r.Group(func(rg chi.Router) {
			if config.Security {
				if config.GatewayMode {
					// Warning: No signature verification will be done on JWT.
					// JWT MUST have been verified before by the API Gateway
					rg.Use(UnverifiedAuthenticator)
				} else {
					rg.Use(func(next http.Handler) http.Handler {
						// jwtauth.Verifier function only verifies header & cookie but not query
						// websocket requests only handles uri parameters, so jwtauth.TokenFromHeader
						// is needed here.
						return jwtauth.Verify(jwtauth.New(jwa.HS256.String(), signingKey, nil),
							jwtauth.TokenFromQuery, jwtauth.TokenFromHeader, jwtauth.TokenFromCookie)(next)
					})
					rg.Use(CustomAuthenticator)
				}
				rg.Use(ContextMiddleware)
			}
			rg.Use(chimiddleware.SetHeader("Content-Type", "application/json"))

			rg.HandleFunc("/log_level", config.LogLevel.ServeHTTP)
			rg.Mount("/engine", engineRouter())

			for _, plugin := range config.PluginCore.Plugins {
				rg.Mount(plugin.Plugin.HandlerPrefix(), plugin.Plugin.Handler())
				rg.HandleFunc(fmt.Sprintf("/plugin%s", plugin.Plugin.HandlerPrefix()), func(w http.ResponseWriter, r *http.Request) {
					render.JSON(w, r, map[string]interface{}{"loaded": true})
				})
				rg.HandleFunc(fmt.Sprintf("/plugin%s/*", plugin.Plugin.HandlerPrefix()), ReverseProxy(plugin.Plugin))
			}

		})

		// Admin Protection routes
		r.Group(func(rg chi.Router) {
			if config.Security {
				if config.GatewayMode {
					// Warning: No signature verification will be done on JWT.
					// JWT MUST have been verified before by the API Gateway
					rg.Use(UnverifiedAuthenticator)
				} else {
					rg.Use(jwtauth.Verifier(jwtauth.New(jwa.HS256.String(), signingKey, nil)))
					rg.Use(CustomAuthenticator)
				}
				// rg.Use(security.AdminAuthentificator)
				rg.Use(ContextMiddleware)
			}
			rg.Use(chimiddleware.SetHeader("Content-Type", "application/json"))

			rg.Mount("/admin", adminRouter())
		})

		// System intra services Protection routes
		r.Group(func(rg chi.Router) {
			//TODO: change to be intra APIs
			// if config.Security {
			// 	rg.Use(jwtauth.Verifier(jwtauth.New(jwa.HS256.String(), signingKey, nil)))
			// 	rg.Use(CustomAuthenticator)
			// 	rg.Use(ContextMiddleware)
			// }
			// rg.Use(chimiddleware.SetHeader("Content-Type", "application/json"))

			rg.Mount("/service", serviceRouter())
		})
	}, nil
}

func buildRoutesV3SAML(config Config) (func(r chi.Router), error) {

	samlConfig := SamlSPMiddlewareConfig{
		MetadataMode:             viper.GetString("AUTHENTICATION_SAML_METADATA_MODE"),
		MetadataFilePath:         viper.GetString("AUTHENTICATION_SAML_METADATA_FILE_PATH"),
		MetadataURL:              viper.GetString("AUTHENTICATION_SAML_METADATA_URL"),
		EnableMemberOfValidation: viper.GetBool("AUTHENTICATION_SAML_ENABLE_MEMBEROF_VALIDATION"),
		AttributeUserID:          viper.GetString("AUTHENTICATION_SAML_ATTRIBUTE_USER_ID"),
		AttributeUserDisplayName: viper.GetString("AUTHENTICATION_SAML_ATTRIBUTE_USER_DISPLAYNAME"),
		AttributeUserMemberOf:    viper.GetString("AUTHENTICATION_SAML_ATTRIBUTE_USER_MEMBEROF"),
		CookieMaxAge:             viper.GetDuration("AUTHENTICATION_SAML_COOKIE_MAX_AGE_DURATION"),
	}
	if ok, err := samlConfig.IsValid(); !ok {
		return nil, err
	}

	spRootURLStr := viper.GetString("AUTHENTICATION_SAML_ROOT_URL")
	entityID := viper.GetString("AUTHENTICATION_SAML_ENTITYID")
	samlKeyFile := viper.GetString("AUTHENTICATION_SAML_KEY_FILE_PATH")
	samlCrtFile := viper.GetString("AUTHENTICATION_SAML_CRT_FILE_PATH")
	samlSPMiddleware, err := NewSamlSP(spRootURLStr, entityID, samlKeyFile, samlCrtFile, samlConfig)
	if err != nil {
		return nil, err
	}

	return func(r chi.Router) {

		// Public routes
		r.Group(func(rg chi.Router) {
			rg.Get("/isalive", handlers.IsAlive)
			rg.Get("/swagger/*", httpSwagger.WrapHandler)
			rg.Handle("/saml/*", samlSPMiddleware)
			rg.Handle("/logout", handlers.LogoutHandler(samlSPMiddleware.Deconnexion))
			r.Get("/authmode", handlers.GetAuthenticationMode)
		})

		// Protected routes
		r.Group(func(rg chi.Router) {
			if config.Security {
				rg.Use(samlSPMiddleware.RequireAccount)
				rg.Use(samlSPMiddleware.ContextMiddleware)
			}
			rg.Use(chimiddleware.SetHeader("Content-Type", "application/json"))

			rg.HandleFunc("/log_level", config.LogLevel.ServeHTTP)
			rg.Mount("/engine", engineRouter())

			for _, plugin := range config.PluginCore.Plugins {
				rg.Mount(plugin.Plugin.HandlerPrefix(), plugin.Plugin.Handler())
				rg.HandleFunc(fmt.Sprintf("/plugin%s", plugin.Plugin.HandlerPrefix()), func(w http.ResponseWriter, r *http.Request) {
					render.JSON(w, r, map[string]interface{}{"loaded": true})
				})
				rg.HandleFunc(fmt.Sprintf("/plugin%s/*", plugin.Plugin.HandlerPrefix()), ReverseProxy(plugin.Plugin))
			}

		})

		// Admin Protection routes
		r.Group(func(rg chi.Router) {
			if config.Security {
				rg.Use(samlSPMiddleware.RequireAccount)
				rg.Use(samlSPMiddleware.ContextMiddleware)
				// rg.Use(samlSPMiddleware.AdminAuthentificator)
			}
			rg.Use(chimiddleware.SetHeader("Content-Type", "application/json"))

			rg.Mount("/admin", adminRouter())
		})

		// System intra services Protection routes
		r.Group(func(rg chi.Router) {
			//TODO: change to be intra APIs
			// if config.Security {
			// 	rg.Use(samlSPMiddleware.RequireAccount)
			// 	rg.Use(samlSPMiddleware.ContextMiddleware)
			// 	rg.Use(security.AdminAuthentificator)
			// }
			rg.Use(chimiddleware.SetHeader("Content-Type", "application/json"))

			rg.Mount("/service", serviceRouter())
		})
	}, nil
}

// ReverseProxy act as a reverse proxy for any plugin http handlers
func ReverseProxy(plugin plugin.MyrteaPlugin) http.HandlerFunc {
	url, _ := url.Parse(fmt.Sprintf("http://localhost:%d", plugin.ServicePort()))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		httputil.NewSingleHostReverseProxy(url).ServeHTTP(w, r)
	})
}

func buildRoutesV3OIDC(config Config) (func(r chi.Router), error) {

	return func(r chi.Router) {
		// Public routes
		r.Group(func(rg chi.Router) {
			rg.Get("/isalive", handlers.IsAlive)
			rg.Get("/swagger/*", httpSwagger.WrapHandler)

			// Endpoints OIDC
			rg.Get("/auth/oidc", handlers.HandleOIDCRedirect)
			rg.Get("/auth/oidc/callback", handlers.HandleOIDCCallback)
			r.Get("/authmode", handlers.GetAuthenticationMode)
		})

		// Protected routes
		r.Group(func(rg chi.Router) {
			if config.Security {
				// Middleware OIDC
				rg.Use(oidcAuth.OIDCMiddleware)
				rg.Use(oidcAuth.ContextMiddleware)

			}
			rg.Use(chimiddleware.SetHeader("Content-Type", "application/json"))

			rg.HandleFunc("/log_level", config.LogLevel.ServeHTTP)
			rg.Mount("/engine", engineRouter())

			for _, plugin := range config.Plugins {
				rg.Mount(plugin.HandlerPrefix(), plugin.Handler())
				rg.HandleFunc(fmt.Sprintf("/plugin%s", plugin.HandlerPrefix()), func(w http.ResponseWriter, r *http.Request) {
					render.JSON(w, r, map[string]interface{}{"loaded": true})
				})
				rg.HandleFunc(fmt.Sprintf("/plugin%s/*", plugin.HandlerPrefix()), ReverseProxy(plugin))
			}
		})

		// Admin Protection routes
		r.Group(func(rg chi.Router) {
			if config.Security {
				rg.Use(oidcAuth.OIDCMiddleware)
				rg.Use(oidcAuth.ContextMiddleware)
			}
			rg.Use(chimiddleware.SetHeader("Content-Type", "application/json"))

			rg.Mount("/admin", adminRouter())
		})

		// System intra services Protection routes
		r.Group(func(rg chi.Router) {
			rg.Use(chimiddleware.SetHeader("Content-Type", "application/json"))

			rg.Mount("/service", serviceRouter())
		})
	}, nil
}
