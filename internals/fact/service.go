package fact

import (
	"context"
	"time"

	"encoding/json"

	"github.com/myrteametrics/myrtea-engine-api/v5/internals/coordinator"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/reader"
	"github.com/myrteametrics/myrtea-sdk/v4/builder"
	"github.com/myrteametrics/myrtea-sdk/v4/elasticsearch"
	"github.com/myrteametrics/myrtea-sdk/v4/engine"
	"github.com/myrteametrics/myrtea-sdk/v4/index"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// Prepare enrich a fact by specifying the offset and the number of hits returned
// It also specify the target index based on the fact model
// The
func Prepare(f *engine.Fact, nhit int, offset int, t time.Time, placeholders map[string]string, update bool) (*builder.EsSearch, error) {
	if f.AdvancedSource != "" {
		var data map[string]interface{}
		var err error
		err = json.Unmarshal([]byte(f.AdvancedSource), &data)
		if err != nil {
			return nil, err
		}

		data["size"] = 0
		if nhit != -1 {
			data["size"] = nhit
		}

		data["from"] = 0
		if offset != -1 {
			data["from"] = offset
		}

		b, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}

		strSource := string(b)
		strSource = builder.ContextualizeDatePlaceholders(strSource, t)
		strSource = builder.ContextualizeTimeZonePlaceholders(strSource, t)
		strSource = builder.ContextualizePlaceholders(strSource, placeholders)

		f.AdvancedSource = strSource
	} else {
		f.ContextualizeDimensions(t, placeholders)
		err := f.ContextualizeCondition(t, placeholders)
		if err != nil {
			return nil, err
		}
	}

	esSearch, err := f.ToElasticQuery(t, placeholders)
	if err != nil {
		return nil, err
	}

	// TODO: remove viper.GetString (and use a real configuration object)
	var indices []string
	if update {
		indices, err = coordinator.GetInstance().LogicalIndex(f.Model).FindIndices(time.Now(), f.CalculationDepth+int64(time.Now().Sub(t).Hours()/24)+5)
	} else {
		indices, err = coordinator.GetInstance().LogicalIndex(f.Model).FindIndices(t, f.CalculationDepth)
	}
	if err != nil {
		return nil, err
	}
	if len(indices) == 0 {
		zap.L().Info("No indices found, fallback on search-all", zap.String("fact", f.Name), zap.String("model", f.Model), zap.Int64("depth", f.CalculationDepth))
		indices = []string{index.BuildAliasName(viper.GetString("INSTANCE_NAME"), f.Model, index.All)}
	}
	esSearch.Indices = indices

	esSearch.Size = 0
	if nhit != -1 {
		esSearch.Size = nhit
	}

	esSearch.Offset = 0
	if offset != -1 {
		esSearch.Offset = offset
	}

	return esSearch, nil
}

// Execute calculate a fact a specific time and returns the result in a standard format
func Execute(esSearch *builder.EsSearch) (*reader.WidgetData, error) {
	search, err := builder.BuildEsSearch(elasticsearch.C(), esSearch)
	if err != nil {
		return nil, err
	}

	res, err := elasticsearch.C().ExecuteSearch(context.Background(), search)
	if err != nil {
		return nil, err
	}

	widgetData, err := reader.Parse(res)
	if err != nil {
		return nil, err
	}

	return widgetData, nil
}
