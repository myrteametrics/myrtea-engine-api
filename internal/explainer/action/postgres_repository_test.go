package action

import (
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/explainer/rootcause"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/models"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/rule"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/situation"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/tests"
)

func dbInit(dbClient *sqlx.DB, t *testing.T) {
	dbDestroy(dbClient, t)

	tests.DBExec(dbClient, tests.CalendarTableV3, t, true)
	tests.DBExec(dbClient, tests.SituationDefinitionTableV1, t, true)
	tests.DBExec(dbClient, tests.FactDefinitionTableV1, t, true)
	tests.DBExec(dbClient, tests.SituationFactsTableV1, t, true)
	tests.DBExec(dbClient, tests.RulesTableV1, t, true)
	tests.DBExec(dbClient, tests.RuleVersionsTableV1, t, true)
	tests.DBExec(dbClient, tests.RefRootCauseTableV1, t, true)
	tests.DBExec(dbClient, tests.RefActionTableV1, t, true)

	sr := situation.NewPostgresRepository(dbClient)
	s1 := situation.Situation{
		Name:  "situation_test_1",
		Facts: []int64{},
	}
	sid1, err := sr.Create(s1)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	s2 := situation.Situation{
		Name:  "situation_test_2",
		Facts: []int64{},
	}
	sid2, err := sr.Create(s2)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	rr := rule.NewPostgresRepository(dbClient)
	r := rule.Rule{
		Name: "rule_test_1",
	}
	ruleID1, err := rr.Create(r)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	rcr := rootcause.NewPostgresRepository(dbClient)
	rc := models.NewRootCause(0, "rc_name_1", "rc_desc_1", sid1, int64(ruleID1))
	_, err = rcr.Create(nil, rc)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	rc2 := models.NewRootCause(0, "rc_name_2", "rc_desc_2", sid2, int64(ruleID1))
	_, err = rcr.Create(nil, rc2)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
}

func dbDestroy(dbClient *sqlx.DB, t *testing.T) {
	tests.DBExec(dbClient, tests.RefActionDropTableV1, t, false)
	tests.DBExec(dbClient, tests.RefRootCauseDropTableV1, t, false)
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

func TestPostgresGet(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroy(db, t)
	dbInit(db, t)
	r := NewPostgresRepository(db)

	var err error
	actionGet, found, err := r.Get(1)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if found {
		t.Error("found an action from nowhere")
	}

	action := models.NewAction(0, "action_name_1", "action_desc_1", 1)
	id, err := r.Create(nil, action)
	if err != nil {
		t.Error(err)
	}

	actionGet, found, err = r.Get(id)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if !found {
		t.Error("action not found")
		t.FailNow()
	}
	if id != actionGet.ID {
		t.Error("invalid action ID")
	}
	if action.Name != actionGet.Name {
		t.Error("invalid action Name")
	}
	if action.Description != actionGet.Description {
		t.Error("invalid action Description")
	}
}

func TestPostgresCreateWithoutID(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroy(db, t)
	dbInit(db, t)
	r := NewPostgresRepository(db)

	var err error
	action := models.NewAction(0, "action_name_1", "action_desc_1", 1)
	id, err := r.Create(nil, action)
	if err != nil {
		t.Error(err)
	}
	actionGet, found, err := r.Get(id)
	if err != nil {
		t.Error(err)
	}
	if !found {
		t.Error("action not found")
	}
	if id != actionGet.ID {
		t.Error("invalid action ID")
	}
	if action.Name != actionGet.Name {
		t.Error("invalid action Name")
	}
	if action.Description != actionGet.Description {
		t.Error("invalid action Description")
	}
}

func TestPostgresCreateIfExists(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroy(db, t)
	dbInit(db, t)
	r := NewPostgresRepository(db)

	var err error
	action := models.NewAction(0, "action_name_1", "action_desc_1", 1)
	id, err := r.Create(nil, action)
	if err != nil {
		t.Error(err)
	}
	action2 := models.NewAction(0, "action_name_1", "action_desc_2", 1)
	_, err = r.Create(nil, action2)
	if err == nil {
		t.Error("Create should not be created")
	}
	actionGet, found, err := r.Get(id)
	if err != nil {
		t.Error(err)
	}
	if !found {
		t.Error("action not found")
	}
	if action.Description != actionGet.Description {
		t.Error("action has been updated while it must not")
	}
}

func TestPostgresUpdate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroy(db, t)
	dbInit(db, t)
	r := NewPostgresRepository(db)

	var err error
	action1 := models.NewAction(0, "action_name_1", "action_desc_1", 1)
	id, err := r.Create(nil, action1)
	if err != nil {
		t.Error(err)
	}

	// Update existing
	action2 := models.NewAction(0, "action_name_2", "action_desc_2", 1)
	err = r.Update(nil, id, action2)
	if err != nil {
		t.Error(err)
	}
	actionGet, found, err := r.Get(id)
	if err != nil {
		t.Error(err)
	}
	if !found {
		t.Error("action not found")
	}
	if id != actionGet.ID {
		t.Error("invalid action ID")
	}
	if action2.Name != actionGet.Name {
		t.Error("invalid action ID")
	}
	if action2.Description != actionGet.Description {
		t.Error("invalid action Description")
	}
}

func TestPostgresUpdateNotExists(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroy(db, t)
	dbInit(db, t)
	r := NewPostgresRepository(db)

	var err error
	action := models.NewAction(0, "action_name_1", "action_desc_1", 1)
	err = r.Update(nil, 1, action)
	if err == nil {
		t.Error("updating a non-existing action should return an error")
	}
	_, found, err := r.Get(1)
	if err != nil {
		t.Error(err)
	}
	if found {
		t.Error("action should not exists")
	}
}

func TestPostgresDelete(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroy(db, t)
	dbInit(db, t)
	r := NewPostgresRepository(db)

	var err error
	action := models.NewAction(0, "action_name_1", "action_desc_1", 1)
	id, err := r.Create(nil, action)
	if err != nil {
		t.Error(err)
	}

	err = r.Delete(nil, id)
	if err != nil {
		t.Error(err)
	}

	_, found, err := r.Get(id)
	if err != nil {
		t.Error(err)
	}
	if found {
		t.Error("action should not exists")
	}
}

func TestPostgresDeleteNotExists(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroy(db, t)
	dbInit(db, t)
	r := NewPostgresRepository(db)

	err := r.Delete(nil, 1)
	if err == nil {
		t.Error("Cannot delete a non-existing action")
	}

	_, found, err := r.Get(1)
	if err != nil {
		t.Error(err)
	}
	if found {
		t.Error("action should not exists")
	}
}

func TestPostgresGetAllByRootCauseID(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroy(db, t)
	dbInit(db, t)
	r := NewPostgresRepository(db)

	var err error
	action := models.NewAction(0, "action_name_1", "action_desc_1", 1)
	_, err = r.Create(nil, action)
	if err != nil {
		t.Error(err)
	}
	action2 := models.NewAction(0, "action_name_2", "action_desc_2", 2)
	_, err = r.Create(nil, action2)
	if err != nil {
		t.Error(err)
	}

	actions, err := r.GetAllByRootCauseID(1)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if actions == nil {
		t.Error("actions map should not be nil")
		t.FailNow()
	}
	if len(actions) == 0 {
		t.Error("No actions returned when it should return 1 action")
		t.FailNow()
	}
	if _, exists := actions[1]; !exists {
		t.Error("Action with id 1 not found")
	}
	if action.Name != actions[1].Name {
		t.Error("Invalid action name")
	}
}

func TestPostgresGetAllByRootCauseIDNotExists(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroy(db, t)
	dbInit(db, t)
	r := NewPostgresRepository(db)

	var err error
	action := models.NewAction(0, "action_name_1", "action_desc_1", 1)
	_, err = r.Create(nil, action)
	if err != nil {
		t.Error(err)
	}
	action2 := models.NewAction(0, "action_name_2", "action_desc_2", 2)
	_, err = r.Create(nil, action2)
	if err != nil {
		t.Error(err)
	}

	actions, err := r.GetAllByRootCauseID(99)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if actions == nil {
		t.Error("actions map should not be nil")
		t.FailNow()
	}
	if len(actions) > 0 {
		t.Error("No actions returned when it should return 1 action")
	}
}

func TestPostgresGetAllBySituationID(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroy(db, t)
	dbInit(db, t)
	r := NewPostgresRepository(db)

	var err error
	action := models.NewAction(0, "action_name_1", "action_desc_1", 1)
	_, err = r.Create(nil, action)
	if err != nil {
		t.Error(err)
	}
	action2 := models.NewAction(0, "action_name_2", "action_desc_2", 2)
	_, err = r.Create(nil, action2)
	if err != nil {
		t.Error(err)
	}

	actions, err := r.GetAllBySituationID(1)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if actions == nil {
		t.Error("actions map should not be nil")
		t.FailNow()
	}
	if len(actions) == 0 {
		t.Error("No actions returned when it should return 1 action")
		t.FailNow()
	}
	if _, exists := actions[1]; !exists {
		t.Error("Action with id 1 not found")
	}
	if action.Name != actions[1].Name {
		t.Error("Invalid action name")
	}
}

func TestPostgresGetAllBySituationIDNotExists(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroy(db, t)
	dbInit(db, t)
	r := NewPostgresRepository(db)

	var err error
	action := models.NewAction(0, "action_name_1", "action_desc_1", 1)
	_, err = r.Create(nil, action)
	if err != nil {
		t.Error(err)
	}
	action2 := models.NewAction(0, "action_name_2", "action_desc_2", 2)
	_, err = r.Create(nil, action2)
	if err != nil {
		t.Error(err)
	}

	actions, err := r.GetAllBySituationID(99)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if actions == nil {
		t.Error("actions map should not be nil")
		t.FailNow()
	}
	if len(actions) > 0 {
		t.Error("No actions returned when it should return 1 action")
	}
}
