package permissions

import (
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/tests"
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

func TestPostgresGet(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroy(db, t)
	dbInit(db, t)
	r := NewPostgresRepository(db)

	permission, found, err := r.Get(uuid.New())
	t.Log(permission)
	t.Log(found)
	t.Log(err)
}

func TestPostgresCreate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroy(db, t)
	dbInit(db, t)
	r := NewPostgresRepository(db)

	newUUID, err := r.Create(Permission{ResourceType: "a", ResourceID: "b", Action: "c"})
	t.Log(newUUID)
	t.Log(err)

	permission, found, err := r.Get(newUUID)
	t.Log(permission)
	t.Log(found)
	t.Log(err)
}

func TestPostgresGetAll(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroy(db, t)
	dbInit(db, t)
	r := NewPostgresRepository(db)

	permissions, err := r.GetAll()
	t.Log(permissions)
	t.Log(err)
}

func TestPostgresGetAllForRole(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroy(db, t)
	dbInit(db, t)
	r := NewPostgresRepository(db)

	permissions, err := r.GetAllForRole(uuid.MustParse("118acbbd-6613-42fa-8b04-b189b7d2adf7"))
	t.Log(permissions)
	t.Log(err)
}
