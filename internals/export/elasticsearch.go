package export

import (
	"fmt"
	"github.com/spf13/viper"
	"time"

	"github.com/myrteametrics/myrtea-engine-api/v5/internals/reader"
	"go.uber.org/zap"
)

func ExportFactHitsFull(factID int64) ([]reader.Hit, error) {
	version := viper.GetInt("ELASTICSEARCH_VERSION")
	switch version {
	case 6:
		return ExportFactHitsFullV6(factID)
	case 7:
		fallthrough
	case 8:
		return ExportFactHitsFullV8(factID)
	default:
		zap.L().Fatal("Unsupported Elasticsearch version", zap.Int("version", version))
		return nil, fmt.Errorf("unsupported Elasticsearch version")
	}
}

func ExportFactHits(ti time.Time, factID int64, placeholders map[string]string, nhit int, offset int) ([]reader.Hit, error) {
	version := viper.GetInt("ELASTICSEARCH_VERSION")
	switch version {
	case 6:
		return ExportFactHitsV6(ti, factID, placeholders, nhit, offset)
	case 7:
		fallthrough
	case 8:
		return ExportFactHitsV8(ti, factID, placeholders, nhit, offset)
	default:
		zap.L().Fatal("Unsupported Elasticsearch version", zap.Int("version", version))
		return nil, fmt.Errorf("unsupported Elasticsearch version")
	}
}
