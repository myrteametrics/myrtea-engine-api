package scheduler

import (
	"fmt"
	"sync"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

// InternalScheduler represents an instance of a scheduler used for fact processing
type InternalScheduler struct {
	mu          sync.RWMutex
	C           *cron.Cron
	Jobs        map[int64]RuntimeJobState
	runningJobs map[int64]bool
	RuleEngine  chan string
}

// FrequencyMode defines the active cron profile of a schedule.
type FrequencyMode string

const (
	FrequencyModeNormal FrequencyMode = "normal"
	FrequencyModeBoost  FrequencyMode = "boost"
)

// RuntimeJobState stores scheduler-only metadata for fast in-memory frequency switching.
type RuntimeJobState struct {
	EntryID    cron.EntryID
	Job        InternalJob
	JobType    string
	Mode       FrequencyMode
	NormalCron string
}

var (
	_globalInternalSchedulerMu sync.RWMutex
	_globalInternalScheduler   *InternalScheduler
	cronParser                 = cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
)

// S is used to access the global fact scheduler singleton
func S() *InternalScheduler {
	_globalInternalSchedulerMu.RLock()
	defer _globalInternalSchedulerMu.RUnlock()

	factScheduler := _globalInternalScheduler
	return factScheduler
}

// ReplaceGlobals affect a new repository to the global fact scheduler singleton
func ReplaceGlobals(scheduler *InternalScheduler) func() {
	_globalInternalSchedulerMu.Lock()
	defer _globalInternalSchedulerMu.Unlock()

	prev := _globalInternalScheduler
	_globalInternalScheduler = scheduler
	return func() { ReplaceGlobals(prev) }
}

// NewScheduler returns a pointer to a new instance of InternalScheduler
func NewScheduler() *InternalScheduler {
	c := cron.New()
	scheduler := &InternalScheduler{
		C:           c,
		Jobs:        make(map[int64]RuntimeJobState),
		runningJobs: make(map[int64]bool),
	}
	return scheduler
}

// AddJobSchedule add a new schedule to the current scheduler
func (s *InternalScheduler) AddJobSchedule(schedule InternalSchedule) error {
	zap.L().Info("Adding new schedule", zap.Any("schedule", schedule))

	if !schedule.Enabled {
		s.RemoveJobSchedule(schedule.ID)
		return nil
	}

	if _, err := cronParser.Parse(schedule.CronExpr); err != nil {
		return err
	}

	isFactBoostManagedSchedule := isFactBoostManagedSchedule(schedule)
	if isFactBoostManagedSchedule {
		if _, err := cronParser.Parse(boostCronFromSchedule(schedule)); err != nil {
			return err
		}
	}

	mode := FrequencyModeNormal
	if prev, ok := s.Jobs[schedule.ID]; ok {
		s.C.Remove(prev.EntryID)
		if isFactBoostManagedSchedule {
			schedule = mergeBoostRuntimeState(schedule, prev)
			switch prev.Mode {
			case FrequencyModeBoost:
				mode = FrequencyModeBoost
			}
		}
	}

	return s.RescheduleJob(schedule, mode)
}

// RemoveJobSchedule add a new job to the current scheduler
func (s *InternalScheduler) RemoveJobSchedule(scheduleID int64) {
	zap.L().Info("Removing schedule", zap.Any("schedule", scheduleID))

	if state, ok := s.Jobs[scheduleID]; ok {
		s.C.Remove(state.EntryID)
		delete(s.Jobs, scheduleID)
	}
}

// RescheduleJob removes an existing schedule and adds it again with a new cron expression
func (s *InternalScheduler) RescheduleJob(schedule InternalSchedule, mode FrequencyMode) error {
	schedule = applyModeToScheduleJob(schedule, mode)
	newCronExpr := resolveCronExpr(schedule, mode)

	zap.L().Info(
		"Rescheduling job",
		zap.Int64("scheduleID", schedule.ID),
		zap.String("cronExpr", newCronExpr),
		zap.String("mode", string(mode)),
	)

	entryID, err := s.C.AddJob(newCronExpr, schedule.Job)
	if err != nil {
		return fmt.Errorf("failed to reschedule job %d: %w", schedule.ID, err)
	}

	s.Jobs[schedule.ID] = buildRuntimeState(schedule, entryID, mode)

	return nil
}

func resolveCronExpr(schedule InternalSchedule, mode FrequencyMode) string {
	switch mode {
	case FrequencyModeBoost:
		return boostCronFromSchedule(schedule)
	case FrequencyModeNormal:
		return schedule.CronExpr
	default:
		return schedule.CronExpr
	}
}

// SwitchJobFrequency switches a schedule between normal and boost cron frequencies at runtime.
// Persisted schedule changes remain managed by repository update flows (e.g. PUT handler + AddJobSchedule).
func (s *InternalScheduler) SwitchJobFrequency(scheduleID int64, mode FrequencyMode) (InternalSchedule, error) {
	state, ok := s.Jobs[scheduleID]
	if !ok {
		return InternalSchedule{}, fmt.Errorf("schedule %d not loaded in runtime scheduler", scheduleID)
	}
	if state.Job == nil {
		return InternalSchedule{}, fmt.Errorf("runtime job not found for schedule %d", scheduleID)
	}

	runtimeSchedule := InternalSchedule{
		ID:       scheduleID,
		Job:      state.Job,
		Enabled:  true,
		CronExpr: state.NormalCron,
		JobType:  state.JobType,
	}

	s.C.Remove(state.EntryID)
	if err := s.RescheduleJob(runtimeSchedule, mode); err != nil {
		return InternalSchedule{}, err
	}

	runtimeSchedule = applyModeToScheduleJob(runtimeSchedule, mode)
	runtimeSchedule.CronExpr = resolveCronExpr(runtimeSchedule, mode)

	return runtimeSchedule, nil
}

func buildRuntimeState(schedule InternalSchedule, entryID cron.EntryID, mode FrequencyMode) RuntimeJobState {
	state := RuntimeJobState{
		EntryID:    entryID,
		Job:        schedule.Job,
		JobType:    schedule.JobType,
		Mode:       mode,
		NormalCron: schedule.CronExpr,
	}

	return state
}

func boostCronFromSchedule(schedule InternalSchedule) string {
	factJob, ok := extractFactCalculationJob(schedule)
	if !ok || !boostConfigured(factJob.JobBoostInfo) {
		return ""
	}
	return factJob.JobBoostInfo.Frequency
}

func mergeBoostRuntimeState(schedule InternalSchedule, previousState RuntimeJobState) InternalSchedule {
	factJob, ok := extractFactCalculationJob(schedule)
	if !ok || !boostConfigured(factJob.JobBoostInfo) {
		return schedule
	}

	boostCopy := *factJob.JobBoostInfo
	prevFactJob, ok := asFactCalculationJob(previousState.Job)
	if ok && boostConfigured(prevFactJob.JobBoostInfo) {
		// Important:
		// if a schedule is edited while boost mode is active, we keep the runtime Used counter.
		// Resetting Used here would lose already-consumed quota and could overrun boost executions.
		boostCopy.Used = prevFactJob.JobBoostInfo.Used
	}

	factJob.JobBoostInfo = &boostCopy
	schedule.Job = factJob
	return schedule
}

func applyModeToScheduleJob(schedule InternalSchedule, mode FrequencyMode) InternalSchedule {
	factJob, ok := extractFactCalculationJob(schedule)
	if !ok || factJob.JobBoostInfo == nil {
		return schedule
	}

	boostCopy := *factJob.JobBoostInfo
	if !boostCopy.Configured {
		boostCopy.Active = false
	} else {
		boostCopy.Active = mode == FrequencyModeBoost
	}
	factJob.JobBoostInfo = &boostCopy
	schedule.Job = factJob

	return schedule
}

// Init loads the job schedules from Data Base
func (s *InternalScheduler) Init() error {
	schedules, err := R().GetAll()
	if err != nil {
		return err
	}
	for _, fs := range schedules {
		if fs.Enabled {
			err := s.AddJobSchedule(fs)
			if err != nil {
				return err
			}
		}

	}
	return nil
}

// ExistingRunningJob check if a job is already running
func (s *InternalScheduler) ExistingRunningJob(scheduleID int64) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, ok := s.runningJobs[scheduleID]
	return ok
}

// AddRunningJob add a job ID to the running job list
func (s *InternalScheduler) AddRunningJob(scheduleID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.runningJobs[scheduleID] = true
}

// RemoveRunningJob remove a job ID to the running job list
func (s *InternalScheduler) RemoveRunningJob(scheduleID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.runningJobs, scheduleID)
}
