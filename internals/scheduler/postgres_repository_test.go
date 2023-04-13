package scheduler

import (
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/tests"
)

func dbInit(dbClient *sqlx.DB, t *testing.T) {
	dbDestroy(dbClient, t)
	tests.DBExec(dbClient, tests.JobSchedulesTableV1, t, true)
}

func dbDestroy(dbClient *sqlx.DB, t *testing.T) {
	tests.DBExec(dbClient, tests.JobSchedulesDropTableV1, t, false)
}

func TestNewPostgresRepository(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	r := NewPostgresRepository(tests.DBClient(t))
	if r == nil {
		t.Error("Notifications Repository is nil")
	}
}

func TestCreateFactJob(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)

	defer dbDestroy(db, t)
	dbInit(db, t)
	r := NewPostgresRepository(db)
	schedule := InternalSchedule{
		Name:     "test-1",
		CronExpr: "0 0 * * * *",
		JobType:  "fact",
		Job: FactCalculationJob{
			FactIds: []int64{1, 2},
		},
		Enabled: true,
	}
	_, err := r.Create(schedule)
	if err != nil {
		t.Error(err)
	}
}

func TestGetJobs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroy(db, t)
	dbInit(db, t)
	r := NewPostgresRepository(db)

	schedule1 := InternalSchedule{
		Name:     "test-1",
		CronExpr: "0 0 * * * *",
		JobType:  "fact",
		Job: FactCalculationJob{
			FactIds: []int64{1, 2},
		},
		Enabled: true,
	}
	_, err := r.Create(schedule1)
	if err != nil {
		t.Error(err)
	}

	schedule2 := InternalSchedule{
		Name:     "test-2",
		CronExpr: "0 0 * * * *",
		JobType:  "fact",
		Job: FactCalculationJob{
			FactIds: []int64{1, 2},
		},
		Enabled: true,
	}
	_, err = r.Create(schedule2)
	if err != nil {
		t.Error(err)
	}

	schedules, err := r.GetAll()
	if err != nil {
		t.Error(err)
	}

	for _, schedule := range schedules {
		switch v := schedule.Job.(type) {
		case FactCalculationJob:
			if v.ScheduleID != schedule.ID {
				t.Error("The scheduleID was not correctly setted in the job ")
			}
		}
	}

}
