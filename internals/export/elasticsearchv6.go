package export

import (
	"time"

	"github.com/myrteametrics/myrtea-engine-api/v5/internals/fact"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/reader"
	"github.com/myrteametrics/myrtea-sdk/v4/engine"
	"go.uber.org/zap"
)

func ExportFactHitsFullV6(f engine.Fact) ([]reader.Hit, error) {
	ti := time.Now()
	placeholders := make(map[string]string)
	nhit := 10000
	offset := 0

	fullHits := make([]reader.Hit, 0)
	for {
		hits, err := ExportFactHitsV6(ti, f, placeholders, nhit, offset)
		if err != nil {
			return nil, err
		}

		fullHits = append(fullHits, hits...)

		if len(hits) < 10000 {
			break
		}
		offset += 10000
	}

	return fullHits, nil
}

func ExportFactHitsV6(ti time.Time, f engine.Fact, placeholders map[string]string, nhit int, offset int) ([]reader.Hit, error) {
	// Change the behaviour of the Fact
	f.Intent.Operator = engine.Select

	widgetData, err := fact.ExecuteFact(ti, f, 0, 0, placeholders, nhit, offset, false)
	if err != nil {
		zap.L().Error("ExecuteFact", zap.Error(err))
	}

	return widgetData.Hits, nil
}
