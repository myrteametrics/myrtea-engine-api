package history

import (
	"testing"
)

func TestGetHistorySituationFacts(t *testing.T) {
	t.SkipNow()

	historySituationsIds := []int64{1, 2, 3, 4, 5}
	builder := HistorySituationFactsBuilder{}.GetHistorySituationFacts(historySituationsIds)

	t.Fail()
	t.Log(builder.ToSql())
}
