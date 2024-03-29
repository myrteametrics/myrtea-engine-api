package fact

import (
	"time"

	"github.com/myrteametrics/myrtea-engine-api/v5/internals/reader"
	"github.com/myrteametrics/myrtea-engine-api/v5/plugins/baseline"
	"go.uber.org/zap"
)

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
