package history

import (
	"testing"
	"time"

	"github.com/myrteametrics/myrtea-engine-api/v5/internals/tests"
)

func TestCompact(t *testing.T) {
	t.SkipNow()

	db := tests.DBClient(t)

	options := GetHistorySituationsOptions{
		SituationID:         -1,
		SituationInstanceID: -1,
		FromTS:              time.Time{},
		ToTS:                time.Now().Add(-1 * 24 * time.Hour),
		ParameterFilters:    make(map[string]interface{}),
	}
	interval := "day"

	New(db).CompactHistory(options, interval)
}

func TestPurge(t *testing.T) {
	t.SkipNow()

	db := tests.DBClient(t)

	options := GetHistorySituationsOptions{
		SituationID:         -1,
		SituationInstanceID: -1,
		FromTS:              time.Time{},
		ToTS:                time.Now().Add(-1 * 24 * time.Hour),
		ParameterFilters:    make(map[string]interface{}),
	}

	New(db).PurgeHistory(options)
}
