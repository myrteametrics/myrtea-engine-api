package export

import (
	"context"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v8/typedapi/core/closepointintime"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/myrteametrics/myrtea-sdk/v4/elasticsearchv8"

	"github.com/myrteametrics/myrtea-engine-api/v5/internals/fact"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/reader"
	"github.com/myrteametrics/myrtea-sdk/v4/engine"
	"go.uber.org/zap"
)

type StreamedExport struct {
	Data      chan []reader.Hit
	ChunkSize int64
}

// NewStreamedExport returns a pointer to a new StreamedExport instance
// One instance per StreamedExport since the channel Data will be closed after export is finished
// TODO: maybe not close channel to handle other exports (ex: combined fact exports) and add an ok channel?
func NewStreamedExport(chunkSize int64) *StreamedExport {
	return &StreamedExport{
		Data:      make(chan []reader.Hit, 10),
		ChunkSize: chunkSize,
	}
}

func (export *StreamedExport) StreamedExportFactHitsFullV8(f engine.Fact) error {
	ti := time.Now()
	placeholders := make(map[string]string)

	// Change the behaviour of the Fact
	f.Intent.Operator = engine.Select

	err := f.ContextualizeCondition(ti, placeholders)
	if err != nil {
		return err
	}

	searchRequest, err := elasticsearchv8.ConvertFactToSearchRequestV8(f, ti, placeholders)
	if err != nil {
		zap.L().Error("ConvertFactToSearchRequestV8 failed", zap.Error(err))
		return err
	}

	indices := fact.FindIndices(f, ti, false)
	indicesStr := strings.Join(indices, ",")

	// handle pit creation
	pit, err := elasticsearchv8.C().OpenPointInTime(indicesStr).KeepAlive("1m").Do(context.Background())
	if err != nil {
		zap.L().Error("OpenPointInTime failed", zap.Error(err))
		return err
	}
	searchRequest.Pit = &types.PointInTimeReference{Id: pit.Id, KeepAlive: "1m"}
	searchRequest.SearchAfter = []types.FieldValue{}
	// searchRequest.TrackTotalHits = false // Speeds up pagination (maybe impl?)

	for {
		response, err := elasticsearchv8.C().Search().
			//Index(indicesStr).
			Size(10000).
			Request(searchRequest).
			Sort("_shard_doc:asc").
			Do(context.Background())
		if err != nil {
			zap.L().Error("ES Search failed", zap.Error(err))
			// TODO: maybe close PIT (defer close function?)
			return err
		}

		// Check if response contains at least one hit
		hitsLen := len(response.Hits.Hits)
		if hitsLen == 0 {
			break
		}

		widgetData, err := reader.ParseV8(response)
		if err != nil {
			return err
		}

		// Handle SearchAfter to paginate
		searchRequest.SearchAfter = response.Hits.Hits[hitsLen-1].Sort
		export.Data <- widgetData.Hits // send data through channel

		// avoids a useless search request
		if len(response.Hits.Hits) < 10000 {
			break
		}
	}

	do, err := elasticsearchv8.C().ClosePointInTime().
		Request(&closepointintime.Request{Id: pit.Id}).
		Do(context.Background())

	if err != nil {
		zap.L().Error("Error during PointInTime closing", zap.Error(err))
	} else if !do.Succeeded {
		zap.L().Warn("Could not close PointInTime")
	}

	close(export.Data)
	return nil
}

func ExportFactHitsFullV8(f engine.Fact) ([]reader.Hit, error) {
	ti := time.Now()
	placeholders := make(map[string]string)
	fullHits := make([]reader.Hit, 0)

	// Change the behaviour of the Fact
	f.Intent.Operator = engine.Select

	err := f.ContextualizeCondition(ti, placeholders)
	if err != nil {
		return nil, err
	}

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
	searchRequest.Pit = &types.PointInTimeReference{Id: pit.Id, KeepAlive: "1m"}
	searchRequest.SearchAfter = []types.FieldValue{}
	// searchRequest.TrackTotalHits = false // Speeds up pagination (maybe impl?)

	for {
		response, err := elasticsearchv8.C().Search().
			//Index(indicesStr).
			Size(10000).
			Request(searchRequest).
			Sort("_shard_doc:asc").
			Do(context.Background())
		if err != nil {
			zap.L().Error("ES Search failed", zap.Error(err))
			// TODO: maybe close PIT (defer close function?)
			return nil, err
		}

		// Check if response contains at least one hit
		hitsLen := len(response.Hits.Hits)
		if hitsLen == 0 {
			break
		}

		widgetData, err := reader.ParseV8(response)
		if err != nil {
			return nil, err
		}

		// Handle SearchAfter to paginate
		searchRequest.SearchAfter = response.Hits.Hits[hitsLen-1].Sort
		fullHits = append(fullHits, widgetData.Hits...)

		// avoids a useless search request
		if len(response.Hits.Hits) < 10000 {
			break
		}
	}

	do, err := elasticsearchv8.C().ClosePointInTime().
		Request(&closepointintime.Request{Id: pit.Id}).
		Do(context.Background())

	if err != nil {
		zap.L().Error("Error during PointInTime closing", zap.Error(err))
	} else if !do.Succeeded {
		zap.L().Warn("Could not close PointInTime")
	}

	return fullHits, nil
}

func ExportFactHitsV8(ti time.Time, f engine.Fact, placeholders map[string]string, nhit int, offset int) ([]reader.Hit, error) {
	// Change the behaviour of the Fact
	f.Intent.Operator = engine.Select

	widgetData, err := fact.ExecuteFact(ti, f, 0, 0, placeholders, nhit, offset, false)
	if err != nil {
		zap.L().Error("ExecuteFact", zap.Error(err))
	}

	return widgetData.Hits, nil
}
