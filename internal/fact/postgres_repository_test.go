package fact

import (
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/tests"
	"github.com/myrteametrics/myrtea-sdk/v5/engine"
)

func dbInit(dbClient *sqlx.DB, t *testing.T) {
	dbDestroy(dbClient, t)

	tests.DBExec(dbClient, tests.FactDefinitionTableV1, t, true)
	tests.DBExec(dbClient, tests.FactHistoryTableV1, t, true)
}

func dbDestroy(dbClient *sqlx.DB, t *testing.T) {
	tests.DBExec(dbClient, tests.FactHistoryDropTableV1, t, false)
	tests.DBExec(dbClient, tests.FactDefinitionDropTableV1, t, false)
}

func TestNewPostgresRepository(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	r := NewPostgresRepository(tests.DBClient(t))
	if r == nil {
		t.Error("Fact Repository is nil")
	}
}

func TestPostgresReplaceGlobal(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	r := NewPostgresRepository(tests.DBClient(t))
	reverse := ReplaceGlobals(r)
	if R() == nil {
		t.Error("Global fact repository is nil")
	}
	reverse()
	if R() != nil {
		t.Error("Global fact repository is not nil after reverse")
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
	factGet, found, err := r.Get(1)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if found {
		t.Error("found a fact from nowhere")
		t.FailNow()
	}

	fact := engine.Fact{Name: "test_name", Comment: "test comment"}
	id, err := r.Create(fact)
	if err != nil {
		t.Error(err)
	}

	factGet, found, err = r.Get(id)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if !found {
		t.Error("fact should be found")
		t.FailNow()
	}
	if id != factGet.ID {
		t.Error("invalid fact ID")
	}
	if fact.Name != factGet.Name {
		t.Error("invalid fact Name")
	}
	if fact.Comment != factGet.Comment {
		t.Error("invalid fact Comment")
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
	fact := engine.Fact{Name: "test_name", Comment: "test comment"}
	id, err := r.Create(fact)
	if err != nil {
		t.Error(err)
	}
	factGet, found, err := r.Get(id)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if !found {
		t.Error("fact should be found")
		t.FailNow()
	}
	if id != factGet.ID {
		t.Error("invalid fact ID")
	}
	if fact.Name != factGet.Name {
		t.Error("invalid fact Name")
	}
	if fact.Comment != factGet.Comment {
		t.Error("invalid fact Comment")
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
	fact := engine.Fact{Name: "test_name", Comment: "test comment"}
	id, err := r.Create(fact)
	if err != nil {
		t.Error(err)
	}
	fact2 := engine.Fact{Name: "test_name", Comment: "test comment 2"}
	_, err = r.Create(fact2)
	if err == nil {
		t.Error("Create should not be created")
	}
	factGet, found, err := r.Get(id)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if !found {
		t.Error("fact should be found")
		t.FailNow()
	}
	if fact.Comment != factGet.Comment {
		t.Error("Fact has been updated while it must not")
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
	id, err := r.Create(engine.Fact{Name: "test_name", Comment: "test comment"})
	if err != nil {
		t.Error(err)
	}

	// Update existing
	fact := engine.Fact{Name: "test_name 2", Comment: "test comment 2"}
	err = r.Update(id, fact)
	if err != nil {
		t.Error(err)
	}
	factGet, found, err := r.Get(id)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if !found {
		t.Error("fact should be found")
		t.FailNow()
	}
	if id != factGet.ID {
		t.Error("invalid fact ID")
	}
	if fact.Name != factGet.Name {
		t.Error("invalid fact ID")
	}
	if fact.Comment != factGet.Comment {
		t.Error("invalid fact Comment")
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
	fact := engine.Fact{Name: "test_name", Comment: "test comment 2"}
	err = r.Update(1, fact)
	if err == nil {
		t.Error("updating a non-existing fact should return an error")
	}
	_, found, err := r.Get(1)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if found {
		t.Error("fact should not exists")
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
	id, err := r.Create(engine.Fact{Name: "test_name", Comment: "test comment"})
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
		t.Error("fact should not exists")
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
		t.Error("Cannot delete a non-existing fact")
	}

	_, found, err := r.Get(1)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if found {
		t.Error("fact should not exists")
		t.FailNow()
	}
}
