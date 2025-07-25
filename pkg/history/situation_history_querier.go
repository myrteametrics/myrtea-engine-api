package history

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/calendar"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/metadata"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type HistorySituationsV4 struct {
	ID                    int64
	SituationID           int64
	SituationName         string
	SituationInstanceID   int64
	SituationInstanceName string
	Ts                    time.Time
	Parameters            map[string]interface{}
	ExpressionFacts       map[string]interface{}
	Metadatas             []metadata.MetaData
	Calendar              *calendar.Calendar
}

// HistoryRecordV4 represents a single and unique situation history entry.
type HistoryRecordV4 struct {
	SituationID         int64
	SituationInstanceID int64
	Ts                  time.Time
	HistoryFacts        []HistoryFactsV4
	Parameters          map[string]interface{}
	ExpressionFacts     map[string]interface{}
	EnableDependsOn     bool
	DependsOnParameters map[string]string
}

type ParamGetFactExprHistory struct {
	FactExpr            string `json:"factExpr"`
	SituationID         int64  `json:"situationId"`
	SituationInstanceID int64  `json:"situationInstanceId"`
}
type ParamGetFactExprHistoryByDate struct {
	ParamGetFactExprHistory
	StartDate string `json:"startDate"` // Expected format: "2006-01-02 15:04:05"
	EndDate   string `json:"endDate"`   // Expected format: "2006-01-02 15:04:05"
}

type GetFactExprHistory struct {
	Results []FactExprResult `json:"results"`
}

type FactExprResult struct {
	Value         float64 `json:"value"`
	FormattedTime string  `json:"formattedTime" example:"2006-01-02 15:04:05"`
}

// OverrideParameters overrides the parameters of the History Record.
func (hr HistoryRecordV4) OverrideParameters(p map[string]interface{}) {
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

func (querier HistorySituationsQuerier) Update(history HistorySituationsV4) error {
	parametersJSON, err := json.Marshal(history.Parameters)
	if err != nil {
		return err
	}

	expressionFactsJSON, err := json.Marshal(history.ExpressionFacts)
	if err != nil {
		return err
	}

	metadatasJSON, err := json.Marshal(history.Metadatas)
	if err != nil {
		return err
	}

	err = querier.ExecUpdate(querier.Builder.Update(history.ID, parametersJSON, expressionFactsJSON, metadatasJSON))
	if err != nil {
		return err
	}

	return nil
}

func (querier HistorySituationsQuerier) ExecUpdate(builder sq.UpdateBuilder) error {
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

func (querier HistorySituationsQuerier) ExecDelete(builder sq.DeleteBuilder) error {
	result, err := builder.RunWith(querier.conn.DB).Exec()
	if err != nil {
		return err
	}

	affectedRows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	zap.L().Info("Auto purge of the table Situation_history_v5", zap.Int64("Number of rows deleted", affectedRows))

	return nil
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
	var (
		rawParameters       []byte
		rawExpressionFacts  []byte
		rawMetadatas        []byte
		calendarId          sql.NullInt64
		calendarName        sql.NullString
		calendarDescription sql.NullString
		calendarTimezone    sql.NullString
	)

	item := HistorySituationsV4{}

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

func (querier *HistorySituationsQuerier) QueryGetFieldsTsMetadatas(ctx context.Context, builder sq.SelectBuilder) (HistorySituationsV4, error) {
	rows, err := builder.RunWith(querier.conn).QueryContext(ctx)
	if err != nil {
		return HistorySituationsV4{}, fmt.Errorf("query execution failed: %w", err)
	}
	defer rows.Close()

	var result HistorySituationsV4
	for rows.Next() {
		var metadatas []metadata.MetaData
		var ts time.Time
		var metadataBytes []byte

		err = rows.Scan(&ts, &metadataBytes)
		if err != nil {
			return HistorySituationsV4{}, fmt.Errorf("row scanning failed: %w", err)
		}

		err = json.Unmarshal(metadataBytes, &metadatas)
		if err != nil {
			return HistorySituationsV4{}, fmt.Errorf("Warning: unable to unmarshal metadatas JSON: %v", err)
		}

		if len(metadatas) == 0 {
			return HistorySituationsV4{}, fmt.Errorf("The latest evaluation of the situation is unknown.")
		}

		result = HistorySituationsV4{Metadatas: metadatas, Ts: ts}
		break //LIMIT 1
	}
	if result.Ts.IsZero() {
		return HistorySituationsV4{}, errors.New("no results found")
	}
	return result, nil
}

func (querier *HistorySituationsQuerier) QueryGetSpecificFactExpr(builder sq.SelectBuilder, formatTime string, factExpr string) (GetFactExprHistory, error) {
	rows, err := builder.RunWith(querier.conn).Query()
	if err != nil {
		return GetFactExprHistory{}, err
	}
	defer rows.Close()

	var results []FactExprResult
	for rows.Next() {
		var resultBytes []byte
		var ts time.Time
		err = rows.Scan(&resultBytes, &ts)
		if err != nil {
			return GetFactExprHistory{}, err
		}

		var parsedResult map[string]interface{}
		err = json.Unmarshal(resultBytes, &parsedResult)
		if err != nil {
			return GetFactExprHistory{}, err
		}

		factRes := FactExprResult{FormattedTime: ts.Format(formatTime)}

		if factExprValue, ok := parsedResult[factExpr]; ok {
			if v, ok := factExprValue.(float64); ok {
				factRes.Value = v
			}
		}
		results = append(results, factRes)
	}

	return GetFactExprHistory{Results: results}, nil
}

func (querier HistorySituationsQuerier) GetTodaysFactExprResultByParameters(param ParamGetFactExprHistory) (GetFactExprHistory, error) {
	builder := querier.Builder.GetTodaysFactExprResultByParameters(param)
	return querier.QueryGetSpecificFactExpr(builder, FormatHourMinute, param.FactExpr)
}

func (querier HistorySituationsQuerier) GetFactExprResultByDate(param ParamGetFactExprHistoryByDate) (GetFactExprHistory, error) {
	builder := querier.Builder.GetFactExprResultByDate(param)
	return querier.QueryGetSpecificFactExpr(builder, FormatDateHourMinute, param.FactExpr)
}

func (param ParamGetFactExprHistory) IsValid() error {
	if param.FactExpr == "" {
		return errors.New("Missing or invalide FactID ")
	}
	if param.SituationID <= 0 {
		return errors.New("Missing  or invalide SituationID")
	}

	if param.SituationInstanceID <= 0 {
		return errors.New("Missing or invalie  SituationInstanceID")
	}
	return nil
}

func (p ParamGetFactExprHistoryByDate) IsValid() error {
	if err := p.ParamGetFactExprHistory.IsValid(); err != nil {
		return err
	}
	if _, err := time.Parse("2006-01-02 15:04:05", p.StartDate); err != nil {
		return errors.New("invalid StartDate format")
	}
	if _, err := time.Parse("2006-01-02 15:04:05", p.EndDate); err != nil {
		return errors.New("invalid EndDate format")
	}
	start, _ := time.Parse("2006-01-02 15:04:05", p.StartDate)
	end, _ := time.Parse("2006-01-02 15:04:05", p.EndDate)
	if start.After(end) {
		return errors.New("StartDate must be before EndDate")
	}
	return nil
}

func (querier HistorySituationsQuerier) GetLatestHistory(situationID int64, situationInstanceID int64) (HistorySituationsV4, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	builder := HistorySituationsBuilder{}
	selectBuilder := builder.GetLatestHistorySituation(situationID, situationInstanceID)
	results, err := querier.QueryGetFieldsTsMetadatas(ctx, selectBuilder)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return HistorySituationsV4{}, errors.New("Timeout Error: The request targeting the 'situation_history_v5' table timed out after 10 seconds.")
		}
		return HistorySituationsV4{}, err
	}
	return results, nil
}
