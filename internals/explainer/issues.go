package explainer

import (
	"fmt"
	"time"

	"github.com/myrteametrics/myrtea-engine-api/v4/internals/explainer/issues"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/fact"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/models"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/reader"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/situation"
	"github.com/myrteametrics/myrtea-sdk/v4/engine"
	"go.uber.org/zap"
)

//CreateIssue function used to create an issue
func CreateIssue(situationID int64, ts time.Time, templateInstanceID int64, rule models.RuleData, name string, level models.IssueLevel, timeout time.Duration, key string) (int64, error) {
	expirationTS := ts.Add(timeout).UTC()

	issue := models.Issue{
		Name:               name,
		Key:                key,
		Level:              level,
		SituationID:        situationID,
		SituationTS:        ts,
		TemplateInstanceID: templateInstanceID,
		ExpirationTS:       expirationTS,
		State:              models.Open,
		Rule:               rule,
	}

	isNew, err := isNewIssue(issue)
	if err != nil {
		zap.L().Error("Cannot search in issue history", zap.String("key", key), zap.Error(err))
		return -1, err
	}
	if !isNew {
		zap.L().Debug("Issue creation skipped (timeout not reached)")
		return 0, nil
	}

	id, err := issues.R().Create(issue)
	if err != nil {
		return -1, err
	}
	return id, nil
}

func isNewIssue(issue models.Issue) (bool, error) {
	issues, err := issues.R().GetCloseToTimeoutByKey(issue.Key, issue.SituationTS)
	if err != nil {
		return false, err
	}
	if len(issues) == 0 {
		return true, nil
	}
	return false, nil
}

//GetFactsHistory get the history of facts for an issue
func GetFactsHistory(issueID int64, groups []int64) ([]models.FrontFactHistory, bool, error) {
	issue, found, err := issues.R().Get(issueID, groups)
	if err != nil {
		return nil, false, err
	}
	if !found {
		zap.L().Warn("Issue doesn't exists or not in groups")
		return nil, false, nil
	}

	history := make([]models.FrontFactHistory, 0)

	record, err := situation.GetFromHistory(issue.SituationID, issue.SituationTS, issue.TemplateInstanceID, false)
	if err != nil {
		return nil, false, err
	}

	if record == nil {
		zap.L().Warn("No record situation history found")
		return nil, false, nil
	}

	//TODO: interval shouldn't hard coded ?
	interval := 3 * time.Hour
	from, to := getFromAndToHistoryDates(issue.SituationTS, interval)

	for factHistoryID, factHistoryTS := range record.FactsIDS {
		factDefinition, found, err := fact.R().Get(factHistoryID)
		if err != nil {
			return nil, false, err
		}
		if !found {
			return nil, false, fmt.Errorf("fact definition not found for id %d", factHistoryID)
		}

		frontFact := models.FrontFactHistory{
			ID:   factHistoryID,
			Name: factDefinition.Name,
		}
		if factDefinition.IsObject {
			items, _, err := fact.GetFactResultFromHistory(factHistoryID, *factHistoryTS, issue.SituationID, 0, false, 0)
			if err != nil {
				return nil, false, err
			}
			attributes := make(map[string]interface{}, 0)
			for k, v := range items.Aggs {
				attributes[k] = v.Value
			}
			factValue := &models.ObjectValue{
				Attributes: attributes,
			}
			frontFact.Deepness = factValue.GetDeepness()
			frontFact.Type = factValue.GetType()
			frontFact.CurrentValue = factValue
			frontFact.History = make(map[time.Time]models.FactValue, 0)
		} else {
			var factCurrentValue models.FactValue
			factValues := make(map[time.Time]models.FactValue)

			items, err := fact.GetFactRangeFromHistory(factHistoryID, issue.SituationID, issue.TemplateInstanceID, from, to)
			if err != nil {
				return nil, false, err
			}
			for ts, item := range items {
				factValue, err := extractFactValue(item, factDefinition)
				if err != nil {
					return nil, false, err
				}

				if factValue.GetType() == "not_supported" {
					factValues = nil
					factCurrentValue = factValue
					break
				}

				if factHistoryTS.Equal(ts) {
					factValue.SetCurrent(true)
					factCurrentValue = factValue
				}

				factValues[ts] = factValue
			}
			frontFact.Deepness = factCurrentValue.GetDeepness()
			frontFact.Type = factCurrentValue.GetType()
			frontFact.CurrentValue = factCurrentValue
			frontFact.History = factValues
		}

		history = append(history, frontFact)
	}

	return history, true, nil
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

func extractFactValue(item reader.Item, factDefinition engine.Fact) (models.FactValue, error) {
	if factDefinition.AdvancedSource != "" {
		return &models.NotSupportedValue{}, nil
	}

	if len(factDefinition.Dimensions) > 0 {
		return &models.NotSupportedValue{}, nil
	}

	keyDocCount := "doc_count"

	if factDefinition.Intent.Operator == engine.Count && factDefinition.Model == factDefinition.Intent.Term {
		if val, ok := item.Aggs[keyDocCount]; ok {
			return &models.SingleValue{Key: keyDocCount, Value: val.Value}, nil
		}

		zap.L().Warn("No value found for key: ", zap.String("KeyAgg", keyDocCount))
		return &models.NotSupportedValue{}, nil
	}

	for aggKey, agg := range item.Aggs {
		if aggKey != keyDocCount {
			return &models.SingleValue{Key: aggKey, Value: agg.Value}, nil
		}
	}

	zap.L().Warn("No value found for key other than doc_count")
	return &models.NotSupportedValue{}, nil
}
