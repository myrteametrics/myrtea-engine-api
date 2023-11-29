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

	// flag to decide wether Gval exp should be process
	// when a fact returns more than 10k hits, the Gval exp is processed only 1 time
	processGval := true

	for {
		hits, err := ExportFactHitsV6(ti, f, placeholders, nhit, offset, processGval)
		if err != nil {
			return nil, err
		}

		fullHits = append(fullHits, hits...)

		// On Elasticsearch v6, this loop does not work. Therefore, we temporarily stop this loop
		// if size equals 10000 and offset also equals 10000.
		break

		// if len(hits) < 10000 {
		// 	break
		// }
		// offset += 10000
		//processGval = false
	}

	return fullHits, nil
}

func ExportFactHitsV6(ti time.Time, f engine.Fact, placeholders map[string]string, nhit int, offset int, processGval ...bool) ([]reader.Hit, error) {
	// Change the behaviour of the Fact
	f.Intent.Operator = engine.Select

	widgetData, err := fact.ExecuteFact(ti, f, 0, 0, placeholders, nhit, offset, false, processGval...)
	if err != nil {
		zap.L().Error("ExecuteFact", zap.Error(err))
		return nil, err
	}

	return widgetData.Hits, nil
}
