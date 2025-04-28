package scheduler

import (
	"errors"
	"time"

	"github.com/myrteametrics/myrtea-engine-api/v5/internal/history"
	"go.uber.org/zap"
)

// CpmpactHistoryJob represent a scheduler job instance which process a group of compact history, and persist the result in postgresql
type CompactHistoryJob struct {
	SituationID         int64             `json:"situationId"`
	SituationInstanceID int64             `json:"situationInstanceId"`
	ParameterFilters    map[string]string `json:"parameterFilters"`
	FromOffset          string            `json:"fromOffset"`
	ToOffset            string            `json:"toOffset"`
	Interval            string            `json:"interval"`
	ScheduleID          int64             `json:"-"`
}

// Internal Job
const (
	Day  string = "day"
	Hour string = "hour"
)

// IsValid checks if an internal schedule job definition is valid and has no missing mandatory fields
func (job CompactHistoryJob) IsValid() (bool, error) {

	var err error
	var fromOffsetDuration time.Duration
	var toOffsetDuration time.Duration

	if fromOffsetDuration, err = parseDuration(job.FromOffset); err != nil {
		return false, errors.New(`Error parsing the Compact's FromOffset `)
	}

	if toOffsetDuration, err = parseDuration(job.ToOffset); err != nil {
		return false, errors.New(`Error parsing the Compact's FromOffset `)
	}

	if toOffsetDuration < fromOffsetDuration {
		return false, errors.New(`FromOffset Duration must be less than ToOffset duration `)
	}

	if job.Interval != Day && job.Interval != Hour {
		return false, errors.New(`the Compact's  Internal value is unknown`)
	}
	return true, nil
}

// Run contains all the business logic of the job
func (job CompactHistoryJob) Run() {

	if S().ExistingRunningJob(job.ScheduleID) {
		zap.L().Info("Skipping Compact ScheduleJob because last execution is still running", zap.Int64("id 	Schedule  ", job.ScheduleID))
		return
	}
	S().AddRunningJob(job.ScheduleID)

	zap.L().Info("Compact history  job started", zap.Int64("id Schedule ", job.ScheduleID))

	fromOffsetDuration, err := parseDuration(job.FromOffset)
	if err != nil {
		zap.L().Info("Error parsing the Compact's FromOffset ", zap.Error(err), zap.Int64("idSchedule", job.ScheduleID))
		S().RemoveRunningJob(job.ScheduleID)
		return
	}

	toOffsetDuration, err := parseDuration(job.ToOffset)
	if err != nil {
		zap.L().Info("Error parsing the Compact's FromOffset ", zap.Error(err), zap.Int64("idSchedule", job.ScheduleID))
		S().RemoveRunningJob(job.ScheduleID)
		return
	}

	if toOffsetDuration < fromOffsetDuration {
		zap.L().Info("the Compact's FromOffset Duration must be less than ToOffset duration ", zap.Error(err), zap.Int64("idSchedule", job.ScheduleID))
		S().RemoveRunningJob(job.ScheduleID)
		return
	}

	options := history.GetHistorySituationsOptions{
		SituationID:         -1,
		SituationInstanceID: -1,
		FromTS:              time.Now().Add(-1 * fromOffsetDuration),
		ToTS:                time.Now().Add(-1 * toOffsetDuration),
		ParameterFilters:    make(map[string]interface{}),
	}

	interval := job.Interval

	err = history.S().CompactHistory(options, interval)
	if err != nil {
		zap.L().Info("Compact History job error", zap.Error(err), zap.Int64("idSchedule", job.ScheduleID))
		S().RemoveRunningJob(job.ScheduleID)
		return
	}

	zap.L().Info("Compact history  job  Ended", zap.Int64("id Schedule", job.ScheduleID))

	S().RemoveRunningJob(job.ScheduleID)
}
