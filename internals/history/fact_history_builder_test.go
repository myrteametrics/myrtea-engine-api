package history

import (
	"testing"
)

func TestGetHistoryFactLast(t *testing.T) {
	t.SkipNow()
	builder := HistoryFactsBuilder{}.GetHistoryFactLast(4, 109, 19)
	t.Fail()
	t.Log(builder.ToSql())
}

func TestGetHistoryFacts(t *testing.T) {
	t.SkipNow()
	var historyFactsIds []int64 = []int64{1, 2, 3}
	builder := HistoryFactsBuilder{}.GetHistoryFacts(historyFactsIds)
	t.Fail()
	t.Log(builder.ToSql())
}
