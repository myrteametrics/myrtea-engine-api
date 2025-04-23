package scheduler

import (
	"encoding/json"
	"errors"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/plugins/baseline"

	"go.uber.org/zap"
)

// BaselineCalculationJob represent a scheduler job instance which process a group of baselines, and persist the result in postgresql
type BaselineCalculationJob struct {
	BaselineIds []int64 `json:"baselineIds"`
	Debug       bool    `json:"debug"`
	ScheduleID  int64   `json:"-"`
}

// IsValid checks if an internal schedule job definition is valid and has no missing mandatory fields
func (job BaselineCalculationJob) IsValid() (bool, error) {
	if job.BaselineIds == nil {
		return false, errors.New("missing BaselineIds")
	}
	if len(job.BaselineIds) <= 0 {
		return false, errors.New("missing BaselineIds")
	}
	return true, nil
}

// Run contains all the business logic of the job
func (job BaselineCalculationJob) Run() {

	if S().ExistingRunningJob(job.ScheduleID) {
		zap.L().Info("Skipping BaselineScheduleJob because last execution is still running", zap.Int64s("ids", job.BaselineIds))
		return
	}
	S().AddRunningJob(job.ScheduleID)

	zap.L().Info("Baseline calculation job started", zap.Int64s("ids", job.BaselineIds))

	pluginBaseline, err := baseline.P()
	if err == nil {
		for _, b := range job.BaselineIds {
			err := pluginBaseline.BaselineService.BuildBaselineValues(b)
			if err != nil {
				zap.L().Error("BuildBaselineValues", zap.Int64("baselineID", b), zap.Error(err))
				S().RemoveRunningJob(job.ScheduleID)
				return
			}
		}
	} else {
		zap.L().Warn("Cannot execute BaselineScheduleJob. Plugin is unavailable")
	}

	zap.L().Info("BaselineScheduleJob Ended", zap.Int64s("ids", job.BaselineIds))

	S().RemoveRunningJob(job.ScheduleID)
}

// UnmarshalJSON unmarshals a quoted json string to a valid BaselineCalculationJob struct
func (job *BaselineCalculationJob) UnmarshalJSON(data []byte) error {
	type Alias BaselineCalculationJob
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(job),
	}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}
	return nil
}
