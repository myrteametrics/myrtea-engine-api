package export

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strings"
	"time"

	"github.com/myrteametrics/myrtea-engine-api/v5/internals/reader"
	"go.uber.org/zap"
)

// WriteConvertHitsToCSV writes hits to CSV
func WriteConvertHitsToCSV(w *csv.Writer, hits []reader.Hit, params CSVParameters, writeHeader bool) error {
	w.Comma = params.Separator

	// avoid to print header when labels are empty
	if writeHeader && len(params.Columns) > 0 {
		w.Write(params.GetColumnsLabel())
	}

	for _, hit := range hits {
		record := make([]string, 0)
		for _, column := range params.Columns {
			value, err := nestedMapLookup(hit.Fields, strings.Split(column.Name, ".")...)
			if err != nil {
				value = ""
			} else if column.Format != "" {
				if date, ok := value.(time.Time); ok {
					value = date.Format(column.Format)
				} else if dateStr, ok := value.(string); ok {
					date, err := parseDate(dateStr)
					if err != nil {
						zap.L().Error("Failed to parse date string:", zap.Any(":", dateStr), zap.Error(err))
					} else {
						value = date.Format(column.Format)
					}
				}
			}
			record = append(record, fmt.Sprintf("%v", value))
		}
		w.Write(record)
	}

	w.Flush()
	return w.Error()
}

// ConvertHitsToCSV converts hits to CSV
func ConvertHitsToCSV(hits []reader.Hit, params CSVParameters, writeHeader bool) ([]byte, error) {
	b := new(bytes.Buffer)
	w := csv.NewWriter(b)
	err := WriteConvertHitsToCSV(w, hits, params, writeHeader)

	if err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

// nestedMapLookup looks up a nested map item
func nestedMapLookup(m map[string]interface{}, ks ...string) (rval interface{}, err error) {
	var ok bool
	if len(ks) == 0 {
		return nil, fmt.Errorf("nestedMapLookup needs at least one key")
	}
	if rval, ok = m[strings.TrimSpace(ks[0])]; !ok {
		return nil, fmt.Errorf("key not found; remaining keys: %v", ks)
	} else if len(ks) == 1 {
		return rval, nil
	} else if m, ok = rval.(map[string]interface{}); !ok {
		return nil, fmt.Errorf("malformed structure at %#v", rval)
	} else {
		return nestedMapLookup(m, ks[1:]...)
	}
}

// parseDate parses a date string
func parseDate(dateStr string) (time.Time, error) {
	formats := []string{
		"2006-01-02T15:04:05.999",
		"2006-01-02T15:4:05.999",
		"2006-01-02T15:04:5.999",
		"2006-01-02T15:4:5.999",
	}

	for _, format := range formats {
		if date, err := time.Parse(format, dateStr); err == nil {
			return date, nil
		}
	}

	return time.Time{}, fmt.Errorf("failed to parse date string: %s", dateStr)
}
