package draft

import (
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/models"
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
	DeleteOldIssueResolutionsDrafts(ts time.Time) error
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
