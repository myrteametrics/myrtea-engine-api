package export

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/myrteametrics/myrtea-engine-api/v5/internals/reader"
	"go.uber.org/zap"
)

func ConvertHitsToCSV(hits []reader.Hit, columns []string, columnsLabel []string, separator rune) ([]byte, error) {
	b := new(bytes.Buffer)
	w := csv.NewWriter(b)
	w.Comma = separator

	dateRegex, err := regexp.Compile(`\d{4}-\d{2}-\d{2}T\d{1,2}:\d{1,2}:\d{1,2}\.\d{3}`)
	if err != nil {
		zap.L().Error("Failed to compile date regex", zap.Error(err))
		return nil, err
	}

	w.Write(columnsLabel)
	for _, hit := range hits {
		record := make([]string, 0)
		for _, column := range columns {
			value, err := nestedMapLookup(hit.Fields, strings.Split(column, ".")...)
			if err != nil {
				value = ""
			} else if strValue, ok := value.(string); ok && dateRegex.MatchString(strValue) {
				t, err := time.Parse("2006-01-02T15:04:05.000", strValue)
				// Si cela échoue, on essayez avec un format qui accepte une ou deux occurrences
				if err != nil {
					t, err = time.Parse("2006-01-02T15:4:5.000", strValue)
					// Si cela échoue également, afficher l'erreur et continuez
					if err != nil {
						zap.L().Info("Unexpected format", zap.String("value", strValue), zap.Error(err))
					} else {
						value = t.Format("2006-01-02 15:04:05") 
					}
				} else {
					value = t.Format("2006-01-02 15:04:05") 
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
