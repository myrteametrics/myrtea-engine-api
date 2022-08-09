package history

import (
	"database/sql"
	"errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

type HistorySituationFactsV4 struct {
	HistorySituationID int64
	HistoryFactID      int64
	FactID             int64
}

type HistorySituationFactsQuerier struct {
	Builder HistorySituationFactsBuilder
	conn    *sqlx.DB
}

func (querier HistorySituationFactsQuerier) GetHistorySituationFacts(historySituationsIds []int64) ([]HistorySituationFactsV4, error) {
	query := querier.Builder.GetHistorySituationFacts(historySituationsIds)
	return querier.Query(query)
}

func (querier HistorySituationFactsQuerier) Execute(builder sq.InsertBuilder) error {
	res, err := builder.RunWith(querier.conn.DB).Exec()
	if err != nil {
		return err
	}

	if count, err := res.RowsAffected(); err != nil {
		return err
	} else if count == 0 {
		return errors.New("no rows inserted")
	}
	return nil
}

func (querier HistorySituationFactsQuerier) Query(builder sq.SelectBuilder) ([]HistorySituationFactsV4, error) {
	rows, err := builder.RunWith(querier.conn.DB).Query()
	if err != nil {
		return make([]HistorySituationFactsV4, 0), err
	}
	defer rows.Close()
	return querier.scanAll(rows)
}

func (querier HistorySituationFactsQuerier) scan(rows *sql.Rows) (HistorySituationFactsV4, error) {
	item := HistorySituationFactsV4{}
	err := rows.Scan(&item.HistorySituationID, &item.HistoryFactID, &item.FactID)
	if err != nil {
		return HistorySituationFactsV4{}, errors.New("couldn't scan the retrieved data: " + err.Error())
	}
	return item, nil
}

func (querier HistorySituationFactsQuerier) scanAll(rows *sql.Rows) ([]HistorySituationFactsV4, error) {
	users := make([]HistorySituationFactsV4, 0)
	for rows.Next() {
		user, err := querier.scan(rows)
		if err != nil {
			return []HistorySituationFactsV4{}, err
		}
		users = append(users, user)
	}
	return users, nil
}

// func (querier HistorySituationFactsQuerier) scanFirst(rows *sql.Rows) (HistorySituationFactsV4, bool, error) {
// 	if rows.Next() {
// 		user, err := querier.scan(rows)
// 		return user, err == nil, err
// 	}
// 	return HistorySituationFactsV4{}, false, nil
// }

// func (querier HistorySituationFactsQuerier) checkRowAffected(result sql.Result, nbRows int64) error {
// 	i, err := result.RowsAffected()
// 	if err != nil {
// 		return err
// 	}
// 	if i != nbRows {
// 		return errors.New("no row deleted (or multiple row deleted) instead of 1 row")
// 	}
// 	return nil
// }
