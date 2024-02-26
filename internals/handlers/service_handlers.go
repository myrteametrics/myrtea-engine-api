package handlers

import (
	"errors"
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
// @Param id query string false "Component to restart"
// @Success 200 {object} ServiceStatus
// @Failure 500 "internal server error"
// @Router /engine/services/{id}/restart [post]
func (sh *ServiceHandler) Restart(w http.ResponseWriter, r *http.Request) {
	s := sh.getComponent(w, r)
	if s == nil {
		return
	}

	// compare LastAction with now and if it's less than 2 minutes, return an error
	if s.GetDefinition().LastAction.Add(2 * time.Minute).After(time.Now()) {
		render.Error(w, r, render.ErrAPITooManyRequests, errors.New("service has been restarted too recently"))
		return
	}

	err := s.Restart()
	if err != nil {
		render.Error(w, r, render.ErrAPIProcessError, err)
		return
	}

	render.OK(w, r)
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
