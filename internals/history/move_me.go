package history

import (
	"errors"

	"github.com/myrteametrics/myrtea-engine-api/v4/internals/situation"
	"github.com/myrteametrics/myrtea-sdk/v4/expression"
)

func ExtractSituationData(situationID int64, situationInstanceID int64) (situation.Situation, map[string]string, error) {
	parameters := make(map[string]string)

	s, found, err := situation.R().Get(situationID)
	if err != nil {
		return situation.Situation{}, make(map[string]string), err
	}
	if !found {
		return situation.Situation{}, make(map[string]string), errors.New("situation not found")
	}
	for k, v := range s.Parameters {
		parameters[k] = v
	}

	if s.IsTemplate {
		si, found, err := situation.R().GetTemplateInstance(situationInstanceID)
		if err != nil {
			return situation.Situation{}, make(map[string]string), err
		}
		if !found {
			return situation.Situation{}, make(map[string]string), errors.New("situation instance not found")
		}
		for k, v := range si.Parameters {
			parameters[k] = v
		}
	}

	return s, parameters, nil
}

func EvaluateExpressionFacts(expressionFacts []situation.ExpressionFact, data map[string]interface{}) map[string]interface{} {
	expressionFactsEvaluated := make(map[string]interface{})
	for _, expressionFact := range expressionFacts {
		result, err := expression.Process(expression.LangEval, expressionFact.Expression, data)
		if err != nil {
			continue
		}
		if expression.IsInvalidNumber(result) {
			continue
		}

		data[expressionFact.Name] = result
		expressionFactsEvaluated[expressionFact.Name] = result
	}
	return expressionFactsEvaluated
}
