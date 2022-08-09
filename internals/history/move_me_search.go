package history

import (
	"sort"
	"time"

	"github.com/myrteametrics/myrtea-engine-api/v4/internals/search"
	"go.uber.org/zap"
)

func ExtractHistoryDataSearch(historySituations []HistorySituationsV4, historySituationFacts []HistorySituationFactsV4, historyFacts []HistoryFactsV4) search.QueryResult {

	mapFacts := make(map[int64]HistoryFactsV4)
	for _, historyFact := range historyFacts {
		mapFacts[historyFact.ID] = historyFact
	}

	mapSituationFact := make(map[int64][]int64)
	for _, historySituationFact := range historySituationFacts {
		if list, exists := mapSituationFact[historySituationFact.HistorySituationID]; exists {
			mapSituationFact[historySituationFact.HistorySituationID] = append(list, historySituationFact.HistoryFactID)
		} else {
			mapSituationFact[historySituationFact.HistorySituationID] = []int64{historySituationFact.HistoryFactID}
		}
	}

	situationRecords := make([]search.SituationHistoryRecord, 0)
	for _, historySituation := range historySituations {

		factIds, exists := mapSituationFact[historySituation.ID]
		if !exists {
			zap.L().Error("SHOULD EXISTS ?!")
		}

		factRecords := make([]search.FactHistoryRecord, 0)
		for _, factId := range factIds {
			factHistory := mapFacts[factId]

			var value interface{}
			var docCount interface{}
			for k, v := range factHistory.Result.Aggs {
				if k == "doc_count" {
					docCount = v.Value
					if value == nil {
						value = v.Value
					}
				} else {
					value = v.Value
				}
			}

			factHistoryRecord := search.FactHistoryRecord{
				FactID:    factHistory.FactID,
				FactName:  factHistory.FactName,
				DateTime:  factHistory.Ts,
				Value:     value,
				DocCount:  docCount,
				Buckets:   factHistory.Result.Buckets,
				Baselines: factHistory.Result.Baselines,
			}
			factRecords = append(factRecords, factHistoryRecord)
		}

		metadatas := make(map[string]interface{})
		for _, metadata := range historySituation.Metadatas {
			metadatas[metadata.Key] = metadata.Value
		}

		parameters := make(map[string]interface{})
		for k, v := range historySituation.Parameters {
			parameters[k] = v
		}

		situationRecord := search.SituationHistoryRecord{
			SituationID:           historySituation.SituationID,
			SituationName:         historySituation.SituationName,
			SituationInstanceID:   historySituation.SituationInstanceID,
			SituationInstanceName: historySituation.SituationInstanceName,
			DateTime:              historySituation.Ts,
			Parameters:            parameters,
			ExpressionFacts:       historySituation.ExpressionFacts,
			MetaData:              metadatas,
			Facts:                 factRecords,
		}

		situationRecords = append(situationRecords, situationRecord)
	}

	// TODO: Is it really mandatory ? or managed by the frontend ?
	sort.SliceStable(situationRecords, func(i, j int) bool {
		return situationRecords[i].DateTime.Before(situationRecords[j].DateTime)
	})

	// TODO: REMOVE THIS ?
	// returns situationRecords []search.SituationHistoryRecord directly ?
	situationHistoryRecord := search.SituationHistoryRecords{
		DateTime:   time.Time{},
		Situations: situationRecords,
	}
	searchResult := []search.SituationHistoryRecords{
		situationHistoryRecord,
	}

	return searchResult
}
