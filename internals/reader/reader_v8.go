package reader

import (
	"encoding/json"

	"github.com/elastic/go-elasticsearch/v8/typedapi/core/search"
	jsoniter "github.com/json-iterator/go"
	"go.uber.org/zap"
)

// ParseV8 parse a elasticsearch SearchResponse (hits and aggregations) and returns a WidgetData
func ParseV8(res *search.Response) (*WidgetData, error) {
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
		var fields map[string]interface{}
		err := json.Unmarshal(hit.Source_, &fields)
		if err != nil {
			zap.L().Warn("Cannot unmarshall Source", zap.Any("source", hit.Source_))
			continue
		}
		hits = append(hits, Hit{ID: hit.Id_, Fields: fields})
	}

	widgetData := WidgetData{
		Hits:       hits,
		Aggregates: item,
	}

	return &widgetData, nil
}
