package history

import (
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
		t.Fatalf("Error initializing DB: %s",err)
	}

	defer db.Close()

	querier := HistoryFactsQuerier{
		Builder: HistoryFactsBuilder{},
		conn:    db,
	}

	param := ParamGetFactHistory {
		FactID: 10000000,
		SituationID: 10000000,
		SituationInstanceID: 10000000,
	}

	value := 44
    historyItem := HistoryFactsV4{
        FactID: param.FactID,
        SituationID: param.SituationID,
        SituationInstanceID: param.SituationInstanceID,
        Ts: time.Now(),
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
        t.Fatalf("Expected 1 result, got %d and resulat : %v ", len(results.Results),results)
    }

    if results.Results[0] != int64(value) {
        t.Fatalf("Retrieved ID does not match inserted ID")
    }
}

