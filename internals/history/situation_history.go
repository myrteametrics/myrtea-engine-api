package history

import (
	"database/sql"
	"encoding/json"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/models"
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
	Metadatas             []models.MetaData
}

type HistorySituationsQuerier struct {
	Builder HistorySituationsBuilder
	conn    *sqlx.DB
}

func (querier HistorySituationsQuerier) GetHistorySituationsIdsLast(options GetHistorySituationsOptions) ([]HistorySituationsV4, error) {
	subQuery, subQueryArgs, err := querier.Builder.GetHistorySituationsIdsLast(options).ToSql()
	if err != nil {
		return make([]HistorySituationsV4, 0), err
	}

	query := querier.Builder.GetHistorySituationsDetails(subQuery, subQueryArgs)
	return querier.execute(query)
}

func (querier HistorySituationsQuerier) GetHistorySituationsIdsByStandardInterval(options GetHistorySituationsOptions, interval string) ([]HistorySituationsV4, error) {
	subQuery, subQueryArgs, err := querier.Builder.GetHistorySituationsIdsByStandardInterval(options, interval).ToSql()
	if err != nil {
		return make([]HistorySituationsV4, 0), err
	}

	query := querier.Builder.GetHistorySituationsDetails(subQuery, subQueryArgs)
	return querier.execute(query)
}

func (querier HistorySituationsQuerier) GetHistorySituationsIdsByCustomInterval(options GetHistorySituationsOptions, referenceDate time.Time, interval time.Duration) ([]HistorySituationsV4, error) {
	subQuery, subQueryArgs, err := querier.Builder.GetHistorySituationsIdsByCustomInterval(options, referenceDate, interval).ToSql()
	if err != nil {
		return make([]HistorySituationsV4, 0), err
	}

	query := querier.Builder.GetHistorySituationsDetails(subQuery, subQueryArgs)
	return querier.execute(query)
}

func (querier HistorySituationsQuerier) execute(builder sq.SelectBuilder) ([]HistorySituationsV4, error) {
	rows, err := builder.RunWith(querier.conn.DB).Query()
	if err != nil {
		return make([]HistorySituationsV4, 0), err
	}
	defer rows.Close()
	return querier.scanAll(rows)
}

func (querier HistorySituationsQuerier) scan(rows *sql.Rows) (HistorySituationsV4, error) {

	var rawParameters []byte
	var rawExpressionFacts []byte
	var rawMetadatas []byte
	item := HistorySituationsV4{}
	err := rows.Scan(&item.ID, &item.SituationID, &item.SituationInstanceID, &item.Ts, &rawParameters, &rawExpressionFacts, &rawMetadatas, &item.SituationName, &item.SituationInstanceName)
	if err != nil {
		return HistorySituationsV4{}, err
	}

	if len(rawParameters) > 0 {
		err = json.Unmarshal(rawParameters, &item.Parameters)
		if err != nil {
			// TODO: add logs !
			return HistorySituationsV4{}, err
		}
	}

	if len(rawExpressionFacts) > 0 {
		err = json.Unmarshal(rawExpressionFacts, &item.ExpressionFacts)
		if err != nil {
			// TODO: add logs !
			return HistorySituationsV4{}, err
		}
	}

	if len(rawMetadatas) > 0 {
		err = json.Unmarshal(rawMetadatas, &item.Metadatas)
		if err != nil {
			// TODO: add logs !
			return HistorySituationsV4{}, err
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
