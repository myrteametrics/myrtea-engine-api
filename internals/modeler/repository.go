package modeler

import (
	"sync"

	"github.com/myrteametrics/myrtea-sdk/v4/modeler"
)

// Repository is a storage interface which can be implemented by multiple backend
// (in-memory map, sql database, in-memory cache, file system, ...)
// It allows standard CRUD operation on situations
type Repository interface {
	Get(id int64) (modeler.Model, bool, error)
	GetByName(name string) (modeler.Model, bool, error)
	Create(model modeler.Model) (int64, error)
	Update(id int64, situation modeler.Model) error
	Delete(id int64) error
	GetAll() (map[int64]modeler.Model, error)
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
