package fact

import (
	"context"
	"errors"
	"github.com/myrteametrics/myrtea-sdk/v4/elasticsearch"
	"strings"
	"time"

	"github.com/myrteametrics/myrtea-engine-api/v5/internals/coordinator"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/reader"
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

	f.ContextualizeDimensions(ti, parameters)
	err := f.ContextualizeCondition(ti, parameters)
	if err != nil {
		return nil, err
	}

	searchRequest, err := elasticsearch.ConvertFactToSearchRequestV8(f, ti, parameters)
	if err != nil {
		zap.L().Error("ConvertFactToSearchRequestV8 failed", zap.Error(err))
		return nil, err
	}
	searchRequest.TrackTotalHits = true

	indices := FindIndices(f, ti, update)
	zap.L().Debug("search", zap.Strings("indices", indices), zap.Any("request", searchRequest))

	response, err := elasticsearch.C().Search().
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

	widgetData, err := reader.Parse(response)
	if err != nil {
		return nil, err
	}

	GetBaselineValues(widgetData, f.ID, situationID, situationInstanceID, ti)

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
