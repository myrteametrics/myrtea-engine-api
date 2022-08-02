package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/explainer"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/explainer/action"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/explainer/draft"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/explainer/issues"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/explainer/rootcause"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/models"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/rule"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/security/permissions"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/security/users"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/situation"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/tests"
	"github.com/myrteametrics/myrtea-sdk/v4/postgres"
	"github.com/myrteametrics/myrtea-sdk/v4/ruleeng"
)

func dbIssueInitRepo(dbClient *sqlx.DB, t *testing.T) {
	dbIssueDestroyRepo(dbClient, t)
	tests.DBExec(dbClient, `CREATE EXTENSION IF NOT EXISTS pgcrypto;`, t, true)
	tests.DBExec(dbClient, tests.UsersTableV1, t, true)
	tests.DBExec(dbClient, tests.CalendarTableV3, t, true)
	tests.DBExec(dbClient, tests.SituationDefinitionTableV1, t, true)
	tests.DBExec(dbClient, tests.FactDefinitionTableV1, t, true)
	tests.DBExec(dbClient, tests.SituationFactsTableV1, t, true)
	tests.DBExec(dbClient, tests.IssuesTableV1, t, true)
	tests.DBExec(dbClient, tests.RulesTableV1, t, true)
	tests.DBExec(dbClient, tests.RuleVersionsTableV1, t, true)
	tests.DBExec(dbClient, tests.SituationRulesTableV1, t, true)
	tests.DBExec(dbClient, tests.RefRootCauseTableV1, t, true)
	tests.DBExec(dbClient, tests.RefActionTableV1, t, true)
	tests.DBExec(dbClient, tests.IssueResolutionDraftTableV1, t, true)
	tests.DBExec(dbClient, tests.IssueResolutionTableV1, t, true)
}

func dbIssueDestroyRepo(dbClient *sqlx.DB, t *testing.T) {
	tests.DBExec(dbClient, tests.IssueResolutionDraftDropTableV1, t, true)
	tests.DBExec(dbClient, tests.IssueResolutionDropTableV1, t, true)
	tests.DBExec(dbClient, tests.RefActionDropTableV1, t, true)
	tests.DBExec(dbClient, tests.RefRootCauseDropTableV1, t, true)
	tests.DBExec(dbClient, tests.SituationRulesDropTableV1, t, true)
	tests.DBExec(dbClient, tests.RuleVersionsDropTableV1, t, true)
	tests.DBExec(dbClient, tests.RulesDropTableV1, t, true)
	tests.DBExec(dbClient, tests.IssuesDropTableV1, t, true)
	tests.DBExec(dbClient, tests.SituationFactsDropTableV1, t, true)
	tests.DBExec(dbClient, tests.SituationDefinitionDropTableV1, t, true)
	tests.DBExec(dbClient, tests.CalendarDropTableV3, t, true)
	tests.DBExec(dbClient, tests.FactDefinitionDropTableV1, t, true)
	tests.DBExec(dbClient, tests.UsersDropTableV1, t, true)
}

func TestGetIssues(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := tests.DBClient(t)
	defer dbIssueDestroyRepo(db, t)
	dbIssueInitRepo(db, t)

	issues.ReplaceGlobals(issues.NewPostgresRepository(db))
	situation.ReplaceGlobals(situation.NewPostgresRepository(db))

	//Situation
	situation1 := situation.Situation{
		Name: "test_name",
	}

	situationID, err := situation.R().Create(situation1)
	if err != nil {
		t.Error(err)
	}

	timestamp := time.Now().Truncate(1 * time.Millisecond).UTC()

	cas := models.CaseInput{
		Name:      "case1",
		Condition: "3 > 2",
		State:     models.Met,
	}

	//Rule
	rule := models.RuleData{
		RuleID:      1,
		RuleVersion: 0,
		CaseName:    cas.Name,
	}

	issue := models.Issue{
		Key:          "key1",
		Level:        models.Critical,
		SituationID:  situationID,
		ExpirationTS: timestamp.Add(24 * time.Hour).UTC(),
		SituationTS:  timestamp,
		Rule:         rule,
		State:        models.Open,
	}

	id, err := issues.R().Create(issue)
	if err != nil {
		t.Error(err)
	}

	// groupsOfUser := make([]groups.GroupOfUser, 0)
	// groupsOfUser = append(groupsOfUser, groups.GroupOfUser{
	// 	ID:       1,
	// 	Name:     "user",
	// 	UserRole: 1,
	// })

	// user := groups.UserWithGroups{
	// 	User:   security.User{},
	// 	Groups: groupsOfUser,
	// }

	req, err := http.NewRequest("GET", "/issues", nil)
	if err != nil {
		t.Fatal(err)
	}

	user := users.UserWithPermissions{Permissions: []permissions.Permission{permissions.New(permissions.TypeSituationIssues, "*", permissions.ActionList), permissions.New(permissions.TypeSituationIssues, "*", permissions.ActionGet)}}
	ctx := context.WithValue(req.Context(), models.ContextKeyUser, user)

	rr := httptest.NewRecorder()
	r := chi.NewRouter()

	r.Get("/issues", GetIssues)
	r.ServeHTTP(rr, req.WithContext(ctx))

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var issues map[int64]*models.Issue
	err = json.Unmarshal(rr.Body.Bytes(), &issues)
	if err != nil {
		t.Errorf("handler returned unexpected body")
	}

	if issues[id] != nil && issues[id].ID != id {
		t.Errorf("handler returned unexpected body: got %v want %v", issues[id].ID, id)
	}
}

func TestGetIssuesStates(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := tests.DBClient(t)
	defer dbIssueDestroyRepo(db, t)
	dbIssueInitRepo(db, t)

	issues.ReplaceGlobals(issues.NewPostgresRepository(db))
	situation.ReplaceGlobals(situation.NewPostgresRepository(db))

	//Situation
	situation1 := situation.Situation{
		Name: "test_name",
	}

	situationID, err := situation.R().Create(situation1)
	if err != nil {
		t.Error(err)
	}

	timestamp := time.Now().Truncate(1 * time.Millisecond).UTC()

	cas := models.CaseInput{
		Name:      "case1",
		Condition: "3 > 2",
		State:     models.Met,
	}

	//Rule
	rule := models.RuleData{
		RuleID:      1,
		RuleVersion: 0,
		CaseName:    cas.Name,
	}

	issue := models.Issue{
		SituationID: situationID,
		Key:         "key1",
		Level:       models.Critical,
		SituationTS: timestamp,
		Rule:        rule,
		State:       models.ClosedFeedback,
	}

	issue2 := models.Issue{
		SituationID: situationID,
		Key:         "key2",
		Level:       models.Critical,
		SituationTS: timestamp,
		Rule:        rule,
		State:       models.Open,
	}

	issue3 := models.Issue{
		SituationID: situationID,
		Key:         "key3",
		Level:       models.Critical,
		SituationTS: timestamp,
		Rule:        rule,
		State:       models.ClosedNoFeedback,
	}

	_, err = issues.R().Create(issue)
	if err != nil {
		t.Error(err)
	}

	_, err = issues.R().Create(issue2)
	if err != nil {
		t.Error(err)
	}

	_, err = issues.R().Create(issue3)
	if err != nil {
		t.Error(err)
	}

	// groupsOfUser := make([]groups.GroupOfUser, 0)
	// groupsOfUser = append(groupsOfUser, groups.GroupOfUser{
	// 	ID:       1,
	// 	Name:     "user",
	// 	UserRole: 1,
	// })

	// user := groups.UserWithGroups{
	// 	User:   security.User{},
	// 	Groups: groupsOfUser,
	// }

	req, err := http.NewRequest("GET", "/issues?states=open,closedfeedback", nil)
	if err != nil {
		t.Fatal(err)
	}

	user := users.UserWithPermissions{Permissions: []permissions.Permission{permissions.New(permissions.TypeSituationIssues, permissions.All, permissions.ActionList), permissions.New(permissions.TypeSituationIssues, permissions.All, permissions.ActionGet)}}
	ctx := context.WithValue(req.Context(), models.ContextKeyUser, user)

	rr := httptest.NewRecorder()
	r := chi.NewRouter()

	r.Get("/issues", GetIssuesByStatesByPage)
	r.ServeHTTP(rr, req.WithContext(ctx))

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var issues models.PaginatedResource
	err = json.Unmarshal(rr.Body.Bytes(), &issues)
	if err != nil {
		t.Errorf("handler returned unexpected body")
		t.Log(rr.Body.String())
	}

	if items := issues.Items.([]interface{}); len(items) != 2 {
		t.Errorf("handler returned unexpected body: more or less issues than expected")
	}
}

func TestGetIssue(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := tests.DBClient(t)
	defer dbIssueDestroyRepo(db, t)
	dbIssueInitRepo(db, t)

	issues.ReplaceGlobals(issues.NewPostgresRepository(db))
	situation.ReplaceGlobals(situation.NewPostgresRepository(db))

	//Situation
	situation1 := situation.Situation{Name: "test_name"}

	situationID, err := situation.R().Create(situation1)
	if err != nil {
		t.Error(err)
	}

	timestamp := time.Now().Truncate(1 * time.Millisecond).UTC()

	cas := models.CaseInput{
		Name:      "",
		Condition: "",
		State:     models.Met,
	}

	//Rule
	rule := models.RuleData{
		RuleID:      1,
		RuleVersion: 0,
		CaseName:    cas.Name,
	}

	issue := models.Issue{
		Key:         "key1",
		SituationID: situationID,
		SituationTS: timestamp,
		Level:       models.Critical,
		Rule:        rule,
		State:       models.Open,
	}

	id, err := issues.R().Create(issue)
	if err != nil {
		t.Error(err)
	}

	// groupsOfUser := make([]groups.GroupOfUser, 0)
	// groupsOfUser = append(groupsOfUser, groups.GroupOfUser{
	// 	ID:       1,
	// 	Name:     "user",
	// 	UserRole: 1,
	// })

	// user := groups.UserWithGroups{
	// 	User:   security.User{},
	// 	Groups: groupsOfUser,
	// }

	req, err := http.NewRequest("GET", "/issues/"+strconv.FormatInt(id, 10), nil)
	if err != nil {
		t.Fatal(err)
	}

	user := users.UserWithPermissions{Permissions: []permissions.Permission{permissions.New(permissions.TypeSituationIssues, "1", permissions.ActionGet)}}
	ctx := context.WithValue(req.Context(), models.ContextKeyUser, user)

	rr := httptest.NewRecorder()
	r := chi.NewRouter()
	r.Get("/issues/{id}", GetIssue)
	r.ServeHTTP(rr, req.WithContext(ctx))

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var issueGet *models.Issue
	err = json.Unmarshal(rr.Body.Bytes(), &issueGet)
	if err != nil {
		t.Errorf("handler returned unexpected body")
	}

	if id != issueGet.ID {
		t.Errorf("handler returned unexpected body: more or less issues than expected")
	}
}

func TestPostIssue(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := tests.DBClient(t)
	defer dbIssueDestroyRepo(db, t)
	dbIssueInitRepo(db, t)

	issues.ReplaceGlobals(issues.NewPostgresRepository(db))
	situation.ReplaceGlobals(situation.NewPostgresRepository(db))

	//Situation
	situation1 := situation.Situation{
		Name: "test_name",
	}
	situationID, err := situation.R().Create(situation1)
	if err != nil {
		t.Error(err)
	}

	timestamp := time.Now().Truncate(1 * time.Millisecond).UTC()

	cas := models.CaseInput{
		Name:      "case1",
		Condition: "3 > 2",
		State:     models.Met,
	}

	//Rule
	rule := models.RuleData{
		RuleID:      1,
		RuleVersion: 0,
		CaseName:    cas.Name,
	}

	issue := models.Issue{
		SituationID: situationID,
		SituationTS: timestamp,
		Rule:        rule,
		State:       models.Open,
	}

	issueData, _ := json.Marshal(issue)

	req, err := http.NewRequest("POST", "/issues", bytes.NewBuffer([]byte(issueData)))
	if err != nil {
		t.Fatal(err)
	}

	user := users.UserWithPermissions{Permissions: []permissions.Permission{permissions.New(permissions.TypeSituationIssues, permissions.All, permissions.ActionCreate)}}
	ctx := context.WithValue(req.Context(), models.ContextKeyUser, user)

	rr := httptest.NewRecorder()
	r := chi.NewRouter()
	r.Post("/issues", PostIssue)
	r.ServeHTTP(rr, req.WithContext(ctx))

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	if rr.Body.String() != "" {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), "")
	}

	id := int64(1)

	issues, err := issues.R().GetAll()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if _, exists := issues[id]; !exists {
		t.Error("Issue 1 should not be nil")
	}
}

func TestIssueLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := tests.DBClient(t)
	defer dbIssueDestroyRepo(db, t)
	dbIssueInitRepo(db, t)

	postgres.ReplaceGlobals(db)
	issues.ReplaceGlobals(issues.NewPostgresRepository(db))
	situation.ReplaceGlobals(situation.NewPostgresRepository(db))
	rule.ReplaceGlobals(rule.NewPostgresRepository(db))
	draft.ReplaceGlobals(draft.NewPostgresRepository(db))
	rootcause.ReplaceGlobals(rootcause.NewPostgresRepository(db))
	action.ReplaceGlobals(action.NewPostgresRepository(db))

	//Situation
	situation1 := situation.Situation{Name: "test_name"}

	situationID, err := situation.R().Create(situation1)
	if err != nil {
		t.Error(err)
	}

	timestamp := time.Now().Truncate(1 * time.Millisecond).UTC()

	rule1 := rule.Rule{
		Name: "rule_1",
		DefaultRule: ruleeng.DefaultRule{Cases: []ruleeng.Case{
			{
				Name:      "case_1",
				Condition: ruleeng.Expression("true"),
				Actions: []ruleeng.ActionDef{{
					Name: ruleeng.Expression(`"create-set"`),
					Parameters: map[string]ruleeng.Expression{
						"key": ruleeng.Expression(`"1"`),
					},
				}},
			},
		}},
		Enabled: true,
	}

	ruleID, _ := rule.R().Create(rule1)
	err = situation.R().SetRules(situationID, []int64{int64(ruleID)})
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	cas := models.CaseInput{
		Name:      "",
		Condition: "",
		State:     models.Met,
	}

	//Rule
	ruleInput := models.RuleData{
		RuleID:      int64(ruleID),
		RuleVersion: 0,
		CaseName:    cas.Name,
	}

	issue1 := models.Issue{
		Key:         "key1",
		SituationID: situationID,
		SituationTS: timestamp,
		Level:       models.Critical,
		Rule:        ruleInput,
		State:       models.Open,
	}

	issue2 := models.Issue{
		Key:         "key2",
		SituationID: situationID,
		SituationTS: timestamp,
		Level:       models.Critical,
		Rule:        ruleInput,
		State:       models.Open,
	}

	issue1ID, err := issues.R().Create(issue1)
	if err != nil {
		t.Error(err)
	}

	issue2ID, err := issues.R().Create(issue2)
	if err != nil {
		t.Error(err)
	}

	// user := groups.UserWithGroups{
	// 	User: security.User{
	// 		FirstName: "user_first_name",
	// 		LastName:  "user_last_name",
	// 	},
	// 	Groups: []groups.GroupOfUser{{
	// 		ID:       1,
	// 		Name:     "group1",
	// 		UserRole: 1,
	// 	}},
	// }

	draft := models.FrontDraft{
		Tree: []*models.FrontRootCause{
			{
				Name:        "rootcause_1",
				Description: "this is the root cause 1",
				Custom:      true,
				Actions: []*models.FrontAction{
					{
						Name:        "action_1",
						Description: "this is the action 1",
						Custom:      true,
					},
				},
			},
		},
	}
	b, _ := json.Marshal(draft)

	user := users.UserWithPermissions{Permissions: []permissions.Permission{permissions.New(permissions.TypeSituationIssues, permissions.All, permissions.ActionGet)}}

	rr := tests.BuildTestHandler(t, "POST", fmt.Sprintf("/issues/%d/draft", issue1ID), string(b), "/issues/{id}/draft", PostIssueDraft, user)
	if rr.Code != http.StatusOK {
		t.Error("Unexpected error posting issue resolution draft")
	}

	getIssue, _, _ := issues.R().Get(issue1ID)
	if *getIssue.AssignedTo != user.Login || getIssue.AssignedAt.IsZero() {
		t.Error("Unexpected recomendation tree")
	}
	recomendation, _ := explainer.GetRecommendationTree(getIssue)
	recomendation.Tree[0].Selected = true
	recomendation.Tree[0].Actions[0].Selected = true
	b, _ = json.Marshal(recomendation)

	rr = tests.BuildTestHandler(t, "POST", fmt.Sprintf("/issues/%d/feedback", issue1ID), string(b), "/issues/{id}/feedback", PostIssueCloseWithFeedback, user)
	if rr.Code != http.StatusOK {
		t.Error("Unexpected error closing issue with feedback")
	}
	getIssue, _, _ = issues.R().Get(issue1ID)
	if getIssue.CloseBy == nil || *getIssue.CloseBy != user.Login ||
		getIssue.ClosedAt == nil || getIssue.ClosedAt.IsZero() || getIssue.State != models.ClosedFeedback {
		t.Error("Unexpected obtained issue")
	}

	rr = tests.BuildTestHandler(t, "POST", fmt.Sprintf("/issues/%d/close", issue2ID), "", "/issues/{id}/close", PostIssueCloseWithoutFeedback, user)
	if rr.Code != http.StatusOK {
		t.Error("Unexpected error closing issue without feedback")
	}
	getIssue, _, _ = issues.R().Get(issue2ID)
	if getIssue.CloseBy == nil || *getIssue.CloseBy != user.Login ||
		getIssue.ClosedAt == nil || getIssue.ClosedAt.IsZero() || getIssue.State != models.ClosedNoFeedback {
		t.Error("Unexpected obtained issue")
	}
}
