package history

import (
	"testing"
)

func TestGetHistoryFactLast(t *testing.T) {
	var factId int64 = 1
	builder := HistoryFactsBuilder{}.GetHistoryFactLast(factId)
	t.Fail()
	t.Log(builder.ToSql())
}

func TestGetHistoryFacts(t *testing.T) {
	var historyFactsIds []int64 = []int64{1, 2, 3}
	builder := HistoryFactsBuilder{}.GetHistoryFacts(historyFactsIds)
	t.Fail()
	t.Log(builder.ToSql())
}
