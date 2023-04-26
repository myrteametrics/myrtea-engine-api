package scheduler

import (
	"errors"
	"strings"
	"time"

	"go.uber.org/zap"
)

// CpmpactHistoryJob represent a scheduler job instance which process a group of compact history, and persist the result in postgresql
type CompactHistoryJob struct {
	SituationID         int64             `json:"situationId"`
	SituationInstanceID int64             `json:"situationInstanceId"`
	ParameterFilters    map[string]string `json:"parameterFilters"`
	FromTS              time.Time         `json:"fromTs"`
	ToTS                time.Time         `json:"toTs"`
	Interval            string            `json:"interval"`
	ScheduleID          int64             `json:"-"`
}

// IsValid checks if an internal schedule job definition is valid and has no missing mandatory fields
func (job CompactHistoryJob) IsValid() (bool, error) {
    
	if job.FromTS.IsZero() {
		return false, errors.New("missing the start date of the compact history")
	}
	if job.ToTS.IsZero(){
		return false, errors.New("missing the end date of the compact history")
	}

	if job.ToTS.Before(job.FromTS) || job.ToTS == job.FromTS{
		return false, errors.New("Start date must be less than end date ")
	}

	if !strings.EqualFold(job.Interval,Day) && !strings.EqualFold(job.Interval,Hour) {
		return false, errors.New(`the Purge's  Internal value is unknown`)
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

	zap.L().Info("Compact history  job  Ended", zap.Int64("id Schedule", job.ScheduleID))

	S().RemoveRunningJob(job.ScheduleID)
}
