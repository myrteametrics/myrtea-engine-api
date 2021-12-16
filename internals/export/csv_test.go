package export

import (
	"testing"

	"github.com/myrteametrics/myrtea-engine-api/v4/internals/reader"
)

func TestConvertHitsToCSV(t *testing.T) {
	hits := []reader.Hit{
		{ID: "1", Fields: map[string]interface{}{"a": "hello", "b": 20, "c": 3.123456, "d": map[string]interface{}{"e": "nested"}}},
		{ID: "2", Fields: map[string]interface{}{"b": 20, "c": 3.123456, "d": map[string]interface{}{"e": "nested"}}},
		{ID: "3", Fields: map[string]interface{}{"a": "hello", "b": 20, "c": 3.123456}},
		{ID: "1", Fields: map[string]interface{}{"a": "hello", "b": 20, "c": 3.123456, "d": map[string]interface{}{"zzz": "nested"}}},
	}
	columns := []string{"a", "b", "c", "d.e"}
	columnsLabel := []string{"Label A", "Label B", "Label C", "Label D.E"}
	csv, err := ConvertHitsToCSV(hits, columns, columnsLabel, ',')
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	t.Log("\n" + string(csv))
}
