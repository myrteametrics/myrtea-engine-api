package action

import (
	"sync"

	"github.com/jmoiron/sqlx"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/model"
)

// Repository is a storage interface which can be implemented by multiple backend
// (in-memory map, sql database, in-memory cache, file system, ...)
// It allows standard CRUD operation on Actions
type Repository interface {
	Get(id int64) (model.Action, bool, error)
	Create(tx *sqlx.Tx, action model.Action) (int64, error)
	Update(tx *sqlx.Tx, id int64, action model.Action) error
	Delete(tx *sqlx.Tx, id int64) error
	GetAll() (map[int64]model.Action, error)
	GetAllByRootCauseID(rootCauseID int64) (map[int64]model.Action, error)
	GetAllBySituationID(situationID int64) (map[int64]model.Action, error)
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
