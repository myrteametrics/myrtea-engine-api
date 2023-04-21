package history

import (
	"testing"
	"time"
)

func TestGetHistorySituationsIdsBase(t *testing.T) {
	t.SkipNow()

	options := GetHistorySituationsOptions{
		SituationID:         -1,
		SituationInstanceID: -1,
		FromTS:              time.Time{},
		ToTS:                time.Time{},
	}
	builder := HistorySituationsBuilder{}.GetHistorySituationsIdsBase(options)

	t.Fail()
	t.Log(builder.ToSql())
}

func TestGetHistorySituationsIdsLast(t *testing.T) {
	t.SkipNow()

	options := GetHistorySituationsOptions{
		SituationID:         -1,
		SituationInstanceID: -1,
		FromTS:              time.Time{},
		ToTS:                time.Time{},
	}
	builder := HistorySituationsBuilder{}.GetHistorySituationsIdsLast(options)

	t.Fail()
	t.Log(builder.ToSql())
}

func TestGetHistorySituationsIdsByStandardInterval(t *testing.T) {
	t.SkipNow()

	options := GetHistorySituationsOptions{
		SituationID:         -1,
		SituationInstanceID: -1,
		FromTS:              time.Time{},
		ToTS:                time.Time{},
	}
	interval := "day"
	builder := HistorySituationsBuilder{}.GetHistorySituationsIdsByStandardInterval(options, interval)

	t.Fail()
	t.Log(builder.ToSql())
}

func TestGetHistorySituationsIdsByCustomInterval(t *testing.T) {
	t.SkipNow()

	options := GetHistorySituationsOptions{
		SituationID:         -1,
		SituationInstanceID: -1,
		FromTS:              time.Time{},
		ToTS:                time.Time{},
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
