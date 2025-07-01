package processor

import (
	"reflect"
	"testing"

	"github.com/myrteametrics/myrtea-sdk/v5/engine"
)

func TestFilter(t *testing.T) {
	doc := map[string]interface{}{
		"a": "value_a",
		"b": map[string]interface{}{
			"c": "value_b_c",
		},
		"d": map[string]interface{}{
			"e": 5,
		},
		"f": map[string]interface{}{
			"g": map[string]interface{}{
				"h": map[string]interface{}{
					"i": map[string]interface{}{
						"j": 5,
					},
					"k": map[string]interface{}{
						"l": 5,
						"m": 10,
						"n": 10,
					},
				},
			},
		},
	}

	testFilterSource(doc, "a", map[string]interface{}{"a": "value_a"}, t)
	// testFilterSource(doc, "b.c", map[string]interface{}{"b": map[string]interface{}{"c": "value_b_c"}}, t)
	// testFilterSource(doc, "d.e", map[string]interface{}{"d": map[string]interface{}{"e": 5}}, t)
}

func testFilterSource(source map[string]interface{}, selectTerm string, expected map[string]interface{}, t *testing.T) {
	f := engine.Fact{Intent: &engine.IntentFragment{Operator: engine.Select, Term: selectTerm}}
	filteredSource := filterSource(f, source)
	if !reflect.DeepEqual(filteredSource, expected) {
		t.Error("invalid filtered source", selectTerm)
		t.Log("Filtered", filteredSource)
		t.Log("Expected", expected)
	}
}

func TestApplyCondition(t *testing.T) {
	doc := map[string]interface{}{
		"a": "value_a",
		"b": map[string]interface{}{
			"c": "value_b_c",
		},
		"d": map[string]interface{}{
			"e": 5,
		},
		"f": map[string]interface{}{
			"g": map[string]interface{}{
				"h": map[string]interface{}{
					"i": map[string]interface{}{
						"j": 5,
					},
					"k": map[string]interface{}{
						"l": 5,
						"m": 10,
						"n": 10,
					},
				},
			},
		},
	}

	testApplyCondition(doc, &engine.LeafConditionFragment{Operator: engine.Exists, Field: "a"}, true, t)
	testApplyCondition(doc, &engine.LeafConditionFragment{Operator: engine.Exists, Field: "b.c"}, true, t)
	testApplyCondition(doc, &engine.LeafConditionFragment{Operator: engine.Exists, Field: "f.g.h.i.j"}, true, t)
	testApplyCondition(doc, &engine.LeafConditionFragment{Operator: engine.Exists, Field: "f.g.h.k.l"}, true, t)
	testApplyCondition(doc, &engine.LeafConditionFragment{Operator: engine.Exists, Field: "f.g.h.k.m"}, true, t)
	testApplyCondition(doc, &engine.LeafConditionFragment{Operator: engine.Exists, Field: "f.g.h.k.n"}, true, t)
	testApplyCondition(doc, &engine.LeafConditionFragment{Operator: engine.Exists, Field: "b"}, false, t)
	testApplyCondition(doc, &engine.LeafConditionFragment{Operator: engine.Exists, Field: "b.c.d"}, false, t)

	testApplyCondition(doc, &engine.LeafConditionFragment{Operator: engine.For, Field: "a", Value: "value_a"}, true, t)
	testApplyCondition(doc, &engine.LeafConditionFragment{Operator: engine.For, Field: "a", Value: "other"}, false, t)
	testApplyCondition(doc, &engine.LeafConditionFragment{Operator: engine.For, Field: "b.c", Value: "value_b_c"}, true, t)
	testApplyCondition(doc, &engine.LeafConditionFragment{Operator: engine.For, Field: "b.c", Value: "other"}, false, t)
	testApplyCondition(doc, &engine.LeafConditionFragment{Operator: engine.For, Field: "b", Value: "value_b_c"}, false, t)
	testApplyCondition(doc, &engine.LeafConditionFragment{Operator: engine.For, Field: "b", Value: "other"}, false, t)
	testApplyCondition(doc, &engine.LeafConditionFragment{Operator: engine.For, Field: "b.c.d", Value: "value_b_c"}, false, t)
	testApplyCondition(doc, &engine.LeafConditionFragment{Operator: engine.For, Field: "b.c.d", Value: "other"}, false, t)

	//testApplyCondition(doc, &engine.LeafConditionFragment{Operator: engine.Between, Field: "d.e", Value: 1, Value2: 10}, true, t)
	//testApplyCondition(doc, &engine.LeafConditionFragment{Operator: engine.Between, Field: "d.e", Value: 8, Value2: 10}, false, t)

	testApplyCondition(doc, &engine.BooleanFragment{Operator: engine.And, Fragments: []engine.ConditionFragment{
		&engine.LeafConditionFragment{Operator: engine.Exists, Field: "a"},
		&engine.LeafConditionFragment{Operator: engine.Exists, Field: "b.c"},
	}}, true, t)
	testApplyCondition(doc, &engine.BooleanFragment{Operator: engine.And, Fragments: []engine.ConditionFragment{
		&engine.LeafConditionFragment{Operator: engine.Exists, Field: "a"},
		&engine.LeafConditionFragment{Operator: engine.Exists, Field: "b.c.d"},
	}}, false, t)

	testApplyCondition(doc, &engine.BooleanFragment{Operator: engine.Or, Fragments: []engine.ConditionFragment{
		&engine.LeafConditionFragment{Operator: engine.Exists, Field: "a"},
		&engine.LeafConditionFragment{Operator: engine.Exists, Field: "b.c"},
	}}, true, t)
	testApplyCondition(doc, &engine.BooleanFragment{Operator: engine.Or, Fragments: []engine.ConditionFragment{
		&engine.LeafConditionFragment{Operator: engine.Exists, Field: "a"},
		&engine.LeafConditionFragment{Operator: engine.Exists, Field: "b.c.d"},
	}}, true, t)
	testApplyCondition(doc, &engine.BooleanFragment{Operator: engine.Or, Fragments: []engine.ConditionFragment{
		&engine.LeafConditionFragment{Operator: engine.Exists, Field: "a.z"},
		&engine.LeafConditionFragment{Operator: engine.Exists, Field: "b.c.d"},
	}}, false, t)

	testApplyCondition(doc, &engine.BooleanFragment{Operator: engine.Not, Fragments: []engine.ConditionFragment{
		&engine.LeafConditionFragment{Operator: engine.Exists, Field: "a"},
		&engine.LeafConditionFragment{Operator: engine.Exists, Field: "b.c"},
	}}, false, t)
	testApplyCondition(doc, &engine.BooleanFragment{Operator: engine.Not, Fragments: []engine.ConditionFragment{
		&engine.LeafConditionFragment{Operator: engine.Exists, Field: "a"},
		&engine.LeafConditionFragment{Operator: engine.Exists, Field: "b.c.d"},
	}}, false, t)
	testApplyCondition(doc, &engine.BooleanFragment{Operator: engine.Not, Fragments: []engine.ConditionFragment{
		&engine.LeafConditionFragment{Operator: engine.Exists, Field: "a.z"},
		&engine.LeafConditionFragment{Operator: engine.Exists, Field: "b.c.d"},
	}}, true, t)
}

func testApplyCondition(doc map[string]interface{}, c engine.ConditionFragment, result bool, t *testing.T) {
	valid := applyCondition(c, doc)
	if valid != result {
		t.Error(c)
	}
}
