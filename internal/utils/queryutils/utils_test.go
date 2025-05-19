package queryutils

import (
	"testing"

	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/model"
)

func TestAppendSearchOptions(t *testing.T) {
	query := "SELECT * FROM mytable"
	params := map[string]interface{}{}
	options := model.SearchOptions{Limit: 10, Offset: 50, SortBy: []model.SortOption{{Field: "field1", Order: model.Asc}}}
	prefix := "mytable"

	var err error
	query, params, err = AppendSearchOptions(query, params, options, prefix)
	if err != nil {
		t.Error(err)
	}
	if query != "SELECT * FROM mytable ORDER BY mytable.field1 ASC LIMIT :limit OFFSET :offset" {
		t.Error("invalid query", query)
	}
	if params["limit"] != 10 {
		t.Error("invalid param limit")
	}
	if params["offset"] != 50 {
		t.Error("invalid param offset")
	}
}
