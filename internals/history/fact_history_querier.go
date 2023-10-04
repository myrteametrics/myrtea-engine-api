package history

import (
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/reader"
	"go.uber.org/zap"
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

type GetFactHistory struct {
	Results []FactResult `json:"results"`
}

type FactResult struct {
	Value int64  `json:"value"`
	Time  time.Duration  `json:"time"`
}

type ParamGetFactHistory struct {
	FactID              int64 `json:"factID"`
	SituationID         int64 `json:"situationId"`
	SituationInstanceID int64 `json:"situationInstanceId"`
}

func (querier HistoryFactsQuerier) Insert(history HistoryFactsV4) (int64, error) {
	resultJSON, err := json.Marshal(history.Result)
	if err != nil {
		return -1, err
	}

	id, err := querier.QueryReturning(querier.Builder.Insert(history, resultJSON))
	if err != nil {
		return -1, err
	}

	return id, nil
}

func (querier HistoryFactsQuerier) Exec(builder sq.InsertBuilder) error {
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

func (querier HistoryFactsQuerier) Update(history HistoryFactsV4) error {
	resultJSON, err := json.Marshal(history.Result)
	if err != nil {
		return err
	}

	err = querier.ExecUpdate(querier.Builder.Update(history.ID, resultJSON))
	if err != nil {
		return err
	}

	return nil
}

func (querier HistoryFactsQuerier) ExecUpdate(builder sq.UpdateBuilder) error {
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

func (querier HistoryFactsQuerier) ExecDelete(builder sq.DeleteBuilder) error {
	result, err := builder.RunWith(querier.conn.DB).Exec()
	if err != nil {
		return err
	}

	affectedRows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	zap.L().Info("Auto purge of the table fact_history_v5", zap.Int64("Number of rows deleted", affectedRows))

	return nil
}


func (querier HistoryFactsQuerier) QueryReturning(builder sq.InsertBuilder) (int64, error) {
	rows, err := builder.RunWith(querier.conn.DB).Query()
	if err != nil {
		return -1, err
	}
	defer rows.Close()

	return querier.scanID(rows)
}

func (querier HistoryFactsQuerier) QueryOne(builder sq.SelectBuilder) (HistoryFactsV4, error) {
	rows, err := builder.RunWith(querier.conn.DB).Query()
	if err != nil {
		return HistoryFactsV4{}, err
	}
	defer rows.Close()

	return querier.scanFirst(rows)
}

func (querier HistoryFactsQuerier) Query(builder sq.SelectBuilder) ([]HistoryFactsV4, error) {
	rows, err := builder.RunWith(querier.conn.DB).Query()
	if err != nil {
		return make([]HistoryFactsV4, 0), err
	}
	defer rows.Close()

	return querier.scanAll(rows)
}

func (querier HistoryFactsQuerier) scanID(rows *sql.Rows) (int64, error) {
	var id int64
	if rows.Next() {
		rows.Scan(&id)
	} else {
		return -1, errors.New("no id returned")
	}

	return id, nil
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

func (querier HistoryFactsQuerier) scanFirst(rows *sql.Rows) (HistoryFactsV4, error) {
	if rows.Next() {
		return querier.scan(rows)
	}

	return HistoryFactsV4{}, nil
}

func (querier *HistoryFactsQuerier) QueryGetSpecificFields(builder sq.SelectBuilder) (GetFactHistory, error) {
    rows, err := builder.RunWith(querier.conn).Query()
    if err != nil {
        return GetFactHistory{}, err
    }
    defer rows.Close()

    var results []FactResult

    for rows.Next() {
        var resultBytes []byte
        var ts time.Time
        err = rows.Scan(&resultBytes, &ts)
        if err != nil {
            return GetFactHistory{}, err
        }

        var parsedResult map[string]map[string]map[string]int64
        err = json.Unmarshal(resultBytes, &parsedResult)
        if err != nil {
            return GetFactHistory{}, err
        }

        duration := time.Duration(ts.Hour())*time.Hour + time.Duration(ts.Minute())*time.Minute + time.Duration(ts.Second())*time.Second
        factRes := FactResult{Time: duration}

		if aggs, ok := parsedResult["aggs"]; ok {
			for key := range aggs {
				if strings.Contains(key, "count") {
					if count, ok := aggs[key]; ok {
						if value, ok := count["value"]; ok {
							factRes.Value = value
							break
						}
					}
				}
			}
		}
		results = append(results, factRes)
    }

    return GetFactHistory{Results: results}, nil
}


func (querier HistoryFactsQuerier) GetTodaysFactResultByParameters(param ParamGetFactHistory) (GetFactHistory, error) {
	builder := querier.Builder.GetTodaysFactResultByParameters(param)
	return querier.QueryGetSpecificFields(builder)
}

func (querier HistoryFactsQuerier) Delete(ID int64) error {
	error := querier.Builder.Delete(ID)
	return querier.ExecDelete(error)
}

func (param ParamGetFactHistory) IsValid() (bool, error) {
	if param.FactID == 0 {
		return false, errors.New("Missing FactID ")
	}

	if param.SituationID == 0 {
		return false, errors.New("Missing SituationID")
	}

	if param.SituationInstanceID == 0 {
		return false, errors.New("Missing SituationInstanceID")
	}
	return true, nil
}

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
