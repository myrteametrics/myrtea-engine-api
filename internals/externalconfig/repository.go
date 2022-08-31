package externalconfig

import (
	"sync"

	"github.com/jmoiron/sqlx"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/models"
)

// Repository is a storage interface which can be implemented by multiple backend
// (in-memory map, sql database, in-memory cache, file system, ...)
// It allows standard CRUD operation on ExternalConfigs
type Repository interface {
	Get(name string) (models.ExternalConfig, bool, error)
	Create(tx *sqlx.Tx, rootCause models.ExternalConfig) error
	Update(tx *sqlx.Tx, name string, rootCause models.ExternalConfig) error
	Delete(tx *sqlx.Tx, name string) error
	GetAll() (map[string]models.ExternalConfig, error)
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
