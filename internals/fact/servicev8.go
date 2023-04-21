package fact

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/myrteametrics/myrtea-engine-api/v5/internals/coordinator"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/reader"
	"github.com/myrteametrics/myrtea-engine-api/v5/plugins/baseline"
	"github.com/myrteametrics/myrtea-sdk/v4/elasticsearchv8"
	"github.com/myrteametrics/myrtea-sdk/v4/engine"
	"github.com/myrteametrics/myrtea-sdk/v4/index"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func ExecuteFactV8(
	ti time.Time,
	f engine.Fact, situationID int64, situationInstanceID int64, parameters map[string]string,
	nhit int, offset int, update bool,
) (*reader.WidgetData, error) {

	f.ContextualizeDimensions(ti, parameters)
	err := f.ContextualizeCondition(ti, parameters)
	if err != nil {
		return nil, err
	}

	searchRequest, err := elasticsearchv8.ConvertFactToSearchRequestV8(f, ti, parameters)
	if err != nil {
		zap.L().Error("ConvertFactToSearchRequestV8 failed", zap.Error(err))
		return nil, err
	}
	searchRequest.TrackTotalHits = true

	indices := FindIndices(f, ti, update)
	zap.L().Debug("search", zap.Strings("indices", indices), zap.Any("request", searchRequest))

	response, err := elasticsearchv8.C().Search().
		Index(strings.Join(indices, ",")).
		From(offset).
		Size(nhit).
		Request(searchRequest).
		Do(context.Background())
	if err != nil {
		zap.L().Error("ES Search failed", zap.Error(err))
		return nil, err
	}
	if response.Shards_.Failed > 0 {
		zap.L().Warn("search", zap.Any("failures", response.Shards_.Failures))
		return nil, errors.New("search failed")
	}

	widgetData, err := reader.ParseV8(response)
	if err != nil {
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

func FindIndices(f engine.Fact, ti time.Time, update bool) []string {

	var indices []string
	var err error

	if update {
		indices, err = coordinator.GetInstance().LogicalIndex(f.Model).FindIndices(time.Now(), f.CalculationDepth+int64(time.Now().Sub(ti).Hours()/24)+5)
	} else {
		indices, err = coordinator.GetInstance().LogicalIndex(f.Model).FindIndices(ti, f.CalculationDepth)
	}
	if err != nil {
		zap.L().Warn("FindIndices", zap.Error(err))
	}

	if len(indices) == 0 {
		zap.L().Info("No indices found, fallback on search-all", zap.String("fact", f.Name), zap.String("model", f.Model), zap.Int64("depth", f.CalculationDepth))
		indices = []string{index.BuildAliasName(viper.GetString("INSTANCE_NAME"), f.Model, index.All)}
	}

	return indices
}
