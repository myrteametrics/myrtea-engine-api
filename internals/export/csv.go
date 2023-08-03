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

func ConvertHitsToCSV(hits []reader.Hit, columns []string, columnsLabel []string, formatColumnsData map[string]string, separator rune) ([]byte, error) {
	b := new(bytes.Buffer)
	w := csv.NewWriter(b)
	w.Comma = separator
	w.Write(columnsLabel)
	for _, hit := range hits {
		record := make([]string, 0)
		for _, column := range columns {
			value, err := nestedMapLookup(hit.Fields, strings.Split(column, ".")...)
			if err != nil {
				value = ""
			} else if format, ok := formatColumnsData[column]; ok {
				if date, ok := value.(time.Time); ok {
					value = date.Format(format)
				} else if dateStr, ok := value.(string); ok {
					date, err := parseDate(dateStr)
					if err != nil {
						zap.L().Error("Failed to parse date string:", zap.Any(":", dateStr), zap.Error(err))
					} else {
						value = date.Format(format)
					}
				}
			}
			record = append(record, fmt.Sprintf("%v", value))
		}
		w.Write(record)
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

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
