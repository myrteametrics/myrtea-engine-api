package search

import (
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/situation"
	"sync"
	"time"
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
