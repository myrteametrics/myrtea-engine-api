package service

import "github.com/myrteametrics/myrtea-engine-api/v5/internals/models"

type Manager struct {
	services map[string]models.Service
}

func NewManager() *Manager {
	return &Manager{
		services: make(map[string]models.Service),
	}
}

// Register adds a service to the manager
func (m Manager) Register(name string, service models.Service) {
	m.services[name] = service
}

// Get returns a service by its name
func (m Manager) Get(name string) (models.Service, bool) {
	service, ok := m.services[name]
	return service, ok
}
