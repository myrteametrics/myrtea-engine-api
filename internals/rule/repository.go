package rule

import (
	"sync"
	"time"
)

// Repository is a storage interface which can be implemented by multiple backend
// (in-memory map, sql database, in-memory cache, file system, ...)
// It allows standard CRUD operation on situations
type Repository interface {
	CheckByName(name string) (bool, error)
	Create(rule Rule) (int64, error)
	Get(id int64) (Rule, bool, error)
	GetByName(name string) (Rule, bool, error)
	Update(rule Rule) error
	Delete(id int64) error
	GetAll() (map[int64]Rule, error)
	GetAllEnabled() (map[int64]Rule, error)
	GetAllModifiedFrom(from time.Time) (map[int64]Rule, error)
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
