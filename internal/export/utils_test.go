package export

import (
	"github.com/myrteametrics/myrtea-sdk/v5/expression"
	"testing"
)

func TestColumnEquals_WithDifferentName(t *testing.T) {
	column1 := Column{Name: "name1", Label: "label", Format: "format"}
	column2 := Column{Name: "name2", Label: "label", Format: "format"}
	expression.AssertEqual(t, column1.Equals(column2), false)
}

func TestColumnEquals_WithDifferentLabel(t *testing.T) {
	column1 := Column{Name: "name", Label: "label1", Format: "format"}
	column2 := Column{Name: "name", Label: "label2", Format: "format"}
	expression.AssertEqual(t, column1.Equals(column2), false)
}

func TestColumnEquals_WithDifferentFormat(t *testing.T) {
	column1 := Column{Name: "name", Label: "label", Format: "format1"}
	column2 := Column{Name: "name", Label: "label", Format: "format2"}
	expression.AssertEqual(t, column1.Equals(column2), false)
}

func TestColumnEquals_WithSameValues(t *testing.T) {
	column1 := Column{Name: "name", Label: "label", Format: "format"}
	column2 := Column{Name: "name", Label: "label", Format: "format"}
	expression.AssertEqual(t, column1.Equals(column2), true)
}

func TestCSVParametersEquals_WithDifferentSeparator(t *testing.T) {
	params1 := CSVParameters{Separator: ",", Limit: 10, Columns: []Column{{Name: "name", Label: "label", Format: "format"}}}
	params2 := CSVParameters{Separator: ";", Limit: 10, Columns: []Column{{Name: "name", Label: "label", Format: "format"}}}
	expression.AssertEqual(t, params1.Equals(params2), false)
}

func TestCSVParametersEquals_WithDifferentLimit(t *testing.T) {
	params1 := CSVParameters{Separator: ",", Limit: 10, Columns: []Column{{Name: "name", Label: "label", Format: "format"}}}
	params2 := CSVParameters{Separator: ",", Limit: 20, Columns: []Column{{Name: "name", Label: "label", Format: "format"}}}
	expression.AssertEqual(t, params1.Equals(params2), false)
}

func TestCSVParametersEquals_WithDifferentColumns(t *testing.T) {
	params1 := CSVParameters{Separator: ",", Limit: 10, Columns: []Column{{Name: "name1", Label: "label", Format: "format"}}}
	params2 := CSVParameters{Separator: ",", Limit: 10, Columns: []Column{{Name: "name2", Label: "label", Format: "format"}}}
	expression.AssertEqual(t, params1.Equals(params2), false)
}

func TestCSVParametersEquals_WithSameValues(t *testing.T) {
	params1 := CSVParameters{Separator: ",", Limit: 10, Columns: []Column{{Name: "name", Label: "label", Format: "format"}}}
	params2 := CSVParameters{Separator: ",", Limit: 10, Columns: []Column{{Name: "name", Label: "label", Format: "format"}}}
	expression.AssertEqual(t, params1.Equals(params2), true)
}

func TestGetColumnsLabel_WithNoColumns(t *testing.T) {
	params := CSVParameters{Separator: ",", Limit: 10, Columns: []Column{}}
	labels := params.GetColumnsLabel()
	expression.AssertEqual(t, len(labels), 0)
}

func TestGetColumnsLabel_WithColumns(t *testing.T) {
	params := CSVParameters{Separator: ",", Limit: 10, Columns: []Column{{Name: "name1", Label: "label1", Format: "format1"}, {Name: "name2", Label: "label2", Format: "format2"}}}
	labels := params.GetColumnsLabel()
	expression.AssertEqual(t, len(labels), 2)
	expression.AssertEqual(t, labels[0], "label1")
	expression.AssertEqual(t, labels[1], "label2")
}
