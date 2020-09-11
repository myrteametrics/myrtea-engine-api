package notification

import (
	"sync"

	"github.com/myrteametrics/myrtea-engine-api/v4/internals/dbutils"
)

// Repository is a storage interface which can be implemented by multiple backend
// (in-memory map, sql database, in-memory cache, file system, ...)
// It allows standard CRUD operation on facts
type Repository interface {
	Create(groups []int64, notif Notification) (int64, error)
	Get(id int64) *FrontNotification
	GetByGroups(groupIds []int64, queryOptionnal dbutils.DBQueryOptionnal) ([]FrontNotification, error)
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
	repository := _globalRepository
	_globalRepositoryMu.RUnlock()
	return repository
}

// ReplaceGlobals affects a new repository to the global repository singleton
func ReplaceGlobals(repository Repository) func() {
	_globalRepositoryMu.Lock()
	prev := _globalRepository
	_globalRepository = repository
	_globalRepositoryMu.Unlock()
	return func() { ReplaceGlobals(prev) }
}
