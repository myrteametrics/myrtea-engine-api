package export

import (
	"time"

	"github.com/myrteametrics/myrtea-engine-api/v4/internals/fact"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/reader"
	"github.com/myrteametrics/myrtea-sdk/v4/builder"
	"github.com/myrteametrics/myrtea-sdk/v4/engine"
	"go.uber.org/zap"
)

func ExportFactHitsFull(factID int64) ([]reader.Hit, error) {
	ti := time.Now()
	placeholders := make(map[string]string)
	nhit := 10000
	offset := 0

	fullHits := make([]reader.Hit, 0)
	for {
		hits, err := ExportFactHits(ti, factID, placeholders, nhit, offset)
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

func ExportFactHits(ti time.Time, factID int64, placeholders map[string]string, nhit int, offset int) ([]reader.Hit, error) {
	f, found, err := fact.R().Get(factID)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, err
	}

	// Change the behaviour of the Fact
	f.Intent.Operator = engine.Select

	pf, err := fact.Prepare(&f, nhit, offset, ti, placeholders, false)
	if err != nil {
		zap.L().Error("Cannot execute fact", zap.Error(err), zap.Any("fact", f))
		return nil, err
	}

	debug := true
	if debug {
		zap.L().Debug("Debugging prepared fact", zap.Any("pf", pf))
		source, _ := builder.BuildEsSearchSource(pf)
		zap.L().Debug("Debugging final elastic query", zap.Any("query", source))
	}

	data, err := fact.Execute(pf)
	if err != nil {
		zap.L().Error("Cannot execute fact", zap.Error(err), zap.Any("prepared-query", pf))
		return nil, err
	}

	return data.Hits, nil
}
