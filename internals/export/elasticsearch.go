package export

import (
	"context"
	"errors"
	es "github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/typedapi/core/search"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/config/esconfig"
	"github.com/myrteametrics/myrtea-sdk/v4/elasticsearch"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v8/typedapi/core/closepointintime"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/fact"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/reader"
	"github.com/myrteametrics/myrtea-sdk/v4/engine"
	"go.uber.org/zap"
)

type StreamedExport struct {
	Data chan []reader.Hit
}

// NewStreamedExport returns a pointer to a new StreamedExport instance
// One instance per StreamedExport since the channel Data will be closed after export is finished
func NewStreamedExport() *StreamedExport {
	return &StreamedExport{
		Data: make(chan []reader.Hit, 10),
	}
}

// DrainChannel Drains the Data channel without processing the remaining values
func (export StreamedExport) DrainChannel() {
	for {
		_, ok := <-export.Data
		if !ok {
			break // Channel is closed and empty
		}
	}
}

// StreamedExportFactHitsFull export data from ElasticSearch to a channel
// Please note that the channel is not closed when this function is executed
// it is building the searchRequest generated by the fact, and the indices
func (export StreamedExport) StreamedExportFactHitsFull(ctx context.Context, f engine.Fact, limit int64, factParameters map[string]string) error {
	ti := time.Now()
	// Change the behaviour of the Fact
	f.Intent.Operator = engine.Select

	err := f.ContextualizeCondition(ti, factParameters)
	if err != nil {
		return err
	}

	searchRequest, err := elasticsearch.ConvertFactToSearchRequestV8(f, ti, factParameters)
	if err != nil {
		zap.L().Error("ConvertFactToSearchRequestV8 failed", zap.Error(err))
		return err
	}

	indices := fact.FindIndices(f, ti, false)
	indicesStr := strings.Join(indices, ",")

	return export.ProcessStreamedExport(ctx, "", indicesStr, searchRequest, limit)
}

// ProcessStreamedExport export data from ElasticSearch to a channel
// using a given elastic client, request and indices to query
// Please note that the channel is not closed when this function is executed
func (export StreamedExport) ProcessStreamedExport(ctx context.Context, elasticClient, indicesStr string, searchRequest *search.Request, limit int64) error {
	cli := elasticsearch.C() // defaults to singleton instance

	if elasticClient != "" {
		config, b, err := esconfig.R().GetByName(elasticClient)
		if err != nil {
			zap.L().Error("Error when getting esconfig from repository",
				zap.String("elasticClient", elasticClient), zap.Error(err))
			return err
		}
		if !b {
			zap.L().Error("Error when getting esconfig from repository",
				zap.String("elasticClient", elasticClient), zap.Error(err))
			return errors.New("selected elasticClient not exists")
		}
		if !config.ExportActivated {
			zap.L().Warn("trying to export on a disabled elasticsearch config", zap.Any("esconfig", config))
			return errors.New("export is disabled for this elasticsearch config")
		}
		if len(config.URLs) == 0 {
			zap.L().Warn("ElasticSearch config urls is empty, cannot export", zap.Any("esconfig", config))
			return errors.New("cannot init an elasticsearch client with missing URLs")
		}

		cli, err = es.NewTypedClient(es.Config{
			Addresses: config.URLs,
		})
		if err != nil {
			zap.L().Error("Cannot init new elasticsearch typed client", zap.Error(err))
			return err
		}
	}

	// handle pit creation
	pit, err := cli.OpenPointInTime(indicesStr).KeepAlive("5m").Do(context.Background())
	if err != nil {
		zap.L().Error("OpenPointInTime failed", zap.Error(err))
		return err
	}

	// close point in time
	defer func() {
		do, err := cli.ClosePointInTime().
			Request(&closepointintime.Request{Id: pit.Id}).
			Do(context.Background())

		if err != nil {
			zap.L().Error("Error during PointInTime closing", zap.Error(err))
		} else if !do.Succeeded {
			zap.L().Warn("Could not close PointInTime")
		}
	}()

	searchRequest.Pit = &types.PointInTimeReference{Id: pit.Id, KeepAlive: "5m"}
	searchRequest.SearchAfter = []types.FieldValue{}
	searchRequest.Sort = append(searchRequest.Sort, "_shard_doc")
	// searchRequest.TrackTotalHits = false // Speeds up pagination (maybe impl?)

	processed := int64(0)
	hasLimit := limit > 0
	var size int

	for {

		if hasLimit && processed >= limit {
			break
		}

		if !hasLimit || limit-processed > 10000 {
			size = 10000
		} else {
			size = int(limit - processed)
		}

		response, err := cli.Search().
			Request(searchRequest).
			Size(size).
			Do(context.Background())
		if err != nil {
			zap.L().Error("ES Search failed", zap.Error(err))
			// TODO: maybe close PIT (defer close function?)
			return err
		}

		if hasLimit {
			processed += int64(size)
		}

		// Check if response contains at least one hit
		hitsLen := len(response.Hits.Hits)
		if hitsLen == 0 {
			break
		}

		widgetData, err := reader.Parse(response)
		if err != nil {
			return err
		}

		// Handle SearchAfter to paginate
		searchRequest.SearchAfter = response.Hits.Hits[hitsLen-1].Sort

		// if ctx was cancelled, stop data pulling
		select {
		case <-ctx.Done():
			return ctx.Err()
		case export.Data <- widgetData.Hits:
			// do nothing
		default:
			return errors.New("StreamedExport channel closed unexceptionally")
		} // send data through channel

		// avoids a useless search request
		if len(response.Hits.Hits) < size {
			break
		}
	}
	return nil
}

func ExportFactHitsFull(f engine.Fact) ([]reader.Hit, error) {
	ti := time.Now()
	placeholders := make(map[string]string)
	fullHits := make([]reader.Hit, 0)

	// Change the behaviour of the Fact
	f.Intent.Operator = engine.Select

	err := f.ContextualizeCondition(ti, placeholders)
	if err != nil {
		return nil, err
	}

	searchRequest, err := elasticsearch.ConvertFactToSearchRequestV8(f, ti, placeholders)
	if err != nil {
		zap.L().Error("ConvertFactToSearchRequestV8 failed", zap.Error(err))
		return nil, err
	}

	indices := fact.FindIndices(f, ti, false)
	indicesStr := strings.Join(indices, ",")

	// handle pit creation
	pit, err := elasticsearch.C().OpenPointInTime(indicesStr).KeepAlive("1m").Do(context.Background())
	if err != nil {
		zap.L().Error("OpenPointInTime failed", zap.Error(err))
		return nil, err
	}
	searchRequest.Pit = &types.PointInTimeReference{Id: pit.Id, KeepAlive: "1m"}
	searchRequest.SearchAfter = []types.FieldValue{}
	// searchRequest.TrackTotalHits = false // Speeds up pagination (maybe impl?)

	for {
		response, err := elasticsearch.C().Search().
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

		widgetData, err := reader.Parse(response)
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

	do, err := elasticsearch.C().ClosePointInTime().
		Request(&closepointintime.Request{Id: pit.Id}).
		Do(context.Background())

	if err != nil {
		zap.L().Error("Error during PointInTime closing", zap.Error(err))
	} else if !do.Succeeded {
		zap.L().Warn("Could not close PointInTime")
	}

	return fullHits, nil
}

// ExportFactHits export data from ElasticSearch and returns a slice of Hits
func ExportFactHits(ti time.Time, f engine.Fact, placeholders map[string]string, nhit int, offset int) ([]reader.Hit, error) {
	// Change the behaviour of the Fact
	f.Intent.Operator = engine.Select

	widgetData, err := fact.ExecuteFact(ti, f, 0, 0, placeholders, nhit, offset, false)
	if err != nil {
		zap.L().Error("ExecuteFact", zap.Error(err))
	}

	return widgetData.Hits, nil
}
