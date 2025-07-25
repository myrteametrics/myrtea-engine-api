package explainer

import (
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/reader"
	"time"

	"github.com/myrteametrics/myrtea-engine-api/v5/internal/explainer/issues"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/fact"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/model"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/history"
	"github.com/myrteametrics/myrtea-sdk/v5/engine"
	"go.uber.org/zap"
)

// CreateIssue function used to create an issue
func CreateIssue(situationHistoryID int64, situationID int64, templateInstanceID int64, ts time.Time, rule model.RuleData, name string, level model.IssueLevel, timeout time.Duration, key string) (int64, error) {
	expirationTS := ts.Add(timeout).UTC()

	issue := model.Issue{
		Name:               name,
		Key:                key,
		Level:              level,
		SituationHistoryID: situationHistoryID,
		SituationID:        situationID,
		SituationTS:        ts,
		TemplateInstanceID: templateInstanceID,
		ExpirationTS:       expirationTS,
		State:              model.Open,
		Rule:               rule,
	}

	//isNew, err := isNewIssue(issue)
	isOpen, err := IsOpenOrDraftIssue(issue.Key)
	if err != nil {
		zap.L().Error("Cannot search in issue history", zap.String("key", key), zap.Error(err))
		return -1, err
	}
	if isOpen {
		//zap.L().Debug("Issue creation skipped (timeout not reached)")
		zap.L().Debug("Issue creation skipped - open/draft issue already existed")
		return 0, nil
	}

	id, err := issues.R().Create(issue)
	if err != nil {
		return -1, err
	}
	return id, nil
}

//func isNewIssue(issue model.Issue) (bool, error) {
//	issues, err := issues.R().GetCloseToTimeoutByKey(issue.Key, issue.SituationTS)
//	if err != nil {
//		return false, err
//	}
//	if len(issues) == 0 {
//		return true, nil
//	}
//	return false, nil
//}

func IsOpenOrDraftIssue(issueKey string) (bool, error) {
	issues, err := issues.R().GetOpenAndDraftIssuesByKey(issueKey)
	if err != nil {
		return false, err
	}
	if len(issues) == 0 {
		return false, nil
	}
	return true, nil
}

// GetFactsHistory get the history of facts for an issue
func GetFactsHistory(issue model.Issue) ([]model.FrontFactHistory, bool, error) {

	// TODO: interval shouldn't hard coded ?
	interval := 3 * time.Hour
	from, to := getFromAndToHistoryDates(issue.SituationTS, interval)

	historySituationFacts, err := history.S().HistorySituationFactsQuerier.Query(history.S().HistorySituationFactsQuerier.Builder.GetHistorySituationFacts([]int64{issue.SituationHistoryID}))
	if err != nil {
		zap.L().Error("", zap.Error(err))
		return nil, false, err
	}

	// Get SituationHistoryV4 from ID
	historySituationIDs, err := history.S().HistorySituationsQuerier.QueryIDs(history.S().HistorySituationsQuerier.Builder.GetHistorySituationsIdsBase(
		history.GetHistorySituationsOptions{
			SituationID:          issue.SituationID,
			SituationInstanceIDs: []int64{issue.TemplateInstanceID},
			FromTS:               from,
			ToTS:                 to,
		},
	))
	if err != nil {
		zap.L().Error("", zap.Error(err))
		return nil, false, err
	}

	// Get All Value of All HistoryFactV4 from all HistorySituationV4 IDs
	historyFacts, _, err := history.S().GetHistoryFactsFromSituationIds(historySituationIDs)
	if err != nil {
		zap.L().Error("", zap.Error(err))
		return nil, false, err
	}

	historyFactGroups := make(map[int64][]history.HistoryFactsV4)
	for _, historyFact := range historyFacts {
		if l, exists := historyFactGroups[historyFact.FactID]; exists {
			historyFactGroups[historyFact.FactID] = append(l, historyFact)
		} else {
			historyFactGroups[historyFact.FactID] = []history.HistoryFactsV4{historyFact}
		}
	}

	// Build frontFacts
	frontFacts := make([]model.FrontFactHistory, 0)
	for factID, historyFacts := range historyFactGroups {
		f, found, err := fact.R().Get(factID)
		if err != nil {
			zap.L().Warn("Get fact", zap.Error(err))
			continue
		}
		if !found {
			zap.L().Warn("fact doesn't exists", zap.Int64("factID", factID))
			continue
		}

		historyFactCurrent := history.HistoryFactsV4{}
		factValues := make(map[time.Time]model.FactValue)
		for _, historyFact := range historyFacts {
			factValue := extractFactValue(historyFact.Result, f)
			if factValue.GetType() == "not_supported" {
				factValues = nil
				historyFactCurrent = historyFact
				break
			}
			for _, historySituationFact := range historySituationFacts {
				if historyFact.ID == historySituationFact.HistoryFactID {
					factValue.SetCurrent(true)
					historyFactCurrent = historyFact
				}
			}
			factValues[historyFact.Ts] = factValue
		}

		factValueCurrent := extractFactValue(historyFactCurrent.Result, f)

		frontFact := model.FrontFactHistory{
			ID:           historyFactCurrent.FactID,
			Name:         historyFactCurrent.FactName,
			Deepness:     factValueCurrent.GetDeepness(),
			Type:         factValueCurrent.GetType(),
			CurrentValue: factValueCurrent,
			History:      factValues,
		}
		frontFacts = append(frontFacts, frontFact)
	}

	return frontFacts, true, nil
}

func getFromAndToHistoryDates(ts time.Time, interval time.Duration) (time.Time, time.Time) {
	tsNow := time.Now().Truncate(1 * time.Millisecond).UTC()

	tsFrom := ts.Add(-interval)
	tsTo := ts.Add(interval)

	if tsTo.After(tsNow) {
		tsTo = tsNow
	}

	return tsFrom, tsTo
}

func extractFactValue(item reader.Item, factDefinition engine.Fact) model.FactValue {
	if factDefinition.AdvancedSource != "" {
		return &model.NotSupportedValue{}
	}

	if len(factDefinition.Dimensions) > 0 {
		return &model.NotSupportedValue{}
	}

	keyDocCount := "doc_count"

	if factDefinition.Intent.Operator == engine.Count && factDefinition.Model == factDefinition.Intent.Term {
		if val, ok := item.Aggs[keyDocCount]; ok {
			return &model.SingleValue{Key: keyDocCount, Value: val.Value}
		}

		zap.L().Warn("No value found for key: ", zap.String("KeyAgg", keyDocCount), zap.Any("fact", factDefinition), zap.Any("item", item))
		return &model.NotSupportedValue{}
	}

	for aggKey, agg := range item.Aggs {
		if aggKey != keyDocCount {
			return &model.SingleValue{Key: aggKey, Value: agg.Value}
		}
	}

	zap.L().Warn("No value found for key other than doc_count", zap.Any("fact", factDefinition), zap.Any("item", item))
	return &model.NotSupportedValue{}
}
