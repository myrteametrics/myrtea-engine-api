package issues

import (
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/groups"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/models"
)

// Repository is a storage interface which can be implemented by multiple backend
// (in-memory map, sql database, in-memory cache, file system, ...)
// It allows standard CRUD operation on Issues
type Repository interface {
	Get(id int64, groups []int64) (models.Issue, bool, error)
	Create(issue models.Issue) (int64, error)
	Update(tx *sqlx.Tx, id int64, issue models.Issue, user groups.UserWithGroups) error
	GetByStates(issueStates []string, groups []int64) (map[int64]models.Issue, error)
	GetByStateByPage(issuesStates []string, options models.SearchOptions, groups []int64) ([]models.Issue, int, error)
	GetCloseToTimeoutByKey(key string, firstSituationTS time.Time) (map[int64]models.Issue, error)
	GetAll(groups []int64) (map[int64]models.Issue, error)
	ChangeState(key string, fromStates []models.IssueState, toState models.IssueState) error
	ChangeStateBetweenDates(key string, fromStates []models.IssueState, toState models.IssueState, from time.Time, to time.Time) error
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
