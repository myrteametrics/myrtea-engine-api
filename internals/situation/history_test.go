package situation

import (
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/models"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/tests"
	"github.com/myrteametrics/myrtea-sdk/v4/postgres"
)

func dbInit(dbClient *sqlx.DB, t *testing.T) {
	dbDestroy(dbClient, t)
	tests.DBExec(dbClient, tests.SituationHistoryTableV1, t, true)
}

func dbDestroy(dbClient *sqlx.DB, t *testing.T) {
	tests.DBExec(dbClient, tests.SituationHistoryDropTableV1, t, true)
}

func TestPersistSituation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}

	db := tests.DBClient(t)
	defer dbDestroy(db, t)
	dbInit(db, t)
	postgres.ReplaceGlobals(db)

	id := int64(1)
	ts := time.Date(2019, time.June, 14, 12, 30, 15, 0, time.UTC)
	r := HistoryRecord{ID: id, TS: ts, FactsIDS: make(map[int64]*time.Time, 0)}
	err := Persist(r, false)
	if err != nil {
		t.Error(err)
	}

	r2 := HistoryRecord{ID: id, TS: ts, FactsIDS: make(map[int64]*time.Time, 0)}
	err = Persist(r2, false)
	if err == nil {
		t.Error("Should not be able to persist the same row two time")
	}
}

func TestGetSituation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}

	db := tests.DBClient(t)
	defer dbDestroy(db, t)
	dbInit(db, t)
	postgres.ReplaceGlobals(db)

	id := int64(1)
	ts := time.Date(2019, time.June, 14, 12, 30, 15, 0, time.UTC)
	r := HistoryRecord{ID: id, TS: ts, FactsIDS: make(map[int64]*time.Time, 0)}
	err := Persist(r, false)
	if err != nil {
		t.Error(err)
	}

	record, err := GetFromHistory(id, ts, 0, false)
	if err != nil {
		t.Error(err)
	}
	if record == nil {
		t.Error("situation history record is nil")
	}
	if record.ID != r.ID {
		t.Error("Invalid record ID")
	}
}

func TestGetSituationInvalid(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}

	db := tests.DBClient(t)
	defer dbDestroy(db, t)
	dbInit(db, t)
	postgres.ReplaceGlobals(db)

	ts := time.Date(2019, time.June, 14, 12, 30, 15, 0, time.UTC)
	r := HistoryRecord{ID: 1, TS: ts, FactsIDS: make(map[int64]*time.Time, 0)}
	err := Persist(r, false)
	if err != nil {
		t.Error(err)
	}

	record, err := GetFromHistory(2, ts, 0, false)
	if err != nil {
		t.Error(err)
	}
	if record != nil {
		t.Error("situation history record should be nil")
	}

	record, err = GetFromHistory(1, time.Date(2019, time.June, 14, 21, 59, 59, 0, time.UTC), 0, false)
	if err != nil {
		t.Error(err)
	}
	if record != nil {
		t.Error("situation history record should be nil")
	}
}

func TestGetSituationNoPostgres(t *testing.T) {
	ts := time.Date(2019, time.June, 14, 12, 30, 15, 0, time.UTC)
	record, err := GetFromHistory(1, ts, 0, true)
	if err == nil {
		t.Error("Should not be able to query history with a uninitialized postgresql")
	}
	if record != nil {
		t.Error("No record should be found")
	}
}

func TestGetSituationClosest(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}

	db := tests.DBClient(t)
	defer dbDestroy(db, t)
	dbInit(db, t)
	postgres.ReplaceGlobals(db)

	id := int64(1)

	ts1 := time.Date(2019, time.June, 14, 12, 30, 15, 0, time.UTC)
	r1 := HistoryRecord{ID: id, TS: ts1, FactsIDS: map[int64]*time.Time{1: nil}}
	err := Persist(r1, false)
	if err != nil {
		t.Error(err)
	}

	ts2 := time.Date(2019, time.June, 14, 12, 35, 15, 0, time.UTC)
	r2 := HistoryRecord{ID: id, TS: ts2, FactsIDS: map[int64]*time.Time{2: nil}}
	err = Persist(r2, false)
	if err != nil {
		t.Error(err)
	}

	// TS Before R1 and R2 (returns nil)
	record, err := GetFromHistory(id, time.Date(2019, time.June, 14, 7, 00, 00, 0, time.UTC), 0, true)
	if err != nil {
		t.Error(err)
	}
	if record != nil {
		t.Error("situation history record should be nil")
	}

	// TS Between R1 and R2 (returns R1)
	record, err = GetFromHistory(id, time.Date(2019, time.June, 14, 12, 32, 59, 0, time.UTC), 0, true)
	if err != nil {
		t.Error(err)
	}
	if record == nil {
		t.Error("situation history record is nil")
	}
	if _, ok := record.FactsIDS[1]; !ok {
		t.Error("Invalid record ID")
	}

	// TS After R1 and R2 (returns R2)
	record, err = GetFromHistory(id, time.Date(2019, time.June, 14, 15, 59, 59, 0, time.UTC), 0, true)
	if err != nil {
		t.Error(err)
	}
	if record == nil {
		t.Error("situation history record is nil")
	}
	if _, ok := record.FactsIDS[2]; !ok {
		t.Error("Invalid record ID")
	}
}

func TestUpdateHistoryMetadata(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}

	db := tests.DBClient(t)
	defer dbDestroy(db, t)
	dbInit(db, t)
	postgres.ReplaceGlobals(db)

	id := int64(1)
	ts1 := time.Date(2019, time.June, 14, 12, 30, 15, 0, time.UTC)
	r1 := HistoryRecord{ID: int64(id), TS: ts1, FactsIDS: map[int64]*time.Time{1: nil}}

	err := Persist(r1, false)
	if err != nil {
		t.Error(err)
	}

	metaDatas := []models.MetaData{
		{
			Key:         "key1",
			Value:       "value1",
			RuleID:      1,
			RuleVersion: 0,
			CaseName:    "case1",
		},
	}

	err = UpdateHistoryMetadata(id, ts1, 0, metaDatas)
	if err != nil {
		t.Error(err)
	}

	// TODO: check metadatas ?
}

func TestUpdateHistoryMetadataInvalid(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}

	db := tests.DBClient(t)
	defer dbDestroy(db, t)
	dbInit(db, t)
	postgres.ReplaceGlobals(db)

	metaDatas := []models.MetaData{
		{
			Key:         "key1",
			Value:       "value1",
			RuleID:      1,
			RuleVersion: 0,
			CaseName:    "case1",
		},
	}

	err := UpdateHistoryMetadata(1, time.Date(2019, time.June, 14, 12, 30, 15, 0, time.UTC), 0, metaDatas)
	if err == nil {
		t.Error("situation doesn't exists and could not be updated")
	}
}
