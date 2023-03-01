package scheduler

import (
	"encoding/json"
	"errors"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

// InternalJob embed the external "standard" cron job with some additionnal data
type InternalJob interface {
	cron.Job
	IsValid() (bool, error)
}

// InternalSchedule wrap a schedule
type InternalSchedule struct {
	ID       int64       `json:"id"`
	Name     string      `json:"name"`
	CronExpr string      `json:"cronexpr" example:"0 */15 * * *"`
	JobType  string      `json:"jobtype" enums:"fact,baseline"`
	Job      InternalJob `json:"job"`
}

// IsValid checks if an internal schedule definition is valid and has no missing mandatory fields
func (schedule *InternalSchedule) IsValid() (bool, error) {
	if schedule.Name == "" {
		return false, errors.New("missing Name")
	}
	if schedule.CronExpr == "" {
		return false, errors.New("missing CronExpr")
	}
	if _, err := cronParser.Parse(schedule.CronExpr); err != nil {
		return false, errors.New("invalid CronExpr" + err.Error())
	}
	if schedule.JobType == "" {
		return false, errors.New("missing JobType")
	}
	if schedule.JobType != "fact" && schedule.JobType != "baseline" {
		// if schedule.JobType != "fact" {
		return false, errors.New("invalid JobType")
	}
	if schedule.Job == nil {
		return false, errors.New("missing Name")
	}
	if ok, err := schedule.Job.IsValid(); !ok {
		return false, errors.New("job is invalid:" + err.Error())
	}
	return true, nil
}

// UnmarshalJSON unmarshals a json object as a InternalSchedule
func (schedule *InternalSchedule) UnmarshalJSON(data []byte) error {
	type Alias InternalSchedule
	aux := &struct {
		Job *json.RawMessage `json:"job"`
		*Alias
	}{
		Alias: (*Alias)(schedule),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	job, err := UnmarshalInternalJob(aux.JobType, *aux.Job, aux.ID)
	if err != nil {
		return err
	}
	schedule.Job = job
	return nil
}

// UnmarshalInternalJob unmarshal a fact from a json string
func UnmarshalInternalJob(t string, b json.RawMessage, scheduleID int64) (InternalJob, error) {
	var job InternalJob
	var err error
	switch t {
	case "fact":
		var tJob FactCalculationJob
		err = json.Unmarshal(b, &tJob)
		tJob.ScheduleID = scheduleID
		job = tJob

	case "baseline":
		var tJob BaselineCalculationJob
		err = json.Unmarshal(b, &tJob)
		tJob.ScheduleID = scheduleID
		job = tJob

	default:
		zap.L().Error("unknown internal job type", zap.String("type", t))
	}
	return job, err
}
