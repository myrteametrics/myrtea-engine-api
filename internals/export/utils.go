package export

import "github.com/myrteametrics/myrtea-engine-api/v5/internals/notifier/notification"

type CSVParameters struct {
	Columns       []Column `json:"columns"`
	Separator     string   `json:"separator"`
	Limit         int64    `json:"limit"`
	ListSeparator string   `json:"listSeparator`
}

type Column struct {
	Name   string `json:"name"`
	Label  string `json:"label"`
	Format string `json:"format" default:""`
}

// Equals compares two Column
func (p Column) Equals(column Column) bool {
	if p.Name != column.Name {
		return false
	}
	if p.Label != column.Label {
		return false
	}
	if p.Format != column.Format {
		return false
	}
	return true
}

// Equals compares two CSVParameters
func (p CSVParameters) Equals(params CSVParameters) bool {
	if p.Separator != params.Separator {
		return false
	}
	if p.Limit != params.Limit {
		return false
	}
	for i, column := range p.Columns {
		if !column.Equals(params.Columns[i]) {
			return false
		}
	}
	return true
}

// GetColumnsLabel returns the label of the columns
func (p CSVParameters) GetColumnsLabel() []string {
	columns := make([]string, 0)
	for _, column := range p.Columns {
		columns = append(columns, column.Label)
	}
	return columns
}

// createExportNotification creates an export notification using given parameters
func createExportNotification(status int, item *WrapperItem) ExportNotification {
	return ExportNotification{
		BaseNotification: notification.BaseNotification{
			Id:         0,
			IsRead:     false,
			Type:       "ExportNotification",
			Persistent: false,
		},
		Export: *item,
		Status: status,
	}
}
