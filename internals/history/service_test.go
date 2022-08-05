package history

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/myrteametrics/myrtea-engine-api/v4/internals/tests"
)

func TestQuery3(t *testing.T) {
	t.Fail()
	db := tests.DBClient(t)
	historyService := New(db)

	options := GetHistorySituationsOptions{
		SituationID:         4,
		SituationInstanceID: -1,
		FromTS:              time.Date(2022, time.July, 1, 0, 0, 0, 0, time.UTC),
		ToTS:                time.Time{},
	}
	interval := "day"

	// Fetch situations history
	historySituations, err := historyService.historySituationsQuerier.GetHistorySituationsIdsByStandardInterval(options, interval)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	// Fetch facts history
	historySituationsIds := make([]int64, 0)
	for _, item := range historySituations {
		historySituationsIds = append(historySituationsIds, item.ID)
	}

	historyFacts, historySituationFacts, err := historyService.historyFactsQuerier.GetHistoryFactsFromSituation(historyService.historySituationFactsQuerier, historySituationsIds)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	// Extract results
	result := historyService.ExtractData(historySituations, historySituationFacts, historyFacts)
	b, _ := json.Marshal(result)
	t.Log(string(b))
}
