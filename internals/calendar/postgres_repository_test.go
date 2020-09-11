package calendar

import (
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/tests"
	"github.com/myrteametrics/myrtea-sdk/v4/postgres"
)

func dbCalendarInit(dbClient *sqlx.DB, t *testing.T) {
	dbCalendarDestroy(dbClient, t)
	tests.DBExec(dbClient, tests.CalendarTableV3, t, true)
	tests.DBExec(dbClient, tests.CalendarUnionTableV3, t, true)
}

func dbCalendarDestroy(dbClient *sqlx.DB, t *testing.T) {
	tests.DBExec(dbClient, tests.CalendarUnionDropTableV3, t, true)
	tests.DBExec(dbClient, tests.CalendarDropTableV3, t, true)
}

func TestNewPostgresRepository(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	r := NewPostgresRepository(tests.DBClient(t))
	if r == nil {
		t.Error("Calendar Repository is nil")
	}
}

func TestPostgresReplaceGlobal(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	r := NewPostgresRepository(tests.DBClient(t))
	reverse := ReplaceGlobals(r)
	if R() == nil {
		t.Error("Global Calendar repository is nil")
	}
	reverse()
	if R() != nil {
		t.Error("Global Calendar repository is not nil after reverse")
	}
}

func TestPostgresCreateAndGet(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}

	db := tests.DBClient(t)
	postgres.ReplaceGlobals(db)
	defer dbCalendarDestroy(db, t)
	dbCalendarInit(db, t)

	calendarR := NewPostgresRepository(db)
	ReplaceGlobals(calendarR)

	calendarGet, found, err := calendarR.Get(1)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if found {
		t.Error("found a calendar from nowhere")
	}

	period := Period{}
	calend1 := Calendar{
		Name:        "c1",
		Description: "calendar 1",
		Periods:     []Period{period},
	}

	id1, err := calendarR.Create(calend1)
	if err != nil {
		t.Error(err)
	}

	calend2 := Calendar{
		Name:             "c2",
		Description:      "calendar 2",
		Periods:          []Period{period},
		UnionCalendarIDs: []int64{id1},
	}

	id2, err := calendarR.Create(calend2)
	if err != nil {
		t.Error(err)
	}

	calendarGet, found, err = calendarR.Get(id2)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if !found {
		t.Error("Calendar doesn't exists after the creation")
		t.FailNow()
	}
	if id2 != calendarGet.ID || calendarGet.UnionCalendarIDs[0] != id1 {
		t.Error("Create and Get with Unions error")
	}
}

func TestPostgresUpdate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}

	db := tests.DBClient(t)
	postgres.ReplaceGlobals(db)
	defer dbCalendarDestroy(db, t)
	dbCalendarInit(db, t)

	calendarR := NewPostgresRepository(db)
	ReplaceGlobals(calendarR)

	//Create
	period := Period{}
	calend1 := Calendar{
		Name:        "c1",
		Description: "calendar 1",
		Periods:     []Period{period},
	}

	id1, err := calendarR.Create(calend1)
	if err != nil {
		t.Error(err)
	}

	calend2 := Calendar{
		Name:        "c2",
		Description: "calendar 2",
		Periods:     []Period{period},
	}

	id2, err := calendarR.Create(calend2)
	if err != nil {
		t.Error(err)
	}

	name := "c3"
	calend3 := Calendar{
		Name:             name,
		Description:      "calendar 3",
		Periods:          []Period{period},
		UnionCalendarIDs: []int64{id1},
	}

	id3, err := calendarR.Create(calend3)
	if err != nil {
		t.Error(err)
	}

	//Update
	calend3.Name = "new_name"
	calend3.ID = id3
	calend3.UnionCalendarIDs[0] = id2
	err = calendarR.Update(calend3)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	//Get
	calendarGet, found, err := calendarR.Get(id3)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if !found {
		t.Error("Calendar doesn't exists after the creation")
		t.FailNow()
	}
	if id3 != calendarGet.ID {
		t.Error("invalid calendar ID")
	}
	if name == calend3.Name || calendarGet.UnionCalendarIDs[0] != id2 {
		t.Error("the update on the calendar didn't work")
	}
}

func TestPostgresDelete(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}

	db := tests.DBClient(t)
	postgres.ReplaceGlobals(db)
	defer dbCalendarDestroy(db, t)
	dbCalendarInit(db, t)

	calendarR := NewPostgresRepository(db)
	ReplaceGlobals(calendarR)

	//Create
	name := "my_calendar_test"
	period := Period{}
	calend := Calendar{
		Name:        name,
		Description: "this is my calendar",
		Periods:     []Period{period},
	}

	id, err := calendarR.Create(calend)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	//Delete
	err = calendarR.Delete(id)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	//Get
	_, found, err := calendarR.Get(id)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if found {
		t.Error("Calendar exists after deletion")
		t.FailNow()
	}
}

func TestPostgresCreateAndGetAll(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}

	db := tests.DBClient(t)
	postgres.ReplaceGlobals(db)
	defer dbCalendarDestroy(db, t)
	dbCalendarInit(db, t)

	calendarR := NewPostgresRepository(db)
	ReplaceGlobals(calendarR)

	_, found, err := calendarR.Get(1)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if found {
		t.Error("found a calendar from nowhere")
	}

	period := Period{}
	calend1 := Calendar{
		Name:        "c1",
		Description: "calendar 1",
		Periods:     []Period{period},
	}

	id1, err := calendarR.Create(calend1)
	if err != nil {
		t.Error(err)
	}

	calend2 := Calendar{
		Name:             "c2",
		Description:      "calendar 2",
		Periods:          []Period{period},
		UnionCalendarIDs: []int64{id1},
	}

	_, err = calendarR.Create(calend2)
	if err != nil {
		t.Error(err)
	}

	calendarsGetAll, err := calendarR.GetAll()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if len(calendarsGetAll) != 2 {
		t.Error(err)
		t.FailNow()
	}
}

func TestPostgresGetAllModified(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}

	db := tests.DBClient(t)
	postgres.ReplaceGlobals(db)
	defer dbCalendarDestroy(db, t)
	dbCalendarInit(db, t)

	calendarR := NewPostgresRepository(db)
	ReplaceGlobals(calendarR)

	period := Period{}
	calend1 := Calendar{
		Name:        "c1",
		Description: "calendar 1",
		Periods:     []Period{period},
	}

	id1, err := calendarR.Create(calend1)
	if err != nil {
		t.Error(err)
	}

	time.Sleep(10 * time.Millisecond)
	timestamp := time.Now()

	calend2 := Calendar{
		Name:             "c2",
		Description:      "calendar 2",
		Periods:          []Period{period},
		UnionCalendarIDs: []int64{id1},
	}

	id2, err := calendarR.Create(calend2)
	if err != nil {
		t.Error(err)
	}

	calendarsGetAll, err := calendarR.GetAllModifiedFrom(timestamp)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if len(calendarsGetAll) != 1 {
		t.Error("The number of calendars obtained is not as expected")
	}
	if id2 != calendarsGetAll[id2].ID {
		t.Error("The calendar obtained is different to the inserted calendar")
	}
}
