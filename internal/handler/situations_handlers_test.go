package handler

import (
	"encoding/json"
	"fmt"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/security/permissions"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/security/users"
	situation2 "github.com/myrteametrics/myrtea-engine-api/v5/pkg/situation"
	"net/http"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/fact"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/tests"
	"github.com/myrteametrics/myrtea-sdk/v5/engine"
)

func situationDbInit(dbClient *sqlx.DB, t *testing.T) {
	situationDbDestroy(dbClient, t)
	tests.DBExec(dbClient, tests.CalendarTableV3, t, true)
	tests.DBExec(dbClient, tests.SituationDefinitionTableV1, t, true)
	tests.DBExec(dbClient, tests.SituationTemplateInstancesTableV1, t, true)
	tests.DBExec(dbClient, tests.FactDefinitionTableV1, t, true)
	tests.DBExec(dbClient, tests.SituationFactsTableV1, t, true)
}

func situationDbDestroy(dbClient *sqlx.DB, t *testing.T) {
	tests.DBExec(dbClient, tests.SituationFactsDropTableV1, t, true)
	tests.DBExec(dbClient, tests.FactDefinitionDropTableV1, t, true)
	tests.DBExec(dbClient, tests.SituationTemplateInstancesDropTableV1, t, true)
	tests.DBExec(dbClient, tests.SituationDefinitionDropTableV1, t, true)
	tests.DBExec(dbClient, tests.CalendarDropTableV3, t, false)
}

func TestPostSituation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}

	db := tests.DBClient(t)
	defer situationDbDestroy(db, t)
	situationDbInit(db, t)

	situation2.ReplaceGlobals(situation2.NewPostgresRepository(db))
	fact.ReplaceGlobals(fact.NewPostgresRepository(db))

	factID, err := fact.R().Create(engine.Fact{})

	s := situation2.Situation{
		Name:  "test_situation",
		Facts: []int64{factID},
	}
	b, _ := json.Marshal(s)

	user := users.UserWithPermissions{Permissions: []permissions.Permission{permissions.New(permissions.TypeSituation, "*", permissions.ActionCreate)}}
	rr := tests.BuildTestHandler(t, "POST", "/situations", string(b), "/situations", PostSituation, user)
	tests.CheckTestHandler(t, rr, http.StatusOK, `{"id":1,"name":"test_situation","facts":[1],"expressionFacts":null,"calendarId":0,"parameters":null,"isTemplate":false,"isObject":false}`+"\n")

	situations, err := situation2.R().GetAll()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if _, exists := situations[1]; !exists {
		t.Error("Situation 1 should not be nil")
	}
}

func TestPutSituationTemplateInstances(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer situationDbDestroy(db, t)
	situationDbInit(db, t)

	situation2.ReplaceGlobals(situation2.NewPostgresRepository(db))

	//create situations
	s1ID, _ := situation2.R().Create(situation2.Situation{Name: "Situation1", IsTemplate: true})

	//Situation template instances
	instance1 := situation2.TemplateInstance{Name: "Instance 1"}
	instance2 := situation2.TemplateInstance{Name: "Instance 2"}
	instance3 := situation2.TemplateInstance{Name: "Instance 3"}

	user := users.UserWithPermissions{Permissions: []permissions.Permission{permissions.New(permissions.TypeSituation, "*", permissions.ActionUpdate)}}

	//Put situation template instances
	data, _ := json.Marshal([]situation2.TemplateInstance{instance1, instance2})
	rr := tests.BuildTestHandler(t, "PUT", "/situations/"+fmt.Sprint(s1ID)+"/instances", string(data), "/situations/{id}/instances", PutSituationTemplateInstances, user)
	tests.CheckTestHandler(t, rr, http.StatusOK, ``)

	getInstances1, _ := situation2.R().GetAllTemplateInstances(s1ID)
	if _, ok := getInstances1[1]; !ok {
		t.Errorf("The template instance %s was not added to the situation template instance", instance1.Name)
	}
	if _, ok := getInstances1[2]; !ok {
		t.Errorf("The template instance %s was not added to the situation template instance", instance2.Name)
	}

	//Post situation template instance
	data, _ = json.Marshal(instance3)
	instance3.ID = 3
	instance3.SituationID = 1
	expectedData, _ := json.Marshal(instance3)
	rr = tests.BuildTestHandler(t, "POST", "/situations/"+fmt.Sprint(s1ID)+"/instances", string(data), "/situations/{id}/instances", PostSituationTemplateInstance, user)
	tests.CheckTestHandler(t, rr, http.StatusOK, string(expectedData)+"\n")

	getInstances2, _ := situation2.R().GetAllTemplateInstances(s1ID)
	if _, ok := getInstances2[3]; !ok {
		t.Errorf("The template instance %s was not added to the situation template instance", instance3.Name)
	}

	if _, ok := getInstances2[1]; !ok {
		t.Errorf("The template instance %s was removed from the situation template instance", instance1.Name)
	}
	if _, ok := getInstances2[2]; !ok {
		t.Errorf("The template instance %s was removed from the situation template instance", instance2.Name)
	}

	//Put situation template instances (remove instancetemplate)
	instance1.ID = 1
	instance1.SituationID = 1
	data, _ = json.Marshal([]situation2.TemplateInstance{instance1, instance3})
	rr = tests.BuildTestHandler(t, "PUT", "/situations/"+fmt.Sprint(s1ID)+"/instances", string(data), "/situations/{id}/instances", PutSituationTemplateInstances, user)
	tests.CheckTestHandler(t, rr, http.StatusOK, ``)

	getInstances3, _ := situation2.R().GetAllTemplateInstances(s1ID)
	if _, ok := getInstances3[1]; !ok {
		t.Errorf("The template instance %s was removed from the situation template instance", instance1.Name)
	}
	if _, ok := getInstances3[2]; ok {
		t.Errorf("The template instance %s was not removed from the situation template instance", instance2.Name)
	}
	if _, ok := getInstances3[3]; !ok {
		t.Errorf("The template instance %s was removed from the situation template instance", instance3.Name)
	}

	//Put situation template instances (remove and add instancetemplate)
	instance4 := situation2.TemplateInstance{Name: "Instance 4"}
	data, _ = json.Marshal([]situation2.TemplateInstance{instance4, instance3})
	rr = tests.BuildTestHandler(t, "PUT", "/situations/"+fmt.Sprint(s1ID)+"/instances", string(data), "/situations/{id}/instances", PutSituationTemplateInstances, user)
	tests.CheckTestHandler(t, rr, http.StatusOK, ``)

	getInstances4, _ := situation2.R().GetAllTemplateInstances(s1ID)
	if _, ok := getInstances4[1]; ok {
		t.Errorf("The template instance %s was not removed from the situation template instance", instance1.Name)
	}
	if _, ok := getInstances4[2]; ok {
		t.Errorf("The template instance %s was not removed from the situation template instance", instance2.Name)
	}
	if _, ok := getInstances4[3]; !ok {
		t.Errorf("The template instance %s was not removed from the situation template instance", instance3.Name)
	}
	if _, ok := getInstances4[4]; !ok {
		t.Errorf("The template instance %s was not added to the situation template instance", instance4.Name)
	}
}
