package scheduler

import (
	"errors"
	"time"

	"github.com/myrteametrics/myrtea-engine-api/v5/internal/history"
	"go.uber.org/zap"
)

// PurgeHistoryJob represent a scheduler job instance which process a group of Purge history, and persist the result in postgresql
type PurgeHistoryJob struct {
	SituationID         int64             `json:"situationId"`
	SituationInstanceID int64             `json:"situationInstanceId"`
	ParameterFilters    map[string]string `json:"parameterFilters"`
	DeleteBeforeTs      string            `json:"deleteBeforeTs"`
	ScheduleID          int64             `json:"-"`
}

// IsValid checks if an internal schedule job definition is valid and has no missing mandatory fields
func (job PurgeHistoryJob) IsValid() (bool, error) {

	if _, err := parseDuration(job.DeleteBeforeTs); err != nil {
		return false, errors.New(`Error parsing the  Purge's DeleteBeforeTs `)
	}

	return true, nil
}

// Run contains all the business logic of the job
func (job PurgeHistoryJob) Run() {

	if S().ExistingRunningJob(job.ScheduleID) {
		zap.L().Info("Skipping Purge ScheduleJob because last execution is still running", zap.Int64("id 	Schedule  ", job.ScheduleID))
		return
	}
	S().AddRunningJob(job.ScheduleID)

	zap.L().Info("Purge history  job started", zap.Int64("id Schedule ", job.ScheduleID))

	DeleteBeforeTsDuration, err := parseDuration(job.DeleteBeforeTs)
	if err != nil {
		zap.L().Info("Error parsing the Purge's DeleteBeforeTs ", zap.Error(err), zap.Int64("idSchedule", job.ScheduleID))
		S().RemoveRunningJob(job.ScheduleID)
		return
	}

	options := history.GetHistorySituationsOptions{
		SituationID:         -1,
		SituationInstanceID: -1,
		ParameterFilters:    make(map[string]string),
		DeleteBeforeTs:      time.Now().Add(-1 * DeleteBeforeTsDuration),
	}

	err = history.S().PurgeHistory(options)

	if err != nil {
		zap.L().Info("Purge History job error", zap.Error(err), zap.Int64("idSchedule", job.ScheduleID))
		S().RemoveRunningJob(job.ScheduleID)
		return
	}

	zap.L().Info("Purge history  job  Ended", zap.Int64("id Schedule", job.ScheduleID))

	S().RemoveRunningJob(job.ScheduleID)
}
