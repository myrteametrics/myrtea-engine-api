package scheduler

import (
	"sync"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

// InternalScheduler represents an instance of a scheduler used for fact processing
type InternalScheduler struct {
	mu          sync.RWMutex
	C           *cron.Cron
	Jobs        map[int64]cron.EntryID
	runningJobs map[int64]bool
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
		Jobs:        make(map[int64]cron.EntryID, 0),
		runningJobs: make(map[int64]bool, 0),
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
		if (fs.Enabled){
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
