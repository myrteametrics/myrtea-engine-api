package evaluator

import (
	"github.com/myrteametrics/myrtea-sdk/v5/ruleeng"
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
