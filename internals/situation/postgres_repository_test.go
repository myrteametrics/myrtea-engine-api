package situation

import (
	"fmt"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/fact"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/tests"
	"github.com/myrteametrics/myrtea-sdk/v4/engine"
)

func dbInitRepo(dbClient *sqlx.DB, t *testing.T) {
	dbDestroyRepo(dbClient, t)

	tests.DBExec(dbClient, tests.CalendarTableV3, t, true)

	_, err := dbClient.Exec(tests.SituationDefinitionTableV1)
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
	_, err = dbClient.Exec(tests.RulesTableV1)
	if err != nil {
		t.Error(err)
	}
	_, err = dbClient.Exec(tests.RuleVersionsTableV1)
	if err != nil {
		t.Error(err)
	}
	_, err = dbClient.Exec(tests.SituationRulesTableV1)
	if err != nil {
		t.Error(err)
	}
}

func dbDestroyRepo(dbClient *sqlx.DB, t *testing.T) {
	_, err := dbClient.Exec(tests.SituationRulesDropTableV1)
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

	_, err = dbClient.Exec(tests.RuleVersionsDropTableV1)
	if err != nil {
		t.Error(err)
	}
	_, err = dbClient.Exec(tests.RulesDropTableV1)
	if err != nil {
		t.Error(err)
	}

	tests.DBExec(dbClient, tests.CalendarDropTableV3, t, true)
}

func insertRule(dbClient *sqlx.DB, t *testing.T, ruleID int64) {
	dt := time.Now().Truncate(1 * time.Millisecond).UTC()
	query := `INSERT INTO rules_v1 VALUES (:id, :name, :enabled, :calendar_id, :last_modified)`
	_, err := dbClient.NamedExec(query, map[string]interface{}{
		"id":            ruleID,
		"name":          "rule" + fmt.Sprint(ruleID),
		"enabled":       true,
		"calendar_id":   nil,
		"last_modified": dt,
	})
	if err != nil {
		t.Error(err)
	}

	query = `INSERT INTO rule_versions_v1 VALUES(:rule_id, 0, '{}', :creation_datetime)`
	_, err = dbClient.NamedExec(query, map[string]interface{}{
		"id":                ruleID,
		"rule_id":           ruleID,
		"creation_datetime": dt,
	})

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
		t.Error("situation Repository is nil")
	}
}

func TestPostgresReplaceGlobal(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	r := NewPostgresRepository(tests.DBClient(t))
	reverse := ReplaceGlobals(r)
	if R() == nil {
		t.Error("Global situation repository is nil")
	}
	reverse()
	if R() != nil {
		t.Error("Global situation repository is not nil after reverse")
	}
}

func TestPostgresGet(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroyRepo(db, t)
	dbInitRepo(db, t)
	r := NewPostgresRepository(db)

	var err error

	situationGet, found, err := r.Get(1)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if found {
		t.Error("found a situation from nowhere")
	}

	situation := Situation{Name: "test_name"}
	id, err := r.Create(situation)
	if err != nil {
		t.Error(err)
	}

	situationGet, found, err = r.Get(id)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if !found {
		t.Error("Situation doesn't exists after the creation")
		t.FailNow()
	}
	if id != situationGet.ID {
		t.Error("invalid Situation ID")
	}
}

func TestPostgresCreate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroyRepo(db, t)
	dbInitRepo(db, t)
	r := NewPostgresRepository(db)
	factR := fact.NewPostgresRepository(db)

	fact1 := engine.Fact{Name: "test_name1", Comment: "test comment"}
	factID1, err := factR.Create(fact1)
	fact2 := engine.Fact{Name: "test_name2", Comment: "test comment"}
	factID2, err := factR.Create(fact2)

	params := map[string]string{"key": "value"}

	situation := Situation{Name: "test", Facts: []int64{factID1, factID2}, Parameters: params}
	id, err := r.Create(situation)
	if err != nil {
		t.Error(err)
	}
	situationGet, found, err := r.Get(id)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if !found {
		t.Error("situation not found")
		t.FailNow()
	}
	if situationGet.ID != id {
		t.Error("invalid situation value")
	}
	if situationGet.Facts[0] != factID1 || situationGet.Facts[1] != factID2 {
		t.Error("The situation facts are not as expected")
	}
	if situationGet.Parameters["key"] != "value" {
		t.Error("The situation parameters are not as expected")
	}

}

func TestPostgresCreateMultiple(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroyRepo(db, t)
	dbInitRepo(db, t)
	r := NewPostgresRepository(db)

	var err error
	situation := Situation{Name: "test"}
	id1, err := r.Create(situation)
	if err != nil {
		t.Error(err)
	}
	situationGet, found, err := r.Get(id1)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if !found {
		t.Error("situation not found")
		t.FailNow()
	}
	if id1 != situationGet.ID {
		t.Error("invalid ID")
	}

	situation2 := Situation{Name: "test2"}
	id2, err := r.Create(situation2)
	if err != nil {
		t.Error(err)
	}
	situation2Get, found, err := r.Get(id2)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if !found {
		t.Error("situation not found")
		t.FailNow()
	}
	if situation2Get.ID != id2 {
		t.Error("invalid ID")
	}
}

func TestPostgresUpdate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroyRepo(db, t)
	dbInitRepo(db, t)
	situationR := NewPostgresRepository(db)
	factR := fact.NewPostgresRepository(db)

	var err error

	fact1 := engine.Fact{Name: "test_name1", Comment: "test comment"}
	id1, err := factR.Create(fact1)
	fact2 := engine.Fact{Name: "test_name2", Comment: "test comment"}
	id2, err := factR.Create(fact2)

	situation := Situation{Name: "test1", Facts: []int64{id1}}
	id, err := situationR.Create(situation)
	if err != nil {
		t.Error(err)
	}
	situation2 := Situation{Name: "test2", Facts: []int64{id2}}
	err = situationR.Update(id, situation2)
	if err != nil {
		t.Error(err)
	}
	situationGet, found, err := situationR.Get(id)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if !found {
		t.Error("situation not found")
		t.FailNow()
	}
	if situationGet.Name != "test2" {
		t.Error("Couldn't update the situation")
	}
	if situationGet.Facts[0] != 2 {
		t.Error("Couldn't update the situation")
	}
}

func TestPostgresUpdateNotExists(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroyRepo(db, t)
	dbInitRepo(db, t)
	r := NewPostgresRepository(db)

	var err error
	situation := Situation{Name: "test"}
	err = r.Update(1, situation)
	if err == nil {
		t.Error("updating a non-existing situation should return an error")
	}
}

func TestPostgresDelete(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroyRepo(db, t)
	dbInitRepo(db, t)
	r := NewPostgresRepository(db)

	var err error

	situation := Situation{Name: "test"}
	id, err := r.Create(situation)
	if err != nil {
		t.Error(err)
	}

	err = r.Delete(id)
	if err != nil {
		t.Error(err)
	}

	_, found, err := r.Get(id)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if found {
		t.Error("Situation should not exists")
	}
}

func TestPostgresDeleteNotExists(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroyRepo(db, t)
	dbInitRepo(db, t)
	r := NewPostgresRepository(db)

	err := r.Delete(1)
	if err == nil {
		t.Error("Cannot delete a non-existing situation")
	}

	_, found, err := r.Get(1)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if found {
		t.Error("situation should not exists")
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

	factR := fact.NewPostgresRepository(db)

	fact1 := engine.Fact{Name: "test_name1", Comment: "test comment"}
	factID1, err := factR.Create(fact1)
	fact2 := engine.Fact{Name: "test_name2", Comment: "test comment"}
	factID2, err := factR.Create(fact2)

	params := map[string]string{"key": "value"}

	situation := Situation{Name: "test", Facts: []int64{factID1}, Parameters: params}
	s1ID, err := r.Create(situation)
	if err != nil {
		t.Error(err)
	}
	situation2 := Situation{Name: "test2", Facts: []int64{factID2}, Parameters: params}
	s2ID, err := r.Create(situation2)
	if err != nil {
		t.Error(err)
	}

	situations, err := r.GetAll()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if len(situations) != 2 {
		t.Error("wrong situations count")
	}
	if situations[s1ID].Facts[0] != factID1 || situations[s1ID].Parameters["key"] != "value" {
		t.Error("The situation " + fmt.Sprint(s1ID) + " is not as expected")
	}
	if situations[s2ID].Facts[0] != factID2 || situations[s2ID].Parameters["key"] != "value" {
		t.Error("The situation " + fmt.Sprint(s2ID) + " is not as expected")
	}
}

func TestPostgresGetAllEmpty(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroyRepo(db, t)
	dbInitRepo(db, t)
	r := NewPostgresRepository(db)

	situations, err := r.GetAll()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if situations == nil {
		t.Error("situation should not be nil")
	}
	if len(situations) != 0 {
		t.Error("wrong situations count")
	}
}

func TestSetAndGetRules(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroyRepo(db, t)
	dbInitRepo(db, t)
	r := NewPostgresRepository(db)

	situation := Situation{Name: "test"}
	situationID, err := r.Create(situation)
	if err != nil {
		t.Error(err)
	}

	rulesIDs := []int64{1, 2}
	for _, id := range rulesIDs {
		insertRule(db, t, id)
	}
	r.SetRules(situationID, rulesIDs)

	getRuleIDs, err := r.GetRules(situationID)
	if err != nil {
		t.Error(err)
	}
	if len(getRuleIDs) != len(rulesIDs) {
		t.Error("The number of ruleIDs obtained is not as expected")
	} else {
		for i, id := range rulesIDs {
			if getRuleIDs[i] != id {
				t.Error("The list of ruleIDs obtained is not as expected")
			}
		}
	}

}
