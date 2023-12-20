package export

import (
	"bytes"
	csv2 "encoding/csv"
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
	params := CSVParameters{
		Columns: []Column{
			{Name: "a", Label: "Label A", Format: ""},
			{Name: "b", Label: "Label B", Format: ""},
			{Name: "c", Label: "Label C", Format: ""},
			{Name: "d.e", Label: "Label D.E", Format: ""},
			{Name: "date", Label: "Date", Format: "02/01/2006"},
		},
		Separator: ",",
	}
	csv, err := ConvertHitsToCSV(hits, params, true)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	t.Log("\n" + string(csv))
}

func TestWriteConvertHitsToCSV(t *testing.T) {
	hits := []reader.Hit{
		{ID: "1", Fields: map[string]interface{}{"a": "hello", "b": 20, "c": 3.123456, "d": map[string]interface{}{"e": "nested"}, "date": "2023-06-30T10:42:59.500"}},
		{ID: "2", Fields: map[string]interface{}{"b": 20, "c": 3.123456, "d": map[string]interface{}{"e": "nested"}, "date": "2023-06-30T10:42:59.500"}},
		{ID: "3", Fields: map[string]interface{}{"a": "hello", "b": 20, "c": 3.123456, "date": "2023-06-30T10:42:59.500"}},
		{ID: "1", Fields: map[string]interface{}{"a": "hello", "b": 20, "c": 3.123456, "d": map[string]interface{}{"zzz": "nested"}, "date": "2023-06-30T10:42:59.500"}},
	}
	params := CSVParameters{
		Columns: []Column{
			{Name: "a", Label: "Label A", Format: ""},
			{Name: "b", Label: "Label B", Format: ""},
			{Name: "c", Label: "Label C", Format: ""},
			{Name: "d.e", Label: "Label D.E", Format: ""},
			{Name: "date", Label: "Date", Format: "02/01/2006"},
		},
		Separator: ",",
	}
	b := new(bytes.Buffer)
	w := csv2.NewWriter(b)
	err := WriteConvertHitsToCSV(w, hits, params, true)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	t.Log("\n" + string(b.Bytes()))
}

func TestNestedMapLookup_WithEmptyKeys(t *testing.T) {
	_, err := nestedMapLookup(map[string]interface{}{}, "")
	if err == nil {
		t.FailNow()
	}
}

func TestNestedMapLookup_WithNonExistentKey(t *testing.T) {
	_, err := nestedMapLookup(map[string]interface{}{"a": "hello"}, "b")
	if err == nil {
		t.FailNow()
	}
}

func TestNestedMapLookup_WithNestedNonExistentKey(t *testing.T) {
	_, err := nestedMapLookup(map[string]interface{}{"a": map[string]interface{}{"b": "hello"}}, "a", "c")
	if err == nil {
		t.FailNow()
	}
}

func TestNestedMapLookup_WithNestedKey(t *testing.T) {
	val, err := nestedMapLookup(map[string]interface{}{"a": map[string]interface{}{"b": "hello"}}, "a", "b")
	if err != nil || val != "hello" {
		t.Error(err)
		t.FailNow()
	}
}

func TestParseDate_WithInvalidFormat(t *testing.T) {
	_, err := parseDate("2023-06-30")
	if err == nil {
		t.FailNow()
	}
}

func TestParseDate_WithValidFormat(t *testing.T) {
	_, err := parseDate("2023-06-30T10:42:59.500")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
}

func TestConvertHitsToCSV_WithEmptyHits(t *testing.T) {
	_, err := ConvertHitsToCSV([]reader.Hit{}, CSVParameters{}, true)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
}
