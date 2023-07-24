package export

import (
	"testing"

	"github.com/myrteametrics/myrtea-engine-api/v5/internals/reader"
)

func TestConvertHitsToCSV(t *testing.T) {
	hits := []reader.Hit{
		{ID: "1", Fields: map[string]interface{}{"a": "hello", "b": 20, "c": 3.123456, "d": map[string]interface{}{"e": "nested"}, "date": "2023-06-30T10:42:59.500"}},
		{ID: "2", Fields: map[string]interface{}{"b": 20, "c": 3.123456, "d": map[string]interface{}{"e": "nested"}, "date": "2023-06-30T10:42:59.500"}},
		{ID: "3", Fields: map[string]interface{}{"a": "hello", "b": 20, "c": 3.123456, "date": "2023-06-30T10:42:59.500"}},
		{ID: "1", Fields: map[string]interface{}{"a": "hello", "b": 20, "c": 3.123456, "d": map[string]interface{}{"zzz": "nested"}, "date": "2023-06-30T10:42:59.500"}},
	}
	columns := []string{"a", "b", "c", "d.e", "date"}
	columnsLabel := []string{"Label A", "Label B", "Label C", "Label D.E", "Date"}
	formateColumnsData := map[string]string{
		"date": "02/01/2006",
	}
	csv, err := ConvertHitsToCSV(hits, columns, columnsLabel, formateColumnsData, ',')
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	t.Log("\n" + string(csv))
}
