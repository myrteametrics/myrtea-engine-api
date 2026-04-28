package scheduler

import (
	"testing"

	"github.com/myrteametrics/myrtea-engine-api/v5/internal/model"
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
		Enabled:  true,
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
		Enabled:  true,
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

func TestRescheduleJob(t *testing.T) {
	fs := InternalSchedule{
		ID:       1,
		Name:     "test",
		CronExpr: "*/15 * * * *",
		JobType:  "fact",
		Enabled:  true,
		Job: FactCalculationJob{
			FactIds: []int64{1, 2},
		},
	}

	s := NewScheduler()
	if err := s.AddJobSchedule(fs); err != nil {
		t.Fatal(err)
	}
	oldEntryID := s.Jobs[fs.ID].EntryID

	if err := s.RescheduleJob(fs, FrequencyModeNormal); err != nil {
		t.Fatal(err)
	}

	newState, ok := s.Jobs[fs.ID]
	if !ok {
		t.Fatal("rescheduled job is missing in scheduler job map")
	}
	if newState.EntryID == oldEntryID {
		t.Fatal("rescheduled job must have a new entry ID")
	}
	if len(s.C.Entries()) != 1 {
		t.Fatalf("expected exactly one cron entry after reschedule, got %d", len(s.C.Entries()))
	}
}

func TestRescheduleJobInvalidCron(t *testing.T) {
	fs := InternalSchedule{
		ID:       1,
		Name:     "test",
		CronExpr: "*/15 * * * *",
		JobType:  "fact",
		Enabled:  true,
		Job: FactCalculationJob{
			FactIds: []int64{1, 2},
		},
	}

	s := NewScheduler()
	if err := s.AddJobSchedule(fs); err != nil {
		t.Fatal(err)
	}
	oldEntryID := s.Jobs[fs.ID].EntryID

	fs.CronExpr = "bad cron"
	if err := s.RescheduleJob(fs, FrequencyModeNormal); err == nil {
		t.Fatal("expected error while rescheduling with invalid cron")
	}

	newState, ok := s.Jobs[fs.ID]
	if !ok {
		t.Fatal("existing job should remain after failed reschedule")
	}
	if newState.EntryID != oldEntryID {
		t.Fatal("existing entry ID should not change when reschedule fails")
	}
	if len(s.C.Entries()) != 1 {
		t.Fatalf("expected exactly one cron entry after failed reschedule, got %d", len(s.C.Entries()))
	}
}

func TestSwitchJobFrequencyBoostAndRevert(t *testing.T) {
	s := NewScheduler()

	schedule := InternalSchedule{
		ID:       42,
		Name:     "fact-switch",
		CronExpr: "*/15 * * * *",
		JobType:  "fact",
		Enabled:  true,
		Job: FactCalculationJob{
			FactIds: []int64{1, 2},
			JobBoostInfo: &model.JobBoostInfo{
				Configured: true,
				JobID:      "42",
				Frequency:  "*/2 * * * *",
				Quota:      10,
			},
		},
	}

	if err := s.AddJobSchedule(schedule); err != nil {
		t.Fatal(err)
	}

	s.SwitchJobFrequency(schedule.ID, FrequencyModeBoost)

	boostState := s.Jobs[schedule.ID]
	if boostState.Mode != FrequencyModeBoost {
		t.Fatalf("expected runtime mode boost, got %s", boostState.Mode)
	}
	boostStateJob, ok := boostState.Job.(FactCalculationJob)
	if !ok || boostStateJob.JobBoostInfo == nil {
		t.Fatal("expected runtime fact job with boost info after boost switch")
	}
	if !boostStateJob.JobBoostInfo.Active {
		t.Fatal("expected runtime boost info active after boost switch")
	}

	s.SwitchJobFrequency(schedule.ID, FrequencyModeNormal)

	normalState := s.Jobs[schedule.ID]
	if normalState.Mode != FrequencyModeNormal {
		t.Fatalf("expected runtime mode normal, got %s", normalState.Mode)
	}
	if normalState.NormalCron != "*/15 * * * *" {
		t.Fatalf("expected runtime normal cron, got %s", normalState.NormalCron)
	}
	normalStateJob, ok := normalState.Job.(FactCalculationJob)
	if !ok || normalStateJob.JobBoostInfo == nil {
		t.Fatal("expected runtime fact job with boost info after normal switch")
	}
	if normalStateJob.JobBoostInfo.Active {
		t.Fatal("expected runtime boost info inactive after normal switch")
	}
}

func TestAddJobScheduleKeepsCurrentModeAndUsed(t *testing.T) {
	s := NewScheduler()

	initial := InternalSchedule{
		ID:       50,
		Name:     "fact-keep-mode",
		CronExpr: "*/15 * * * *",
		JobType:  "fact",
		Enabled:  true,
		Job: FactCalculationJob{
			FactIds: []int64{1},
			JobBoostInfo: &model.JobBoostInfo{
				Configured: true,
				JobID:      "50",
				Frequency:  "*/2 * * * *",
				Quota:      10,
				Used:       0,
			},
		},
	}

	if err := s.AddJobSchedule(initial); err != nil {
		t.Fatal(err)
	}
	s.SwitchJobFrequency(initial.ID, FrequencyModeBoost)

	// Simulate runtime progress while boosted.
	state := s.Jobs[initial.ID]
	stateJob, ok := state.Job.(FactCalculationJob)
	if !ok {
		t.Fatal("expected runtime FactCalculationJob")
	}
	if stateJob.JobBoostInfo == nil {
		t.Fatal("expected runtime boost info")
	}
	stateBoost := *stateJob.JobBoostInfo
	stateBoost.Used = 3
	stateJob.JobBoostInfo = &stateBoost
	state.Job = stateJob
	s.Jobs[initial.ID] = state

	updatedSchedule := InternalSchedule{
		ID:       50,
		Name:     "fact-keep-mode",
		CronExpr: "*/30 * * * *", // edited normal cron
		JobType:  "fact",
		Enabled:  true,
		Job: FactCalculationJob{
			FactIds: []int64{1},
			JobBoostInfo: &model.JobBoostInfo{
				Configured: true,
				JobID:      "50",
				Frequency:  "*/1 * * * *", // edited boost cron
				Quota:      20,            // edited quota
				Used:       0,             // should keep runtime used
			},
		},
	}

	if err := s.AddJobSchedule(updatedSchedule); err != nil {
		t.Fatal(err)
	}

	after := s.Jobs[initial.ID]
	if after.Mode != FrequencyModeBoost {
		t.Fatalf("expected mode to stay boost after update, got %s", after.Mode)
	}
	if after.NormalCron != "*/30 * * * *" {
		t.Fatalf("expected updated normal cron retained, got %s", after.NormalCron)
	}
	afterJob, ok := after.Job.(FactCalculationJob)
	if !ok || afterJob.JobBoostInfo == nil {
		t.Fatal("expected runtime FactCalculationJob with boost info")
	}
	if afterJob.JobBoostInfo.Frequency != "*/1 * * * *" {
		t.Fatalf("expected updated boost cron after update, got %s", afterJob.JobBoostInfo.Frequency)
	}
	if afterJob.JobBoostInfo.Quota != 20 {
		t.Fatalf("expected updated quota 20, got %d", afterJob.JobBoostInfo.Quota)
	}
	if afterJob.JobBoostInfo.Used != 0 {
		t.Fatalf("expected runtime used preserved (0), got %d", afterJob.JobBoostInfo.Used)
	}
	if !afterJob.JobBoostInfo.Active {
		t.Fatal("expected boost info to remain active while runtime mode is boost")
	}
}
