package handlers

import (
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/security/permissions"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/service"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/utils/httputil"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type ServiceHandler struct {
	Manager        *service.Manager
	restartTimeout time.Duration
	reloadTimeout  time.Duration
}

func NewServiceHandler(manager *service.Manager) *ServiceHandler {
	return &ServiceHandler{Manager: manager, restartTimeout: 5 * time.Minute, reloadTimeout: 2 * time.Minute}
}

// GetServices godoc
//
//	@Summary		Get all services
//	@Description	Get all services
//	@Tags			Services
//	@Produce		json
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{array}	service.Definition
//	@Failure		401	"missing permission"
//	@Failure		500	"internal server error"
//	@Router			/engine/services [get]
func (sh *ServiceHandler) GetServices(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeService, permissions.All, permissions.ActionList)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	services := make([]service.Definition, 0)

	for _, s := range sh.Manager.GetAll() {
		services = append(services, *s.GetDefinition())
	}

	httputil.JSON(w, r, services)
}

// Restart godoc
//
//	@Summary		Restart the service
//	@Description	Restart the service
//	@Tags			Service
//	@Produce		json
//	@Param			id	query	string	false	"Service to restart"
//	@Success		200	"service was restarted successfully"
//	@Failure		401	"missing permission"
//	@Failure		429	"too recently"
//	@Failure		500	"internal server error"
//	@Router			/engine/services/{id}/restart [post]
func (sh *ServiceHandler) Restart(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeService, permissions.All, "restart")) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	s := sh.getComponent(w, r)
	if s == nil {
		return
	}

	def := s.GetDefinition()

	// compare LastRestart with now and if it's less than restartTimeout duration, return an error
	if !def.LastRestart.IsZero() && time.Now().Sub(def.LastRestart) <= sh.restartTimeout {
		zap.L().Info("restart too early", zap.String("service", def.Name))
		httputil.Error(w, r, httputil.ErrAPITooManyRequests, errors.New("service has been restarted too recently"))
		return
	}

	// set timeout before calling restart to avoid multiple restarts one after each-other
	def.LastRestart = time.Now()

	code, err := s.Restart()
	if err != nil {
		// if an error occurs, set time to 0
		def.LastRestart = time.Time{}
		zap.L().Error("error happened during restart", zap.Error(err), zap.String("service", s.GetDefinition().Name))
		httputil.Error(w, r, httputil.ErrAPIProcessError, err)
		return
	}

	httputil.JSON(w, r, *s.GetDefinition())
	w.WriteHeader(code)
}

// Reload godoc
//
//	@Summary		Reload the service
//	@Description	Reload the service
//	@Tags			Service
//	@Produce		json
//	@Param			id			query	string	false	"Service to reload"
//	@Param			component	query	string	false	"Component to reload"
//	@Success		200			"component reloaded"
//	@Failure		401			"missing permission"
//	@Failure		429			"too recently"
//	@Failure		500			"internal server error"
//	@Router			/engine/services/{id}/reload/{component} [post]
func (sh *ServiceHandler) Reload(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeService, permissions.All, "reload")) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	s := sh.getComponent(w, r)
	if s == nil {
		return
	}
	component := chi.URLParam(r, "component")

	if !s.GetDefinition().HasComponent(component) {
		zap.L().Warn("Component not found", zap.String("component", component))
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFound, fmt.Errorf("component '%s' not found", component))
		return
	}

	// compare LastReload with now and if it's less than reloadTimeout duration, return an error
	if time.Now().Sub(s.GetDefinition().LastReload) <= sh.reloadTimeout {
		httputil.Error(w, r, httputil.ErrAPITooManyRequests, errors.New("service has been reloaded too recently"))
		return
	}

	// set timeout before calling reload to avoid multiple reloads one after each-other
	def := s.GetDefinition()
	def.LastReload = time.Now()

	code, err := s.Reload(component)
	if err != nil || code == 0 {
		// if an error occurs, set time to 0
		def.LastReload = time.Time{}
		zap.L().Error("error happened during reload", zap.Error(err), zap.String("service", s.GetDefinition().Name))
		httputil.Error(w, r, httputil.ErrAPIProcessError, err)
		return
	}

	httputil.JSON(w, r, *s.GetDefinition())
	w.WriteHeader(code)
}

// GetStatus godoc
//
//	@Summary		GetStatus of a service
//	@Description	GetStatus of a service
//	@Tags			Service
//	@Produce		json
//	@Param			id	query		string	false	"Component to get status from"
//	@Success		200	{object}	service.Status
//	@Failure		401	"missing permission"
//	@Failure		500	"internal server error"
//	@Router			/engine/services/{id}/status [get]
func (sh *ServiceHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeService, permissions.All, permissions.ActionGet)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	s := sh.getComponent(w, r)
	if s == nil {
		return
	}

	httputil.JSON(w, r, s.GetStatus())
}

// getComponent returns the component from the manager
func (sh *ServiceHandler) getComponent(w http.ResponseWriter, r *http.Request) service.Service {
	id := chi.URLParam(r, "id")

	uid, err := uuid.Parse(id)
	if err != nil {
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return nil
	}

	s, ok := sh.Manager.Get(uid)
	if !ok {
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFound, errors.New("service not found"))
		return nil
	}
	return s
}
