package history

import (
	"testing"
	"time"

	"github.com/myrteametrics/myrtea-engine-api/v5/internal/situation"
	"github.com/myrteametrics/myrtea-sdk/v5/expression"
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

func TestGetTodayTimeRange(t *testing.T) {
	todayStart, tomorrowStart := getTodayTimeRange()

	now := time.Now().UTC().Truncate(24 * time.Hour)
	expectedTodayStart := now.Format("2006-01-02 15:04:05")
	expectedTomorrowStart := now.Add(24 * time.Hour).Format("2006-01-02 15:04:05")

	if todayStart != expectedTodayStart {
		t.Errorf("Expected today's start to be %s, but got %s", expectedTodayStart, todayStart)
	}

	if tomorrowStart != expectedTomorrowStart {
		t.Errorf("Expected tomorrow's start to be %s, but got %s", expectedTomorrowStart, tomorrowStart)
	}
}

func TestGetStartDate30DaysAgo(t *testing.T) {
	date30DaysAgo := getStartDate30DaysAgo()

	expectedDate := time.Now().UTC().AddDate(0, 0, -30).Format("2006-01-02 15:04:05")

	if date30DaysAgo != expectedDate {
		t.Errorf("Expected date 30 days ago to be %s, but got %s", expectedDate, date30DaysAgo)
	}
}
