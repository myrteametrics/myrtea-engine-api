package scheduler

import (
	"sync"
)

// Repository is a storage interface which can be implemented by multiple backend
// (in-memory map, sql database, in-memory cache, file system, ...)
// It allows standard CRUD operation on situations
type Repository interface {
	Create(schedule InternalSchedule) (int64, error)
	Get(id int64) (InternalSchedule, bool, error)
	Update(schedule InternalSchedule) error
	Delete(id int64) error
	GetAll() (map[int64]InternalSchedule, error)
	refreshNextIdGen() (int64, bool, error)
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

// ReplaceGlobalRepository affect a new repository to the global repository singleton
func ReplaceGlobalRepository(repository Repository) func() {
	_globalRepositoryMu.Lock()
	defer _globalRepositoryMu.Unlock()

	prev := _globalRepository
	_globalRepository = repository
	return func() { ReplaceGlobalRepository(prev) }
}
