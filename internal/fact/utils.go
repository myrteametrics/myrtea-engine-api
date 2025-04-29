package fact

import (
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/plugins/baseline"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/reader"
	"time"

	"go.uber.org/zap"
)

// GetBaselineValues fetches the baseline values for a given fact ID and situation instance ID.
func GetBaselineValues(widgetData *reader.WidgetData, factId int64, situationID int64, situationInstanceID int64, ti time.Time) {
	pluginBaseline, err := baseline.P()
	if err == nil {
		values, err := pluginBaseline.BaselineService.GetBaselineValues(-1, factId, situationID, situationInstanceID, ti)
		if err != nil {
			zap.L().Error("Cannot fetch fact baselines", zap.Int64("id", factId), zap.Error(err))
			return
		}
		widgetData.Aggregates.Baselines = values
	}
}
