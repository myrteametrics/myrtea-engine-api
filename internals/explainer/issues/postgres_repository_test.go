package issues

import (
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/models"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/security/users"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/situation"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/tests"
)

func dbInitRepo(dbClient *sqlx.DB, t *testing.T) {
	dbDestroyRepo(dbClient, t)
	_, err := dbClient.Exec(`CREATE EXTENSION IF NOT EXISTS pgcrypto;`)
	if err != nil {
		t.Error(err)
	}

	_, err = dbClient.Exec(tests.UsersTableV1)
	if err != nil {
		t.Error(err)
	}

	tests.DBExec(dbClient, tests.CalendarTableV3, t, true)

	_, err = dbClient.Exec(tests.SituationDefinitionTableV1)
	if err != nil {
		t.Error(err)
	}

	_, err = dbClient.Exec(tests.FactDefinitionTableV1)
	if err != nil {
		t.Error(err)
	}

	_, err = dbClient.Exec(tests.SituationFactsTableV1)
	if err != nil {
		t.Error(err)
	}

	_, err = dbClient.Exec(tests.IssuesTableV1)
	if err != nil {
		t.Error(err)
	}
}

func dbDestroyRepo(dbClient *sqlx.DB, t *testing.T) {
	_, err := dbClient.Exec(tests.IssuesDropTableV1)
	if err != nil {
		t.Error(err)
	}
	_, err = dbClient.Exec(tests.SituationFactsDropTableV1)
	if err != nil {
		t.Error(err)
	}
	_, err = dbClient.Exec(tests.FactDefinitionDropTableV1)
	if err != nil {
		t.Error(err)
	}
	_, err = dbClient.Exec(tests.SituationDefinitionDropTableV1)
	if err != nil {
		t.Error(err)
	}

	tests.DBExec(dbClient, tests.CalendarDropTableV3, t, false)

	_, err = dbClient.Exec(tests.UsersDropTableV1)
	if err != nil {
		t.Error(err)
	}
}

func TestNewPostgresRepository(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	r := NewPostgresRepository(tests.DBClient(t))
	if r == nil {
		t.Error("Issue Repository is nil")
	}
}

func TestPostgresReplaceGlobal(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	r := NewPostgresRepository(tests.DBClient(t))
	reverse := ReplaceGlobals(r)
	if R() == nil {
		t.Error("Global Issue repository is nil")
	}
	reverse()
	if R() != nil {
		t.Error("Global Issue repository is not nil after reverse")
	}
}

func TestPostgresCreateAndGet(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroyRepo(db, t)
	dbInitRepo(db, t)
	r := NewPostgresRepository(db)
	s := situation.NewPostgresRepository(db)

	var err error
	groups := []int64{1, 2}
	issueGet, found, err := r.Get(1)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if found {
		t.Error("found a issue from nowhere")
	}

	//Situation
	situation1 := situation.Situation{Name: "test_name", Groups: groups}
	situationID, err := s.Create(situation1)
	if err != nil {
		t.Error(err)
	}

	timestamp := time.Now().Truncate(1 * time.Millisecond).UTC()

	//Rule
	rule := models.RuleData{
		RuleID: 1,
	}

	issue := models.Issue{
		SituationID: situationID,
		SituationTS: timestamp,
		Rule:        rule,
		State:       models.Open,
	}
	id, err := r.Create(issue)
	if err != nil {
		t.Error(err)
	}

	issueGet, found, err = r.Get(id)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if !found {
		t.Error("Issue doesn't exists after the creation")
		t.FailNow()
	}
	if id != issueGet.ID {
		t.Error("invalid issue ID")
	}
}

func TestPostgresUpdate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroyRepo(db, t)
	dbInitRepo(db, t)
	r := NewPostgresRepository(db)
	s := situation.NewPostgresRepository(db)

	//Situation
	groupList := []int64{1, 2}
	situation1 := situation.Situation{Name: "test_name", Groups: groupList}
	situationID, err := s.Create(situation1)
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
		Name:         "name1",
		Level:        models.Critical,
		SituationID:  situationID,
		ExpirationTS: timestamp.Add(24 * time.Hour).UTC(),
		SituationTS:  timestamp,
		Rule:         rule,
		State:        models.Open,
	}

	id, err := r.Create(issue)
	if err != nil {
		t.Error(err)
	}

	issue.State = models.ClosedNoFeedback

	err = r.Update(nil, id, issue, users.User{})
	if err != nil {
		t.Error(err)
	}

	issueGet, found, err := r.Get(id)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if !found {
		t.Error("Issue doesn't exists after the creation")
		t.FailNow()
	}
	if issueGet.State == models.Open {
		t.Error("Issue not properly updated")
	}
}

func TestPostgresGetByStates(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroyRepo(db, t)
	dbInitRepo(db, t)
	r := NewPostgresRepository(db)
	s := situation.NewPostgresRepository(db)

	//Situation1
	groups := []int64{1, 2}
	situation1 := situation.Situation{Name: "test_name1", Groups: groups}
	situation1ID, err := s.Create(situation1)
	if err != nil {
		t.Error(err)
	}

	//Situation2
	situation2 := situation.Situation{Name: "test_name2", Groups: groups}
	situation2ID, err := s.Create(situation2)
	if err != nil {
		t.Error(err)
	}

	timestamp := time.Now().Truncate(1 * time.Millisecond).UTC()

	issue1 := models.Issue{
		SituationID: situation1ID,
		SituationTS: timestamp,
		State:       models.Open,
	}

	issue2 := models.Issue{
		SituationID: situation2ID,
		SituationTS: timestamp,
		State:       models.ClosedFeedback,
	}

	_, err = r.Create(issue1)
	if err != nil {
		t.Error(err)
	}

	_, err = r.Create(issue2)
	if err != nil {
		t.Error(err)
	}

	issues, err := r.GetByStates([]string{models.ClosedFeedback.String()})
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if len(issues) != 1 {
		t.Error("Failed issues not filtered by states")
	}
}

func TestPostgresGetAll(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroyRepo(db, t)
	dbInitRepo(db, t)
	r := NewPostgresRepository(db)
	s := situation.NewPostgresRepository(db)

	//Situation1
	groups := []int64{1, 2}
	situation1 := situation.Situation{Name: "test_name1", Groups: groups}
	situation1ID, err := s.Create(situation1)
	if err != nil {
		t.Error(err)
	}

	//Situation2
	situation2 := situation.Situation{Name: "test_name2", Groups: groups}
	situation2ID, err := s.Create(situation2)
	if err != nil {
		t.Error(err)
	}

	timestamp := time.Now().Truncate(1 * time.Millisecond).UTC()

	issue1 := models.Issue{
		SituationID: situation1ID,
		SituationTS: timestamp,
		State:       models.Open,
	}

	issue2 := models.Issue{
		SituationID: situation2ID,
		SituationTS: timestamp,
		State:       models.Open,
	}

	_, err = r.Create(issue1)
	if err != nil {
		t.Error(err)
	}

	_, err = r.Create(issue2)
	if err != nil {
		t.Error(err)
	}

	issues, err := r.GetAll()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if len(issues) != 2 {
		t.Error("Failed to get all issues")
	}
}

func TestPostgresGetByStateByPage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroyRepo(db, t)
	dbInitRepo(db, t)
	r := NewPostgresRepository(db)
	s := situation.NewPostgresRepository(db)

	//Situation1
	groups := []int64{1, 2}
	situation1 := situation.Situation{Name: "test_name1", Groups: groups}
	situation1ID, err := s.Create(situation1)
	if err != nil {
		t.Error(err)
	}

	//Situation2
	situation2 := situation.Situation{Name: "test_name2", Groups: groups}
	situation2ID, err := s.Create(situation2)
	if err != nil {
		t.Error(err)
	}

	timestamp := time.Now().Truncate(1 * time.Millisecond).UTC()

	_, err = r.Create(models.Issue{
		SituationID: situation1ID,
		SituationTS: timestamp,
		State:       models.Open,
	})
	if err != nil {
		t.Error(err)
	}

	_, err = r.Create(models.Issue{
		SituationID: situation1ID,
		SituationTS: timestamp,
		State:       models.Open,
	})
	if err != nil {
		t.Error(err)
	}

	_, err = r.Create(models.Issue{
		SituationID: situation1ID,
		SituationTS: timestamp,
		State:       models.Open,
	})
	if err != nil {
		t.Error(err)
	}

	_, err = r.Create(models.Issue{
		SituationID: situation2ID,
		SituationTS: timestamp,
		State:       models.Open,
	})

	_, err = r.Create(models.Issue{
		SituationID: situation2ID,
		SituationTS: timestamp,
		State:       models.ClosedFeedback,
	})

	if err != nil {
		t.Error(err)
	}

	issues, total, err := r.GetByStateByPage([]string{"open"}, models.SearchOptions{Limit: 2, Offset: 0})
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if total != 4 {
		t.Error("invalid total")
	}
	if len(issues) != 2 {
		t.Error("Failed to get all issues")
	}
	if issues[0].ID != 1 || issues[1].ID != 2 {
		t.Error("invalid issue ID")
		t.Log(issues)
	}

	issues, total, err = r.GetByStateByPage([]string{"open"}, models.SearchOptions{Limit: 2, Offset: 2})
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if total != 4 {
		t.Error("invalid total")
	}
	if len(issues) != 2 {
		t.Error("Failed to get all issues")
	}
	if issues[0].ID != 3 || issues[1].ID != 4 {
		t.Error("invalid issue ID")
		t.Log(issues)
	}

	issues, total, err = r.GetByStateByPage([]string{"open"}, models.SearchOptions{Limit: 3, Offset: 3})
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if total != 4 {
		t.Error("invalid total")
	}
	if len(issues) != 1 {
		t.Error("Failed to get all issues")
	}
	if issues[0].ID != 4 {
		t.Error("invalid issue ID")
		t.Log(issues)
	}

	issues, total, err = r.GetByStateByPage([]string{"open"}, models.SearchOptions{Limit: 2, Offset: 4})
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if total != 4 {
		t.Error("invalid total")
	}
	if len(issues) != 0 {
		t.Error("Failed to get all issues")
	}
}
