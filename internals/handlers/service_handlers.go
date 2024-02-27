package handlers

import (
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/handlers/render"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/service"
	"net/http"
	"time"
)

type ServiceHandler struct {
	Manager *service.Manager
}

func NewServiceHandler(manager *service.Manager) *ServiceHandler {
	return &ServiceHandler{Manager: manager}
}

// GetServices godoc
// @Summary Get all services
// @Description Get all services
// @Tags Services
// @Produce json
// @Security Bearer
// @Success 200 {array} models.ServiceDefinition
// @Failure 500 "internal server error"
// @Router /engine/services [get]
func (sh *ServiceHandler) GetServices(w http.ResponseWriter, r *http.Request) {
	var services []service.Definition

	for _, s := range sh.Manager.GetAll() {
		services = append(services, *s.GetDefinition())
	}

	render.JSON(w, r, services)
}

// Restart godoc
// @Summary Restart the service
// @Description Restart the service
// @Tags Service
// @Produce json
// @Param id query string false "Service to restart"
// @Success 200 "service was restarted successfully"
// @Failure 429 "too recently"
// @Failure 500 "internal server error"
// @Router /engine/services/{id}/restart [post]
func (sh *ServiceHandler) Restart(w http.ResponseWriter, r *http.Request) {
	s := sh.getComponent(w, r)
	if s == nil {
		return
	}

	// compare LastRestart with now and if it's less than 2 minutes, return an error
	if time.Now().Sub(s.GetDefinition().LastRestart) > 2*time.Minute {
		render.Error(w, r, render.ErrAPITooManyRequests, errors.New("service has been restarted too recently"))
		return
	}

	code, err := s.Restart()
	if err != nil || code == 0 {
		render.Error(w, r, render.ErrAPIProcessError, err)
		return
	}

	w.WriteHeader(code)
}

// Reload godoc
// @Summary Reload the service
// @Description Reload the service
// @Tags Service
// @Produce json
// @Param id query string false "Service to reload"
// @Param component query string false "Component to reload"
// @Success 200 {object} ServiceStatus
// @Failure 500 "internal server error"
// @Router /engine/services/{id}/reload/{component} [post]
func (sh *ServiceHandler) Reload(w http.ResponseWriter, r *http.Request) {
	s := sh.getComponent(w, r)
	if s == nil {
		return
	}
	component := chi.URLParam(r, "component")

	if !s.GetDefinition().HasComponent(component) {
		render.Error(w, r, render.ErrAPIDBResourceNotFound, fmt.Errorf("component '%s' not found", component))
		return
	}

	// compare LastReload with now and if it's less than 2 minutes, return an error
	if time.Now().Sub(s.GetDefinition().LastReload) > 2*time.Minute {
		render.Error(w, r, render.ErrAPITooManyRequests, errors.New("service has been reloaded too recently"))
		return
	}

	code, err := s.Reload(component)
	if err != nil || code == 0 {
		render.Error(w, r, render.ErrAPIProcessError, err)
		return
	}

	w.WriteHeader(code)
}

// GetStatus godoc
// @Summary GetStatus of a service
// @Description GetStatus of a service
// @Tags Service
// @Produce json
// @Param id query string false "Component to get status from"
// @Success 200 {object} models.ServiceStatus
// @Failure 500 "internal server error"
// @Router /engine/services/{id}/status [get]
func (sh *ServiceHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	s := sh.getComponent(w, r)
	if s == nil {
		return
	}

	render.JSON(w, r, s.GetStatus())
}

// getComponent returns the component from the manager
func (sh *ServiceHandler) getComponent(w http.ResponseWriter, r *http.Request) service.Service {
	id := chi.URLParam(r, "id")

	uid, err := uuid.Parse(id)
	if err != nil {
		render.Error(w, r, render.ErrAPIParsingInteger, err)
		return nil
	}

	s, ok := sh.Manager.Get(uid)
	if !ok {
		render.Error(w, r, render.ErrAPIDBResourceNotFound, errors.New("service not found"))
		return nil
	}
	return s
}
