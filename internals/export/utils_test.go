package export

import (
	"github.com/myrteametrics/myrtea-sdk/v4/expression"
	"testing"
)

func TestEquals(t *testing.T) {
	p1 := CSVParameters{}
	p2 := CSVParameters{}
	expression.AssertEqual(t, p1.Equals(p2), false)
	expression.AssertEqual(t, p1.Equals(p1), true)

	// make a full test with all variables in parameters filled
	params3 := CSVParameters{
		Columns:           []string{"col1", "col2"},
		ColumnsLabel:      []string{"col1", "col2"},
		FormatColumnsData: map[string]string{"col1": "format1", "col2": "format2"},
		Separator:         ';',
		Limit:             10,
		ChunkSize:         100,
	}
	expression.AssertEqual(t, params3.Equals(p2), false)
	expression.AssertEqual(t, params3.Equals(params3), true)

	// test separator
	p1 = CSVParameters{Separator: ';'}
	p2 = CSVParameters{Separator: ','}
	expression.AssertEqual(t, p1.Equals(p2), false)

	// test limit
	p1 = CSVParameters{Limit: 10}
	p2 = CSVParameters{Limit: 101}
	expression.AssertEqual(t, p1.Equals(p2), false)

	// test chunk size
	p1 = CSVParameters{ChunkSize: 100}
	p2 = CSVParameters{ChunkSize: 10}
	expression.AssertEqual(t, p1.Equals(p2), false)

	// test columns size
	p1 = CSVParameters{Columns: []string{"col1", "col2"}}
	p2 = CSVParameters{Columns: []string{"col1", "col2", "col3"}}
	expression.AssertEqual(t, p1.Equals(p2), false)

	// test columns values
	p1 = CSVParameters{Columns: []string{"col1", "col2"}}
	p2 = CSVParameters{Columns: []string{"col1", "col3"}}
	expression.AssertEqual(t, p1.Equals(p2), false)

	// test columnsLabel size
	p1 = CSVParameters{ColumnsLabel: []string{"col1", "col2"}}
	p2 = CSVParameters{ColumnsLabel: []string{"col1", "col2", "col3"}}
	expression.AssertEqual(t, p1.Equals(p2), false)

	// test columnsLabel values
	p1 = CSVParameters{ColumnsLabel: []string{"col1", "col2"}}
	p2 = CSVParameters{ColumnsLabel: []string{"col1", "col3"}}
	expression.AssertEqual(t, p1.Equals(p2), false)

	// test formatColumnsData size
	p1 = CSVParameters{FormatColumnsData: map[string]string{"col1": "format1", "col2": "format2"}}
	p2 = CSVParameters{FormatColumnsData: map[string]string{"col1": "format1", "col2": "format2", "col3": "format3"}}
	expression.AssertEqual(t, p1.Equals(p2), false)

	// test formatColumnsData values
	p1 = CSVParameters{FormatColumnsData: map[string]string{"col1": "format1", "col2": "format2"}}
	p2 = CSVParameters{FormatColumnsData: map[string]string{"col1": "format1", "col2": "format3"}}
	expression.AssertEqual(t, p1.Equals(p2), false)

	// test formatColumnsData keys
	p1 = CSVParameters{FormatColumnsData: map[string]string{"col1": "format1", "col2": "format2"}}
	p2 = CSVParameters{FormatColumnsData: map[string]string{"col1": "format1", "col3": "format2"}}
	expression.AssertEqual(t, p1.Equals(p2), false)
}
