package scheduler

import (
	"errors"
	"strings"
	"time"

	"github.com/myrteametrics/myrtea-engine-api/v5/internals/history"
	"go.uber.org/zap"
)

// PurgeHistoryJob represent a scheduler job instance which process a group of Purge history, and persist the result in postgresql
type PurgeHistoryJob struct {
	SituationID         int64             `json:"situationId"`
	SituationInstanceID int64             `json:"situationInstanceId"`
	ParameterFilters    map[string]string `json:"parameterFilters"`
	FromOffset          string            `json:"fromOffset"`
	ToOffset            string            `json:"toOffset"`
	Interval            string            `json:"interval"`
	ScheduleID          int64             `json:"-"`
}

// IsValid checks if an internal schedule job definition is valid and has no missing mandatory fields
func (job PurgeHistoryJob) IsValid() (bool, error) {

	var err error
	var fromOffsetDuration time.Duration
	var toOffsetDuration time.Duration

	if fromOffsetDuration, err = parseDuration(job.FromOffset); err != nil {
		return false, errors.New(`Error parsing the  Purge's FromOffset `)
	}

	if toOffsetDuration, err = parseDuration(job.ToOffset); err != nil {
		return false, errors.New(`Error parsing the  Purge's FromOffset `)
	}

	if toOffsetDuration < fromOffsetDuration  {
		return false, errors.New(`FromOffset Duration must be less than ToOffset duration `)
	}

	if !strings.EqualFold(job.Interval, Day) && !strings.EqualFold(job.Interval, Hour) {
		return false, errors.New(`the Purge's  Internal value is unknown`)
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

	// calculer Duration
	fromOffsetDuration, _ := parseDuration(job.FromOffset)
	toOffsetDuration, _ := parseDuration(job.ToOffset)

	options := history.GetHistorySituationsOptions{
		SituationID:         -1,
		SituationInstanceID: -1,
		FromTS:              time.Now().Add(-1 * fromOffsetDuration),
		ToTS:                time.Now().Add(-1 * toOffsetDuration),
		ParameterFilters:    make(map[string]string),
	}

	interval := job.Interval

	history.S().PurgeHistory(options, interval)

	zap.L().Info("Purge history  job  Ended", zap.Int64("id Schedule", job.ScheduleID))

	S().RemoveRunningJob(job.ScheduleID)
}
