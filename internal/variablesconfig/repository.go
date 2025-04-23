package variablesconfig

import (
	"sync"

	"github.com/myrteametrics/myrtea-engine-api/v5/internal/models"
)

// Repository is a storage interface which can be implemented by multiple backend
// (in-memory map, sql database, in-memory cache, file system, ...)
// It allows standard CRUD operation on VariablesConfig
type Repository interface {
	Get(id int64) (models.VariablesConfig, bool, error)
	GetByKey(key string) (models.VariablesConfig, bool, error)
	Create(variable models.VariablesConfig) (int64, error)
	Update(id int64, variable models.VariablesConfig) error
	Delete(id int64) error
	GetAll() ([]models.VariablesConfig, error)
	GetAllAsMap() (map[string]interface{}, error)
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
