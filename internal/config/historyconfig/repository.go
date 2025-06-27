package historyconfig

import (
	"github.com/spf13/viper"
	"sync"
	"time"
)

// Repository is a storage interface which can be implemented by multiple backends
// (in-memory map, sql database, in-memory cache, file system, ...)
// It allows standard CRUD operations on config history entries
type Repository interface {
	Create(history ConfigHistory) (int64, error)
	Get(id int64) (ConfigHistory, bool, error)
	GetAll() (map[int64]ConfigHistory, error)
	GetAllFromInterval(from time.Time, to time.Time) (map[int64]ConfigHistory, error)
	GetAllByType(historyType string) (map[int64]ConfigHistory, error)
	GetAllByUser(user string) (map[int64]ConfigHistory, error)
	Delete(id int64) error
	DeleteOldest() error
}

var (
	_globalRepositoryMu sync.RWMutex
	_globalRepository   Repository
	maxHistoryRecords   int // Maximum number of history records to keep
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

	maxHistoryRecords = viper.GetInt("MAX_CONFIG_HISTORY_RECORDS")
	return func() { ReplaceGlobals(prev) }
}
