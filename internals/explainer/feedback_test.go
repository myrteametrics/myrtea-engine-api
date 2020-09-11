package explainer

import (
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/explainer/action"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/explainer/draft"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/explainer/issues"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/explainer/rootcause"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/groups"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/models"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/rule"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/situation"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/tests"
)

func createIssue(repo issues.Repository, issue models.Issue, t *testing.T, failNow bool) int64 {
	id, err := repo.Create(issue)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	return id
}

func createRC(repo rootcause.Repository, rc models.RootCause, t *testing.T, failNow bool) int64 {
	id, err := repo.Create(nil, rc)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	return id
}
func createAction(repo action.Repository, action models.Action, t *testing.T, failNow bool) int64 {
	id, err := repo.Create(nil, action)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	return id
}
func createIssueResolution(dbClient *sqlx.DB, issueID int64, rootCauseID int64, actionID int64, t *testing.T, failNow bool) {
	tx, err := dbClient.Beginx()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	err = persistIssueResolutionStat(tx, issueID, rootCauseID, actionID)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	err = tx.Commit()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
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

	situation.ReplaceGlobals(situation.NewPostgresRepository(dbClient))
	s := situation.Situation{
		Groups: []int64{1, 2},
		Name:   "situation_1",
		Facts:  []int64{},
	}
	situationID1, err := situation.R().Create(s)
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
	issueID1 := createIssue(issues.R(), issue, t, true)
	issueID2 := createIssue(issues.R(), issue, t, true)
	_ = createIssue(issues.R(), issue, t, true)

	rcr := rootcause.NewPostgresRepository(dbClient)
	rootcause.ReplaceGlobals(rcr)
	rootCauseID1 := createRC(rcr, models.NewRootCause(0, "rc_1", "rc_desc_1", situationID1, int64(ruleID1)), t, true)
	rootCauseID2 := createRC(rcr, models.NewRootCause(0, "rc_2", "rc_desc_2", situationID1, int64(ruleID1)), t, true)

	ar := action.NewPostgresRepository(dbClient)
	action.ReplaceGlobals(ar)
	_ = createAction(ar, models.NewAction(0, "action_1", "action_desc_1", rootCauseID1), t, true)
	actionID2 := createAction(ar, models.NewAction(0, "action_2", "action_desc_2", rootCauseID1), t, true)
	actionID3 := createAction(ar, models.NewAction(0, "action_3", "action_desc_3", rootCauseID2), t, true)
	actionID4 := createAction(ar, models.NewAction(0, "action_4", "action_desc_4", rootCauseID2), t, true)

	createIssueResolution(dbClient, issueID1, rootCauseID2, actionID2, t, true)
	createIssueResolution(dbClient, issueID1, rootCauseID2, actionID3, t, true)
	createIssueResolution(dbClient, issueID2, rootCauseID2, actionID3, t, true)
	createIssueResolution(dbClient, issueID2, rootCauseID2, actionID4, t, true)

	dr := draft.NewPostgresRepository(dbClient)
	draft.ReplaceGlobals(dr)
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

func TestFeedbackWithoutSelectedRootCause(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroy(db, t)
	dbInit(db, t)

	recommendation := models.FrontRecommendation{
		Tree: []*models.FrontRootCause{
			{ID: 2, Name: "rc_2", Description: "rc_desc_2", Selected: false, Custom: false, Occurrence: 0, UsageRate: 0, ClusteringScore: -1,
				Actions: []*models.FrontAction{
					{ID: 3, Name: "action_3", Description: "action_desc_3", Selected: false, Custom: false, Occurrence: 0, UsageRate: 0},
					{ID: 4, Name: "action_4", Description: "action_desc_4", Selected: false, Custom: false, Occurrence: 0, UsageRate: 0},
				}},
			{ID: 1, Name: "rc_1", Description: "rc_desc_1", Selected: false, Custom: false, Occurrence: 2, UsageRate: 1, ClusteringScore: -1,
				Actions: []*models.FrontAction{
					{ID: 1, Name: "action_1", Description: "action_desc_1", Selected: false, Custom: false, Occurrence: 2, UsageRate: 1},
					{ID: 2, Name: "action_2", Description: "action_desc_2", Selected: false, Custom: false, Occurrence: 1, UsageRate: 0.5},
				}},
		},
	}

	err := CloseIssueWithFeedback(db, 3, recommendation, []int64{1, 2}, groups.UserWithGroups{})
	if err == nil {
		t.Error("Should return an error 'A feedback must have one rootcause selected'")
	}
}

func TestFeedbackWithMoreThanOneRootCause(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroy(db, t)
	dbInit(db, t)

	recommendation := models.FrontRecommendation{
		Tree: []*models.FrontRootCause{
			{ID: 2, Name: "rc_2", Description: "rc_desc_2", Selected: true, Custom: false, Occurrence: 0, UsageRate: 0, ClusteringScore: -1,
				Actions: []*models.FrontAction{
					{ID: 3, Name: "action_3", Description: "action_desc_3", Selected: false, Custom: false, Occurrence: 0, UsageRate: 0},
					{ID: 4, Name: "action_4", Description: "action_desc_4", Selected: false, Custom: false, Occurrence: 0, UsageRate: 0},
				}},
			{ID: 1, Name: "rc_1", Description: "rc_desc_1", Selected: true, Custom: false, Occurrence: 2, UsageRate: 1, ClusteringScore: -1,
				Actions: []*models.FrontAction{
					{ID: 1, Name: "action_1", Description: "action_desc_1", Selected: false, Custom: false, Occurrence: 2, UsageRate: 1},
					{ID: 2, Name: "action_2", Description: "action_desc_2", Selected: false, Custom: false, Occurrence: 1, UsageRate: 0.5},
				}},
		},
	}

	err := CloseIssueWithFeedback(db, 3, recommendation, []int64{1, 2}, groups.UserWithGroups{})
	if err == nil {
		t.Error("Should return an error 'A feedback can't have multiple selected rootcause'")
	}
}

func TestFeedbackWithoutSelectedAction(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroy(db, t)
	dbInit(db, t)

	recommendation := models.FrontRecommendation{
		Tree: []*models.FrontRootCause{
			{ID: 2, Name: "rc_2", Description: "rc_desc_2", Selected: true, Custom: false, Occurrence: 0, UsageRate: 0, ClusteringScore: -1,
				Actions: []*models.FrontAction{
					{ID: 3, Name: "action_3", Description: "action_desc_3", Selected: false, Custom: false, Occurrence: 0, UsageRate: 0},
					{ID: 4, Name: "action_4", Description: "action_desc_4", Selected: false, Custom: false, Occurrence: 0, UsageRate: 0},
				}},
			{ID: 1, Name: "rc_1", Description: "rc_desc_1", Selected: false, Custom: false, Occurrence: 2, UsageRate: 1, ClusteringScore: -1,
				Actions: []*models.FrontAction{
					{ID: 1, Name: "action_1", Description: "action_desc_1", Selected: false, Custom: false, Occurrence: 2, UsageRate: 1},
					{ID: 2, Name: "action_2", Description: "action_desc_2", Selected: false, Custom: false, Occurrence: 1, UsageRate: 0.5},
				}},
		},
	}

	err := CloseIssueWithFeedback(db, 3, recommendation, []int64{1, 2}, groups.UserWithGroups{})
	if err == nil {
		t.Error("Should return an error 'A feedback must have at least one action selected'")
	}
}

func TestFeedbackOneExistingAction(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroy(db, t)
	dbInit(db, t)

	recommendation := models.FrontRecommendation{
		Tree: []*models.FrontRootCause{
			{ID: 2, Name: "rc_2", Description: "rc_desc_2", Selected: true, Custom: false, Occurrence: 0, UsageRate: 0, ClusteringScore: -1,
				Actions: []*models.FrontAction{
					{ID: 3, Name: "action_3", Description: "action_desc_3", Selected: true, Custom: false, Occurrence: 0, UsageRate: 0},
					{ID: 4, Name: "action_4", Description: "action_desc_4", Selected: false, Custom: false, Occurrence: 0, UsageRate: 0},
				}},
			{ID: 1, Name: "rc_1", Description: "rc_desc_1", Selected: false, Custom: false, Occurrence: 2, UsageRate: 1, ClusteringScore: -1,
				Actions: []*models.FrontAction{
					{ID: 1, Name: "action_1", Description: "action_desc_1", Selected: false, Custom: false, Occurrence: 2, UsageRate: 1},
					{ID: 2, Name: "action_2", Description: "action_desc_2", Selected: false, Custom: false, Occurrence: 1, UsageRate: 0.5},
				}},
		},
	}

	err := CloseIssueWithFeedback(db, 3, recommendation, []int64{1, 2}, groups.UserWithGroups{})
	if err != nil {
		t.Error(err)
	}
}

func TestFeedbackMultipleExistingActions(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroy(db, t)
	dbInit(db, t)

	recommendation := models.FrontRecommendation{
		Tree: []*models.FrontRootCause{
			{ID: 2, Name: "rc_2", Description: "rc_desc_2", Selected: true, Custom: false, Occurrence: 0, UsageRate: 0, ClusteringScore: -1,
				Actions: []*models.FrontAction{
					{ID: 3, Name: "action_3", Description: "action_desc_3", Selected: true, Custom: false, Occurrence: 0, UsageRate: 0},
					{ID: 4, Name: "action_4", Description: "action_desc_4", Selected: true, Custom: false, Occurrence: 0, UsageRate: 0},
				}},
			{ID: 1, Name: "rc_1", Description: "rc_desc_1", Selected: false, Custom: false, Occurrence: 2, UsageRate: 1, ClusteringScore: -1,
				Actions: []*models.FrontAction{
					{ID: 1, Name: "action_1", Description: "action_desc_1", Selected: false, Custom: false, Occurrence: 2, UsageRate: 1},
					{ID: 2, Name: "action_2", Description: "action_desc_2", Selected: false, Custom: false, Occurrence: 1, UsageRate: 0.5},
				}},
		},
	}

	err := CloseIssueWithFeedback(db, 3, recommendation, []int64{1, 2}, groups.UserWithGroups{})
	if err != nil {
		t.Error(err)
	}
}

func TestFeedbackCustomRootCauseAndOneAction(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroy(db, t)
	dbInit(db, t)

	recommendation := models.FrontRecommendation{
		Tree: []*models.FrontRootCause{
			{ID: 2, Name: "rc_2", Description: "rc_desc_2", Selected: false, Custom: false, Occurrence: 0, UsageRate: 0, ClusteringScore: -1,
				Actions: []*models.FrontAction{
					{ID: 3, Name: "action_3", Description: "action_desc_3", Selected: false, Custom: false, Occurrence: 0, UsageRate: 0},
					{ID: 4, Name: "action_4", Description: "action_desc_4", Selected: false, Custom: false, Occurrence: 0, UsageRate: 0},
				}},
			{ID: 1, Name: "rc_1", Description: "rc_desc_1", Selected: false, Custom: false, Occurrence: 2, UsageRate: 1, ClusteringScore: -1,
				Actions: []*models.FrontAction{
					{ID: 1, Name: "action_1", Description: "action_desc_1", Selected: false, Custom: false, Occurrence: 2, UsageRate: 1},
					{ID: 2, Name: "action_2", Description: "action_desc_2", Selected: false, Custom: false, Occurrence: 1, UsageRate: 0.5},
				}},
			{Name: "new_custom_rc_1", Description: "new_custom_rc_desc_1", Selected: true, Custom: true,
				Actions: []*models.FrontAction{
					{Name: "new_custom_action_1", Description: "new_custom_action_desc_1", Selected: true, Custom: true},
				}},
		},
	}

	err := CloseIssueWithFeedback(db, 3, recommendation, []int64{1, 2}, groups.UserWithGroups{})
	if err != nil {
		t.Error(err)
	}
}

func TestFeedbackCustomRootCauseAndMultipleActions(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroy(db, t)
	dbInit(db, t)

	recommendation := models.FrontRecommendation{
		Tree: []*models.FrontRootCause{
			{ID: 2, Name: "rc_2", Description: "rc_desc_2", Selected: false, Custom: false, Occurrence: 0, UsageRate: 0, ClusteringScore: -1,
				Actions: []*models.FrontAction{
					{ID: 3, Name: "action_3", Description: "action_desc_3", Selected: false, Custom: false, Occurrence: 0, UsageRate: 0},
					{ID: 4, Name: "action_4", Description: "action_desc_4", Selected: false, Custom: false, Occurrence: 0, UsageRate: 0},
				}},
			{ID: 1, Name: "rc_1", Description: "rc_desc_1", Selected: false, Custom: false, Occurrence: 2, UsageRate: 1, ClusteringScore: -1,
				Actions: []*models.FrontAction{
					{ID: 1, Name: "action_1", Description: "action_desc_1", Selected: false, Custom: false, Occurrence: 2, UsageRate: 1},
					{ID: 2, Name: "action_2", Description: "action_desc_2", Selected: false, Custom: false, Occurrence: 1, UsageRate: 0.5},
				}},
			{Name: "new_custom_rc_1", Description: "new_custom_rc_desc_1", Selected: true, Custom: true,
				Actions: []*models.FrontAction{
					{Name: "new_custom_action_1", Description: "new_custom_action_desc_1", Selected: true, Custom: true},
					{Name: "new_custom_action_2", Description: "new_custom_action_desc_2", Selected: true, Custom: true},
					{Name: "new_custom_action_3", Description: "new_custom_action_desc_3", Selected: true, Custom: true},
				}},
		},
	}

	err := CloseIssueWithFeedback(db, 3, recommendation, []int64{1, 2}, groups.UserWithGroups{})
	if err != nil {
		t.Error(err)
	}
}

func TestFeedbackCustomAndExistingMultipleActions(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroy(db, t)
	dbInit(db, t)

	recommendation := models.FrontRecommendation{
		Tree: []*models.FrontRootCause{
			{ID: 2, Name: "rc_2", Description: "rc_desc_2", Selected: true, Custom: false, Occurrence: 0, UsageRate: 0, ClusteringScore: -1,
				Actions: []*models.FrontAction{
					{ID: 3, Name: "action_3", Description: "action_desc_3", Selected: false, Custom: false, Occurrence: 0, UsageRate: 0},
					{ID: 4, Name: "action_4", Description: "action_desc_4", Selected: true, Custom: false, Occurrence: 0, UsageRate: 0},
					{Name: "new_custom_action_1", Description: "new_custom_action_desc_1", Selected: true, Custom: true},
					{Name: "new_custom_action_2", Description: "new_custom_action_desc_2", Selected: false, Custom: true},
				}},
			{ID: 1, Name: "rc_1", Description: "rc_desc_1", Selected: false, Custom: false, Occurrence: 2, UsageRate: 1, ClusteringScore: -1,
				Actions: []*models.FrontAction{
					{ID: 1, Name: "action_1", Description: "action_desc_1", Selected: false, Custom: false, Occurrence: 2, UsageRate: 1},
					{ID: 2, Name: "action_2", Description: "action_desc_2", Selected: false, Custom: false, Occurrence: 1, UsageRate: 0.5},
				}},
		},
	}

	err := CloseIssueWithFeedback(db, 3, recommendation, []int64{1, 2}, groups.UserWithGroups{})
	if err != nil {
		t.Error(err)
	}
}
