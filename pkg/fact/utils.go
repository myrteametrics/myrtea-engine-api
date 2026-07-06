package fact

import (
	"time"

	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/plugins/baseline"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/reader"

	"go.uber.org/zap"
)

// GetBaselineValues fetches the baseline values for a given fact ID and situation instance ID.
//
// Deprecated: superseded by GetMatrixProfileResults. Kept in place, but no longer called
// from the fact execution path.
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

// GetMatrixProfileResults fetches the matrix profile results (status + matching score) for
// every matrix profile definition linked to the given situation instance.
func GetMatrixProfileResults(widgetData *reader.WidgetData, situationID int64, situationInstanceID int64, ti time.Time) {
	pluginBaseline, err := baseline.P()
	if err == nil {
		values, err := pluginBaseline.BaselineService.GetMatrixProfileResults(situationID, situationInstanceID, ti)
		if err != nil {
			zap.L().Error("Cannot fetch matrix profile results", zap.Int64("situationID", situationID), zap.Int64("situationInstanceID", situationInstanceID), zap.Error(err))
			return
		}
		widgetData.Aggregates.MatrixProfiles = values
	}
}
