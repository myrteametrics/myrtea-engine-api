package action

import (
	"sync"

	"github.com/jmoiron/sqlx"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/models"
)

// Repository is a storage interface which can be implemented by multiple backend
// (in-memory map, sql database, in-memory cache, file system, ...)
// It allows standard CRUD operation on Actions
type Repository interface {
	Get(id int64) (models.Action, bool, error)
	Create(tx *sqlx.Tx, action models.Action) (int64, error)
	Update(tx *sqlx.Tx, id int64, action models.Action) error
	Delete(tx *sqlx.Tx, id int64) error
	GetAll() (map[int64]models.Action, error)
	GetAllByRootCauseID(rootCauseID int64) (map[int64]models.Action, error)
	GetAllBySituationID(situationID int64) (map[int64]models.Action, error)
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
