package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/security/permissions"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/security/users"
	situation2 "github.com/myrteametrics/myrtea-engine-api/v5/pkg/situation"
	"net/http"
	"strconv"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/rule"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/tests"
	"github.com/myrteametrics/myrtea-sdk/v5/postgres"
)

func dbInit(dbClient *sqlx.DB, t *testing.T) {
	dbDestroy(dbClient, t)
	tests.DBExec(dbClient, tests.CalendarTableV3, t, true)
	tests.DBExec(dbClient, tests.RulesTableV1, t, true)
	tests.DBExec(dbClient, tests.RuleVersionsTableV1, t, true)
	tests.DBExec(dbClient, tests.SituationDefinitionTableV1, t, true)
	tests.DBExec(dbClient, tests.SituationRulesTableV1, t, true)
	tests.DBExec(dbClient, tests.FactDefinitionTableV1, t, true)
	tests.DBExec(dbClient, tests.SituationFactsTableV1, t, true)
}

func dbDestroy(dbClient *sqlx.DB, t *testing.T) {
	tests.DBExec(dbClient, tests.SituationFactsDropTableV1, t, true)
	tests.DBExec(dbClient, tests.SituationRulesDropTableV1, t, true)
	tests.DBExec(dbClient, tests.SituationDefinitionDropTableV1, t, true)
	tests.DBExec(dbClient, tests.FactDefinitionDropTableV1, t, true)
	tests.DBExec(dbClient, tests.RuleVersionsDropTableV1, t, false)
	tests.DBExec(dbClient, tests.RulesDropTableV1, t, false)
	tests.DBExec(dbClient, tests.CalendarDropTableV3, t, true)
}

func TestGetRules(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroy(db, t)
	dbInit(db, t)
	rule.ReplaceGlobals(rule.NewPostgresRepository(db))

	rule1 := rule.Rule{}
	json.Unmarshal(dataRule1, &rule1)
	id1, _ := rule.R().Create(rule1)

	rule2 := rule.Rule{}
	json.Unmarshal(dataRule2, &rule2)
	id2, _ := rule.R().Create(rule2)

	user := users.UserWithPermissions{Permissions: []permissions.Permission{permissions.New(permissions.TypeRule, permissions.All, permissions.ActionList), permissions.New(permissions.TypeRule, permissions.All, permissions.ActionGet)}}
	rr := tests.BuildTestHandler(t, "GET", "/rules", ``, "/rules", GetRules, user)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	rule1.ID = id1
	rule2.ID = id2

	mapRules := []rule.Rule{rule1, rule2}

	rulesData, _ := json.Marshal(mapRules)
	expected := string(rulesData) + "\n"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}

func TestPostRule(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}

	db := tests.DBClient(t)
	defer dbDestroy(db, t)
	dbInit(db, t)
	rule.ReplaceGlobals(rule.NewPostgresRepository(db))

	user := users.UserWithPermissions{Permissions: []permissions.Permission{permissions.New(permissions.TypeRule, permissions.All, permissions.ActionCreate)}}
	rr := tests.BuildTestHandler(t, "POST", "/rules", string(dataRule1), "/rules", PostRule, user)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	_, found, err := rule.R().Get(1)
	if !found {
		t.Error("rule not found")
	}
	if err != nil {
		t.Error(err)
	}

}

func TestUpdateRules(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroy(db, t)
	dbInit(db, t)
	rule.ReplaceGlobals(rule.NewPostgresRepository(db))

	oldRule1 := rule.Rule{}
	json.Unmarshal(dataRule1, &oldRule1)
	id1, _ := rule.R().Create(oldRule1)

	newRule1 := rule.Rule{}
	json.Unmarshal(dataRule2, &newRule1)
	user := users.UserWithPermissions{Permissions: []permissions.Permission{permissions.New(permissions.TypeRule, permissions.All, permissions.ActionUpdate)}}
	rr := tests.BuildTestHandler(t, "PUT", "/rules/"+strconv.FormatInt(id1, 10), string(dataRule2), "/rules/{id}", PutRule, user)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	getRule1, _, _ := rule.R().Get(id1)
	if getRule1.Name != newRule1.Name || getRule1.Version != 1 || getRule1.SameCasesAs(oldRule1) {
		t.Errorf("The rule was not properly updated")
		t.Log(getRule1)
		t.Log(newRule1)
	}
}

func TestPostRulesSituations(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroy(db, t)
	dbInit(db, t)
	postgres.ReplaceGlobals(db)
	rule.ReplaceGlobals(rule.NewPostgresRepository(db))
	situation2.ReplaceGlobals(situation2.NewPostgresRepository(db))

	///create rules
	rule1 := rule.Rule{}
	json.Unmarshal(dataRule1, &rule1)
	r1ID, _ := rule.R().Create(rule1)

	rule2 := rule.Rule{}
	json.Unmarshal(dataRule2, &rule2)
	r2ID, _ := rule.R().Create(rule2)

	//create situations
	s1ID, _ := situation2.R().Create(situation2.Situation{Name: "Situation1"})
	s2ID, _ := situation2.R().Create(situation2.Situation{Name: "Situation2"})
	s3ID, _ := situation2.R().Create(situation2.Situation{Name: "Situation3"})

	//Post new situations to rules
	situationIDs := []int64{s1ID, s2ID}
	data, _ := json.Marshal(situationIDs)

	user := users.UserWithPermissions{Permissions: []permissions.Permission{permissions.New(permissions.TypeRule, permissions.All, permissions.ActionUpdate)}}
	rr := tests.BuildTestHandler(t, "POST", "/rules/"+fmt.Sprint(r1ID)+"/situations", string(data), "/rules/{id}/situations", PostRuleSituations, user)
	tests.CheckTestHandler(t, rr, http.StatusOK, ``)

	getSituatations, _ := situation2.R().GetAllByRuleID(int64(r1ID), false)
	if _, ok := getSituatations[s1ID]; !ok {
		t.Errorf("The rule %d was not added to the rule list of the situation %d", r1ID, s1ID)
	}
	if _, ok := getSituatations[s2ID]; !ok {
		t.Errorf("The rule %d was not added to the rule list of the situation %d", r1ID, s2ID)
	}

	rr = tests.BuildTestHandler(t, "POST", "/rules/"+fmt.Sprint(r2ID)+"/situations", string(data), "/rules/{id}/situations", PostRuleSituations, user)
	tests.CheckTestHandler(t, rr, http.StatusOK, ``)

	getSituatations, _ = situation2.R().GetAllByRuleID(int64(r2ID), false)
	if _, ok := getSituatations[s1ID]; !ok {
		t.Errorf("The rule %d was not added to the rule list of the situation %d", r1ID, s1ID)
	}
	if _, ok := getSituatations[s2ID]; !ok {
		t.Errorf("The rule %d was not added to the rule list of the situation %d", r1ID, s2ID)
	}

	//validate rules order
	rulesS1, _ := situation2.R().GetRules(s1ID)
	rulesS2, _ := situation2.R().GetRules(s1ID)

	if rulesS1[0] != int64(r1ID) || rulesS1[1] != int64(r2ID) || rulesS2[0] != int64(r1ID) || rulesS2[1] != int64(r2ID) {
		t.Errorf("The execution order of the rules is not as expected")
	}

	//Post new + existing situations to rules
	situationIDs = []int64{s1ID, s2ID, s3ID}
	data, _ = json.Marshal(situationIDs)

	rr = tests.BuildTestHandler(t, "POST", "/rules/"+fmt.Sprint(r1ID)+"/situations", string(data), "/rules/{id}/situations", PostRuleSituations, user)
	tests.CheckTestHandler(t, rr, http.StatusOK, ``)

	getSituatations, _ = situation2.R().GetAllByRuleID(int64(r1ID), false)
	if _, ok := getSituatations[s1ID]; !ok {
		t.Errorf("The rule %d was not added to the rule list of the situation %d", r1ID, s1ID)
	}
	if _, ok := getSituatations[s2ID]; !ok {
		t.Errorf("The rule %d was not added to the rule list of the situation %d", r1ID, s2ID)
	}
	if _, ok := getSituatations[s3ID]; !ok {
		t.Errorf("The rule %d was not added to the rule list of the situation %d", r1ID, s3ID)
	}

	//Remove situations from rules
	situationIDs = []int64{s1ID, s3ID}
	data, _ = json.Marshal(situationIDs)

	rr = tests.BuildTestHandler(t, "POST", "/rules/"+fmt.Sprint(r1ID)+"/situations", string(data), "/rules/{id}/situations", PostRuleSituations, user)
	tests.CheckTestHandler(t, rr, http.StatusOK, ``)

	getSituatations, _ = situation2.R().GetAllByRuleID(int64(r1ID), false)
	if _, ok := getSituatations[s1ID]; !ok {
		t.Errorf("The rule %d was not added to the rule list of the situation %d", r1ID, s1ID)
	}
	if _, ok := getSituatations[s2ID]; ok {
		t.Errorf("The rule %d was not removed from the rule list of the situation %d", r1ID, s2ID)
	}
	if _, ok := getSituatations[s3ID]; !ok {
		t.Errorf("The rule %d was not added to the rule list of the situation %d", r1ID, s3ID)
	}
}

var dataRule1 = []byte(`{
	"name": "rule1",
	"description": "this is the rule 1",
	"cases": [
	  {
		"name": "case1",
		"condition": "A > B",
		"actions": [
		  {
			"name": "\"set\"",
			"parameters": {
			  "key": "\"value\""
			}
		  }
		]
	  }
	],
	"enabled": true
}`)

var dataRule2 = []byte(`{
	"name": "rule2",
	"description": "this is the rule 2",
	"cases": [
	  {
		"name": "case1",
		"condition": "B < C",
		"actions": [
		  {
			"name": "\"Action\"",
			"parameters": {
			  "key": "\"value\""
			}
		  }
		]
	  }
	],
	"enabled": true
}`)
