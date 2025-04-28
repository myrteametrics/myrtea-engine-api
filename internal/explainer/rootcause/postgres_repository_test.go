package rootcause

import (
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/situation"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/models"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/rule"
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

	var err error

	sr := situation.NewPostgresRepository(dbClient)
	s := situation.Situation{
		Name:  "situation_test_1",
		Facts: []int64{},
	}
	_, err = sr.Create(s)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	rr := rule.NewPostgresRepository(dbClient)
	r := rule.Rule{
		Name: "rule_test_1",
	}
	_, err = rr.Create(r)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

}

func dbDestroy(dbClient *sqlx.DB, t *testing.T) {
	tests.DBExec(dbClient, tests.RefRootCauseDropTableV1, t, false)
	tests.DBExec(dbClient, tests.RuleVersionsDropTableV1, t, true)
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
		t.Error("rc Repository is nil")
	}
}

func TestPostgresReplaceGlobal(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	r := NewPostgresRepository(tests.DBClient(t))
	reverse := ReplaceGlobals(r)
	if R() == nil {
		t.Error("Global rc repository is nil")
	}
	reverse()
	if R() != nil {
		t.Error("Global rc repository is not nil after reverse")
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
	_, found, err := r.Get(1)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if found {
		t.Error("found a rc from nowhere")
	}

	rc := models.NewRootCause(0, "rc_name_1", "rc_desc_1", 1, 1)
	id, err := r.Create(nil, rc)
	if err != nil {
		t.Error(err)
	}

	rcGet, found, err := r.Get(1)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if !found {
		t.Error("rootcause not found")
	}
	if id != rcGet.ID {
		t.Error("invalid rc ID")
	}
	if rc.Name != rcGet.Name {
		t.Error("invalid rc Name")
	}
	if rc.Description != rcGet.Description {
		t.Error("invalid rc Description")
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
	rc := models.NewRootCause(0, "rc_name_1", "rc_desc_1", 1, 1)
	id, err := r.Create(nil, rc)
	if err != nil {
		t.Error(err)
	}
	rcGet, found, err := r.Get(1)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if !found {
		t.Error("rootcause not found")
	}
	if id != rcGet.ID {
		t.Error("invalid rc ID")
	}
	if rc.Name != rcGet.Name {
		t.Error("invalid rc Name")
	}
	if rc.Description != rcGet.Description {
		t.Error("invalid rc Description")
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
	rc := models.NewRootCause(0, "rc_name_1", "rc_desc_1", 1, 1)
	id, err := r.Create(nil, rc)
	if err != nil {
		t.Error(err)
	}
	rc2 := models.NewRootCause(0, "rc_name_1", "rc_desc_2", 1, 1)
	_, err = r.Create(nil, rc2)
	if err == nil {
		t.Error("Create should not be created")
	}
	rcGet, found, err := r.Get(id)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if !found {
		t.Error("rootcause not found")
	}
	if rc.Description != rcGet.Description {
		t.Error("rc has been updated while it must not")
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
	rc1 := models.NewRootCause(0, "rc_name_1", "rc_desc_1", 1, 1)
	id, err := r.Create(nil, rc1)
	if err != nil {
		t.Error(err)
	}

	// Update existing
	rc2 := models.NewRootCause(0, "rc_name_2", "rc_desc_2", 1, 1)
	err = r.Update(nil, id, rc2)
	if err != nil {
		t.Error(err)
	}
	rcGet, found, err := r.Get(1)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if !found {
		t.Error("rootcause not found")
	}
	if id != rcGet.ID {
		t.Error("invalid rc ID")
	}
	if rc2.Name != rcGet.Name {
		t.Error("invalid rc ID")
	}
	if rc2.Description != rcGet.Description {
		t.Error("invalid rc Description")
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
	rc := models.NewRootCause(0, "rc_name_1", "rc_desc_1", 1, 1)
	err = r.Update(nil, 1, rc)
	if err == nil {
		t.Error("updating a non-existing rc should return an error")
	}
	_, found, err := r.Get(1)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if found {
		t.Error("rc should not exists")
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
	rc := models.NewRootCause(0, "rc_name_1", "rc_desc_1", 1, 1)
	id, err := r.Create(nil, rc)
	if err != nil {
		t.Error(err)
	}

	err = r.Delete(nil, id)
	if err != nil {
		t.Error(err)
	}

	_, found, err := r.Get(1)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if found {
		t.Error("rc should not exists")
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
		t.Error("Cannot delete a non-existing rc")
	}

	_, found, err := r.Get(1)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if found {
		t.Error("rc should not exists")
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
	rc := models.NewRootCause(0, "rc_name_1", "rc_desc_1", 1, 1)
	_, err = r.Create(nil, rc)
	if err != nil {
		t.Error(err)
	}
	rc2 := models.NewRootCause(0, "rc_name_2", "rc_desc_2", 1, 1)
	_, err = r.Create(nil, rc2)
	if err != nil {
		t.Error(err)
	}

	rootCauses, err := r.GetAllBySituationID(1)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if rootCauses == nil {
		t.Error("rootcauses map should not be nil")
		t.FailNow()
	}
	if len(rootCauses) == 0 {
		t.Error("No rootcauses returned when it should return 1 rootcause")
		t.FailNow()
	}
	if _, exists := rootCauses[1]; !exists {
		t.Error("Action with id 1 not found")
	}
	if rc.Name != rootCauses[1].Name {
		t.Error("Invalid rootcause name")
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
	rc := models.NewRootCause(0, "rc_name_1", "rc_desc_1", 1, 1)
	_, err = r.Create(nil, rc)
	if err != nil {
		t.Error(err)
	}
	rc2 := models.NewRootCause(0, "rc_name_2", "rc_desc_2", 1, 1)
	_, err = r.Create(nil, rc2)
	if err != nil {
		t.Error(err)
	}

	rootCauses, err := r.GetAllBySituationID(99)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if rootCauses == nil {
		t.Error("rootcauses map should not be nil")
		t.FailNow()
	}
	if len(rootCauses) > 0 {
		t.Error("No rootcauses returned when it should return 1 rootcause")
	}
}
