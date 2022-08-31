package history

// Flatten situation data (old and new facts + parameters)
func (service HistoryService) ExtractFactData(historyFactsNew []HistoryFactsV4, existingFactIDs []int64) ([]HistoryFactsV4, map[string]interface{}, error) {
	historySituationFlattenData := make(map[string]interface{})
	historyFactsAll := make([]HistoryFactsV4, 0)

	for _, historyFactNew := range historyFactsNew {
		historyFactNewData, err := historyFactNew.Result.ToAbstractMap()
		if err != nil {
			return make([]HistoryFactsV4, 0), make(map[string]interface{}), err
		}
		historySituationFlattenData[historyFactNew.FactName] = historyFactNewData
		historyFactsAll = append(historyFactsAll, historyFactNew)
	}

	for _, factId := range existingFactIDs {
		exists := false
		for _, historyFactNew := range historyFactsNew {
			if factId == historyFactNew.FactID {
				exists = true
			}
		}
		if !exists {
			historyFactLast, err := service.HistoryFactsQuerier.QueryOne(service.HistoryFactsQuerier.Builder.GetHistoryFactLast(factId))
			if err != nil {
				return make([]HistoryFactsV4, 0), make(map[string]interface{}), err
			}

			historyFactLastData, err := historyFactLast.Result.ToAbstractMap()
			if err != nil {
				return make([]HistoryFactsV4, 0), make(map[string]interface{}), err
			}

			historySituationFlattenData[historyFactLast.FactName] = historyFactLastData
			historyFactsAll = append(historyFactsAll, historyFactLast)
		}
	}
	return historyFactsAll, historySituationFlattenData, nil
}
