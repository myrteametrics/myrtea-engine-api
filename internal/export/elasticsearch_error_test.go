package export

import (
	"errors"
	"testing"

	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/myrteametrics/myrtea-sdk/v5/expression"
)

func TestFormatElasticError(t *testing.T) {
	reason := "all shards failed"
	root := "Failed to parse query [foo:]"
	esErr := &types.ElasticsearchError{
		Status: 400,
		ErrorCause: types.ErrorCause{
			Type:   "search_phase_execution_exception",
			Reason: &reason,
			RootCause: []types.ErrorCause{
				{Type: "query_shard_exception", Reason: &root},
				{Type: "query_shard_exception", Reason: &root}, // duplicates (one per shard) are merged
			},
		},
	}
	got := formatElasticError(esErr).Error()
	expression.AssertEqual(t, got, "elasticsearch error (status 400, search_phase_execution_exception): query_shard_exception: Failed to parse query [foo:]")

	plain := errors.New("boom")
	expression.AssertEqual(t, formatElasticError(plain), plain)
}
