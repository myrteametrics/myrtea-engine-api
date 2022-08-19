package roles

import (
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/security/permissions"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/tests"
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

// func TestPostgresGet(t *testing.T) {
// 	if testing.Short() {
// 		t.Skip("skipping postgresql test in short mode")
// 	}
// 	db := tests.DBClient(t)
// 	defer dbDestroy(db, t)
// 	dbInit(db, t)
// 	r := NewPostgresRepository(db)

// 	permission, found, err := r.Get(uuid.New())
// 	t.Log(permission)
// 	t.Log(found)
// 	t.Log(err)
// }

// func TestPostgresCreate(t *testing.T) {
// 	if testing.Short() {
// 		t.Skip("skipping postgresql test in short mode")
// 	}
// 	db := tests.DBClient(t)
// 	defer dbDestroy(db, t)
// 	dbInit(db, t)
// 	r := NewPostgresRepository(db)

// 	newUUID, err := r.Create(Permission{ResourceType: "a", ResourceID: "b", Action: "c"})
// 	t.Log(newUUID)
// 	t.Log(err)

// 	permission, found, err := r.Get(newUUID)
// 	t.Log(permission)
// 	t.Log(found)
// 	t.Log(err)
// }

func TestSetRolePermissions(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroy(db, t)
	dbInit(db, t)
	r := NewPostgresRepository(db)
	rp := permissions.NewPostgresRepository(db)

	roleUUID, _ := r.Create(Role{})
	permissionUUID1, _ := rp.Create(permissions.Permission{})
	permissionUUID2, _ := rp.Create(permissions.Permission{})
	permissionUUID3, _ := rp.Create(permissions.Permission{})

	err := r.SetRolePermissions(roleUUID, []uuid.UUID{permissionUUID1, permissionUUID2, permissionUUID3})
	t.Log(err)
}
