package externalconfig

import (
	"sync"

	"github.com/myrteametrics/myrtea-engine-api/v5/internals/models"
)

// Repository is a storage interface which can be implemented by multiple backend
// (in-memory map, sql database, in-memory cache, file system, ...)
// It allows standard CRUD operation on ExternalConfigs
type Repository interface {
	Get(id int64) (models.ExternalConfig, bool, error)
	GetByName(name string) (models.ExternalConfig, bool, error)
	Create(config models.ExternalConfig) (int64, error)
	Update(id int64, config models.ExternalConfig) error
	Delete(id int64) error
	GetAll() (map[int64]models.ExternalConfig, error)
	GetAllOldVersions(id int64) ([]models.ExternalConfig, error)
}

var (
	_globalRepositoryMu sync.RWMutex
	_globalRepository   Repository
)

// R is used to access the global repository singleton
func R() Repository {
	_globalRepositoryMu.RLock()
	defer _globalRepositoryMu.RUnlock()

	repository := _globalRepository
	return repository
}

// ReplaceGlobals affect a new repository to the global repository singleton
func ReplaceGlobals(repository Repository) func() {
	_globalRepositoryMu.Lock()
	defer _globalRepositoryMu.Unlock()

	prev := _globalRepository
	_globalRepository = repository
	return func() { ReplaceGlobals(prev) }
}
