package users

import (
	"sync"

	"github.com/myrteametrics/myrtea-sdk/v4/security"
)

// Repository is a storage interface which can be implemented by multiple backend
// (in-memory map, sql database, in-memory cache, file system, ...)
// It allows standard CRUD operation on facts
type Repository interface {
	Get(id int64) (security.User, bool, error)
	Create(user security.UserWithPassword) (int64, error)
	Update(user security.User) error
	Delete(id int64) error
	GetAll() (map[int64]security.User, error)
	GetUsersOfGroup(groupID int64) (map[int64]UserOfGroup, error)
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
