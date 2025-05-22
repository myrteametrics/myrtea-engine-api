package calendar

import (
	"sync"
	"time"
)

// Repository is a storage interface which can be implemented by multiple backend
// (in-memory map, sql database, in-memory cache, file system, ...)
// It allows standard CRUD operation on calendars
type Repository interface {
	Get(id int64) (Calendar, bool, error)
	Create(calendar Calendar) (int64, error)
	Update(calendar Calendar) error
	Delete(id int64) error
	GetAll() (map[int64]Calendar, error)
	GetAllModifiedFrom(from time.Time) (map[int64]Calendar, error)
	GetSituationCalendar(id int64) (Calendar, bool, error)
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
