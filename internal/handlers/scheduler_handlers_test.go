package handlers

import (
	"encoding/json"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/security/users"
	"net/http"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/scheduler"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/security/permissions"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/tests"
)

func dbSchedulerInit(dbClient *sqlx.DB, t *testing.T) {
	dbSchedulerDestroy(dbClient, t)
	tests.DBExec(dbClient, tests.JobSchedulesTableV1, t, true)
}

func dbSchedulerDestroy(dbClient *sqlx.DB, t *testing.T) {
	tests.DBExec(dbClient, tests.JobSchedulesDropTableV1, t, false)
}

func TestGetJobs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbSchedulerDestroy(db, t)
	dbSchedulerInit(db, t)

	schedulerR := scheduler.NewPostgresRepository(db)
	scheduler.ReplaceGlobalRepository(schedulerR)

	job1 := scheduler.InternalSchedule{
		Name:     "facts_1_2",
		CronExpr: "0/5 * * * *",
		JobType:  "fact",
		Job: scheduler.FactCalculationJob{
			FactIds: []int64{1, 2},
		},
	}
	jobID1, err := schedulerR.Create(job1)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	job1.ID = jobID1

	job2 := scheduler.InternalSchedule{
		Name:     "facts_1_3",
		CronExpr: "0 0/20 * * *",
		JobType:  "fact",
		Job: scheduler.FactCalculationJob{
			FactIds: []int64{1, 2},
		},
	}
	jobID2, err := schedulerR.Create(job2)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	job2.ID = jobID2

	user := users.UserWithPermissions{Permissions: []permissions.Permission{permissions.New(permissions.TypeScheduler, "*", permissions.ActionList)}}
	rr := tests.BuildTestHandler(t, "POST", "/jobs", ``, "/jobs", GetJobSchedules, user)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	mapJobs := []scheduler.InternalSchedule{job1, job2}

	jobsData, _ := json.Marshal(mapJobs)
	expected := string(jobsData) + "\n"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}
