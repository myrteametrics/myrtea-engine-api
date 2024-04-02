package export

import (
	"context"
	"fmt"
	"time"

	"github.com/myrteametrics/myrtea-sdk/v4/engine"
	"github.com/spf13/viper"

	"github.com/myrteametrics/myrtea-engine-api/v5/internals/reader"
	"go.uber.org/zap"
)

func ExportFactHitsFull(f engine.Fact) ([]reader.Hit, error) {
	version := viper.GetInt("ELASTICSEARCH_VERSION")
	switch version {
	case 6:
		return ExportFactHitsFullV6(f)
	case 7:
		fallthrough
	case 8:
		return ExportFactHitsFullV8(f)
	default:
		zap.L().Fatal("Unsupported Elasticsearch version", zap.Int("version", version))
		return nil, fmt.Errorf("unsupported Elasticsearch version")
	}
}

func (export StreamedExport) StreamedExportFactHitsFull(ctx context.Context, f engine.Fact, limit int64, placeholders map[string]string) error {
	version := viper.GetInt("ELASTICSEARCH_VERSION")
	switch version {
	case 7:
		fallthrough
	case 8:
		return export.StreamedExportFactHitsFullV8(ctx, f, limit, placeholders)
	default:
		// No fatal here, 6 is unsupported
		//zap.L().Fatal("Unsupported Elasticsearch version", zap.Int("version", version))
		return fmt.Errorf("export is only supported in Elasticsearch version >= 7")
	}
}

func ExportFactHits(ti time.Time, f engine.Fact, placeholders map[string]string, nhit int, offset int) ([]reader.Hit, error) {
	version := viper.GetInt("ELASTICSEARCH_VERSION")
	switch version {
	case 6:
		return ExportFactHitsV6(ti, f, placeholders, nhit, offset)
	case 7:
		fallthrough
	case 8:
		return ExportFactHitsFullV8(f)
	default:
		zap.L().Fatal("Unsupported Elasticsearch version", zap.Int("version", version))
		return nil, fmt.Errorf("unsupported Elasticsearch version")
	}
}
