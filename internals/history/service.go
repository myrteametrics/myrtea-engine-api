package history

import (
	"sync"

	"github.com/jmoiron/sqlx"
)

var (
	_globalHistoryServiceMu sync.RWMutex
	_globalHistoryService   HistoryService
)

type HistoryService struct {
	HistorySituationsQuerier     HistorySituationsQuerier
	HistorySituationFactsQuerier HistorySituationFactsQuerier
	HistoryFactsQuerier          HistoryFactsQuerier
}

func New(db *sqlx.DB) HistoryService {
	return HistoryService{
		HistorySituationsQuerier:     HistorySituationsQuerier{conn: db, Builder: HistorySituationsBuilder{}},
		HistorySituationFactsQuerier: HistorySituationFactsQuerier{conn: db, Builder: HistorySituationFactsBuilder{}},
		HistoryFactsQuerier:          HistoryFactsQuerier{conn: db, Builder: HistoryFactsBuilder{}},
	}
}

// R is used to access the global service singleton
func S() HistoryService {
	_globalHistoryServiceMu.RLock()
	defer _globalHistoryServiceMu.RUnlock()

	service := _globalHistoryService
	return service
}

// ReplaceGlobals affect a new service to the global service singleton
func ReplaceGlobals(service HistoryService) func() {
	_globalHistoryServiceMu.Lock()
	defer _globalHistoryServiceMu.Unlock()

	prev := _globalHistoryService
	_globalHistoryService = service
	return func() { ReplaceGlobals(prev) }
}
