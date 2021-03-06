package fact

import (
	"sync"

	"github.com/myrteametrics/myrtea-sdk/v4/engine"
)

// Repository is a storage interface which can be implemented by multiple backend
// (in-memory map, sql database, in-memory cache, file system, ...)
// It allows standard CRUD operation on facts
type Repository interface {
	//TODO: Get ride of the pointers for the fact and situations

	Get(id int64) (engine.Fact, bool, error)
	GetByName(name string) (engine.Fact, bool, error)
	Create(fact engine.Fact) (int64, error)
	Update(id int64, fact engine.Fact) error
	Delete(id int64) error
	GetAll() (map[int64]engine.Fact, error)
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
