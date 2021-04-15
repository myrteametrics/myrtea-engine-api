package explainer

import (
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/explainer/issues"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/fact"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/models"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/reader"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/situation"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/tests"
	"github.com/myrteametrics/myrtea-sdk/v4/engine"
	"github.com/myrteametrics/myrtea-sdk/v4/postgres"
)

func dbInitRepo(dbClient *sqlx.DB, t *testing.T) {
	dbDestroyRepo(dbClient, t)
	_, err := dbClient.Exec(`CREATE EXTENSION IF NOT EXISTS pgcrypto;`)
	if err != nil {
		t.Error(err)
	}

	tests.DBExec(dbClient, tests.CalendarTableV3, t, true)
	tests.DBExec(dbClient, tests.UsersTableV1, t, true)
	tests.DBExec(dbClient, tests.SituationDefinitionTableV1, t, true)
	tests.DBExec(dbClient, tests.SituationHistoryTableV1, t, true)
	tests.DBExec(dbClient, tests.FactDefinitionTableV1, t, true)
	tests.DBExec(dbClient, tests.FactHistoryTableV1, t, true)
	tests.DBExec(dbClient, tests.SituationFactsTableV1, t, true)
	tests.DBExec(dbClient, tests.IssuesTableV1, t, true)
}

func dbDestroyRepo(dbClient *sqlx.DB, t *testing.T) {
	tests.DBExec(dbClient, tests.IssuesDropTableV1, t, true)
	tests.DBExec(dbClient, tests.SituationFactsDropTableV1, t, true)
	tests.DBExec(dbClient, tests.FactHistoryDropTableV1, t, true)
	tests.DBExec(dbClient, tests.FactDefinitionDropTableV1, t, true)
	tests.DBExec(dbClient, tests.SituationHistoryDropTableV1, t, true)
	tests.DBExec(dbClient, tests.SituationDefinitionDropTableV1, t, true)
	tests.DBExec(dbClient, tests.UsersDropTableV1, t, true)
	tests.DBExec(dbClient, tests.CalendarDropTableV3, t, true)
}

func TestCreateIssueWithoutTimeout(t *testing.T) {
	//Init Postgres repo for Issues
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroyRepo(db, t)
	dbInitRepo(db, t)

	repo := issues.NewPostgresRepository(db)
	issues.ReplaceGlobals(repo)

	s := situation.NewPostgresRepository(db)

	//Situation
	groups := []int64{1, 2}
	situation1 := situation.Situation{Name: "test_name", Groups: groups}
	situationID, err := s.Create(situation1)
	if err != nil {
		t.Error(err)
	}

	timestamp := time.Now().Truncate(1 * time.Millisecond).UTC()

	rule := models.RuleData{
		RuleID:      1,
		RuleVersion: 0,
		CaseName:    "case1",
	}

	timeout := time.Hour * 24

	_, err = CreateIssue(situationID, timestamp, 0, rule, "issue1", models.Critical, timeout, "key1")
	if err != nil {
		t.Error(err)
	}
}

func TestCreateIssueWithTimeout(t *testing.T) {
	//Init Postgres repo for Issues
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroyRepo(db, t)
	dbInitRepo(db, t)

	repo := issues.NewPostgresRepository(db)
	issues.ReplaceGlobals(repo)

	s := situation.NewPostgresRepository(db)

	//Situation
	groups := []int64{1, 2}
	situation1 := situation.Situation{Name: "test_name", Groups: groups}
	situationID, err := s.Create(situation1)
	if err != nil {
		t.Error(err)
	}

	//Rule
	rule := models.RuleData{
		RuleID: 1,
	}

	timestamp1 := time.Now().Truncate(1 * time.Millisecond).UTC()
	timeout := 24 * time.Hour
	expirationTS := timestamp1.Add(timeout).UTC()

	issueName := "issue1"
	key1 := "my_rule_case_key1"
	key2 := "my_rule_case_key2"

	//First issue
	issue1 := models.Issue{
		Name:         issueName,
		Key:          key1,
		Level:        models.Critical,
		SituationID:  situationID,
		SituationTS:  timestamp1,
		ExpirationTS: expirationTS,
		Rule:         rule,
		State:        models.Open,
		CreationTS:   timestamp1,
	}

	_, err = repo.Create(issue1)
	if err != nil {
		t.Error(err)
	}

	//Testing case1
	timestamp2 := timestamp1.Add(5 * time.Hour)
	id1, err := CreateIssue(situationID, timestamp2, 0, rule, issueName, models.Critical, timeout, key1)
	if err != nil {
		t.Error(err)
	}

	if id1 != 0 {
		t.Error("Issue has been created when it shouldn't be, because is within the expiration time and with the same key")
	}

	//Testing case2
	id2, err := CreateIssue(situationID, timestamp2, 0, rule, issueName, models.Critical, timeout, key2)
	if err != nil {
		t.Error(err)
	}

	if id2 <= 0 {
		t.Error("Issue has not been created when it should have, within expiration date but not the same key")
	}

	//Testing case3
	timestamp3 := timestamp1.Add(timeout + (1 * time.Hour))
	id3, err := CreateIssue(situationID, timestamp3, 0, rule, issueName, models.Critical, timeout, key1)
	if err != nil {
		t.Error(err)
	}

	if id3 <= 0 {
		t.Error("Issue has not been created when it should have, has the same key but not in the expiration date")
	}
}

func TestCreateIssueWithTimeoutAndState(t *testing.T) {
	//Init Postgres repo for Issues
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroyRepo(db, t)
	dbInitRepo(db, t)

	repo := issues.NewPostgresRepository(db)
	issues.ReplaceGlobals(repo)

	s := situation.NewPostgresRepository(db)

	//Situation
	groups := []int64{1, 2}
	situation1 := situation.Situation{Name: "test_name", Groups: groups}
	situationID, err := s.Create(situation1)
	if err != nil {
		t.Error(err)
	}

	//Rule
	rule := models.RuleData{
		RuleID: 1,
	}

	timestamp1 := time.Now().Truncate(1 * time.Millisecond).UTC()
	timeout := 24 * time.Hour
	expirationTS := timestamp1.Add(timeout).UTC()

	issueName := "issue1"
	key1 := "my_rule_case_key1"

	//First issue
	issue1 := models.Issue{
		Name:         issueName,
		Key:          key1,
		Level:        models.Critical,
		SituationID:  situationID,
		SituationTS:  timestamp1,
		ExpirationTS: expirationTS,
		Rule:         rule,
		State:        models.ClosedNoFeedback,
		CreationTS:   timestamp1,
	}

	_, err = repo.Create(issue1)
	if err != nil {
		t.Error(err)
	}

	//Testing case1
	timestamp2 := timestamp1.Add(5 * time.Hour)
	id1, err := CreateIssue(situationID, timestamp2, 0, rule, issueName, models.Critical, timeout, key1)
	if err != nil {
		t.Error(err)
	}

	if id1 <= 0 {
		t.Error("Issue has not been created when it should have, is within the expiration date but the first issue is closed")
	}
}

func TestIssueGetFactsHistory(t *testing.T) {
	//Init Postgres repo for Issues
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroyRepo(db, t)
	dbInitRepo(db, t)

	repo := issues.NewPostgresRepository(db)
	issues.ReplaceGlobals(repo)

	f := fact.NewPostgresRepository(db)
	fact.ReplaceGlobals(f)

	s := situation.NewPostgresRepository(db)
	situation.ReplaceGlobals(s)

	postgres.ReplaceGlobals(db)

	idFact1, err := f.Create(engine.Fact{
		Name: "fact_test_1",
		Intent: &engine.IntentFragment{
			Name:     "count",
			Operator: engine.Count,
			Term:     "test_model",
		},
		Model: "test_model",
	})

	idFact2, err := f.Create(engine.Fact{
		Name: "fact_test_2",
		Intent: &engine.IntentFragment{
			Name:     "count",
			Operator: engine.Count,
			Term:     "test_model",
		},
		Model: "test_model",
	})

	idFact3, err := f.Create(engine.Fact{Name: "fact_test_3"})

	if err != nil {
		t.Error(err)
	}

	//Situation
	groups := []int64{1, 2}
	situation1 := situation.Situation{
		Name:   "test_name",
		Facts:  []int64{idFact1, idFact2, idFact3},
		Groups: groups,
	}

	situationID, err := s.Create(situation1)
	if err != nil {
		t.Error(err)
	}

	//Fact History
	rangeInterval := 6
	delta := time.Minute * 10
	timestamp := time.Now().Truncate(1 * time.Millisecond).UTC()
	issueDate := timestamp.Add(-(delta * time.Duration(rangeInterval/2)))
	from := timestamp.Add(-(delta * time.Duration(rangeInterval)))

	for i := 1; i <= rangeInterval; i++ {
		ts := from.Add(delta * time.Duration(i))

		item1 := &reader.Item{
			Aggs: map[string]*reader.ItemAgg{
				"doc_count": {Value: i * 10},
			},
		}

		item2 := &reader.Item{
			Aggs: map[string]*reader.ItemAgg{
				"doc_count": {Value: i * 5},
			},
		}

		fact.PersistFactResult(idFact1, ts, 0, 0, item1, true)
		fact.PersistFactResult(idFact2, ts, 0, 0, item2, true)

		factHistory := make(map[int64]*time.Time, 0)

		factHistory[idFact1] = &ts
		factHistory[idFact2] = &ts

		//Situation History
		historyRecord := situation.HistoryRecord{
			ID:       situationID,
			FactsIDS: factHistory,
			TS:       ts,
		}
		situation.Persist(historyRecord, false)
	}

	//Rule
	rule := models.RuleData{
		RuleID: 1,
	}

	timeout := 24 * time.Hour
	expirationTS := issueDate.Add(timeout).UTC()

	issueName := "issue1"
	key1 := "my_rule_case_key1"

	//First issue
	issue1 := models.Issue{
		Name:         issueName,
		Key:          key1,
		Level:        models.Critical,
		SituationID:  situationID,
		SituationTS:  issueDate,
		ExpirationTS: expirationTS,
		Rule:         rule,
		State:        models.ClosedNoFeedback,
		CreationTS:   issueDate,
	}

	issueID, err := repo.Create(issue1)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	issue1.ID = issueID

	//Facts and their history for the issue
	issueFactsHistory, found, err := GetFactsHistory(issue1)
	if err != nil || !found {
		t.Error(err)
		t.FailNow()
	}

	for _, iFactHistory := range issueFactsHistory {
		if iFactHistory.ID == idFact1 && len(iFactHistory.History) <= 0 {
			t.Error("Facts history 1 for the issue got wrong !")
			t.FailNow()
		}

		if iFactHistory.ID == idFact2 && len(iFactHistory.History) <= 0 {
			t.Error("Facts history 2 for the issue got wrong !")
			t.FailNow()
		}

		if iFactHistory.ID == idFact3 && len(iFactHistory.History) > 0 {
			t.Error("Facts history 3 for the issue got wrong !")
			t.FailNow()
		}
	}
}
