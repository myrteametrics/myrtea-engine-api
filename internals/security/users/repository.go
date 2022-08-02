package users

import (
	"sync"

	uuid "github.com/google/uuid"
)

// Repository is a storage interface which can be implemented by multiple backend
// (in-memory map, sql database, in-memory cache, file system, ...)
// It allows standard CRUD operation on facts
type Repository interface {
	Get(uuid uuid.UUID) (User, bool, error)
	Create(user UserWithPassword) (uuid.UUID, error)
	Update(user User) error
	UpdateWithPassword(user UserWithPassword) error
	Delete(uuid uuid.UUID) error
	GetAll() ([]User, error)

	GetAllForRole(roleUUID uuid.UUID) ([]User, error)

	SetUserRoles(userUUID uuid.UUID, roleUUIDs []uuid.UUID) error
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
