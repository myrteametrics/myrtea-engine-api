package history

import (
	"reflect"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/reader"
	"github.com/myrteametrics/myrtea-sdk/v4/postgres"
)

func TestGetHistoryFactLast(t *testing.T) {
	t.SkipNow()

	builder := HistoryFactsBuilder{}.GetHistoryFactLast(4, 109, 19)

	t.Fail()
	t.Log(builder.ToSql())
}

func TestGetHistoryFacts(t *testing.T) {
	t.SkipNow()

	historyFactsIds := []int64{1, 2, 3}
	builder := HistoryFactsBuilder{}.GetHistoryFacts(historyFactsIds)

	t.Fail()
	t.Log(builder.ToSql())
}

func initialiseDB() (*sqlx.DB, error) {

	credentials := postgres.Credentials{
		URL:      "localhost",
		Port:     "5432",
		DbName:   "postgres",
		User:     "postgres",
		Password: "postgres",
	}

	db, err := postgres.DbConnection(credentials)
	if err != nil {
		return nil, err
	}
	return db, nil
}
func TestGetByCriteria(t *testing.T) {

	// Initiate DB connection
	db, err := initialiseDB()

	if err != nil {
		t.Fatalf("Error initializing DB: %s", err)
	}

	defer db.Close()

	querier := HistoryFactsQuerier{
		Builder: HistoryFactsBuilder{},
		conn:    db,
	}

	param := ParamGetFactHistory{
		FactID:              10000000,
		SituationID:         10000000,
		SituationInstanceID: 10000000,
	}

	value := 44
	historyItem := HistoryFactsV4{
		FactID:              param.FactID,
		SituationID:         param.SituationID,
		SituationInstanceID: param.SituationInstanceID,
		Ts:                  time.Now(),
		Result: reader.Item{
			Aggs: map[string]*reader.ItemAgg{
				"count": {
					Value: value,
				},
			},
		},
	}

	insertedID, err := querier.Insert(historyItem)
	if err != nil {
		t.Fatalf("Error inserting: %s", err)
	}
	if insertedID <= 0 {
		t.Fatalf("Invalid ID returned after insert")
	}

	defer func() {
		err := querier.Delete(insertedID)
		if err != nil {
			t.Fatalf("Error cleaning up: %s", err)
		}
	}()

	results, err := querier.GetTodaysFactResultByParameters(param)
	if err != nil {
		t.Fatalf("Error retrieving: %s", err)
	}

	if len(results.Results) != 1 {
		t.Fatalf("Expected 1 result, got %d and resulat : %v ", len(results.Results), results)
	}

	if results.Results[0].Value != int64(value) {
		t.Fatalf("Retrieved ID does not match inserted ID")
	}
}

func TestGetTodaysFactResultByParameters(t *testing.T) {
	builder := HistoryFactsBuilder{}

	param := ParamGetFactHistory{
		FactID:              123,
		SituationID:         456,
		SituationInstanceID: 789,
	}

	expectedSQL := `SELECT result, ts FROM fact_history_v5 WHERE fact_id = $1 AND situation_id = $2 AND situation_instance_id = $3 AND ts >= $4::timestamptz AND ts < $5::timestamptz`

	todayStart, tomorrowStart := getTodayTimeRange()

	expectedArgs := []interface{}{
		param.FactID,
		param.SituationID,
		param.SituationInstanceID,
		todayStart,
		tomorrowStart,
	}

	sql, args, err := builder.GetTodaysFactResultByParameters(param).ToSql()
	if err != nil {
		t.Fatalf("Failed to build SQL: %v", err)
	}

	if expectedSQL != sql {
		t.Errorf("Expected SQL to be %s but got %s", expectedSQL, sql)
	}

	if !reflect.DeepEqual(expectedArgs, args) {
		t.Errorf("Expected args to be %v, but got %v", expectedArgs, args)
	}
}
