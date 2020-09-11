package groups

import (
	"sync"
)

// Repository is a storage interface which can be implemented by multiple backend
// (in-memory map, sql database, in-memory cache, file system, ...)
// It allows standard CRUD operation on facts
type Repository interface {
	Get(id int64) (Group, bool, error)
	Create(group Group) (int64, error)
	Update(group Group) error
	Delete(id int64) error
	GetAll() (map[int64]Group, error)

	GetMembership(userID int64, groupID int64) (Membership, bool, error)
	CreateMembership(membership Membership) error
	UpdateMembership(membership Membership) error
	DeleteMembership(userID int64, groupID int64) error

	GetGroupsOfUser(userID int64) ([]GroupOfUser, error)
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
