package tasker

import (
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/calendar"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/evaluator"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/explainer/issues"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/fact"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/models"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/reader"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/rule"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/situation"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/tests"
	"github.com/myrteametrics/myrtea-sdk/v4/engine"
	"github.com/myrteametrics/myrtea-sdk/v4/postgres"
	"github.com/myrteametrics/myrtea-sdk/v4/ruleeng"
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
	tests.DBExec(dbClient, tests.IssuesTableV1, t, true)
}

func dbDestroy(dbClient *sqlx.DB, t *testing.T) {
	tests.DBExec(dbClient, tests.IssuesDropTableV1, t, true)
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

func TestNewTasker(t *testing.T) {
	tasker := NewTasker()
	if tasker == nil {
		t.Error("tasker constructor returns nil")
	}
}

func TestReplaceGlobal(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	tasker := NewTasker()
	reverse := ReplaceGlobals(tasker)
	if T() == nil {
		t.Error("Global tasker is nil")
	}
	reverse()
	if T() != nil {
		t.Error("Global situation repository is not nil after reverse")
	}
}

// func TestPersisDataPerform(t *testing.T) {
// 	if testing.Short() {
// 		t.Skip("skipping postgresql test in short mode")
// 	}

// 	db := tests.DBClient(t)
// 	defer dbDestroy(db, t)
// 	dbInit(db, t)
// 	postgres.ReplaceGlobals(db)

// 	id := int64(1)
// 	ts1 := time.Date(2019, time.June, 14, 12, 30, 15, 0, time.UTC)
// 	r1 := situation.HistoryRecord{
// 		ID: id,
// 		TS: ts1,
// 		FactsIDS: []situation.FactHistoryID{
// 			situation.FactHistoryID{
// 				ID: 1,
// 				TS: &ts1,
// 			},
// 		},
// 	}

// 	err := situation.Persist(r1, false)
// 	if err != nil {
// 		t.Error(err)
// 	}

// 	exp1, _ := expressions.New("3 + 4")
// 	exp2, _ := expressions.New("(3 + 4) * 2")

// 	definitions := []taskdef.Definition{
// 		&taskdef.SetTaskDefinition{
// 			Key:   "key1",
// 			Value: exp1,
// 		},
// 		&taskdef.SetTaskDefinition{
// 			Key:   "key2",
// 			Value: exp2,
// 		},
// 	}

// 	tasks := InstantiateTasks(definitions, knowledge.NewBase())

// 	if len(tasks) != 1 {
// 		t.Error("The number of Tasks is not as expected")
// 	} else {
// 		tasks[0].Perform(
// 			"key-1",
// 			models.InputTask{
// 				SituationID: id,
// 				TS:          ts1,
// 				Rule: models.RuleInput{
// 					RuleID:      1,
// 					RuleVersion: 0,
// 					CasesInput: []models.CaseInput{
// 						models.CaseInput{
// 							Name:      "case1",
// 							Condition: "true",
// 							State:     models.Met,
// 						},
// 					},
// 				},
// 			},
// 		)

// 		getMetaDatas, _ := situation.GetHistoryMetadata(id, ts1)
// 		if len(getMetaDatas) != 1 {
// 			t.Error("The number of MetaDatas is not as expected")
// 		} else {
// 			tasks[0].Perform(
// 				"key-2",
// 				models.InputTask{
// 					SituationID: id,
// 					TS:          ts1,
// 					Rule: models.RuleInput{
// 						RuleID:      2,
// 						RuleVersion: 0,
// 						CasesInput: []models.CaseInput{
// 							models.CaseInput{
// 								Name:      "case1",
// 								Condition: "true",
// 								State:     models.Met,
// 							},
// 						},
// 					},
// 				},
// 			)
// 			getMetaDatas, _ = situation.GetHistoryMetadata(id, ts1)
// 			if len(getMetaDatas) != 2 {
// 				t.Error("The number of MetaDatas is not as expected")
// 			}
// 		}
// 	}
// }

func TestIssueTasks(t *testing.T) {

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
	issues.ReplaceGlobals(issues.NewPostgresRepository(db))

	calendar.ReplaceGlobals(calendar.NewPostgresRepository(db))
	calendar.Init()

	factID, _ := fact.R().Create(engine.Fact{Name: "fact_test_1"})

	s := situation.Situation{Name: "situation_test", Groups: []int64{0}, Facts: []int64{factID}}
	sID, err := situation.R().Create(s)
	if err != nil {
		t.Error(err)
	}

	ruleID, _ := rule.R().Create(rule.Rule{
		Name: "rule_test",
		DefaultRule: ruleeng.DefaultRule{Cases: []ruleeng.Case{
			{
				Name:      "case1",
				Condition: ruleeng.Expression("fact_test_1.aggs.agg0.value > 5"),
				Actions: []ruleeng.ActionDef{{
					Name: ruleeng.Expression(`"create-issue"`),
					Parameters: map[string]ruleeng.Expression{
						"id":             ruleeng.Expression(`"task1"`),
						"name":           ruleeng.Expression(`"Issue_test"`),
						"level":          ruleeng.Expression(`"info"`),
						"timeout":        ruleeng.Expression(`"10s"`),
						"isNotification": ruleeng.Expression(`false`),
					},
				}},
			},
			{
				Name:      "case2",
				Condition: ruleeng.Expression("fact_test_1.aggs.agg0.value <= 5"),
				Actions: []ruleeng.ActionDef{{
					Name: ruleeng.Expression(`"close-today-issues"`),
					Parameters: map[string]ruleeng.Expression{
						"id": ruleeng.Expression(`"task1"`),
					},
				}},
			},
		}},
		Enabled: true,
	})

	situation.R().SetRules(sID, []int64{int64(ruleID)})

	ReplaceGlobals(NewTasker())
	T().StartBatchProcessor()
	defer T().StopBatchProcessor()

	t1 := time.Now()
	situations := []evaluator.SituationToEvaluate{
		{ID: sID, TS: t1},
	}
	item := &reader.Item{
		Key:         "fact_test_1",
		KeyAsString: "fact_test_1",
		Aggs: map[string]*reader.ItemAgg{
			"agg0":      {Value: 10},
			"doc_count": {Value: 1},
		},
		Buckets: nil,
	}
	sh := situation.HistoryRecord{
		ID:       sID,
		TS:       t1,
		FactsIDS: map[int64]*time.Time{factID: &t1},
	}
	fact.PersistFactResult(1, t1, 0, 0, item, true)
	situation.Persist(sh, false)

	evaluations, _ := evaluator.EvaluateSituations(situations, "standart")
	taskBatchs := make([]TaskBatch, 0)
	for _, evaluation := range evaluations {
		taskBatchs = append(taskBatchs, TaskBatch{
			Context: map[string]interface{}{
				"situationID":        evaluation.ID,
				"ts":                 evaluation.TS,
				"templateInstanceID": evaluation.TemplateInstanceID,
			},
			Agenda: evaluation.Agenda,
		})
	}
	T().BatchReceiver <- taskBatchs
	time.Sleep(100 * time.Millisecond)

	createdIssues, _ := issues.R().GetAll()
	if createdIssues[1].Key != "1-1-task1" || createdIssues[1].State != models.Open {
		t.Errorf("The created Issue is not as expected")
	}

	tasks := evaluations[0].Agenda
	if len(tasks) != 1 {
		t.Errorf("The number of tasks is not as expected")
	}
	if tasks[0].GetName() != "create-issue" {
		t.Errorf("The task is not as expected")
	}

	t1 = time.Now()
	situations = []evaluator.SituationToEvaluate{
		{ID: sID, TS: t1},
	}
	item = &reader.Item{
		Key:         "fact_test_1",
		KeyAsString: "fact_test_1",
		Aggs: map[string]*reader.ItemAgg{
			"agg0":      {Value: 2},
			"doc_count": {Value: 1},
		},
		Buckets: nil,
	}
	sh = situation.HistoryRecord{
		ID:       sID,
		TS:       t1,
		FactsIDS: map[int64]*time.Time{factID: &t1},
	}
	fact.PersistFactResult(1, t1, 0, 0, item, true)
	situation.Persist(sh, false)

	evaluations, _ = evaluator.EvaluateSituations(situations, "standart")
	for _, evaluation := range evaluations {
		taskBatchs = append(taskBatchs, TaskBatch{
			Context: map[string]interface{}{
				"situationID":        evaluation.ID,
				"ts":                 evaluation.TS,
				"templateInstanceID": evaluation.TemplateInstanceID,
			},
			Agenda: evaluation.Agenda,
		})
	}
	T().BatchReceiver <- taskBatchs
	time.Sleep(100 * time.Millisecond)

	tasks = evaluations[0].Agenda
	if len(tasks) != 1 {
		t.Errorf("The number of tasks is not as expected")
	}
	if tasks[0].GetName() != "close-today-issues" {
		t.Errorf("The task is not as expected")
	}

	createdIssues, _ = issues.R().GetAll()
	if createdIssues[1].Key != "1-1-task1" || createdIssues[1].State != models.ClosedDiscard {
		t.Errorf("The created Issue is not as expected")
	}
}

func TestTimezoneInCloseIssueTask(t *testing.T) {

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
	issues.ReplaceGlobals(issues.NewPostgresRepository(db))

	calendar.ReplaceGlobals(calendar.NewPostgresRepository(db))
	calendar.Init()

	factID, _ := fact.R().Create(engine.Fact{Name: "fact_test_1"})

	s := situation.Situation{Name: "situation_test", Groups: []int64{0}, Facts: []int64{factID}}
	sID, err := situation.R().Create(s)
	if err != nil {
		t.Error(err)
	}

	rule1 := rule.Rule{
		Name: "rule_test",
		DefaultRule: ruleeng.DefaultRule{Cases: []ruleeng.Case{
			{
				Name:      "case1",
				Condition: ruleeng.Expression("fact_test_1.aggs.agg0.value > 5"),
				Actions: []ruleeng.ActionDef{{
					Name: ruleeng.Expression(`"create-issue"`),
					Parameters: map[string]ruleeng.Expression{
						"id":             ruleeng.Expression(`"task1"`),
						"name":           ruleeng.Expression(`"Issue_test"`),
						"level":          ruleeng.Expression(`"info"`),
						"timeout":        ruleeng.Expression(`"10s"`),
						"isNotification": ruleeng.Expression(`false`),
					},
				}},
			},
			{
				Name:      "case2",
				Condition: ruleeng.Expression("fact_test_1.aggs.agg0.value <= 5"),
				Actions: []ruleeng.ActionDef{{
					Name: ruleeng.Expression(`"close-today-issues"`),
					Parameters: map[string]ruleeng.Expression{
						"id": ruleeng.Expression(`"task1"`),
					},
				}},
			},
		}},
		Enabled: true,
	}
	ruleID, _ := rule.R().Create(rule1)
	rule1.ID = ruleID

	situation.R().SetRules(sID, []int64{int64(ruleID)})

	ReplaceGlobals(NewTasker())
	T().StartBatchProcessor()
	defer T().StopBatchProcessor()

	loc, _ := time.LoadLocation("Europe/Paris")
	now := time.Now().In(loc)
	t1 := time.Date(now.Year(), now.Month(), now.Day(), 0, 10, 0, 0, loc)
	situations := []evaluator.SituationToEvaluate{
		{ID: sID, TS: t1},
	}
	item := &reader.Item{
		Key:         "fact_test_1",
		KeyAsString: "fact_test_1",
		Aggs: map[string]*reader.ItemAgg{
			"agg0":      {Value: 10},
			"doc_count": {Value: 1},
		},
		Buckets: nil,
	}
	sh := situation.HistoryRecord{
		ID:       sID,
		TS:       t1,
		FactsIDS: map[int64]*time.Time{factID: &t1},
	}
	fact.PersistFactResult(1, t1, 0, 0, item, true)
	situation.Persist(sh, false)

	evaluations, _ := evaluator.EvaluateSituations(situations, "standart")
	taskBatchs := make([]TaskBatch, 0)
	for _, evaluation := range evaluations {
		taskBatchs = append(taskBatchs, TaskBatch{
			Context: map[string]interface{}{
				"situationID":        evaluation.ID,
				"ts":                 evaluation.TS,
				"templateInstanceID": evaluation.TemplateInstanceID,
			},
			Agenda: evaluation.Agenda,
		})
	}
	T().BatchReceiver <- taskBatchs
	time.Sleep(100 * time.Millisecond)

	createdIssues, _ := issues.R().GetAll()
	if createdIssues[1].Key != "1-1-task1" || createdIssues[1].State != models.Open {
		t.Errorf("The created Issue is not as expected")
	}

	tasks := evaluations[0].Agenda
	if len(tasks) != 1 {
		t.Errorf("The number of tasks is not as expected")
	}
	if tasks[0].GetName() != "create-issue" {
		t.Errorf("The task is not as expected")
	}

	tests.DBExec(db, `UPDATE issues_v1 SET created_at = '`+t1.UTC().Format("2006-01-02 15:04:05.000Z")+`' WHERE id = 1`, t, true)

	t1 = time.Now()
	situations = []evaluator.SituationToEvaluate{
		{ID: sID, TS: t1},
	}
	item = &reader.Item{
		Key:         "fact_test_1",
		KeyAsString: "fact_test_1",
		Aggs: map[string]*reader.ItemAgg{
			"agg0":      {Value: 2},
			"doc_count": {Value: 1},
		},
		Buckets: nil,
	}
	sh = situation.HistoryRecord{
		ID:       sID,
		TS:       t1,
		FactsIDS: map[int64]*time.Time{factID: &t1},
	}
	fact.PersistFactResult(1, t1, 0, 0, item, true)
	situation.Persist(sh, false)

	evaluations, _ = evaluator.EvaluateSituations(situations, "standart")
	taskBatchs = make([]TaskBatch, 0)
	for _, evaluation := range evaluations {
		taskBatchs = append(taskBatchs, TaskBatch{
			Context: map[string]interface{}{
				"situationID":        evaluation.ID,
				"ts":                 evaluation.TS,
				"templateInstanceID": evaluation.TemplateInstanceID,
			},
			Agenda: evaluation.Agenda,
		})
	}
	T().BatchReceiver <- taskBatchs
	time.Sleep(100 * time.Millisecond)

	tasks = evaluations[0].Agenda
	if len(tasks) != 1 {
		t.Errorf("The number of tasks is not as expected")
	}
	if tasks[0].GetName() != "close-today-issues" {
		t.Errorf("The task is not as expected")
	}

	createdIssues, _ = issues.R().GetAll()
	if createdIssues[1].Key != "1-1-task1" || createdIssues[1].State != models.Open {
		t.Errorf("The created Issue is not as expected")
	}

	rule1.Cases[1].Actions[0].Parameters["timezone"] = "`Europe/Paris`"
	_ = rule.R().Update(rule1)

	evaluations, _ = evaluator.EvaluateSituations(situations, "standart")
	taskBatchs = make([]TaskBatch, 0)
	for _, evaluation := range evaluations {
		taskBatchs = append(taskBatchs, TaskBatch{
			Context: map[string]interface{}{
				"situationID":        evaluation.ID,
				"ts":                 evaluation.TS,
				"templateInstanceID": evaluation.TemplateInstanceID,
			},
			Agenda: evaluation.Agenda,
		})
	}
	T().BatchReceiver <- taskBatchs
	time.Sleep(100 * time.Millisecond)

	tasks = evaluations[0].Agenda
	if len(tasks) != 1 {
		t.Errorf("The number of tasks is not as expected")
	}
	if tasks[0].GetName() != "close-today-issues" {
		t.Errorf("The task is not as expected")
	}

	createdIssues, _ = issues.R().GetAll()
	if createdIssues[1].Key != "1-1-task1" || createdIssues[1].State != models.ClosedDiscard {
		t.Errorf("The created Issue is not as expected")
	}
}
