package export

import (
	"testing"

	"github.com/myrteametrics/myrtea-sdk/v4/engine"
)

func TestExportFactHits(t *testing.T) {
	f := engine.Fact{
		ID:               1,
		Name:             "test",
		IsObject:         false,
		IsTemplate:       false,
		Model:            "",
		CalculationDepth: 2,
		Intent:           &engine.IntentFragment{Name: "", Operator: engine.Count, Term: ""},
		Dimensions:       []*engine.DimensionFragment{{Name: "", Operator: engine.By, Term: "", Size: 10000}},
		Condition: &engine.BooleanFragment{Operator: engine.And, Fragments: []engine.ConditionFragment{
			&engine.LeafConditionFragment{Operator: engine.Exists, Field: ""},
			&engine.LeafConditionFragment{Operator: engine.Exists, Field: ""},
		}},
	}
	_ = f
	// hits, err := QueryFactHits(1)
	// if err != nil {
	// 	t.Log(err)
	// 	t.FailNow()
	// }
	// t.Log(hits)
}
