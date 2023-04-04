package fact

import (
	"context"
	"fmt"
	"time"

	"encoding/json"

	"github.com/myrteametrics/myrtea-engine-api/v5/internals/coordinator"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/reader"
	"github.com/myrteametrics/myrtea-engine-api/v5/plugins/baseline"
	"github.com/myrteametrics/myrtea-sdk/v4/builder"
	"github.com/myrteametrics/myrtea-sdk/v4/elasticsearchv6"
	"github.com/myrteametrics/myrtea-sdk/v4/engine"
	"github.com/myrteametrics/myrtea-sdk/v4/index"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func ExecuteFact(
	ti time.Time,
	f engine.Fact, situationID int64, situationInstanceID int64, parameters map[string]string,
	nhit int, offset int, update bool,
) (*reader.WidgetData, error) {

	version := viper.GetInt("ELASTICSEARCH_VERSION")
	switch version {
	case 6:
		return ExecuteFactV6(ti, f, situationID, situationInstanceID, parameters, nhit, offset, update)
	case 7:
		fallthrough
	case 8:
		return ExecuteFactV8(ti, f, situationID, situationInstanceID, parameters, nhit, offset, update)
	default:
		zap.L().Fatal("Unsupported Elasticsearch version", zap.Int("version", version))
		return nil, fmt.Errorf("Unsupported Elasticsearch version")
	}
}

func ExecuteFactV6(
	ti time.Time,
	f engine.Fact, situationID int64, situationInstanceID int64, parameters map[string]string,
	nhit int, offset int, update bool,
) (*reader.WidgetData, error) {
	pf, err := PrepareV6(&f, nhit, offset, ti, parameters, update)
	if err != nil {
		zap.L().Error("Cannot prepare fact", zap.Int64("id", f.ID), zap.Any("fact", f), zap.Error(err))
		return nil, err
	}

	widgetData, err := ExecuteV6(pf)
	if err != nil {
		zap.L().Error("Cannot execute fact", zap.Int64("id", f.ID), zap.Any("pf", pf), zap.Error(err))
		return nil, err
	}

	pluginBaseline, err := baseline.P()
	if err == nil {
		values, err := pluginBaseline.BaselineService.GetBaselineValues(-1, f.ID, situationID, situationInstanceID, ti)
		if err != nil {
			zap.L().Error("Cannot fetch fact baselines", zap.Int64("id", f.ID), zap.Error(err))
		}
		widgetData.Aggregates.Baselines = values
	}

	return widgetData, nil
}

// PrepareV6 enrich a fact by specifying the offset and the number of hits returned
// It also specify the target index based on the fact model
func PrepareV6(f *engine.Fact, nhit int, offset int, t time.Time, parameters map[string]string, update bool) (*builder.EsSearch, error) {
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
		strSource = builder.ContextualizePlaceholders(strSource, parameters)

		f.AdvancedSource = strSource
	} else {
		f.ContextualizeDimensions(t, parameters)
		err := f.ContextualizeCondition(t, parameters)
		if err != nil {
			return nil, err
		}
	}

	esSearch, err := f.ToElasticQuery(t, parameters)
	if err != nil {
		return nil, err
	}

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

// ExecuteV6 calculate a fact a specific time and returns the result in a standard format
func ExecuteV6(esSearch *builder.EsSearch) (*reader.WidgetData, error) {
	search, err := builder.BuildEsSearch(elasticsearchv6.C(), esSearch)
	if err != nil {
		return nil, err
	}

	res, err := elasticsearchv6.C().ExecuteSearch(context.Background(), search)
	if err != nil {
		return nil, err
	}

	widgetData, err := reader.Parse(res)
	if err != nil {
		return nil, err
	}

	return widgetData, nil
}
