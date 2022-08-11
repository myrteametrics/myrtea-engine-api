package evaluator

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/calendar"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/fact"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/reader"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/rule"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/situation"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/tests"
	"github.com/myrteametrics/myrtea-sdk/v4/engine"
	"github.com/myrteametrics/myrtea-sdk/v4/postgres"
)

func dbInit(dbClient *sqlx.DB, t *testing.T) {
	dbDestroy(dbClient, t)
	tests.DBExec(dbClient, tests.CalendarTableV3, t, true)
	tests.DBExec(dbClient, tests.SituationDefinitionTableV1, t, true)
	tests.DBExec(dbClient, tests.FactDefinitionTableV1, t, true)
	tests.DBExec(dbClient, tests.SituationFactsTableV1, t, true)
	tests.DBExec(dbClient, tests.RulesTableV1, t, true)
	tests.DBExec(dbClient, tests.RuleVersionsTableV1, t, true)
	tests.DBExec(dbClient, tests.SituationRulesTableV1, t, true)
	tests.DBExec(dbClient, tests.SituationHistoryTableV1, t, true)
	tests.DBExec(dbClient, tests.FactHistoryTableV1, t, true)
}

func dbDestroy(dbClient *sqlx.DB, t *testing.T) {
	tests.DBExec(dbClient, tests.SituationRulesDropTableV1, t, true)
	tests.DBExec(dbClient, tests.SituationFactsDropTableV1, t, true)
	tests.DBExec(dbClient, tests.FactDefinitionDropTableV1, t, true)
	tests.DBExec(dbClient, tests.SituationDefinitionDropTableV1, t, true)
	tests.DBExec(dbClient, tests.RuleVersionsDropTableV1, t, true)
	tests.DBExec(dbClient, tests.RulesDropTableV1, t, true)
	tests.DBExec(dbClient, tests.SituationHistoryDropTableV1, t, true)
	tests.DBExec(dbClient, tests.FactHistoryDropTableV1, t, true)
	tests.DBExec(dbClient, tests.CalendarDropTableV3, t, true)
}

func TestEvaluator(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	tests.CheckDebugLogs(t)

	db := tests.DBClient(t)
	defer dbDestroy(db, t)
	dbInit(db, t)
	postgres.ReplaceGlobals(db)
	fact.ReplaceGlobals(fact.NewPostgresRepository(db))
	rule.ReplaceGlobals(rule.NewPostgresRepository(db))
	situation.ReplaceGlobals(situation.NewPostgresRepository(db))
	calendar.ReplaceGlobals(calendar.NewPostgresRepository(db))

	calendar.Init()

	fact1ID, _ := fact.R().Create(engine.Fact{Name: "fact_test_1"})

	s := situation.Situation{Name: "test"}

	sID, err := situation.R().Create(s)
	if err != nil {
		t.Error(err)
	}

	t1, _ := time.Parse("2006-01-02T15:04:05", "2019-06-14T15:48:50")
	situations := []SituationToEvaluate{
		{ID: sID, TS: t1},
	}

	item := &reader.Item{
		Key:         "fact_test_1",
		KeyAsString: "fact_test_1",
		Aggs: map[string]*reader.ItemAgg{
			"agg0":      {Value: 10},
			"doc_count": {Value: 10},
		},
		Buckets: nil,
	}
	sh := situation.HistoryRecord{
		ID:       sID,
		TS:       t1,
		FactsIDS: map[int64]*time.Time{fact1ID: &t1},
	}

	fact.PersistFactResult(1, t1, 0, 0, item, true)
	situation.Persist(sh, false)

	var rule1 rule.Rule
	json.Unmarshal([]byte(ruleStr), &rule1)
	rule1ID, _ := rule.R().Create(rule1)

	situation.R().SetRules(sID, []int64{int64(rule1ID)})
	time.Sleep(500 * time.Millisecond)

	evaluatedSituations, err := EvaluateSituations(situations, "Standard")

	if len(evaluatedSituations) != 1 {
		t.Errorf("The number os evaluated situation is not as expected")
	}
	agenda := evaluatedSituations[0].Agenda
	if agenda[0].GetName() != "set" || agenda[1].GetName() != "set" || agenda[2].GetName() != "notify" {
		t.Errorf("The actions names are not ar expected")
	}

	if agenda[0].GetParameters()["status.A"].(float64) != float64(3) || agenda[1].GetParameters()["status.B"].(float64) != float64(5) || agenda[2].GetParameters()["id"].(string) != "notify-1" {
		t.Errorf("The actions parameters are not ar expected")
	}
}

var ruleStr = `{	
	"name": "rule1",
	"description": "this is the rule 1",
	"cases": [
	  {
		"name": "case1",
		"condition": "fact_test_1.aggs.agg0.value == fact_test_1.aggs.doc_count.value",
		"actions": [
		  {
			"name": "\"set\"",
			"parameters": {
			  "status.A": "1 + 2"
			}
		  },
		  {
			"name": "\"set\"",
			"parameters": {
			  "status.B": "2 + 3"
			}
		  },
		  {
			"name": "\"notify\"",
			"parameters": {
			  "id": "\"notify-1\"",
			  "level": "\"info\"",
			  "title": "\"my_title\"",
			  "description": "\"my_description\"",
			  "timeout": "\"10s\"",
			  "groups": "[1,2]"
			}
		  }
		]
	  }
	],
	"enabled": true
  }`
