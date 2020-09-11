package evaluator

import (
	"errors"
	"fmt"
	"time"

	"github.com/myrteametrics/myrtea-engine-api/v4/internals/calendar"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/fact"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/reader"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/rule"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/situation"
	"github.com/myrteametrics/myrtea-sdk/v4/engine"
	"github.com/myrteametrics/myrtea-sdk/v4/ruleeng"
)

var (
	_globalREngine = make(map[string]*ruleeng.RuleEngine, 0)
	_lastUpdate    = make(map[string]time.Time, 0)
)

//EvaluateSituations evaluates a slice of situations and return a slice with the evaluated situations
func EvaluateSituations(situations []SituationToEvaluate, engineID string) ([]EvaluatedSituation, error) {
	var err error
	if _, ok := _globalREngine[engineID]; !ok {
		err = InitEngine(engineID)
		if err != nil {
			return nil, err
		}
	}
	err = UpdateEngine(engineID)
	if err != nil {
		return nil, err
	}

	evaluatedSituations := make([]EvaluatedSituation, 0)
	for _, s := range situations {

		ruleIDs, err := situation.R().GetRules(s.ID)
		if err != nil {
			return nil, fmt.Errorf("Error geting rules for situation instance (%d, %s, %d): %s", s.ID, s.TS, s.TemplateInstanceID, err.Error())
		}

		sKnowledge, err := getSituationKnowledge(s)
		if err != nil {
			return nil, fmt.Errorf("Error geting knowledge for situation instance (%d, %s, %d): %s", s.ID, s.TS, s.TemplateInstanceID, err.Error())
		}

		_globalREngine[engineID].Reset()
		_globalREngine[engineID].GetKnowledgeBase().SetFacts(sKnowledge)

		ruleIDsInt := make([]int64, 0)
		for _, id := range ruleIDs {
			found, valid, _ := calendar.CBase().InPeriodFromCalendarID(id, s.TS)
			if !found || valid {
				ruleIDsInt = append(ruleIDsInt, id)
			}
		}
		_globalREngine[engineID].ExecuteRules(ruleIDsInt)
		agenda := _globalREngine[engineID].GetResults()

		situation.SetAsEvaluated(s.ID, s.TS, s.TemplateInstanceID)

		if agenda != nil {
			evaluatedSituations = append(evaluatedSituations, EvaluatedSituation{
				ID:                 s.ID,
				TS:                 s.TS,
				TemplateInstanceID: s.TemplateInstanceID,
				Agenda:             agenda,
			})
		}

	}

	return evaluatedSituations, nil
}

//EvaluateObjectSituations evaluates a slice of situations and return a slice with the evaluated situations
func EvaluateObjectSituations(situations []SituationToEvaluate, factObject engine.Fact, objects []map[string]interface{}, engineID string) ([]EvaluatedSituation, error) {
	var err error
	if _, ok := _globalREngine[engineID]; !ok {
		err = InitEngine(engineID)
		if err != nil {
			return nil, err
		}
	}
	err = UpdateEngine(engineID)
	if err != nil {
		return nil, err
	}

	evaluatedSituations := make([]EvaluatedSituation, 0)
	for _, s := range situations {

		ruleIDs, err := situation.R().GetRules(s.ID)
		if err != nil {
			return nil, fmt.Errorf("Error geting rules for situation instance (%d, %s, %d): %s", s.ID, s.TS, s.TemplateInstanceID, err.Error())
		}

		historyRecord, sKnowledge, err := buildObjectSituationHistoryRecord(s, factObject.ID)
		if err != nil {
			return nil, fmt.Errorf("Error building situation instance history record(%d, %s, %d): %s", s.ID, s.TS, s.TemplateInstanceID, err.Error())
		}

		ruleIDsInt := make([]int64, 0)
		for _, id := range ruleIDs {
			found, valid, _ := calendar.CBase().InPeriodFromCalendarID(id, s.TS)
			if !found || valid {
				ruleIDsInt = append(ruleIDsInt, id)
			}
		}

		for _, object := range objects {

			_globalREngine[engineID].Reset()
			_globalREngine[engineID].GetKnowledgeBase().SetFacts(sKnowledge)
			_globalREngine[engineID].GetKnowledgeBase().InsertFact(factObject.Name, object)

			_globalREngine[engineID].ExecuteRules(ruleIDsInt)

			agenda := _globalREngine[engineID].GetResults()
			if len(agenda) > 0 {

				ts := time.Now().UTC()
				err = persistObjectSituationHistoryRecord(historyRecord, factObject.ID, object, ts)
				if err != nil {
					return nil, fmt.Errorf("Error persisting situation instance history record(%d, %s, %d): %s", s.ID, s.TS, s.TemplateInstanceID, err.Error())
				}

				evaluatedSituations = append(evaluatedSituations, EvaluatedSituation{
					ID:                 s.ID,
					TS:                 ts,
					TemplateInstanceID: s.TemplateInstanceID,
					Agenda:             agenda,
				})
			}
		}
	}

	return evaluatedSituations, nil
}

//InitEngine inits an engine if it does not exist
func InitEngine(engineID string) error {
	if _, ok := _globalREngine[engineID]; ok {
		return fmt.Errorf("Engine with ID %s already exists", engineID)
	}

	_globalREngine[engineID] = ruleeng.NewRuleEngine()
	_lastUpdate[engineID] = time.Now()

	rules, err := rule.R().GetAll()
	if err != nil {
		return errors.New("Couldn't read rules " + err.Error())
	}

	for _, rule := range rules {
		if rule.Enabled {
			_globalREngine[engineID].InsertRule(rule)
		}
	}

	return nil
}

//UpdateEngine updates an engine if it exists
func UpdateEngine(engineID string) error {
	if _, ok := _globalREngine[engineID]; !ok {
		return fmt.Errorf("Engine with ID %s does not exist", engineID)
	}

	now := time.Now()
	rules, err := rule.R().GetAllModifiedFrom(_lastUpdate[engineID])
	if err != nil {
		return errors.New("Couldn't read modified rules " + err.Error())
	}
	_lastUpdate[engineID] = now
	for _, rule := range rules {
		if rule.Enabled {
			_globalREngine[engineID].InsertRule(rule)
		} else {
			_globalREngine[engineID].RemoveRule(rule.ID)
		}
	}

	return nil
}

func getSituationKnowledge(situationInstance SituationToEvaluate) (map[string]interface{}, error) {

	situationData := make(map[string]interface{}, 0)
	record, err := situation.GetFromHistory(situationInstance.ID, situationInstance.TS, situationInstance.TemplateInstanceID, false)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, errors.New("situation was not found in the history")
	}

	for factID, factTS := range record.FactsIDS {
		f, found, err := fact.R().Get(factID)
		if err != nil {
			return nil, err
		}
		if !found {
			return nil, fmt.Errorf("Fact not found with id=%d", factID)
		}
		if factTS == nil {
			return nil, fmt.Errorf("At least one fact has never been calculated, id=%d, name=%s", f.ID, f.Name)
		}

		item, _, err := fact.GetFactResultFromHistory(factID, *factTS, situationInstance.ID, situationInstance.TemplateInstanceID, false, -1)
		if err != nil {
			return nil, err
		}
		itemData, err := item.ToAbstractMap()
		if err != nil {
			return nil, err
		}

		situationData[f.Name] = itemData
	}
	for key, value := range record.Parameters {
		situationData[key] = value
	}
	for key, value := range record.EvaluatedExpressionFacts {
		situationData[key] = value
	}
	return situationData, nil
}

func buildObjectSituationHistoryRecord(situationInstance SituationToEvaluate, factObjectID int64) (situation.HistoryRecord, map[string]interface{}, error) {

	historyRecord := situation.HistoryRecord{
		ID:                 situationInstance.ID,
		TS:                 situationInstance.TS,
		TemplateInstanceID: situationInstance.TemplateInstanceID,
		FactsIDS:           make(map[int64]*time.Time, 0),
		Parameters:         situationInstance.Parameters,
	}

	situationData := make(map[string]interface{}, 0)

	for _, factID := range situationInstance.Facts {
		if factID != factObjectID {
			f, found, err := fact.R().Get(factID)
			if err != nil {
				return situation.HistoryRecord{}, nil, err
			}
			if !found {
				return situation.HistoryRecord{}, nil, fmt.Errorf("Fact not found with id=%d", factID)
			}

			item, ts, err := fact.GetFactResultFromHistory(factID, situationInstance.TS, situationInstance.ID, 0, true, -1)
			if err != nil {
				return situation.HistoryRecord{}, nil, err
			}
			itemData, err := item.ToAbstractMap()
			if err != nil {
				return situation.HistoryRecord{}, nil, err
			}

			historyRecord.FactsIDS[factID] = &ts
			situationData[f.Name] = itemData
		}
	}

	return historyRecord, situationData, nil

}

func persistObjectSituationHistoryRecord(historyRecord situation.HistoryRecord, factObjectID int64, object map[string]interface{}, ts time.Time) error {
	objectItems := make(map[string]*reader.ItemAgg, 0)
	for k, v := range object {
		objectItems[k] = &reader.ItemAgg{Value: v}
	}
	item := &reader.Item{
		Key:  object["id"].(string),
		Aggs: objectItems,
	}

	err := fact.PersistFactResult(factObjectID, ts, 0, 0, item, true)
	if err != nil {
		return fmt.Errorf("Error persisting factObject (%d): %s", factObjectID, err.Error())
	}

	historyRecord.FactsIDS[factObjectID] = &ts
	historyRecord.TS = ts

	err = situation.Persist(historyRecord, true)
	if err != nil {
		return fmt.Errorf("Error persisting historyrecord: %s", err.Error())
	}

	return nil
}
