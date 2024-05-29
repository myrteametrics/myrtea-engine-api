package reader

import (
	"encoding/json"
	"fmt"
	"github.com/myrteametrics/myrtea-engine-api/v5/plugins/baseline"

	"github.com/elastic/go-elasticsearch/v8/typedapi/core/search"
	jsoniter "github.com/json-iterator/go"
	"go.uber.org/zap"
)

// WidgetData is a standard api response format for fact
type WidgetData struct {
	Hits       []Hit `json:"hits"`
	Aggregates *Item `json:"aggregates"`
}

// Hit is used to represent a object Hit
type Hit struct {
	ID     string                 `json:"id"`
	Fields map[string]interface{} `json:"fields"`
}

// Item is used to represent a calculated aggregate and it's sub-aggregate
type Item struct {
	Key         string                            `json:"key,omitempty"`
	KeyAsString string                            `json:"key-as-string,omitempty"`
	Aggs        map[string]*ItemAgg               `json:"aggs,omitempty"`
	Buckets     map[string][]*Item                `json:"buckets,omitempty"`
	Baselines   map[string]baseline.BaselineValue `json:"baselines,omitempty"`
}

// ToAbstractMap convert an Item in an abstract map[string]interface{}
func (item *Item) ToAbstractMap() (map[string]interface{}, error) {
	b, err := json.Marshal(item)
	if err != nil {
		return nil, err
	}

	var data map[string]interface{}
	err = json.Unmarshal(b, &data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// ItemAgg is used to represent a single aggregate value
type ItemAgg struct {
	Value interface{} `json:"value,omitempty"`
}

// EnrichWithTotalHits enrich an Item with a new key doc_count giving the total number of hits
func EnrichWithTotalHits(item *Item, totalHits int64) *Item {
	itemAgg := &ItemAgg{}
	itemAgg.Value = totalHits
	if item.Aggs == nil {
		item.Aggs = make(map[string]*ItemAgg)
	}
	item.Aggs["doc_count"] = itemAgg
	return item
}

// ParseAggs parse the Aggregations part of an elasticsearch result
// It is used recursively to parse sub aggs
func ParseAggs(item map[string]interface{}) *Item {
	itm := &Item{}
	for key, _item := range item {
		switch key {
		case "key_as_string":
			itm.Key = _item.(string)

		case "key":
			if itm.Key == "" {
				itm.Key = fmt.Sprintf("%v", _item)
			}

		case "doc_count":
			if itm.Aggs == nil {
				itm.Aggs = make(map[string]*ItemAgg)
			}
			itm.Aggs["doc_count"] = &ItemAgg{_item}

		default:
			item := _item.(map[string]interface{})
			if _subItem, ok := item["value"]; ok {
				if itm.Aggs == nil {
					itm.Aggs = make(map[string]*ItemAgg)
				}
				itm.Aggs[key] = &ItemAgg{_subItem}
			} else if _subItems, ok := item["buckets"]; ok {
				if itm.Buckets == nil {
					itm.Buckets = make(map[string][]*Item)
				}
				subItems := _subItems.([]interface{})
				var buckets []*Item
				for _, _subItem := range subItems {
					subItem := _subItem.(map[string]interface{})
					bucket := ParseAggs(subItem)
					buckets = append(buckets, bucket)
					itm.Buckets[key] = buckets
				}
			} else {
				if itm.Buckets == nil {
					itm.Buckets = make(map[string][]*Item)
				}
				var buckets []*Item
				bucket := ParseAggs(item)
				if key != "meta" {
					buckets = append(buckets, bucket)
					itm.Buckets[key] = buckets
				}
			}
		}
	}
	return itm
}

// Parse an elasticsearch search.Response (hits and aggregations) and returns a WidgetData
func Parse(res *search.Response) (*WidgetData, error) {
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
