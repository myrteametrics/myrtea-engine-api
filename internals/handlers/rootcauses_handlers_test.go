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
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/explainer/rootcause"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/models"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/rule"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/situation"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/tests"
)

func rootCausesDBInit(dbClient *sqlx.DB, t *testing.T) {
	rootCausesDBDestroy(dbClient, t)
	tests.DBExec(dbClient, tests.CalendarTableV3, t, true)
	tests.DBExec(dbClient, tests.SituationDefinitionTableV1, t, true)
	tests.DBExec(dbClient, tests.FactDefinitionTableV1, t, true)
	tests.DBExec(dbClient, tests.SituationFactsTableV1, t, true)
	tests.DBExec(dbClient, tests.RulesTableV1, t, true)
	tests.DBExec(dbClient, tests.RuleVersionsTableV1, t, true)
	tests.DBExec(dbClient, tests.RefRootCauseTableV1, t, true)

	situationR := situation.NewPostgresRepository(dbClient)
	situation.ReplaceGlobals(situationR)
	s := situation.Situation{
		Name:  "situation_test_1",
		Facts: []int64{},
	}
	_, err := situationR.Create(s)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	rr := rule.NewPostgresRepository(dbClient)
	r := rule.Rule{Name: "rule_test_1"}
	_, err = rr.Create(r)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	rootcausesR := rootcause.NewPostgresRepository(dbClient)
	rootcause.ReplaceGlobals(rootcausesR)

}

func rootCausesDBDestroy(dbClient *sqlx.DB, t *testing.T) {
	tests.DBExec(dbClient, tests.RefRootCauseDropTableV1, t, false)
	tests.DBExec(dbClient, tests.RuleVersionsDropTableV1, t, true)
	tests.DBExec(dbClient, tests.RulesDropTableV1, t, true)
	tests.DBExec(dbClient, tests.SituationFactsDropTableV1, t, false)
	tests.DBExec(dbClient, tests.FactDefinitionDropTableV1, t, false)
	tests.DBExec(dbClient, tests.SituationDefinitionDropTableV1, t, false)
	tests.DBExec(dbClient, tests.CalendarDropTableV3, t, false)
}

func TestGetRootCauses(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer rootCausesDBDestroy(db, t)
	rootCausesDBInit(db, t)

	var err error
	rootcause1 := models.NewRootCause(0, "rootcause1", "rootcause_desc_1", 1, 1)
	rootcause1.ID, err = rootcause.R().Create(nil, rootcause1)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	rootcause2 := models.NewRootCause(0, "rootcause2", "rootcause_desc_2", 1, 1)
	rootcause2.ID, err = rootcause.R().Create(nil, rootcause2)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	req, err := http.NewRequest("GET", "/rootcauses", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	r := chi.NewRouter()
	r.Get("/rootcauses", GetRootCauses)
	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	mapRootCauses := []models.RootCause{rootcause1, rootcause2}
	usersData, _ := json.Marshal(mapRootCauses)

	expected := string(usersData) + "\n"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}

func TestGetRootCause(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer rootCausesDBDestroy(db, t)
	rootCausesDBInit(db, t)

	rootcause1 := models.NewRootCause(0, "rootcause1", "rootcause_desc_1", 1, 1)
	rootcause1.ID, _ = rootcause.R().Create(nil, rootcause1)

	req, err := http.NewRequest("GET", "/rootcauses/"+strconv.FormatInt(rootcause1.ID, 10), nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	r := chi.NewRouter()
	r.Get("/rootcauses/{id}", GetRootCause)
	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	userData, _ := json.Marshal(rootcause1)
	expected := string(userData) + "\n"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}

func TestPostRootCause(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer rootCausesDBDestroy(db, t)
	rootCausesDBInit(db, t)

	rootcause1 := models.NewRootCause(0, "rootcause1", "rootcause_desc_1", 1, 1)
	rootcauseData, _ := json.Marshal(rootcause1)

	req, err := http.NewRequest("POST", "/rootcauses", bytes.NewBuffer(rootcauseData))
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	r := chi.NewRouter()
	r.Post("/rootcauses", PostRootCause)
	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	rootcauseGet, found, err := rootcause.R().Get(1)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if !found {
		t.Errorf("RootCause with id 1 not found")
		t.FailNow()
	}
	if rootcauseGet.Name != rootcause1.Name {
		t.Errorf("The user rootcause was not inserted correctly")
	}
}

func TestPutRootCause(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer rootCausesDBDestroy(db, t)
	rootCausesDBInit(db, t)

	rootcause1 := models.NewRootCause(0, "rootcause1", "rootcause_desc_1", 1, 1)
	rootcause1.ID, _ = rootcause.R().Create(nil, rootcause1)
	rootcause1.Name = "newName"

	rootcauseData, _ := json.Marshal(rootcause1)
	req, err := http.NewRequest("PUT", "/rootcauses/"+strconv.FormatInt(rootcause1.ID, 10), bytes.NewBuffer(rootcauseData))
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	r := chi.NewRouter()
	r.Put("/rootcauses/{id}", PutRootCause)
	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	rootcauseGet, found, err := rootcause.R().Get(1)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if !found {
		t.Errorf("RootCause with id 1 not found")
		t.FailNow()
	}
	if rootcauseGet.Name != rootcause1.Name {
		t.Errorf("The user rootcause was not updated correctly")
	}
}

func TestDeleteRootCause(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer rootCausesDBDestroy(db, t)
	rootCausesDBInit(db, t)

	rootcause1 := models.NewRootCause(0, "rootcause1", "rootcause_desc_1", 1, 1)
	rootcause1.ID, _ = rootcause.R().Create(nil, rootcause1)

	req, err := http.NewRequest("DELETE", "/rootcauses/"+strconv.FormatInt(rootcause1.ID, 10), nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	r := chi.NewRouter()
	r.Delete("/rootcauses/{id}", DeleteRootCause)
	r.ServeHTTP(rr, req)

	_, found, err := rootcause.R().Get(1)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if found {
		t.Errorf("RootCause with id 1 found while it should not")
	}
}
