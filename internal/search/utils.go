package search

import (
	"encoding/json"
	"sort"

	"github.com/myrteametrics/myrtea-engine-api/v5/internal/models"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/reader"
)

func extractFactHistoryRecordValues(rawResults []byte, out *FactHistoryRecord, downSamplingOperation string) error {

	if downSamplingOperation == "" || downSamplingOperation == "first" || downSamplingOperation == "latest" {
		var item reader.Item
		err := json.Unmarshal(rawResults, &item)
		if err != nil {
			return err
		}

		for k, v := range item.Aggs {
			if k == "doc_count" {
				out.DocCount = v.Value
				if out.Value == nil {
					out.Value = v.Value
				}
			} else {
				out.Value = v.Value
			}
		}
		out.Buckets = item.Buckets
		out.Baselines = item.Baselines

	} else {
		var itemList []reader.Item
		err := json.Unmarshal(rawResults, &itemList)
		if err != nil {
			return err
		}
		var dataList []map[string]interface{}
		for _, item := range itemList {
			data := make(map[string]interface{}, 0)
			for k, v := range item.Aggs {
				if k == "doc_count" {
					data["doc_count"] = v.Value
					if data["value"] == nil {
						data["value"] = v.Value
					}
				} else {
					data["value"] = v.Value
				}
			}
			dataList = append(dataList, data)
		}

		aggregation := make(map[string]interface{}, 0)
		aggregate(dataList, downSamplingOperation, aggregation)

		if v, ok := aggregation["value"]; ok {
			out.Value = v
		}
		if v, ok := aggregation["doc_count"]; ok {
			out.DocCount = v
		}
	}

	return nil
}

func extractMetaData(rawMetadatas []byte, out map[string]interface{}, metaDataSource interface{}, downSamplingOperation string) error {
	if rawMetadatas == nil {
		return nil
	}
	var keys []string
	switch value := metaDataSource.(type) {
	case bool:
		if !value {
			return nil
		}
	case string:
		keys = append(keys, value)
	case []string:
		keys = append(keys, value...)
	}

	var metadatas []models.MetaData

	if downSamplingOperation == "" || downSamplingOperation == "first" || downSamplingOperation == "latest" {
		err := json.Unmarshal(rawMetadatas, &metadatas)
		if err != nil {
			return err
		}
		if len(keys) == 0 {
			for _, metadata := range metadatas {
				out[metadata.Key] = metadata.Value
			}
		} else {
			sort.Strings(keys)
			for _, metadata := range metadatas {
				i := sort.SearchStrings(keys, metadata.Key)
				if i < len(keys) && keys[i] == metadata.Key {
					out[metadata.Key] = metadata.Value
				}
			}
		}
	} else {
		var metadatasList [][]models.MetaData
		var dataList []map[string]interface{}
		err := json.Unmarshal(rawMetadatas, &metadatasList)
		if err != nil {
			return err
		}
		for _, m := range metadatasList {
			data := make(map[string]interface{}, 0)
			if len(keys) == 0 {
				for _, metadata := range m {
					data[metadata.Key] = metadata.Value
				}
			} else {
				sort.Strings(keys)
				for _, metadata := range m {
					i := sort.SearchStrings(keys, metadata.Key)
					if i < len(keys) && keys[i] == metadata.Key {
						data[metadata.Key] = metadata.Value
					}
				}
			}
			dataList = append(dataList, data)
		}

		aggregate(dataList, downSamplingOperation, out)
	}
	return nil
}

func extractExpressionFacts(rawExpressionFacts []byte, out map[string]interface{}, expressionFactsSource interface{}, downSamplingOperation string) error {
	if rawExpressionFacts == nil {
		return nil
	}
	var keys []string
	switch value := expressionFactsSource.(type) {
	case bool:
		if !value {
			return nil
		}
	case string:
		keys = append(keys, value)
	case []string:
		keys = append(keys, value...)
	}

	if len(keys) == 0 {
		if downSamplingOperation == "" || downSamplingOperation == "first" || downSamplingOperation == "latest" {
			err := json.Unmarshal(rawExpressionFacts, &out)
			if err != nil {
				return err
			}
		} else {
			var expressionFactsList []map[string]interface{}
			err := json.Unmarshal(rawExpressionFacts, &expressionFactsList)
			if err != nil {
				return err
			}
			for key, value := range expressionFactsList[0] {
				out[key] = value
			}
		}

	} else {
		var expressionFacts map[string]interface{}
		if downSamplingOperation == "" || downSamplingOperation == "first" || downSamplingOperation == "latest" {
			err := json.Unmarshal(rawExpressionFacts, &expressionFacts)
			if err != nil {
				return err
			}
		} else {
			var expressionFactsList []map[string]interface{}
			err := json.Unmarshal(rawExpressionFacts, &expressionFactsList)
			if err != nil {
				return err
			}
			expressionFacts = expressionFactsList[0]
		}

		sort.Strings(keys)
		for key, value := range expressionFacts {
			i := sort.SearchStrings(keys, key)
			if i < len(keys) && keys[i] == key {
				out[key] = value
			}
		}

	}
	return nil
}

func extractParameters(rawParameters []byte, out map[string]interface{}, parametersSource interface{}, downSamplingOperation string) error {
	if rawParameters == nil {
		return nil
	}
	var keys []string
	switch value := parametersSource.(type) {
	case bool:
		if !value {
			return nil
		}
	case string:
		keys = append(keys, value)
	case []string:
		keys = append(keys, value...)
	}

	if len(keys) == 0 {
		if downSamplingOperation == "" || downSamplingOperation == "first" || downSamplingOperation == "latest" {
			err := json.Unmarshal(rawParameters, &out)
			if err != nil {
				return err
			}
		} else {
			var parametersList []map[string]interface{}
			err := json.Unmarshal(rawParameters, &parametersList)
			if err != nil {
				return err
			}
			for key, value := range parametersList[0] {
				out[key] = value
			}
		}

	} else {
		var parameters map[string]interface{}
		if downSamplingOperation == "" || downSamplingOperation == "first" || downSamplingOperation == "latest" {
			err := json.Unmarshal(rawParameters, &parameters)
			if err != nil {
				return err
			}
		} else {
			var parametersList []map[string]interface{}
			err := json.Unmarshal(rawParameters, &parametersList)
			if err != nil {
				return err
			}
			parameters = parametersList[0]
		}

		sort.Strings(keys)
		for key, value := range parameters {
			i := sort.SearchStrings(keys, key)
			if i < len(keys) && keys[i] == key {
				out[key] = value
			}
		}

	}
	return nil
}

func aggregate(dataList []map[string]interface{}, operation string, out map[string]interface{}) {
	switch operation {
	case "sum":
		for _, data := range dataList {
			for key, value := range data {
				if _, ok := out[key]; !ok {
					out[key] = value
				} else {
					val1, ok1 := out[key].(float64)
					val2, ok2 := value.(float64)
					if ok1 && ok2 {
						out[key] = val1 + val2
					}
				}
			}
		}
	case "max":
		for _, data := range dataList {
			for key, value := range data {
				if _, ok := out[key]; !ok {
					out[key] = value
				} else {
					val1, ok1 := out[key].(float64)
					val2, ok2 := value.(float64)
					if ok1 && ok2 && val2 > val1 {
						out[key] = val2
					}
				}
			}
		}
	case "min":
		for _, data := range dataList {
			for key, value := range data {
				if _, ok := out[key]; !ok {
					out[key] = value
				} else {
					val1, ok1 := out[key].(float64)
					val2, ok2 := value.(float64)
					if ok1 && ok2 && val2 < val1 {
						out[key] = val2
					}
				}
			}
		}
	case "avg":
		sum := make(map[string]float64, 0)
		count := make(map[string]float64, 0)
		for _, data := range dataList {
			for key, value := range data {
				if _, ok := sum[key]; !ok {
					val, ok := value.(float64)
					if ok {
						sum[key] = val
						count[key] = 1
					}
				} else {
					val, ok := value.(float64)
					if ok {
						sum[key] = sum[key] + val
						count[key] = count[key] + 1
					}
				}
			}
		}
		for key, value := range sum {
			out[key] = value / count[key]
		}
	}
}
