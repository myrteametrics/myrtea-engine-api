package evaluator

import (
	"errors"
	"fmt"

	"github.com/myrteametrics/myrtea-engine-api/v4/internals/calendar"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/fact"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/history"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/rule"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/situation"
	"github.com/myrteametrics/myrtea-sdk/v4/expression"
	"github.com/myrteametrics/myrtea-sdk/v4/ruleeng"
	"go.uber.org/zap"
)

func BuildLocalRuleEngine(engineID string) (*ruleeng.RuleEngine, error) {
	var err error
	if _, ok := GetEngine(engineID); !ok {
		err = InitEngine(engineID)
		if err != nil {
			return nil, err
		}
	}
	err = UpdateEngine(engineID)
	if err != nil {
		return nil, err
	}

	localRuleEngine, err := CloneEngine(engineID, true, false)
	if err != nil {
		return nil, err
	}

	return localRuleEngine, nil
}

func EvaluateRules(ruleEngine *ruleeng.RuleEngine, knowledgeBase map[string]interface{}, ruleIDs []int64) []ruleeng.Action {
	ruleEngine.Reset()
	ruleEngine.GetKnowledgeBase().SetFacts(knowledgeBase)
	ruleEngine.ExecuteRules(ruleIDs)
	return ruleEngine.GetResults()
}

//EvaluateSituations evaluates a slice of situations and return a slice with the evaluated situations
func EvaluateSituationsV2(situations []history.HistorySituationsV4, engineID string) ([]EvaluatedSituation, error) {
	var err error
	if _, ok := GetEngine(engineID); !ok {
		err = InitEngine(engineID)
		if err != nil {
			return nil, err
		}
	}
	err = UpdateEngine(engineID)
	if err != nil {
		return nil, err
	}

	localRuleEngine, err := CloneEngine(engineID, true, false)
	if err != nil {
		return nil, err
	}

	evaluatedSituations := make([]EvaluatedSituation, 0)
	for _, s := range situations {

		ruleIDs, err := situation.R().GetRules(s.ID)
		if err != nil {
			return nil, fmt.Errorf("error geting rules for situation instance (%d, %s, %d): %s", s.ID, s.Ts, s.SituationInstanceID, err.Error())
		}

		sKnowledge, err := GetSituationKnowledgeV2(s)
		if err != nil {
			return nil, fmt.Errorf("error geting knowledge for situation instance (%d, %s, %d): %s", s.ID, s.Ts, s.SituationInstanceID, err.Error())
		}
		//Add date keywords in sKnowledge
		for key, value := range expression.GetDateKeywords(s.Ts) {
			sKnowledge[key] = value
		}

		localRuleEngine.Reset()
		localRuleEngine.GetKnowledgeBase().SetFacts(sKnowledge)

		ruleIDsInt := make([]int64, 0)
		for _, id := range ruleIDs {
			r, found, err := rule.R().Get(id)
			if err != nil {
				zap.L().Error("Get Rule", zap.Int64("id", id), zap.Error(err))
				continue
			}
			if !found {
				zap.L().Warn("Rule is missing", zap.Int64("id", id))
				continue
			}

			cfound, valid, _ := calendar.CBase().InPeriodFromCalendarID(int64(r.CalendarID), s.Ts)
			if !cfound || valid {
				ruleIDsInt = append(ruleIDsInt, id)
			}
		}

		localRuleEngine.ExecuteRules(ruleIDsInt)
		agenda := localRuleEngine.GetResults()

		situation.SetAsEvaluated(s.ID, s.Ts, s.SituationInstanceID)

		if agenda != nil {
			evaluatedSituations = append(evaluatedSituations, EvaluatedSituation{
				ID:                 s.ID,
				TS:                 s.Ts,
				TemplateInstanceID: s.SituationInstanceID,
				Agenda:             agenda,
			})
		}

	}

	return evaluatedSituations, nil
}

func GetSituationKnowledgeV2(historySituation history.HistorySituationsV4) (map[string]interface{}, error) {

	situationData := make(map[string]interface{}, 0)
	record, err := situation.GetFromHistory(historySituation.ID, historySituation.Ts, historySituation.SituationInstanceID, false)
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

		item, _, err := fact.GetFactResultFromHistory(factID, *factTS, historySituation.ID, historySituation.SituationInstanceID, false, -1)
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
