package history

import (
	"database/sql"
	"encoding/json"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/reader"
)

type HistoryFactsQuerier struct {
	Builder HistoryFactsBuilder
	conn    *sqlx.DB
}

type HistoryFactsV4 struct {
	ID                  int64
	FactID              int64
	FactName            string
	SituationID         int64
	SituationInstanceID int64
	Ts                  time.Time
	Result              reader.Item
}

func (querier HistoryFactsQuerier) GetHistoryFacts(historyFactsIds []int64) ([]HistoryFactsV4, error) {
	query := querier.Builder.GetHistoryFacts(historyFactsIds)
	return querier.execute(query)
}

func (querier HistoryFactsQuerier) GetHistoryFactsFromSituation(sfq HistorySituationFactsQuerier, historySituationsIds []int64) ([]HistoryFactsV4, []HistorySituationFactsV4, error) {
	historySituationFacts, err := sfq.GetHistorySituationFacts(historySituationsIds)
	if err != nil {
		return nil, nil, err
	}

	historyFactsIds := make([]int64, 0)
	for _, item := range historySituationFacts {
		historyFactsIds = append(historyFactsIds, item.HistoryFactsID)
	}

	query := querier.Builder.GetHistoryFacts(historyFactsIds)
	historyFacts, err := querier.execute(query)
	if err != nil {
		return nil, nil, err
	}

	return historyFacts, historySituationFacts, nil
}

func (querier HistoryFactsQuerier) execute(builder sq.SelectBuilder) ([]HistoryFactsV4, error) {
	rows, err := builder.RunWith(querier.conn.DB).Query()
	if err != nil {
		return make([]HistoryFactsV4, 0), err
	}
	defer rows.Close()
	return querier.scanAll(rows)
}

func (querier HistoryFactsQuerier) scan(rows *sql.Rows) (HistoryFactsV4, error) {
	var rawResult []byte
	item := HistoryFactsV4{}
	err := rows.Scan(&item.ID, &item.FactID, &item.SituationID, &item.SituationInstanceID, &item.Ts, &rawResult, &item.FactName)
	if err != nil {
		return HistoryFactsV4{}, err
	}

	err = json.Unmarshal(rawResult, &item.Result)
	if err != nil {
		return HistoryFactsV4{}, err
	}

	return item, nil
}

func (querier HistoryFactsQuerier) scanAll(rows *sql.Rows) ([]HistoryFactsV4, error) {
	users := make([]HistoryFactsV4, 0)
	for rows.Next() {
		user, err := querier.scan(rows)
		if err != nil {
			return []HistoryFactsV4{}, err
		}
		users = append(users, user)
	}
	return users, nil
}

// func (querier HistoryFactsQuerier) scanFirst(rows *sql.Rows) (HistoryFactsV4, bool, error) {
// 	if rows.Next() {
// 		user, err := querier.scan(rows)
// 		return user, err == nil, err
// 	}
// 	return HistoryFactsV4{}, false, nil
// }

// func (querier HistoryFactsQuerier) checkRowAffected(result sql.Result, nbRows int64) error {
// 	i, err := result.RowsAffected()
// 	if err != nil {
// 		return err
// 	}
// 	if i != nbRows {
// 		return errors.New("no row deleted (or multiple row deleted) instead of 1 row")
// 	}
// 	return nil
// }
