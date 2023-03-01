package scheduler

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/myrteametrics/myrtea-engine-api/v5/plugins/baseline"
	"github.com/myrteametrics/myrtea-sdk/v4/expression"
	"go.uber.org/zap"
)

// BaselineCalculationJob represent a scheduler job instance which process a group of baselines, and persist the result in postgresql
type BaselineCalculationJob struct {
	BaselineIds    []int64 `json:"baselineIds"`
	From           string  `json:"from,omitempty"`
	To             string  `json:"to,omitempty"`
	LastDailyValue bool    `json:"lastDailyValue,omitempty"`
	Debug          bool    `json:"debug"`
	ScheduleID     int64   `json:"-"`
}

// // ResolveFromAndTo resolves the expressions in parameters From and To
func (job *BaselineCalculationJob) ResolveFromAndTo(t time.Time) (time.Time, time.Time, error) {

	var from time.Time
	var to time.Time

	if job.From == "" && job.To == "" {
		return from, to, nil
	}
	if job.From == "" || job.To == "" {
		return from, to, errors.New("missing From or To Parameter")
	}

	variables := expression.GetDateKeywords(t)
	result, err := expression.Process(expression.LangEval, job.From, variables)
	if err != nil {
		zap.L().Error("Error processing From expression in baseline calculation job", zap.Error(err))
		return from, to, err
	}
	from, err = time.ParseInLocation(timeLayout, result.(string), time.UTC)
	if err != nil {
		zap.L().Error("Error parsing From expression result as datetime in baseline calculation job", zap.Error(err))
		return from, to, err
	}

	result, err = expression.Process(expression.LangEval, job.To, variables)
	if err != nil {
		zap.L().Error("Error processing To expression in baseline calculation job", zap.Error(err))
		return from, to, err
	}
	to, err = time.ParseInLocation(timeLayout, result.(string), time.UTC)
	if err != nil {
		zap.L().Error("Error parsing To expression result as datetime in baseline calculation job", zap.Error(err))
		return from, to, err
	}

	if job.LastDailyValue {
		from = time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, from.Location())
		to = time.Date(to.Year(), to.Month(), to.Day(), 23, 59, 59, 0, to.Location())
	}

	return from, to, nil
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
	if _, _, err := job.ResolveFromAndTo(time.Now()); err != nil {
		return err
	}
	return nil
}
