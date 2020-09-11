package modeler

import (
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/tests"
	"github.com/myrteametrics/myrtea-sdk/v4/modeler"
)

func dbInit(dbClient *sqlx.DB, t *testing.T) {
	dbDestroy(dbClient, t)
	tests.DBExec(dbClient, tests.ModelTableV1, t, true)
}

func dbDestroy(dbClient *sqlx.DB, t *testing.T) {
	tests.DBExec(dbClient, tests.ModelDropTableV1, t, false)
}

func TestNewPostgresRepository(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	r := NewPostgresRepository(tests.DBClient(t))
	if r == nil {
		t.Error("Model Repository is nil")
	}
}

func TestPostgresReplaceGlobal(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	r := NewPostgresRepository(tests.DBClient(t))
	reverse := ReplaceGlobals(r)
	if R() == nil {
		t.Error("Global model repository is nil")
	}
	reverse()
	if R() != nil {
		t.Error("Global model repository is not nil after reverse")
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
	modelGet, found, err := r.Get(1)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if found {
		t.Error("found a model from nowhere")
		t.FailNow()
	}

	model := modeler.Model{Name: "test_name", Synonyms: []string{"model_synonym"}, Fields: []modeler.Field{}}
	id, err := r.Create(model)
	if err != nil {
		t.Error(err)
	}

	modelGet, found, err = r.Get(id)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if !found {
		t.Error("model should be found")
		t.FailNow()
	}
	if id != modelGet.ID {
		t.Error("invalid model ID")
	}
	if model.Name != modelGet.Name {
		t.Error("invalid model Name")
	}
	if model.Synonyms[0] != modelGet.Synonyms[0] {
		t.Error("invalid model Synonyms")
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
	model := modeler.Model{Name: "test_name", Synonyms: []string{"model_synonym"}, Fields: []modeler.Field{}}
	id, err := r.Create(model)
	if err != nil {
		t.Error(err)
	}
	modelGet, found, err := r.Get(id)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if !found {
		t.Error("model should be found")
		t.FailNow()
	}
	if id != modelGet.ID {
		t.Error("invalid model ID")
	}
	if model.Name != modelGet.Name {
		t.Error("invalid model Name")
	}
	if model.Synonyms[0] != modelGet.Synonyms[0] {
		t.Error("invalid model Synonyms")
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
	model := modeler.Model{Name: "test_name", Synonyms: []string{"model_synonym"}, Fields: []modeler.Field{}}
	id, err := r.Create(model)
	if err != nil {
		t.Error(err)
	}
	model2 := modeler.Model{Name: "test_name", Synonyms: []string{"model_synonym"}, Fields: []modeler.Field{}}
	_, err = r.Create(model2)
	if err == nil {
		t.Error("Create should not be created")
	}
	modelGet, found, err := r.Get(id)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if !found {
		t.Error("model should be found")
		t.FailNow()
	}
	if model.Synonyms[0] != modelGet.Synonyms[0] {
		t.Error("Model has been updated while it must not")
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
	id, err := r.Create(modeler.Model{Name: "test_name", Synonyms: []string{"model_synonym"}, Fields: []modeler.Field{}})
	if err != nil {
		t.Error(err)
	}

	// Update existing
	model := modeler.Model{Name: "test_name 2", Synonyms: []string{"model_synonym"}, Fields: []modeler.Field{}}
	err = r.Update(id, model)
	if err != nil {
		t.Error(err)
	}
	modelGet, found, err := r.Get(id)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if !found {
		t.Error("model should be found")
		t.FailNow()
	}
	if id != modelGet.ID {
		t.Error("invalid model ID")
	}
	if model.Name != modelGet.Name {
		t.Error("invalid model ID")
	}
	if model.Synonyms[0] != modelGet.Synonyms[0] {
		t.Error("invalid model Synonyms")
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
	model := modeler.Model{Name: "test_name", Synonyms: []string{"model_synonym"}, Fields: []modeler.Field{}}
	err = r.Update(1, model)
	if err == nil {
		t.Error("updating a non-existing model should return an error")
	}
	_, found, err := r.Get(1)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if found {
		t.Error("model should not exists")
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
	id, err := r.Create(modeler.Model{Name: "test_name", Synonyms: []string{"model_synonym"}, Fields: []modeler.Field{}})
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
		t.Error("model should not exists")
		t.FailNow()
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

	err := r.Delete(1)
	if err == nil {
		t.Error("Cannot delete a non-existing model")
	}

	_, found, err := r.Get(1)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if found {
		t.Error("model should not exists")
		t.FailNow()
	}
}
