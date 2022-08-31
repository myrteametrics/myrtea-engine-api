package history

import (
	"sync"
	"time"

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

func (service HistoryService) GetHistorySituationsIdsLast(options GetHistorySituationsOptions) ([]HistorySituationsV4, error) {
	subQuery, subQueryArgs, err := service.HistorySituationsQuerier.Builder.GetHistorySituationsIdsLast(options).ToSql()
	if err != nil {
		return make([]HistorySituationsV4, 0), err
	}
	return service.HistorySituationsQuerier.Query(service.HistorySituationsQuerier.Builder.GetHistorySituationsDetails(subQuery, subQueryArgs))
}

func (service HistoryService) GetHistorySituationsIdsByStandardInterval(options GetHistorySituationsOptions, interval string) ([]HistorySituationsV4, error) {
	subQuery, subQueryArgs, err := service.HistorySituationsQuerier.Builder.GetHistorySituationsIdsByStandardInterval(options, interval).ToSql()
	if err != nil {
		return make([]HistorySituationsV4, 0), err
	}
	return service.HistorySituationsQuerier.Query(service.HistorySituationsQuerier.Builder.GetHistorySituationsDetails(subQuery, subQueryArgs))
}

func (service HistoryService) GetHistorySituationsIdsByCustomInterval(options GetHistorySituationsOptions, interval time.Duration, referenceDate time.Time) ([]HistorySituationsV4, error) {
	subQuery, subQueryArgs, err := service.HistorySituationsQuerier.Builder.GetHistorySituationsIdsByCustomInterval(options, interval, referenceDate).ToSql()
	if err != nil {
		return make([]HistorySituationsV4, 0), err
	}
	return service.HistorySituationsQuerier.Query(service.HistorySituationsQuerier.Builder.GetHistorySituationsDetails(subQuery, subQueryArgs))
}

func (service HistoryService) GetHistoryFactsFromSituation(historySituations []HistorySituationsV4) ([]HistoryFactsV4, []HistorySituationFactsV4, error) {
	historySituationsIds := make([]int64, 0)
	for _, item := range historySituations {
		historySituationsIds = append(historySituationsIds, item.ID)
	}
	return service.GetHistoryFactsFromSituationIds(historySituationsIds)
}

func (service HistoryService) GetHistoryFactsFromSituationIds(historySituationsIds []int64) ([]HistoryFactsV4, []HistorySituationFactsV4, error) {
	historySituationFacts, err := service.HistorySituationFactsQuerier.Query(service.HistorySituationFactsQuerier.Builder.GetHistorySituationFacts(historySituationsIds))
	if err != nil {
		return nil, nil, err
	}

	historyFactsIds := make([]int64, 0)
	for _, item := range historySituationFacts {
		historyFactsIds = append(historyFactsIds, item.HistoryFactID)
	}

	historyFacts, err := service.HistoryFactsQuerier.Query(service.HistoryFactsQuerier.Builder.GetHistoryFacts(historyFactsIds))
	if err != nil {
		return nil, nil, err
	}

	return historyFacts, historySituationFacts, nil
}
