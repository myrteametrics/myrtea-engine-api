package draft

import (
	"fmt"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/explainer/action"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/explainer/issues"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/explainer/rootcause"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/models"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/rule"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/situation"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/tests"
)

// repoCreateIssue create a new issue in the specified repository
func repoCreateIssue(repo issues.Repository, issue models.Issue, t *testing.T, failNow bool) int64 {
	id, err := repo.Create(issue)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	return id
}

// repoCreateRC create a new in rootcause the specified repository
func repoCreateRC(repo rootcause.Repository, rc models.RootCause, t *testing.T, failNow bool) int64 {
	id, err := repo.Create(nil, rc)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	return id
}

// repoCreateAction create a new in action the specified repository
func repoCreateAction(repo action.Repository, action models.Action, t *testing.T, failNow bool) int64 {
	id, err := repo.Create(nil, action)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	return id
}

func repoCreateDraft(r Repository, t *testing.T) models.FrontRecommendation {
	newDraft := models.FrontRecommendation{
		Tree: []*models.FrontRootCause{
			{ID: 1, Name: "rc_1", Description: "rc_desc_1", Selected: false, Custom: false, Actions: []*models.FrontAction{
				{ID: 1, Name: "action_1", Description: "action_desc_1", Selected: false, Custom: false},
				{ID: 2, Name: "action_2", Description: "action_desc_2", Selected: false, Custom: false},
			}},
			{ID: 2, Name: "rc_2", Description: "rc_desc_2", Selected: false, Custom: false, Actions: []*models.FrontAction{
				{ID: 3, Name: "action_3", Description: "action_desc_3", Selected: true, Custom: false},
				{ID: 4, Name: "action_4", Description: "action_desc_4", Selected: true, Custom: false},
			}},
		},
	}
	err := r.Create(nil, 1, newDraft)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	return newDraft
}

func dbInit(dbClient *sqlx.DB, t *testing.T) {
	dbDestroy(dbClient, t)

	tests.DBExec(dbClient, tests.CalendarTableV3, t, true)
	tests.DBExec(dbClient, tests.SituationDefinitionTableV1, t, true)
	tests.DBExec(dbClient, tests.FactDefinitionTableV1, t, true)
	tests.DBExec(dbClient, tests.SituationFactsTableV1, t, true)
	tests.DBExec(dbClient, tests.RulesTableV1, t, true)
	tests.DBExec(dbClient, tests.RuleVersionsTableV1, t, true)
	tests.DBExec(dbClient, tests.IssuesTableV1, t, true)
	tests.DBExec(dbClient, tests.RefRootCauseTableV1, t, true)
	tests.DBExec(dbClient, tests.RefActionTableV1, t, true)
	tests.DBExec(dbClient, tests.IssueResolutionTableV1, t, true)
	tests.DBExec(dbClient, tests.IssueResolutionDraftTableV1, t, true)

	sr := situation.NewPostgresRepository(dbClient)
	situation.ReplaceGlobals(sr)
	s := situation.Situation{
		Name:  "situation_1",
		Facts: []int64{},
	}
	situationID1, err := sr.Create(s)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	rule.ReplaceGlobals(rule.NewPostgresRepository(dbClient))
	r := rule.Rule{
		Name: "rule_test_1",
	}
	ruleID1, err := rule.R().Create(r)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	issues.ReplaceGlobals(issues.NewPostgresRepository(dbClient))
	issue := models.Issue{
		SituationID: situationID1,
		SituationTS: time.Now().Truncate(1 * time.Millisecond).UTC(),
		Rule:        models.RuleData{RuleID: 1},
		State:       models.Open,
	}
	_ = repoCreateIssue(issues.R(), issue, t, true)
	_ = repoCreateIssue(issues.R(), issue, t, true)
	_ = repoCreateIssue(issues.R(), issue, t, true)

	rcr := rootcause.NewPostgresRepository(dbClient)
	rootcause.ReplaceGlobals(rcr)
	rootCauseID1 := repoCreateRC(rcr, models.NewRootCause(0, "rc_1", "rc_desc_1", situationID1, int64(ruleID1)), t, true)
	rootCauseID2 := repoCreateRC(rcr, models.NewRootCause(0, "rc_2", "rc_desc_2", situationID1, int64(ruleID1)), t, true)

	ar := action.NewPostgresRepository(dbClient)
	action.ReplaceGlobals(ar)
	_ = repoCreateAction(ar, models.NewAction(0, "action_1", "action_desc_1", rootCauseID1), t, true)
	_ = repoCreateAction(ar, models.NewAction(0, "action_2", "action_desc_2", rootCauseID1), t, true)
	_ = repoCreateAction(ar, models.NewAction(0, "action_3", "action_desc_3", rootCauseID2), t, true)
	_ = repoCreateAction(ar, models.NewAction(0, "action_4", "action_desc_4", rootCauseID2), t, true)

}

func dbDestroy(dbClient *sqlx.DB, t *testing.T) {
	tests.DBExec(dbClient, tests.IssueResolutionDraftDropTableV1, t, false)
	tests.DBExec(dbClient, tests.IssueResolutionDropTableV1, t, false)
	tests.DBExec(dbClient, tests.RefActionDropTableV1, t, false)
	tests.DBExec(dbClient, tests.RefRootCauseDropTableV1, t, false)
	tests.DBExec(dbClient, tests.IssuesDropTableV1, t, false)
	tests.DBExec(dbClient, tests.RuleVersionsDropTableV1, t, false)
	tests.DBExec(dbClient, tests.RulesDropTableV1, t, false)
	tests.DBExec(dbClient, tests.SituationFactsDropTableV1, t, false)
	tests.DBExec(dbClient, tests.FactDefinitionDropTableV1, t, false)
	tests.DBExec(dbClient, tests.SituationDefinitionDropTableV1, t, false)
	tests.DBExec(dbClient, tests.CalendarDropTableV3, t, false)
}

func TestNewPostgresRepository(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	r := NewPostgresRepository(tests.DBClient(t))
	if r == nil {
		t.Error("action repository is nil")
	}
}

func TestPostgresReplaceGlobal(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	r := NewPostgresRepository(tests.DBClient(t))
	reverse := ReplaceGlobals(r)
	if R() == nil {
		t.Error("Global action repository is nil")
	}
	reverse()
	if R() != nil {
		t.Error("Global action repository is not nil after reverse")
	}
}

func TestGetNotExists(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroy(db, t)
	dbInit(db, t)

	r := NewPostgresRepository(db)

	_, found, err := r.Get(1)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if found {
		t.Error("No draft should be found with id 1")
	}
}

func TestCreateGet(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroy(db, t)
	dbInit(db, t)

	r := NewPostgresRepository(db)

	newDraft := repoCreateDraft(r, t)

	draft, found, err := r.Get(1)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if !found {
		t.Error("Draft not properly created")
		t.FailNow()
	}
	if draft.ConcurrencyUUID == "" {
		t.Error("No concurrencyUUID set during creation")
	}
	if len(draft.Tree) != len(newDraft.Tree) {
		t.Error("Invalid rootcause len")
	}
	if draft.Tree[0].ID != newDraft.Tree[0].ID ||
		draft.Tree[0].Name != newDraft.Tree[0].Name ||
		draft.Tree[0].Description != newDraft.Tree[0].Description ||
		draft.Tree[0].Selected != newDraft.Tree[0].Selected ||
		draft.Tree[0].Custom != newDraft.Tree[0].Custom {
		t.Error("Invalid rootcause 1")
	}
	if len(draft.Tree[0].Actions) != len(newDraft.Tree[0].Actions) ||
		len(draft.Tree[1].Actions) != len(newDraft.Tree[1].Actions) {
		t.Error("Invalid rootcauses actions len")
	}
	if draft.Tree[0].Actions[0].ID != newDraft.Tree[0].Actions[0].ID ||
		draft.Tree[0].Actions[0].Name != newDraft.Tree[0].Actions[0].Name ||
		draft.Tree[0].Actions[0].Description != newDraft.Tree[0].Actions[0].Description ||
		draft.Tree[0].Actions[0].Selected != newDraft.Tree[0].Actions[0].Selected ||
		draft.Tree[0].Actions[0].Custom != newDraft.Tree[0].Actions[0].Custom {
		t.Error("Invalid rootcause 1 action 1")
	}
}

func TestUpdate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroy(db, t)
	dbInit(db, t)

	r := NewPostgresRepository(db)

	repoCreateDraft(r, t)
	newDraft, found, err := r.Get(1)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if !found {
		t.Error("Draft not properly created")
		t.FailNow()
	}

	newDraft.Tree[0].Selected = true
	newDraft.Tree[0].Actions[0].Selected = true
	newDraft.Tree[1].Selected = false
	newDraft.Tree[1].Actions[0].Selected = false
	newDraft.Tree[1].Actions[1].Selected = false
	err = r.Update(nil, 1, newDraft)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	draft, found, err := r.Get(1)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if !found {
		t.Error("Draft not found with id 1")
		t.FailNow()
	}
	if draft.ConcurrencyUUID == newDraft.ConcurrencyUUID {
		t.Error("Concurrency UUID not updated after update")
	}
	if len(draft.Tree) != len(newDraft.Tree) {
		t.Error("Invalid rootcause len")
	}
	if draft.Tree[0].ID != newDraft.Tree[0].ID ||
		draft.Tree[0].Name != newDraft.Tree[0].Name ||
		draft.Tree[0].Description != newDraft.Tree[0].Description ||
		draft.Tree[0].Selected != newDraft.Tree[0].Selected ||
		draft.Tree[0].Custom != newDraft.Tree[0].Custom {
		t.Error("Invalid rootcause 1")
	}
	if len(draft.Tree[0].Actions) != len(newDraft.Tree[0].Actions) ||
		len(draft.Tree[1].Actions) != len(newDraft.Tree[1].Actions) {
		t.Error("Invalid rootcauses actions len")
	}
	if draft.Tree[0].Actions[0].ID != newDraft.Tree[0].Actions[0].ID ||
		draft.Tree[0].Actions[0].Name != newDraft.Tree[0].Actions[0].Name ||
		draft.Tree[0].Actions[0].Description != newDraft.Tree[0].Actions[0].Description ||
		draft.Tree[0].Actions[0].Selected != newDraft.Tree[0].Actions[0].Selected ||
		draft.Tree[0].Actions[0].Custom != newDraft.Tree[0].Actions[0].Custom {
		t.Error("Invalid rootcause 1 action 1")
	}
}

func TestUpdateWithoutConcurrencyUUID(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroy(db, t)
	dbInit(db, t)

	r := NewPostgresRepository(db)

	repoCreateDraft(r, t)
	newDraft, found, err := r.Get(1)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if !found {
		t.Error("Draft not found with id 1")
		t.FailNow()
	}
	concurrencyUUIDV1 := newDraft.ConcurrencyUUID

	newDraft.ConcurrencyUUID = ""
	newDraft.Tree[0].Selected = true
	err = r.Update(nil, 1, newDraft)
	if err == nil {
		t.Error("Update should not be possible without concurrencyUUID")
	}

	newDraftV2, found, err := r.Get(1)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if !found {
		t.Error("Draft not found with id 1")
		t.FailNow()
	}
	if newDraftV2.ConcurrencyUUID != concurrencyUUIDV1 {
		t.Error("Draft concurrencyUUID has been updated while it should not")
	}
	if newDraftV2.Tree[0].Selected == true {
		t.Error("Draft tree has been updated while it should not")
	}
}

func TestCheckExists(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroy(db, t)
	dbInit(db, t)

	r := NewPostgresRepository(db)
	repoCreateDraft(r, t)

	exists, err := r.CheckExists(nil, 1)
	if err != nil {
		t.Error(err)
	}
	if !exists {
		t.Error("Draft exists check returns false while it should not")
	}

}

func TestCheckExistsNotExists(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroy(db, t)
	dbInit(db, t)

	r := NewPostgresRepository(db)
	repoCreateDraft(r, t)

	exists, err := r.CheckExists(nil, 2)
	if err != nil {
		t.Error(err)
	}
	if exists {
		t.Error("Draft exists check returns true while it should not")
	}
}

func TestCheckExistsWithUUID(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroy(db, t)
	dbInit(db, t)

	r := NewPostgresRepository(db)
	repoCreateDraft(r, t)

	newDraft, found, err := r.Get(1)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if !found {
		t.Error("Draft not found with id 1")
		t.FailNow()
	}
	concurrencyUUID := newDraft.ConcurrencyUUID

	exists, err := r.CheckExistsWithUUID(nil, 1, concurrencyUUID)
	if err != nil {
		t.Error(err)
	}
	if !exists {
		t.Error("Draft exists check returns true while it should not")
	}
}

func TestCheckExistsWithUUIDNotExists(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroy(db, t)
	dbInit(db, t)

	r := NewPostgresRepository(db)
	repoCreateDraft(r, t)

	newDraft, found, err := r.Get(1)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if !found {
		t.Error("Draft not found with id 1")
		t.FailNow()
	}
	concurrencyUUID := newDraft.ConcurrencyUUID

	exists, err := r.CheckExistsWithUUID(nil, 2, concurrencyUUID)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if exists {
		t.Error("Draft exists check returns true while it should not")
	}
}

func TestCheckExistsWithUUIDInvalidUUID(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroy(db, t)
	dbInit(db, t)

	r := NewPostgresRepository(db)
	repoCreateDraft(r, t)

	newDraft, found, err := r.Get(1)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if !found {
		t.Error("Draft not found with id 1")
		t.FailNow()
	}
	concurrencyUUID := fmt.Sprintf("%s-not-a-uuid", newDraft.ConcurrencyUUID)

	exists, err := r.CheckExistsWithUUID(nil, 1, concurrencyUUID)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if exists {
		t.Error("Draft exists check returns true while it should not")
	}
}

func TestCheckExistsWithUUIDConcurrentUpdate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroy(db, t)
	dbInit(db, t)

	r := NewPostgresRepository(db)
	repoCreateDraft(r, t)

	newDraft, found, err := r.Get(1)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if !found {
		t.Error("Draft not found with id 1")
		t.FailNow()
	}
	concurrencyUUID := newDraft.ConcurrencyUUID

	newDraft.Tree[0].Selected = true
	err = r.Update(nil, 1, newDraft)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	exists, err := r.CheckExistsWithUUID(nil, 1, concurrencyUUID)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if exists {
		t.Error("Draft exists check returns true while it should not")
	}
}
