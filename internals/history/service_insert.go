package history

// Flatten situation data (old and new facts + parameters)
func (service HistoryService) ExtractFactData2(historyFactsNew []HistoryFactsV4, existingFactIDs []int64) ([]HistoryFactsV4, map[string]interface{}, error) {
	historySituationFlattenData := make(map[string]interface{})
	historyFactsAll := make([]HistoryFactsV4, 0)

	for _, historyFactNew := range historyFactsNew {
		historyFactNewData, err := historyFactNew.Result.ToAbstractMap()
		if err != nil {
			// err
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
			query := service.HistoryFactsQuerier.Builder.GetHistoryFactLast(factId)
			historyFactLast, err := service.HistoryFactsQuerier.QueryOne(query)
			if err != nil {
				// err
			}

			historyFactLastData, err := historyFactLast.Result.ToAbstractMap()
			if err != nil {
				// err
			}

			historySituationFlattenData[historyFactLast.FactName] = historyFactLastData
			historyFactsNew = append(historyFactsNew, historyFactLast)
		}
	}
	return historyFactsNew, historySituationFlattenData, nil
}

func (service HistoryService) ExtractFactData(historyFactNew HistoryFactsV4, existingFactIDs []int64) ([]HistoryFactsV4, map[string]interface{}, error) {
	historySituationFlattenData := make(map[string]interface{})
	historyFactsNew := make([]HistoryFactsV4, 0)
	for _, factId := range existingFactIDs {
		if factId == historyFactNew.FactID {
			historyFactNewData, err := historyFactNew.Result.ToAbstractMap()
			if err != nil {
				// err
			}

			historySituationFlattenData[historyFactNew.FactName] = historyFactNewData
			historyFactsNew = append(historyFactsNew, historyFactNew)

		} else {
			query := service.HistoryFactsQuerier.Builder.GetHistoryFactLast(factId)
			historyFactLast, err := service.HistoryFactsQuerier.QueryOne(query)
			if err != nil {
				// err
			}

			historyFactLastData, err := historyFactLast.Result.ToAbstractMap()
			if err != nil {
				// err
			}

			historySituationFlattenData[historyFactLast.FactName] = historyFactLastData
			historyFactsNew = append(historyFactsNew, historyFactLast)
		}
	}
	return historyFactsNew, historySituationFlattenData, nil
}
