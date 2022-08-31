package rootcause

import (
	"sync"

	"github.com/jmoiron/sqlx"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/models"
)

// Repository is a storage interface which can be implemented by multiple backend
// (in-memory map, sql database, in-memory cache, file system, ...)
// It allows standard CRUD operation on RootCauses
type Repository interface {
	Get(id int64) (models.RootCause, bool, error)
	Create(tx *sqlx.Tx, rootCause models.RootCause) (int64, error)
	Update(tx *sqlx.Tx, id int64, rootCause models.RootCause) error
	Delete(tx *sqlx.Tx, id int64) error
	GetAll() (map[int64]models.RootCause, error)
	GetAllBySituationID(situationID int64) (map[int64]models.RootCause, error)
	GetAllBySituationIDRuleID(situationID int64, ruleID int64) (map[int64]models.RootCause, error)
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
