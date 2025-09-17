package template

import (
	sq "github.com/Masterminds/squirrel"
	"sync"
)

// Repository is a storage interface which can be implemented by multiple backends
// (in-memory map, sql database, in-memory cache, file system, ...)
// It allows standard CRUD operations on email templates
type Repository interface {
	Create(template Template) (int64, error)
	Get(id int64) (Template, error)
	GetByName(name string) (Template, error)
	GetAll() ([]Template, error)
	Update(template Template) error
	Delete(id int64) error
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

// newStatement creates a new SQL statement builder with Dollar placeholder format
func newStatement() sq.StatementBuilderType {
	return sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
}
