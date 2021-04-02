package roles

import (
	"sync"

	uuid "github.com/google/uuid"
)

// Repository is a storage interface which can be implemented by multiple backend
// (in-memory map, sql database, in-memory cache, file system, ...)
// It allows standard CRUD operation on facts
type Repository interface {
	Get(uuid uuid.UUID) (Role, bool, error)
	Create(permission Role) (uuid.UUID, error)
	Update(permission Role) error
	Delete(uuid uuid.UUID) error
	GetAll() ([]Role, error)

	GetAllForUser(userUUID uuid.UUID) ([]Role, error)

	SetRolePermissions(roleUUID uuid.UUID, permissionUUIDs []uuid.UUID) error
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
