package users

import (
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/groups"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/tests"
	"github.com/myrteametrics/myrtea-sdk/v4/security"
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

	usersR := NewPostgresRepository(db)

	user := security.UserWithPassword{
		User: security.User{
			Login: "admin",
			Role:  1,
		},
		Password: "pass",
	}
	userID, err := usersR.Create(user)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	userGet, found, err := usersR.Get(userID)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if !found {
		t.Error("User not found")
		t.FailNow()
	}
	if user.Login != userGet.Login || user.Role != userGet.Role {
		t.Error("The user obtained is different to the inserted user")
	}
}

func TestDelete(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroy(db, t)
	dbInit(db, t)

	usersR := NewPostgresRepository(db)

	user := security.UserWithPassword{
		User: security.User{
			Login: "admin",
			Role:  1,
		},
		Password: "pass",
	}
	userID, err := usersR.Create(user)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	userGet, found, err := usersR.Get(userID)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if !found {
		t.Error("user not found")
		t.FailNow()
	}
	if user.Login != userGet.Login || user.Role != userGet.Role {
		t.Error("The user obtained is different to the inserted user")
	}

	err = usersR.Delete(userID)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	userGet, found, err = usersR.Get(userID)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if found {
		t.Error("User found while it should not")
		t.FailNow()
	}
}

func TestUpdate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroy(db, t)
	dbInit(db, t)

	usersR := NewPostgresRepository(db)

	user := security.UserWithPassword{
		User: security.User{
			Login: "admin",
			Role:  1,
		},
		Password: "pass",
	}
	userID, err := usersR.Create(user)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	userGet, found, err := usersR.Get(userID)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if !found {
		t.Error(err)
		t.FailNow()
	}
	if user.Login != userGet.Login || user.Role != userGet.Role {
		t.Error("The user obtained is different to the inserted user")
	}

	user.ID = userID
	user.Login = "newLogin"
	user.Role = 5
	err = usersR.Update(user.User)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	userGet, found, err = usersR.Get(userID)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if !found {
		t.Error(err)
		t.FailNow()
	}
	if user.Login != userGet.Login || user.Role != userGet.Role {
		t.Error("The user was not updated")
	}
}

func TestGetAll(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroy(db, t)
	dbInit(db, t)

	usersR := NewPostgresRepository(db)

	user1 := security.UserWithPassword{
		User: security.User{
			Login: "admin",
			Role:  1,
		},
		Password: "pass1",
	}
	usersR.Create(user1)
	user2 := security.UserWithPassword{
		User: security.User{
			Login: "test",
			Role:  1,
		},
		Password: "pass2",
	}
	usersR.Create(user2)

	users, err := usersR.GetAll()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if len(users) != 2 {
		t.Error("The Number of users is not as expected")
	}
}

func TestGetUsersOfGroup(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroy(db, t)
	dbInit(db, t)

	usersR := NewPostgresRepository(db)
	groupsR := groups.NewPostgresRepository(db)

	user := security.UserWithPassword{
		User: security.User{
			Login: "admin",
			Role:  1,
		},
		Password: "pass",
	}
	userID, err := usersR.Create(user)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	userGet, found, err := usersR.Get(userID)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if !found {
		t.Error(err)
		t.FailNow()
	}
	if user.Login != userGet.Login || user.Role != userGet.Role {
		t.Error("The user obtained is different to the inserted user")
	}

	group := groups.Group{
		Name: "userGroup",
	}
	groupID, err := groupsR.Create(group)
	if err != nil {
		t.Error(err)
	}
	groupGet, found, err := groupsR.Get(groupID)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if !found {
		t.Error(err)
		t.FailNow()
	}
	if group.Name != groupGet.Name {
		t.Error("The group obtained is different to the inserted group")
	}

	membership := groups.Membership{
		UserID:  userGet.ID,
		GroupID: groupGet.ID,
		Role:    1,
	}

	groupsR.CreateMembership(membership)

	users, err := usersR.GetUsersOfGroup(groupGet.ID)
	if err != nil {
		t.Error(err)
	}
	if len(users) != 1 {
		t.Error("The number of users if not as expected")
		t.FailNow()
	}
	if users[userID].ID != userGet.ID || users[userID].Login != userGet.Login ||
		users[userID].Role != userGet.Role || users[userID].RoleInGroup != membership.Role {
		t.Error("The user of group obtained is different to the expected user")
	}
}
