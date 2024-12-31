package history

import (
	"sync"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/explainer/draft"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/explainer/issues"
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

// R is used to access the global service singleton.
func S() HistoryService {
	_globalHistoryServiceMu.RLock()
	defer _globalHistoryServiceMu.RUnlock()

	service := _globalHistoryService

	return service
}

// ReplaceGlobals affect a new service to the global service singleton.
func ReplaceGlobals(service HistoryService) func() {
	_globalHistoryServiceMu.Lock()
	defer _globalHistoryServiceMu.Unlock()

	prev := _globalHistoryService
	_globalHistoryService = service

	return func() { ReplaceGlobals(prev) }
}

func (service HistoryService) GetHistorySituationsIdsLast(options GetHistorySituationsOptions) ([]HistorySituationsV4, error) {
	subQuery, subQueryArgs, err := service.HistorySituationsQuerier.Builder.
		GetHistorySituationsIdsLast(options).
		ToSql()
	if err != nil {
		return make([]HistorySituationsV4, 0), err
	}

	return service.HistorySituationsQuerier.Query(
		service.HistorySituationsQuerier.Builder.GetHistorySituationsDetails(subQuery, subQueryArgs),
	)
}

func (service HistoryService) GetHistorySituationsIdsByStandardInterval(options GetHistorySituationsOptions, interval string) ([]HistorySituationsV4, error) {
	subQuery, subQueryArgs, err := service.HistorySituationsQuerier.Builder.
		GetHistorySituationsIdsByStandardInterval(options, interval).
		ToSql()
	if err != nil {
		return make([]HistorySituationsV4, 0), err
	}

	return service.HistorySituationsQuerier.Query(
		service.HistorySituationsQuerier.Builder.GetHistorySituationsDetails(subQuery, subQueryArgs),
	)
}

func (service HistoryService) GetAllHistorySituationsIdsByStandardInterval(options GetHistorySituationsOptions, interval string) ([]HistorySituationsV4, error) {
	subQuery := service.HistorySituationsQuerier.Builder.
		GetAllHistorySituationsIdsByStandardInterval(options, interval)

	_, _, err := subQuery.ToSql()
	if err != nil {
		return nil, err
	}

	query := service.HistorySituationsQuerier.Builder.GetAllHistorySituationsDetails(subQuery)
	// TODO remove
	//queryString, interfacevalue, errr := query.ToSql()
	////show values
	//zap.L().Info("f", zap.String("queryString", queryString), zap.Any("interfacevalue", interfacevalue), zap.Error(errr))

	return service.HistorySituationsQuerier.Query(
		query,
	)
}

func (service HistoryService) GetHistorySituationsIdsByCustomInterval(options GetHistorySituationsOptions, interval time.Duration, referenceDate time.Time) ([]HistorySituationsV4, error) {
	subQuery, subQueryArgs, err := service.HistorySituationsQuerier.Builder.
		GetHistorySituationsIdsByCustomInterval(options, interval, referenceDate).
		ToSql()
	if err != nil {
		return make([]HistorySituationsV4, 0), err
	}

	return service.HistorySituationsQuerier.Query(
		service.HistorySituationsQuerier.Builder.GetHistorySituationsDetails(subQuery, subQueryArgs),
	)
}

func (service HistoryService) GetHistoryFactsFromSituation(historySituations []HistorySituationsV4) ([]HistoryFactsV4, []HistorySituationFactsV4, error) {
	historySituationsIds := make([]int64, 0)
	for _, item := range historySituations {
		historySituationsIds = append(historySituationsIds, item.ID)
	}

	return service.GetHistoryFactsFromSituationIds(historySituationsIds)
}

func (service HistoryService) GetHistoryFactsFromSituationIds(historySituationsIds []int64) ([]HistoryFactsV4, []HistorySituationFactsV4, error) {
	historySituationFacts, err := service.HistorySituationFactsQuerier.Query(
		service.HistorySituationFactsQuerier.Builder.GetHistorySituationFacts(historySituationsIds),
	)
	if err != nil {
		return nil, nil, err
	}

	historyFactsIds := make([]int64, 0)
	for _, item := range historySituationFacts {
		historyFactsIds = append(historyFactsIds, item.HistoryFactID)
	}

	historyFacts, err := service.HistoryFactsQuerier.Query(
		service.HistoryFactsQuerier.Builder.GetHistoryFacts(historyFactsIds),
	)
	if err != nil {
		return nil, nil, err
	}

	return historyFacts, historySituationFacts, nil
}

func (service HistoryService) PurgeHistory(options GetHistorySituationsOptions) error {
	return service.deleteHistoryPurge(
		service.HistorySituationsQuerier.Builder.GetHistorySituationsIdsBase(options), options,
	)
}

func (service HistoryService) CompactHistory(options GetHistorySituationsOptions, interval string) error {
	return service.deleteHistory(
		service.HistorySituationsQuerier.Builder.GetHistorySituationsIdsByStandardInterval(options, interval),
	)
}

func (service HistoryService) deleteHistory(selector sq.SelectBuilder) error {
	err := service.HistorySituationFactsQuerier.ExecDelete(
		service.HistorySituationFactsQuerier.Builder.DeleteHistoryFrom(selector),
	)
	if err != nil {
		return err
	}

	err = service.HistorySituationsQuerier.ExecDelete(
		service.HistorySituationsQuerier.Builder.DeleteOrphans(),
	)
	if err != nil {
		return err
	}

	err = service.HistoryFactsQuerier.ExecDelete(
		service.HistoryFactsQuerier.Builder.DeleteOrphans(),
	)
	if err != nil {
		return err
	}

	return nil
}
func (service HistoryService) deleteHistoryPurge(selector sq.SelectBuilder, options GetHistorySituationsOptions) error {
	err := service.HistorySituationFactsQuerier.ExecDelete(
		service.HistorySituationFactsQuerier.Builder.DeleteHistoryFrom(selector),
	)
	if err != nil {
		return err
	}

	err = issues.R().DeleteOldIssueDetections(options.DeleteBeforeTs)
	if err != nil {
		return err
	}

	err = issues.R().DeleteOldIssueResolutions(options.DeleteBeforeTs)
	if err != nil {
		return err
	}

	err = draft.R().DeleteOldIssueResolutionsDrafts(options.DeleteBeforeTs)
	if err != nil {
		return err
	}

	err = issues.R().DeleteOldIssues(options.DeleteBeforeTs)
	if err != nil {
		return err
	}

	err = service.HistorySituationsQuerier.ExecDelete(
		service.HistorySituationsQuerier.Builder.DeleteOrphans(),
	)
	if err != nil {
		return err
	}

	err = service.HistoryFactsQuerier.ExecDelete(
		service.HistoryFactsQuerier.Builder.DeleteOrphans(),
	)
	if err != nil {
		return err
	}

	return nil
}
