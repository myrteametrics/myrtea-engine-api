package reader

import (
	"encoding/json"

	jsoniter "github.com/json-iterator/go"
	"github.com/myrteametrics/myrtea-sdk/v4/elasticsearchv8"
	"go.uber.org/zap"
)

// ParseV8 parse a elasticsearch SearchResponse (hits and aggregations) and returns a WidgetData
func ParseV8(res elasticsearchv8.SearchResponse) (*WidgetData, error) {
	item := &Item{}

	// Parse Aggregations
	data, err := jsoniter.Marshal(res.Aggregations)
	if err != nil {
		zap.L().Error("LoadKPI.MarshalAggregation:", zap.Error(err))
		return nil, err
	}
	if string(data) != "null" && string(data) != "{}" {
		aggs := make(map[string]interface{})
		err := json.Unmarshal(data, &aggs)
		if err != nil {
			return nil, err
		}
		item = ParseAggs(aggs)
	}
	item = EnrichWithTotalHits(item, res.Hits.Total.Value)

	// Parse Hits
	hits := make([]Hit, 0)
	for _, hit := range res.Hits.Hits {
		fields, ok := hit.Source_.(map[string]interface{})
		if ok {
			hits = append(hits, Hit{ID: hit.Id_, Fields: fields})
		}
	}

	widgetData := WidgetData{
		Hits:       hits,
		Aggregates: item,
	}

	return &widgetData, nil
}
