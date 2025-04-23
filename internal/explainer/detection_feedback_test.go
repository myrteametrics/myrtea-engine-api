package explainer

// import (
// 	"testing"
// 	"time"

// 	"github.com/jmoiron/sqlx"
// 	"github.com/myrteametrics/myrtea-engine-api/v5/internals/explainer/issues"
// 	"github.com/myrteametrics/myrtea-engine-api/v5/internals/groups"
// 	"github.com/myrteametrics/myrtea-engine-api/v5/internals/models"
// 	"github.com/myrteametrics/myrtea-engine-api/v5/internals/situation"
// 	"github.com/myrteametrics/myrtea-engine-api/v5/internals/tests"
// 	"github.com/myrteametrics/myrtea-engine-api/v5/internals/users"
// 	"github.com/myrteametrics/myrtea-sdk/v5/security"
// )

// func dbInitDetection(dbClient *sqlx.DB, t *testing.T) {

// 	dbDestroyDetection(dbClient, t)
// 	tests.DBExec(dbClient, tests.UsersTableV1, t, true)
// 	tests.DBExec(dbClient, tests.SituationDefinitionTableV1, t, true)
// 	tests.DBExec(dbClient, tests.FactDefinitionTableV1, t, true)
// 	tests.DBExec(dbClient, tests.SituationFactsTableV1, t, true)
// 	tests.DBExec(dbClient, tests.IssuesTableV1, t, true)
// 	tests.DBExec(dbClient, tests.IssueFeedbackTableV3, t, false)

// 	ur := users.NewPostgresRepository(dbClient)
// 	ur.Create(security.UserWithPassword{User: security.User{Login: "a"}})
// 	ur.Create(security.UserWithPassword{User: security.User{Login: "b"}})

// 	sr := situation.NewPostgresRepository(dbClient)
// 	situation.ReplaceGlobals(sr)
// 	s := situation.Situation{
// 		Groups: []int64{1, 2},
// 		Name:   "situation_1",
// 		Facts:  []int64{},
// 	}
// 	situationID1, err := sr.Create(s)
// 	if err != nil {
// 		t.Error(err)
// 		t.FailNow()
// 	}

// 	ir := issues.NewPostgresRepository(dbClient)
// 	issues.ReplaceGlobals(ir)
// 	issue := models.Issue{
// 		SituationID: situationID1,
// 		SituationTS: time.Now().Truncate(1 * time.Millisecond).UTC(),
// 		State:       models.Open,
// 	}
// 	createIssue(ir, issue, t, true)
// 	createIssue(ir, issue, t, true)
// 	createIssue(ir, issue, t, true)
// }

// func dbDestroyDetection(dbClient *sqlx.DB, t *testing.T) {
// 	tests.DBExec(dbClient, tests.IssueFeedbackDropTableV3, t, false)
// 	tests.DBExec(dbClient, tests.IssuesDropTableV1, t, false)
// 	tests.DBExec(dbClient, tests.SituationFactsDropTableV1, t, false)
// 	tests.DBExec(dbClient, tests.FactDefinitionDropTableV1, t, false)
// 	tests.DBExec(dbClient, tests.SituationDefinitionDropTableV1, t, false)
// 	tests.DBExec(dbClient, tests.UsersDropTableV1, t, true)
// }

// func checkRating(db *sqlx.DB, t *testing.T, issueID int64, userID int64, rating int, expectedAvg float64) {
// 	err := AddIssueDetectionFeedback(db, issueID, userID, rating, []int64{1})
// 	if err != nil {
// 		t.Error(err)
// 		t.FailNow()
// 	}
// 	issue, _, _ := issues.R().Get(issueID, groups.GetTokenAllGroups())
// 	if issue.DetectionRatingAvg != expectedAvg {
// 		t.Error("invalid average rating", issue.DetectionRatingAvg)
// 		t.FailNow()
// 	}
// }

// func TestAddIssueDetectionFeedback(t *testing.T) {
// 	if testing.Short() {
// 		t.Skip("skipping postgresql test in short mode")
// 	}
// 	db := tests.DBClient(t)
// 	defer dbDestroyDetection(db, t)
// 	dbInitDetection(db, t)

// 	issue, _, _ := issues.R().Get(1, groups.GetTokenAllGroups())
// 	if issue.DetectionRatingAvg != -1 {
// 		t.Error("invalid default average rating", issue.DetectionRatingAvg)
// 		t.FailNow()
// 	}

// 	tx, _ := db.Beginx()
// 	avg, err := calculateDetectionRatingAverage(tx, 1)
// 	if err != nil {
// 		tx.Rollback()
// 		t.Error(err)
// 		t.FailNow()
// 	}
// 	if avg != -1 {
// 		tx.Rollback()
// 		t.Error("invalid default average rating", issue.DetectionRatingAvg)
// 		t.FailNow()
// 	}
// 	tx.Commit()

// 	checkRating(db, t, 1, 1, 2, 2)

// 	checkRating(db, t, 2, 1, 5, 5)
// 	checkRating(db, t, 2, 2, 2, 3.5)
// 	checkRating(db, t, 2, 2, 4, 4.5) // Erase previous feedback of user 2
// }
