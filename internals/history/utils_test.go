package history

import (
	"testing"

	"github.com/myrteametrics/myrtea-engine-api/v5/internals/situation"
	"github.com/myrteametrics/myrtea-sdk/v4/expression"
)

func TestEvaluateExpressionFactsChain(t *testing.T) {
	expressionFacts := []situation.ExpressionFact{
		{Name: "expf1", Expression: "fact1.aggs.doc_count.value + fact2.aggs.doc_count.value"},
		{Name: "expf2", Expression: "expf1 + fact3.aggs.doc_count.value"},
		{Name: "expf3", Expression: "expf1 + expf2"},
	}
	data := map[string]interface{}{
		"fact1": map[string]interface{}{"aggs": map[string]interface{}{"doc_count": map[string]interface{}{"value": 10}}},
		"fact2": map[string]interface{}{"aggs": map[string]interface{}{"doc_count": map[string]interface{}{"value": 10}}},
		"fact3": map[string]interface{}{"aggs": map[string]interface{}{"doc_count": map[string]interface{}{"value": 10}}},
	}

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

		t.Log(data)
		t.Log(expressionFactsEvaluated)
	}
}
