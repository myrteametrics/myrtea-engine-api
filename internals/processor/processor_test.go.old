package processor

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/calendar"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/explainer/issues"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/fact"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/notifier"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/notifier/notification"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/reader"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/rule"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/situation"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/tasker"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/tests"
	"github.com/myrteametrics/myrtea-sdk/v4/engine"
	skd_models "github.com/myrteametrics/myrtea-sdk/v4/models"
	"github.com/myrteametrics/myrtea-sdk/v4/postgres"
	"github.com/myrteametrics/myrtea-sdk/v4/ruleeng"
)

func dbInitRepo(dbClient *sqlx.DB, t *testing.T) {
	dbDestroyRepo(dbClient, t)
	tests.DBExec(dbClient, `CREATE EXTENSION IF NOT EXISTS pgcrypto;`, t, true)
	tests.DBExec(dbClient, tests.UsersTableV1, t, true)
	tests.DBExec(dbClient, tests.CalendarTableV3, t, true)
	tests.DBExec(dbClient, tests.SituationDefinitionTableV1, t, true)
	tests.DBExec(dbClient, tests.SituationHistoryTableV1, t, true)
	tests.DBExec(dbClient, tests.FactDefinitionTableV1, t, true)
	tests.DBExec(dbClient, tests.SituationFactsTableV1, t, true)
	tests.DBExec(dbClient, tests.FactHistoryTableV1, t, true)
	tests.DBExec(dbClient, tests.RulesTableV1, t, true)
	tests.DBExec(dbClient, tests.RuleVersionsTableV1, t, true)
	tests.DBExec(dbClient, tests.IssuesTableV1, t, true)
	tests.DBExec(dbClient, tests.SituationRulesTableV1, t, true)
}

func dbDestroyRepo(dbClient *sqlx.DB, t *testing.T) {
	tests.DBExec(dbClient, tests.SituationRulesDropTableV1, t, false)
	tests.DBExec(dbClient, tests.IssuesDropTableV1, t, false)
	tests.DBExec(dbClient, tests.FactHistoryDropTableV1, t, false)
	tests.DBExec(dbClient, tests.SituationFactsDropTableV1, t, false)
	tests.DBExec(dbClient, tests.FactDefinitionDropTableV1, t, false)
	tests.DBExec(dbClient, tests.SituationHistoryDropTableV1, t, false)
	tests.DBExec(dbClient, tests.SituationDefinitionDropTableV1, t, false)
	tests.DBExec(dbClient, tests.UsersDropTableV1, t, false)
	tests.DBExec(dbClient, tests.RuleVersionsDropTableV1, t, false)
	tests.DBExec(dbClient, tests.RulesDropTableV1, t, false)
	tests.DBExec(dbClient, tests.CalendarDropTableV3, t, false)
}

func TestEvalProcessor(t *testing.T) {
	//Init Postgres repo for Issues
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	tests.CheckDebugLogs(t)

	db := tests.DBClient(t)
	defer dbDestroyRepo(db, t)
	dbInitRepo(db, t)

	postgres.ReplaceGlobals(db)
	issues.ReplaceGlobals(issues.NewPostgresRepository(db))
	fact.ReplaceGlobals(fact.NewPostgresRepository(db))
	situation.ReplaceGlobals(situation.NewPostgresRepository(db))
	rule.ReplaceGlobals(rule.NewPostgresRepository(db))
	notifier.ReplaceGlobals(notifier.NewNotifier())
	notification.ReplaceGlobals(notification.NewPostgresRepository(db))
	tasker.ReplaceGlobals(tasker.NewTasker())
	calendar.ReplaceGlobals(calendar.NewPostgresRepository(db))

	calendar.Init()

	timestamp := time.Now().Truncate(1 * time.Millisecond).UTC()

	idFact1, err := fact.R().Create(engine.Fact{Name: "fact_test_1"})
	idFact2, err := fact.R().Create(engine.Fact{Name: "fact_test_2"})
	idFact3, err := fact.R().Create(engine.Fact{Name: "object", IsObject: true})

	if err != nil {
		t.Error(err)
	}

	//Situation
	situation1 := situation.Situation{
		Name:  "test_name",
		Facts: []int64{idFact1, idFact2, idFact3},
	}

	situationID, err := situation.R().Create(situation1)
	if err != nil {
		t.Error(err)
	}

	//Fact History
	rangeInterval := 5 //number of days
	from := timestamp.Add(-(time.Hour * 24 * time.Duration(rangeInterval)))

	for i := 0; i < rangeInterval; i++ {
		ts := from.Add(time.Hour * 24 * time.Duration(i))

		item1 := &reader.Item{
			Aggs: map[string]*reader.ItemAgg{
				"doc_count": {Value: 10},
			},
		}

		item2 := &reader.Item{
			Aggs: map[string]*reader.ItemAgg{
				"doc_count": {Value: 5},
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
			TS:       ts,
			FactsIDS: factHistory,
		}
		situation.Persist(historyRecord, false)
	}

	task1 := ruleeng.ActionDef{
		Name: ruleeng.Expression(`"create-issue"`),
		Parameters: map[string]ruleeng.Expression{
			"id":             ruleeng.Expression(`"key_task_test"`),
			"name":           ruleeng.Expression(`"create-issue"`),
			"level":          ruleeng.Expression(`"critical"`),
			"timeout":        ruleeng.Expression(`"12h"`),
			"isNotification": ruleeng.Expression(`false`),
		},
	}
	ruleID, _ := rule.R().Create(rule.Rule{
		Name: "testRule1",
		DefaultRule: ruleeng.DefaultRule{Cases: []ruleeng.Case{
			{
				Name:      "case1",
				Condition: ruleeng.Expression("object.f1 > object.f2 && fact_test_1.aggs.doc_count.value == 10"),
				Actions:   []ruleeng.ActionDef{task1},
			},
		}},
		Enabled: true,
	})
	situation.R().SetRules(situationID, []int64{int64(ruleID)})

	//Test
	documents := make([]skd_models.Document, 0)

	var object1 map[string]interface{}
	var object2 map[string]interface{}
	data1 := `{"f3":2,"f2":1}`
	data2 := `{"f1":2,"f2":1}`
	json.Unmarshal([]byte(data1), &object1)
	json.Unmarshal([]byte(data2), &object2)

	documents = append(documents,
		skd_models.Document{ID: "id_object1", Source: map[string]interface{}{"f3": 2, "f2": 1}},   // no match (missing f1)
		skd_models.Document{ID: "id_object2", Source: map[string]interface{}{"f1": 2, "f2": 1}},   // match
		skd_models.Document{ID: "id_object3", Source: map[string]interface{}{"f1": 2, "f2": 100}}, // no match (not f1 > f2)
		skd_models.Document{ID: "id_object4", Source: map[string]interface{}{"f1": 20, "f2": 1}},  // match
		skd_models.Document{ID: "id_object5", Source: map[string]interface{}{"f1": 3, "f2": 1}},   // match
		skd_models.Document{ID: "id_object6", Source: map[string]interface{}{"f1": 19, "f2": 1}},  // match
	)

	tasker.T().StartBatchProcessor()

	err = ReceiveObjects("object", documents)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	tasker.T().StopBatchProcessor()

	var count int
	postgres.DB().Get(&count, "select count(*) from fact_history_v1 where id = $1", idFact3)
	if count != 4 {
		t.Error("invalid fact history rows")
		t.FailNow()
	}

	issues, err := issues.R().GetAll()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if len(issues) != 1 {
		// 4 matched rule, but 1 issue + 3 skipped
		t.Error("Failed to get all issues")
		t.Log(issues)
	}
}
