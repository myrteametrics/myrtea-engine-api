package users

import (
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/tests"
)

func dbInit(dbClient *sqlx.DB, t *testing.T) {
	dbDestroy(dbClient, t)
	tests.DBExec(dbClient, `CREATE EXTENSION IF NOT EXISTS pgcrypto;`, t, true)
	tests.DBExec(dbClient, tests.UsersTableV1, t, true)
	tests.DBExec(dbClient, tests.GroupsTableV1, t, true)
	tests.DBExec(dbClient, tests.UserMembershipsTableV1, t, true)
}
func dbDestroy(dbClient *sqlx.DB, t *testing.T) {
	tests.DBExec(dbClient, tests.UserMembershipsDropTableV1, t, false)
	tests.DBExec(dbClient, tests.GroupsDropTableV1, t, false)
	tests.DBExec(dbClient, tests.UsersDropTableV1, t, false)
}

func TestCreate_Get(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroy(db, t)
	dbInit(db, t)

	r := NewPostgresRepository(db)
	_ = r
}
