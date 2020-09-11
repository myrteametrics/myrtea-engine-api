package draft

import (
	"sync"

	"github.com/jmoiron/sqlx"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/models"
)

// Repository is a storage interface which can be implemented by multiple backend
// (in-memory map, sql database, in-memory cache, file system, ...)
// It allows standard CRUD operation on RootCauses
type Repository interface {
	Get(issueID int64) (models.FrontDraft, bool, error)
	Create(tx *sqlx.Tx, issueID int64, draft models.FrontDraft) error
	Update(tx *sqlx.Tx, issueID int64, draft models.FrontDraft) error
	// Delete(tx *sqlx.Tx, id int64) error)
	CheckExists(tx *sqlx.Tx, issueID int64) (bool, error)
	CheckExistsWithUUID(tx *sqlx.Tx, issueID int64, uuid string) (bool, error)
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
