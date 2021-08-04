package export

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strings"

	"github.com/myrteametrics/myrtea-engine-api/v4/internals/reader"
)

func ConvertHitsToCSV(hits []reader.Hit, columns []string, columnsLabel []string, separator rune) ([]byte, error) {
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
