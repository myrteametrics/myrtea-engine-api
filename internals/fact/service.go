package fact

import (
	"context"
	"errors"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8/typedapi/core/deletebyquery"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/conflicts"
	"github.com/myrteametrics/myrtea-sdk/v5/elasticsearch"
	"strings"
	"time"

	"github.com/myrteametrics/myrtea-engine-api/v5/internals/coordinator"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/reader"
	"github.com/myrteametrics/myrtea-sdk/v5/engine"
	"github.com/myrteametrics/myrtea-sdk/v5/index"
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

func ExecuteFactDeleteQuery(
	ti time.Time,
	f engine.Fact,
) (*deletebyquery.Response, error) {

	parameters := make(map[string]string, 0)
	f.ContextualizeDimensions(ti, parameters)
	err := f.ContextualizeCondition(ti, parameters)
	if err := f.ContextualizeCondition(ti, parameters); err != nil {
		return nil, fmt.Errorf("failed to contextualize condition: %w", err)
	}

	searchRequest, err := elasticsearch.ConvertFactToSearchRequestV8(f, ti, parameters)
	if err != nil {
		zap.L().Error("ConvertFactToSearchRequestV8 failed", zap.Error(err))
		return nil, fmt.Errorf("failed to convert fact to search request: %w", err)
	}

	indices := FindIndices(f, ti, false)
	if len(indices) == 0 {
		return nil, errors.New("no indices found for the fact")
	}

	zap.L().Debug("Preparing to execute DeleteByQuery",
		zap.Strings("indices", indices),
		zap.Any("request", searchRequest))

	response, err := elasticsearch.C().DeleteByQuery(strings.Join(indices, ",")).
		Query(searchRequest.Query).
		Conflicts(conflicts.Proceed). // Ignore les conflits (insertion/suppression)
		Do(context.Background())

	if err != nil {
		zap.L().Error("ES DeleteByQuery execution failed", zap.Error(err))
		return nil, fmt.Errorf("failed to execute DeleteByQuery: %w", err)
	}

	// Check for failures in the response
	if len(response.Failures) > 0 {
		zap.L().Warn("DeleteByQuery encountered shard failures", zap.Any("failures", response.Failures))
		return nil, errors.New("one or more shards failed during DeleteByQuery")
	}

	// Check if the request timed out
	if response.TimedOut != nil && *response.TimedOut {
		zap.L().Warn("DeleteByQuery timed out")
		return nil, errors.New("delete by query operation timed out")
	}

	zap.L().Info("DeleteByQuery completed successfully",
		zap.Int64p("deleted", response.Deleted),
		zap.Int64p("version_conflicts", response.VersionConflicts))

	return response, nil
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
