package apikey

import (
	"github.com/google/uuid"
	"sync"
)

type Repository interface {
	Get(uuid uuid.UUID) (APIKey, bool, error)
	Create(apiKey APIKey) (CreateResponse, error)
	Update(apiKey APIKey) error
	Delete(uuid uuid.UUID) error
	GetAll() ([]APIKey, error)
	GetAllForRole(roleUUID uuid.UUID) ([]APIKey, error)
	Validate(keyValue string) (APIKey, bool, error)
	Deactivate(uuid uuid.UUID) error
}

var (
	_globalRepositoryMu sync.RWMutex
	_globalRepository   Repository
)

// R is used to access the global repository singleton
func R() Repository {
	_globalRepositoryMu.RLock()
	repository := _globalRepository
	_globalRepositoryMu.RUnlock()
	return repository
}

// ReplaceGlobals affect a new repository to the global repository singleton
func ReplaceGlobals(repository Repository) func() {
	_globalRepositoryMu.Lock()
	prev := _globalRepository
	_globalRepository = repository
	_globalRepositoryMu.Unlock()
	return func() { ReplaceGlobals(prev) }
}
