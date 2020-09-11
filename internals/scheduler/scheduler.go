package scheduler

import (
	"sync"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

// InternalScheduler represents an instance of a scheduler used for fact processing
type InternalScheduler struct {
	C           *cron.Cron
	Jobs        map[int64]cron.EntryID
	RunningJobs map[int64]bool
	RuleEngine  chan string
}

var (
	_globalInternalSchedulerMu sync.RWMutex
	_globalInternalScheduler   *InternalScheduler
	cronParser                 = cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
)

// S is used to access the global fact scheduler singleton
func S() *InternalScheduler {
	_globalInternalSchedulerMu.RLock()
	factScheduler := _globalInternalScheduler
	_globalInternalSchedulerMu.RUnlock()
	return factScheduler
}

// ReplaceGlobals affect a new repository to the global fact scheduler singleton
func ReplaceGlobals(scheduler *InternalScheduler) func() {
	_globalInternalSchedulerMu.Lock()
	prev := _globalInternalScheduler
	_globalInternalScheduler = scheduler
	_globalInternalSchedulerMu.Unlock()
	return func() { ReplaceGlobals(prev) }
}

// NewScheduler returns a pointer to a new instance of InternalScheduler
func NewScheduler() *InternalScheduler {
	c := cron.New()
	scheduler := &InternalScheduler{
		C:           c,
		Jobs:        make(map[int64]cron.EntryID, 0),
		RunningJobs: make(map[int64]bool, 0),
	}
	return scheduler
}

// AddJobSchedule add a new schedule to the current scheduler
func (s *InternalScheduler) AddJobSchedule(schedule InternalSchedule) error {
	zap.L().Info("Adding new schedule", zap.Any("schedule", schedule))

	if entryID, ok := s.Jobs[schedule.ID]; ok {
		s.C.Remove(entryID)
	}

	entryID, err := s.C.AddJob(schedule.CronExpr, schedule.Job)
	if err != nil {
		return err
	}
	s.Jobs[schedule.ID] = entryID

	return nil
}

// RemoveJobSchedule add a new job to the current scheduler
func (s *InternalScheduler) RemoveJobSchedule(scheduleID int64) {
	zap.L().Info("Removing schedule", zap.Any("schedule", scheduleID))

	if entryID, ok := s.Jobs[scheduleID]; ok {
		s.C.Remove(entryID)
		delete(s.Jobs, scheduleID)
	}
}

// Init loads the job schedules from Data Base
func (s *InternalScheduler) Init() error {
	schedules, err := R().GetAll()
	if err != nil {
		return err
	}
	for _, fs := range schedules {
		err := s.AddJobSchedule(fs)
		if err != nil {
			return err
		}
	}
	return nil
}
