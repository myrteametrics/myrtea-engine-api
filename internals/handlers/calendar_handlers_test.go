package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/calendar"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/groups"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/tests"
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

func TestGetCalendars(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	db := tests.DBClient(t)

	calendarR := calendar.NewPostgresRepository(db)
	calendar.ReplaceGlobals(calendarR)
	defer dbCalendarDestroy(db, t)
	dbCalendarInit(db, t)

	//Create
	name := "my_calendar_test"
	period := calendar.Period{}
	calend := calendar.Calendar{
		Name:        name,
		Description: "this is my calendar",
		Periods:     []calendar.Period{period},
		Enabled:     true,
	}

	id, err := calendarR.Create(calend)
	if err != nil {
		t.Error(err)
	}

	rr := tests.BuildTestHandler(t, "GET", "/calendars", "", "/calendars", GetCalendars, groups.UserWithGroups{})

	var calendars []calendar.Calendar
	err = json.Unmarshal(rr.Body.Bytes(), &calendars)
	if err != nil {
		t.Errorf("handler returned unexpected body")
	}

	if calendars[0].ID != id {
		t.Errorf("handler returned unexpected body: got %v want %v", string(calendars[id].ID), string(id))
	}
}

func TestGetCalendar(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	db := tests.DBClient(t)

	calendarR := calendar.NewPostgresRepository(db)
	calendar.ReplaceGlobals(calendarR)
	defer dbCalendarDestroy(db, t)
	dbCalendarInit(db, t)

	//Create
	name := "my_calendar_test"
	period := calendar.Period{}
	calend := calendar.Calendar{
		Name:        name,
		Description: "this is my calendar",
		Periods:     []calendar.Period{period},
	}

	id, err := calendarR.Create(calend)
	if err != nil {
		t.Error(err)
	}

	rr := tests.BuildTestHandler(t, "GET", "/calendars/"+strconv.FormatInt(id, 10), "", "/calendars/{id}", GetCalendar, groups.UserWithGroups{})

	var calendarGet *calendar.Calendar
	err = json.Unmarshal(rr.Body.Bytes(), &calendarGet)
	if err != nil {
		t.Errorf("handler returned unexpected body")
	}

	if id != calendarGet.ID {
		t.Errorf("handler returned unexpected body: more or less calendars than expected")
	}
}

func TestPostCalendar(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	db := tests.DBClient(t)

	calendarR := calendar.NewPostgresRepository(db)
	calendar.ReplaceGlobals(calendarR)
	defer dbCalendarDestroy(db, t)
	dbCalendarInit(db, t)

	name := "my_calendar_test"
	period := calendar.Period{}
	calend := calendar.Calendar{
		Name:        name,
		Description: "this is my calendar",
		Periods:     []calendar.Period{period},
	}

	calendarData, _ := json.Marshal(calend)

	rr := tests.BuildTestHandler(t, "POST", "/calendars", string(calendarData), "/calendars", PostCalendar, groups.UserWithGroups{})

	var calendarPost *calendar.Calendar
	err := json.Unmarshal(rr.Body.Bytes(), &calendarPost)
	if err != nil {
		t.Errorf("handler returned unexpected body")
	}

	if calend.Name != calendarPost.Name {
		t.Errorf("handler returned unexpected body: more or less calendars than expected")
	}
}

func TestPutCalendar(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	db := tests.DBClient(t)

	calendarR := calendar.NewPostgresRepository(db)
	calendar.ReplaceGlobals(calendarR)
	defer dbCalendarDestroy(db, t)
	dbCalendarInit(db, t)

	//Create
	name := "my_calendar_test"
	period := calendar.Period{}
	calend := calendar.Calendar{
		Name:        name,
		Description: "this is my calendar",
		Periods:     []calendar.Period{period},
	}

	id, err := calendarR.Create(calend)
	if err != nil {
		t.Error(err)
	}

	calend.Name = "anothername"
	calendarData, _ := json.Marshal(calend)

	rr := tests.BuildTestHandler(t, "PUT", "/calendars/"+strconv.FormatInt(id, 10), string(calendarData), "/calendars/{id}", PutCalendar, groups.UserWithGroups{})

	var calendarPut *calendar.Calendar
	err = json.Unmarshal(rr.Body.Bytes(), &calendarPut)
	if err != nil {
		t.Errorf("handler returned unexpected body")
	}

	if name == calendarPut.Name {
		t.Errorf("handler returned unexpected body")
	}
}

func TestDeleteCalendar(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	db := tests.DBClient(t)

	calendarR := calendar.NewPostgresRepository(db)
	calendar.ReplaceGlobals(calendarR)
	defer dbCalendarDestroy(db, t)
	dbCalendarInit(db, t)

	//Create
	name := "my_calendar_test"
	period := calendar.Period{}
	calend := calendar.Calendar{
		Name:        name,
		Description: "this is my calendar",
		Periods:     []calendar.Period{period},
	}

	id, err := calendarR.Create(calend)
	if err != nil {
		t.Error(err)
	}

	rr := tests.BuildTestHandler(t, "DELETE", "/calendars/"+strconv.FormatInt(id, 10), "", "/calendars/{id}", DeleteCalendar, groups.UserWithGroups{})

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	//Get
	_, found, err := calendarR.Get(id)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if found {
		t.Errorf("Calendar exists after deletion")
		t.FailNow()
	}
}
