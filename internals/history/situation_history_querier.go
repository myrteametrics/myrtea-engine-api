package history

import (
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/calendar"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/models"
	"go.uber.org/zap"
)

type HistorySituationsV4 struct {
	ID                    int64
	SituationID           int64
	SituationName         string
	SituationInstanceID   int64
	SituationInstanceName string
	Ts                    time.Time
	Parameters            map[string]string
	ExpressionFacts       map[string]interface{}
	Metadatas             []models.MetaData
	Calendar              *calendar.Calendar
}

// HistoryRecordV4 represents a single and unique situation history entry
type HistoryRecordV4 struct {
	SituationID         int64
	SituationInstanceID int64
	Ts                  time.Time
	HistoryFacts        []HistoryFactsV4
	Parameters          map[string]string
	ExpressionFacts     map[string]interface{}
}

// OverrideParameters overrides the parameters of the History Record
func (hr HistoryRecordV4) OverrideParameters(p map[string]string) {
	for key, value := range p {
		hr.Parameters[key] = value
	}
}

type HistorySituationsQuerier struct {
	Builder HistorySituationsBuilder
	conn    *sqlx.DB
}

func (querier HistorySituationsQuerier) Insert(history HistorySituationsV4) (int64, error) {
	parametersJSON, err := json.Marshal(history.Parameters)
	if err != nil {
		return -1, err
	}

	expressionFactsJSON, err := json.Marshal(history.ExpressionFacts)
	if err != nil {
		return -1, err
	}

	metadatasJSON, err := json.Marshal(history.Metadatas)
	if err != nil {
		return -1, err
	}

	id, err := querier.QueryReturning(querier.Builder.Insert(history, parametersJSON, expressionFactsJSON, metadatasJSON))
	if err != nil {
		return -1, err
	}
	return id, nil
}

func (querier HistorySituationsQuerier) QueryReturning(builder sq.InsertBuilder) (int64, error) {
	rows, err := builder.RunWith(querier.conn.DB).Query()
	if err != nil {
		return -1, err
	}
	defer rows.Close()
	return querier.scanID(rows)
}

func (querier HistorySituationsQuerier) Query(builder sq.SelectBuilder) ([]HistorySituationsV4, error) {
	rows, err := builder.RunWith(querier.conn.DB).Query()
	if err != nil {
		return make([]HistorySituationsV4, 0), err
	}
	defer rows.Close()
	return querier.scanAll(rows)
}

func (querier HistorySituationsQuerier) QueryIDs(builder sq.SelectBuilder) ([]int64, error) {
	rows, err := builder.RunWith(querier.conn.DB).Query()
	if err != nil {
		return make([]int64, 0), err
	}
	defer rows.Close()
	return querier.scanAllIDs(rows)
}

func (querier HistorySituationsQuerier) scanAllIDs(rows *sql.Rows) ([]int64, error) {
	ids := make([]int64, 0)
	for rows.Next() {
		var id int64
		err := rows.Scan(&id)
		if err != nil {
			return []int64{}, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (querier HistorySituationsQuerier) scanID(rows *sql.Rows) (int64, error) {
	var id int64
	if rows.Next() {
		rows.Scan(&id)
	} else {
		return -1, errors.New("no id returned")
	}
	return id, nil
}

func (querier HistorySituationsQuerier) scan(rows *sql.Rows) (HistorySituationsV4, error) {
	var rawParameters []byte
	var rawExpressionFacts []byte
	var rawMetadatas []byte
	item := HistorySituationsV4{}
	//
	var calendarId sql.NullInt64
	var calendarName sql.NullString
	var calendarDescription sql.NullString
	var calendarTimezone sql.NullString

	err := rows.Scan(&item.ID, &item.SituationID, &item.SituationInstanceID, &item.Ts, &rawParameters,
		&rawExpressionFacts, &rawMetadatas, &item.SituationName, &item.SituationInstanceName,
		&calendarId, &calendarName, &calendarDescription, &calendarTimezone)
	if err != nil {
		return HistorySituationsV4{}, err
	}

	if len(rawParameters) > 0 {
		err = json.Unmarshal(rawParameters, &item.Parameters)
		if err != nil {
			zap.L().Error("Unmarshal", zap.Error(err))
			return HistorySituationsV4{}, err
		}
	}

	if len(rawExpressionFacts) > 0 {
		err = json.Unmarshal(rawExpressionFacts, &item.ExpressionFacts)
		if err != nil {
			zap.L().Error("Unmarshal", zap.Error(err))
			return HistorySituationsV4{}, err
		}
	}

	if len(rawMetadatas) > 0 {
		err = json.Unmarshal(rawMetadatas, &item.Metadatas)
		if err != nil {
			zap.L().Error("Unmarshal", zap.Error(err))
			return HistorySituationsV4{}, err
		}
	}

	if calendarId.Valid && calendarName.Valid && calendarDescription.Valid && calendarTimezone.Valid {
		item.Calendar = &calendar.Calendar{
			ID:          calendarId.Int64,
			Name:        calendarName.String,
			Description: calendarDescription.String,
			Timezone:    calendarTimezone.String,
		}
	}

	return item, nil
}

func (querier HistorySituationsQuerier) scanAll(rows *sql.Rows) ([]HistorySituationsV4, error) {
	users := make([]HistorySituationsV4, 0)
	for rows.Next() {
		user, err := querier.scan(rows)
		if err != nil {
			return []HistorySituationsV4{}, err
		}
		users = append(users, user)
	}
	return users, nil
}

// func (querier HistorySituationsQuerier) scanFirst(rows *sql.Rows) (HistorySituationsV4, bool, error) {
// 	if rows.Next() {
// 		user, err := querier.scan(rows)
// 		return user, err == nil, err
// 	}
// 	return HistorySituationsV4{}, false, nil
// }

// func (querier HistorySituationsQuerier) checkRowAffected(result sql.Result, nbRows int64) error {
// 	i, err := result.RowsAffected()
// 	if err != nil {
// 		return err
// 	}
// 	if i != nbRows {
// 		return errors.New("no row deleted (or multiple row deleted) instead of 1 row")
// 	}
// 	return nil
// }
