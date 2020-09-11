package search

import (
	"sync"
	"time"

	"github.com/myrteametrics/myrtea-engine-api/v4/internals/situation"
)

// Repository is a storage interface which can be implemented by multiple backend
// (in-memory map, sql database, in-memory cache, file system, ...)
// It allows standard CRUD operation on situations
type Repository interface {
	GetSituationHistoryRecords(s situation.Situation, templateInstanceID int64, t time.Time, start time.Time, end time.Time,
		factSource interface{}, factExpressionsSource interface{}, metaDataSource interface{}, parametersSource interface{}, downSampling DownSampling) (QueryResult, error)
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
