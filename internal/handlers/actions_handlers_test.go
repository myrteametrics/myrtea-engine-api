package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/explainer/action"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/explainer/rootcause"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/models"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/rule"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/situation"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/tests"
)

func actionsDBInit(dbClient *sqlx.DB, t *testing.T) {
	actionsDBDestroy(dbClient, t)
	tests.DBExec(dbClient, tests.CalendarTableV3, t, true)
	tests.DBExec(dbClient, tests.SituationDefinitionTableV1, t, true)
	tests.DBExec(dbClient, tests.FactDefinitionTableV1, t, true)
	tests.DBExec(dbClient, tests.SituationFactsTableV1, t, true)
	tests.DBExec(dbClient, tests.RulesTableV1, t, true)
	tests.DBExec(dbClient, tests.RuleVersionsTableV1, t, true)
	tests.DBExec(dbClient, tests.RefRootCauseTableV1, t, true)
	tests.DBExec(dbClient, tests.RefActionTableV1, t, true)

	situationR := situation.NewPostgresRepository(dbClient)
	situation.ReplaceGlobals(situationR)
	s := situation.Situation{
		Name:  "situation_test_1",
		Facts: []int64{},
	}
	s1ID, err := situationR.Create(s)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	rr := rule.NewPostgresRepository(dbClient)
	r := rule.Rule{Name: "rule_test_1"}
	ruleID1, err := rr.Create(r)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	rootcausesR := rootcause.NewPostgresRepository(dbClient)
	rc1 := models.NewRootCause(0, "rc_1", "rc_desc_1", s1ID, int64(ruleID1))
	_, err = rootcausesR.Create(nil, rc1)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	actionsR := action.NewPostgresRepository(dbClient)
	action.ReplaceGlobals(actionsR)

}

func actionsDBDestroy(dbClient *sqlx.DB, t *testing.T) {
	tests.DBExec(dbClient, tests.RefActionDropTableV1, t, false)
	tests.DBExec(dbClient, tests.RefRootCauseDropTableV1, t, false)
	tests.DBExec(dbClient, tests.RuleVersionsDropTableV1, t, true)
	tests.DBExec(dbClient, tests.RulesDropTableV1, t, true)
	tests.DBExec(dbClient, tests.SituationFactsDropTableV1, t, false)
	tests.DBExec(dbClient, tests.FactDefinitionDropTableV1, t, false)
	tests.DBExec(dbClient, tests.SituationDefinitionDropTableV1, t, false)
	tests.DBExec(dbClient, tests.CalendarDropTableV3, t, true)
}

func TestGetActions(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer actionsDBDestroy(db, t)
	actionsDBInit(db, t)

	action1 := models.Action{
		Name:        "action1",
		Description: "action_desc_1",
		RootCauseID: 1,
	}
	action2 := models.Action{
		Name:        "action2",
		Description: "action_desc_2",
		RootCauseID: 1,
	}
	var err error
	action1.ID, err = action.R().Create(nil, action1)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	action2.ID, err = action.R().Create(nil, action2)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	req, err := http.NewRequest("GET", "/actions", nil)
	if err != nil {
		t.Fatal(err)
		t.FailNow()
	}
	rr := httptest.NewRecorder()
	r := chi.NewRouter()
	r.Get("/actions", GetActions)
	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		t.FailNow()
	}

	mapActions := []models.Action{action1, action2}
	usersData, err := json.Marshal(mapActions)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	expected := string(usersData) + "\n"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}

func TestGetAction(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer actionsDBDestroy(db, t)
	actionsDBInit(db, t)

	action1 := models.Action{
		Name:        "action",
		Description: "action_desc_1",
		RootCauseID: 1,
	}
	action1.ID, _ = action.R().Create(nil, action1)

	req, err := http.NewRequest("GET", "/actions/"+strconv.FormatInt(action1.ID, 10), nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	r := chi.NewRouter()
	r.Get("/actions/{id}", GetAction)
	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	userData, _ := json.Marshal(action1)
	expected := string(userData) + "\n"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}

func TestPostAction(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer actionsDBDestroy(db, t)
	actionsDBInit(db, t)

	action1 := models.Action{
		Name:        "action",
		Description: "action_desc_1",
		RootCauseID: 1,
	}
	actionData, _ := json.Marshal(action1)

	req, err := http.NewRequest("POST", "/actions", bytes.NewBuffer(actionData))
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	r := chi.NewRouter()
	r.Post("/actions", PostAction)
	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	actionGet, found, err := action.R().Get(1)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if !found {
		t.Errorf("Action with id 1 not found")
		t.FailNow()
	}
	if actionGet.Name != action1.Name {
		t.Errorf("The user action was not inserted correctly")
	}
}

func TestPutAction(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer actionsDBDestroy(db, t)
	actionsDBInit(db, t)

	action1 := models.Action{
		Name:        "action",
		Description: "action_desc_1",
		RootCauseID: 1,
	}
	action1.ID, _ = action.R().Create(nil, action1)
	action1.Name = "newName"

	actionData, _ := json.Marshal(action1)
	req, err := http.NewRequest("PUT", "/actions/"+strconv.FormatInt(action1.ID, 10), bytes.NewBuffer(actionData))
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	r := chi.NewRouter()
	r.Put("/actions/{id}", PutAction)
	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	actionGet, found, err := action.R().Get(1)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if !found {
		t.Errorf("Action with id 1 not found")
		t.FailNow()
	}
	if actionGet.Name != action1.Name {
		t.Errorf("The user action was not updated correctly")
	}
}

func TestDeleteAction(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer actionsDBDestroy(db, t)
	actionsDBInit(db, t)

	action1 := models.Action{
		Name:        "action",
		Description: "action_desc_1",
		RootCauseID: 1,
	}
	action1.ID, _ = action.R().Create(nil, action1)

	req, err := http.NewRequest("DELETE", "/actions/"+strconv.FormatInt(action1.ID, 10), nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	r := chi.NewRouter()
	r.Delete("/actions/{id}", DeleteAction)
	r.ServeHTTP(rr, req)

	_, found, err := action.R().Get(1)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if found {
		t.Errorf("Action with id 1 found while it should not")
		t.FailNow()
	}
}
