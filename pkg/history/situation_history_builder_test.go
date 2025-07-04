package history

import (
	"reflect"
	"testing"
	"time"
)

func TestGetHistorySituationsIdsBase(t *testing.T) {
	t.SkipNow()

	options := GetHistorySituationsOptions{
		SituationID:          -1,
		SituationInstanceIDs: []int64{},
		FromTS:               time.Time{},
		ToTS:                 time.Time{},
	}
	builder := HistorySituationsBuilder{}.GetHistorySituationsIdsBase(options)

	t.Fail()
	t.Log(builder.ToSql())
}

func TestGetHistorySituationsIdsLast(t *testing.T) {
	t.SkipNow()

	options := GetHistorySituationsOptions{
		SituationID:          -1,
		SituationInstanceIDs: []int64{},
		FromTS:               time.Time{},
		ToTS:                 time.Time{},
	}
	builder := HistorySituationsBuilder{}.GetHistorySituationsIdsLast(options)

	t.Fail()
	t.Log(builder.ToSql())
}

func TestGetHistorySituationsIdsByStandardInterval(t *testing.T) {
	t.SkipNow()

	options := GetHistorySituationsOptions{
		SituationID:          -1,
		SituationInstanceIDs: []int64{},
		FromTS:               time.Time{},
		ToTS:                 time.Time{},
	}
	interval := "day"
	builder := HistorySituationsBuilder{}.GetHistorySituationsIdsByStandardInterval(options, interval)

	t.Fail()
	t.Log(builder.ToSql())
}

func TestGetHistorySituationsIdsByCustomInterval(t *testing.T) {
	t.SkipNow()

	options := GetHistorySituationsOptions{
		SituationID:          -1,
		SituationInstanceIDs: []int64{},
		FromTS:               time.Time{},
		ToTS:                 time.Time{},
	}
	interval := 48 * time.Hour
	referenceDate := time.Now()
	builder := HistorySituationsBuilder{}.GetHistorySituationsIdsByCustomInterval(options, interval, referenceDate)

	t.Fail()
	t.Log(builder.ToSql())
}

func TestGetHistorySituationsDetails(t *testing.T) {
	t.SkipNow()

	subQueryIds := ""
	subQueryIdsArgs := []interface{}{}
	builder := HistorySituationsBuilder{}.GetHistorySituationsDetails(subQueryIds, subQueryIdsArgs)

	t.Fail()
	t.Log(builder.ToSql())
}

func TestGetLatestHistorySituation(t *testing.T) {
	builder := HistorySituationsBuilder{}

	expectedSQL := "SELECT ts, metadatas FROM situation_history_v5 WHERE situation_id = $1 AND situation_instance_id = $2 AND ts >= $3::timestamptz ORDER BY ts DESC LIMIT 1"
	expectedArgs := []interface{}{
		int64(1),
		int64(12345678),
		getStartDate30DaysAgo(),
	}

	sql, args, err := builder.GetLatestHistorySituation(1, 12345678).ToSql()
	if err != nil {
		t.Fatalf("Failed to build SQL: %v", err)
	}

	if expectedSQL != sql {
		t.Errorf("Expected SQL to be \n%s\n but got \n%s", expectedSQL, sql)
	}

	if !reflect.DeepEqual(expectedArgs, args) {
		t.Errorf("Expected args to be %v, but got %v", expectedArgs, args)
	}
}
