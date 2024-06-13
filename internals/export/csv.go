package export

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/myrteametrics/myrtea-engine-api/v5/internals/reader"
	"go.uber.org/zap"
)

// WriteConvertHitsToCSV writes hits to CSV
func WriteConvertHitsToCSV(w *csv.Writer, hits []reader.Hit, params CSVParameters, writeHeader bool) error {
	if len(params.Separator) == 1 {
		w.Comma, _ = utf8.DecodeRune([]byte(params.Separator))
		if w.Comma == utf8.RuneError {
			w.Comma = ','
		}
	} else {
		w.Comma = ','
	}

	// avoid to print header when labels are empty
	if writeHeader && len(params.Columns) > 0 {
		w.Write(params.GetColumnsLabel())
	}

	for _, hit := range hits {
		record := make([]string, 0)
		for _, column := range params.Columns {
			value, err := nestedMapLookup(hit.Fields, strings.Split(column.Name, ".")...)
			if err != nil {
				record = append(record, "")
				continue
			}

			switch v := value.(type) {
			case []interface{}:
				if params.ListSeparator != "" {
					value = convertSliceToString(v, column.Format, params.ListSeparator)
				}
			default:
				if column.Format != "" {
					value = formatValue(value, column.Format)
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

// formatValue Apply the specified format to a given value.
func formatValue(value interface{}, format string) interface{} {
	switch v := value.(type) {
	case time.Time:
		return v.Format(format)
	case string:
		date, err := parseDate(v)
		if err == nil {
			return date.Format(format)
		}
		zap.L().Error("Failed to parse date string:", zap.Any("value", v), zap.Error(err))
	}
	return value
}

// convertSliceToString Take a slice []interface{}, format and returns the elements separated by custom separator.
func convertSliceToString(slice []interface{}, format string, separator string) string {
	var strValues []string
	for _, elem := range slice {
		if format != "" {
			elem = formatValue(elem, format)
		}
		strValues = append(strValues, fmt.Sprintf("%v", elem))
	}
	return strings.Join(strValues, separator)
}
