package notification

import (
	"sync"

	"github.com/myrteametrics/myrtea-engine-api/v5/internals/dbutils"
)

// Repository is a storage interface which can be implemented by multiple backend
// (in-memory map, sql database, in-memory cache, file system, ...)
// It allows standard CRUD operation on facts
type Repository interface {
	Create(notif Notification) (int64, error)
	Get(id int64) (Notification, error)
	GetAll(queryOptionnal dbutils.DBQueryOptionnal) ([]Notification, error)
	Delete(id int64) error
	UpdateRead(id int64, state bool) error
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

// ReplaceGlobals affects a new repository to the global repository singleton
func ReplaceGlobals(repository Repository) func() {
	_globalRepositoryMu.Lock()
	defer _globalRepositoryMu.Unlock()

	prev := _globalRepository
	_globalRepository = repository
	return func() { ReplaceGlobals(prev) }
}
