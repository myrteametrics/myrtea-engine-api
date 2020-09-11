package scheduler

import (
	"testing"
)

func TestNewScheduler(t *testing.T) {
	s := NewScheduler()
	if s == nil {
		t.Error("FactScheduler is nil")
	}
}

func TestReplaceGlobalScheduler(t *testing.T) {
	ReplaceGlobals(nil)
	s := NewScheduler()
	reverse := ReplaceGlobals(s)
	if S() == nil {
		t.Error("global scheduler is nil")
	}
	reverse()
	if S() != nil {
		t.Error("global scheduler is not nil after reverse")
	}
}

func TestAddJobSchedule(t *testing.T) {
	fs := InternalSchedule{
		Name:     "test",
		CronExpr: "*/15 * * * *",
		JobType:  "fact",
		Job: FactCalculationJob{
			FactIds: []int64{1, 2},
		},
	}
	s := NewScheduler()
	err := s.AddJobSchedule(fs)
	if err != nil {
		t.Error(err)
	}
	if len(s.C.Entries()) == 0 {
		t.Error("New fact schedule not added properly to the cron entries")
	}
}

func TestAddJobScheduleInvalidCron(t *testing.T) {
	fs := InternalSchedule{
		Name:     "test",
		CronExpr: "*/15 * a a aa",
		JobType:  "fact",
		Job: FactCalculationJob{
			FactIds: []int64{1, 2},
		},
	}
	s := NewScheduler()
	err := s.AddJobSchedule(fs)
	if err == nil {
		t.Error("Invalid fact schedule must not be added to the scheduler")
	}
	if len(s.C.Entries()) > 0 {
		t.Error("A fact schedule has been added while it must not")
	}
}
