package export

import (
	"context"
	"github.com/elastic/go-elasticsearch/v8/typedapi/core/closepointintime"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/myrteametrics/myrtea-engine-api/v5/plugins/baseline"
	"github.com/myrteametrics/myrtea-sdk/v4/elasticsearchv8"
	"strings"
	"time"

	"github.com/myrteametrics/myrtea-engine-api/v5/internals/fact"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/reader"
	"github.com/myrteametrics/myrtea-sdk/v4/engine"
	"go.uber.org/zap"
)

func ExportFactHitsFullV8(factID int64) ([]reader.Hit, error) {
	ti := time.Now()
	placeholders := make(map[string]string)
	fullHits := make([]reader.Hit, 0)

	f, found, err := fact.R().Get(factID)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, err
	}

	// Change the behaviour of the Fact
	f.Intent.Operator = engine.Select

	searchRequest, err := elasticsearchv8.ConvertFactToSearchRequestV8(f, ti, placeholders)
	if err != nil {
		zap.L().Error("ConvertFactToSearchRequestV8 failed", zap.Error(err))
		return nil, err
	}

	indices := fact.FindIndices(f, ti, false)
	indicesStr := strings.Join(indices, ",")

	// handle pit creation
	pit, err := elasticsearchv8.C().OpenPointInTime(indicesStr).KeepAlive("1m").Do(context.Background())
	if err != nil {
		zap.L().Error("OpenPointInTime failed", zap.Error(err))
		return nil, err
	}
	searchRequest.Pit.Id = pit.Id
	searchRequest.SearchAfter = []types.FieldValue{}

	for {
		response, err := elasticsearchv8.C().Search().
			Index(indicesStr).
			Size(10000).
			Request(searchRequest).
			Sort("_shard_doc:asc").
			Do(context.Background())
		if err != nil {
			zap.L().Error("ES Search failed", zap.Error(err))
			// TODO: maybe close PIT?
			return nil, err
		}

		widgetData, err := reader.ParseV8(response)
		if err != nil {
			return nil, err
		}

		pluginBaseline, err := baseline.P()
		if err == nil {
			values, err := pluginBaseline.BaselineService.GetBaselineValues(-1, f.ID, 0, 0, ti)
			if err != nil {
				zap.L().Error("Cannot fetch fact baselines", zap.Int64("id", f.ID), zap.Error(err))
			}
			widgetData.Aggregates.Baselines = values
		}

		// Check if response contains at least one hit
		hitsLen := len(response.Hits.Hits)
		if hitsLen == 0 {
			break
		}

		// Handle SearchAfter to paginate
		searchRequest.SearchAfter = response.Hits.Hits[hitsLen-1].Sort
		fullHits = append(fullHits, widgetData.Hits...)

		// TODO: Maybe remove?
		if len(widgetData.Hits) < 10000 {
			break
		}
	}

	do, err := elasticsearchv8.C().ClosePointInTime().
		Request(&closepointintime.Request{Id: pit.Id}).
		Do(context.Background())

	if err != nil {
		return nil, err // TODO:
	}

	if !do.Succeeded {
		zap.L().Warn("Could not close PointInTime")
	}

	return fullHits, nil
}

func ExportFactHitsV8(ti time.Time, factID int64, placeholders map[string]string, nhit int, offset int) ([]reader.Hit, error) {
	f, found, err := fact.R().Get(factID)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, err
	}

	// Change the behaviour of the Fact
	f.Intent.Operator = engine.Select

	widgetData, err := fact.ExecuteFact(ti, f, 0, 0, placeholders, nhit, offset, false)
	if err != nil {
		zap.L().Error("ExecuteFact", zap.Error(err))
	}

	return widgetData.Hits, nil
}
