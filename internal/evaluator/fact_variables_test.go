package evaluator

import (
	"reflect"
	"testing"
	"time"

	"github.com/myrteametrics/myrtea-sdk/v5/engine"
	"github.com/myrteametrics/myrtea-sdk/v5/expression"
)

func TestScanFactConditionVariables(t *testing.T) {
	condition := &engine.BooleanFragment{
		Operator: engine.And,
		Fragments: []engine.ConditionFragment{
			&engine.LeafConditionFragment{Operator: engine.For, Field: "supervision", Value: "supervision"},
			&engine.LeafConditionFragment{Operator: engine.For, Field: "flux", Value: `"distri"`},
			&engine.LeafConditionFragment{Operator: engine.Exists, Field: "source_date"},
			&engine.LeafConditionFragment{Operator: engine.Between, Field: "source_date", Value: "startofday", Value2: "now"},
			&engine.LeafConditionFragment{Operator: engine.OptionalFor, Field: "cible", Value: "cible"},
			&engine.LeafConditionFragment{Operator: engine.From, Field: "sink_date", Value: "known", TimeZone: "zone"},
		},
	}
	parameters := expression.GetDateKeywords(time.Now())
	parameters["known"] = "2026-01-01"

	expected := []FactVariable{
		{Name: "cible", Optional: true},
		{Name: "supervision", Optional: false},
		{Name: "zone", Optional: false},
	}

	variables := ScanFactConditionVariables(condition, parameters)
	if !reflect.DeepEqual(variables, expected) {
		t.Errorf("Expected variables %v, got %v", expected, variables)
	}
}

func TestScanFactConditionVariablesMandatoryWins(t *testing.T) {
	condition := &engine.BooleanFragment{
		Operator: engine.Or,
		Fragments: []engine.ConditionFragment{
			&engine.LeafConditionFragment{Operator: engine.OptionalFor, Field: "a", Value: "shared"},
			&engine.LeafConditionFragment{Operator: engine.For, Field: "b", Value: "shared"},
		},
	}

	expected := []FactVariable{{Name: "shared", Optional: false}}

	variables := ScanFactConditionVariables(condition, map[string]interface{}{})
	if !reflect.DeepEqual(variables, expected) {
		t.Errorf("Expected variables %v, got %v", expected, variables)
	}
}

func TestScanFactConditionVariablesNestedAndFunctions(t *testing.T) {
	condition := &engine.LeafConditionFragment{
		Operator: engine.For,
		Field:    "code",
		Value:    `context.code + "_" + replace(suffix, "-", "_")`,
	}
	parameters := map[string]interface{}{
		"context": map[string]interface{}{"code": "x"},
	}

	expected := []FactVariable{{Name: "suffix", Optional: false}}

	variables := ScanFactConditionVariables(condition, parameters)
	if !reflect.DeepEqual(variables, expected) {
		t.Errorf("Expected variables %v, got %v", expected, variables)
	}
}

func TestScanFactConditionVariablesNone(t *testing.T) {
	condition := &engine.LeafConditionFragment{Operator: engine.Between, Field: "date", Value: "startofday", Value2: "now"}

	variables := ScanFactConditionVariables(condition, expression.GetDateKeywords(time.Now()))
	if len(variables) != 0 {
		t.Errorf("Expected no variable, got %v", variables)
	}
}
