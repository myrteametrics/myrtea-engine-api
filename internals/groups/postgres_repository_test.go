package groups

import (
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/tests"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/users"
	"github.com/myrteametrics/myrtea-sdk/v4/security"
)

func dbInit(dbClient *sqlx.DB, t *testing.T) {

	dbDestroy(dbClient, t)
	_, err := dbClient.Exec(`CREATE EXTENSION IF NOT EXISTS pgcrypto;`)
	if err != nil {
		t.Error(err)
	}

	_, err = dbClient.Exec(tests.UsersTableV1)
	if err != nil {
		t.Error(err)
	}

	_, err = dbClient.Exec(tests.GroupsTableV1)
	if err != nil {
		t.Error(err)
	}

	_, err = dbClient.Exec(tests.UserMembershipsTableV1)
	if err != nil {
		t.Error(err)
	}
}

func dbDestroy(dbClient *sqlx.DB, t *testing.T) {

	_, err := dbClient.Exec(tests.UserMembershipsDropTableV1)
	if err != nil {
		t.Error(err)
	}

	_, err = dbClient.Exec(tests.UsersDropTableV1)
	if err != nil {
		t.Error(err)
	}

	_, err = dbClient.Exec(tests.GroupsDropTableV1)
	if err != nil {
		t.Error(err)
	}
}

func TestCreate_Get(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroy(db, t)
	dbInit(db, t)

	groupsR := NewPostgresRepository(db)

	group := Group{
		Name: "userGroup",
	}
	groupID, err := groupsR.Create(group)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	groupGet, found, err := groupsR.Get(groupID)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if !found {
		t.Error("Group not found after creation")
		t.FailNow()
	}
	if group.Name != groupGet.Name {
		t.Error("The group obtained is different to the inserted group")
	}
}

func TestDelete(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroy(db, t)
	dbInit(db, t)

	groupsR := NewPostgresRepository(db)

	group := Group{
		Name: "userGroup",
	}
	groupID, err := groupsR.Create(group)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	groupGet, found, err := groupsR.Get(groupID)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if !found {
		t.Error("Group not found")
		t.FailNow()
	}
	if group.Name != groupGet.Name {
		t.Error("The group obtained is different to the inserted group")
	}

	err = groupsR.Delete(groupID)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	_, found, err = groupsR.Get(groupID)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if found {
		t.Error("Group found while it should be deleted")
	}
}

func TestUpdate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroy(db, t)
	dbInit(db, t)

	groupsR := NewPostgresRepository(db)

	group := Group{
		Name: "userGroup",
	}
	groupID, err := groupsR.Create(group)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	groupGet, found, err := groupsR.Get(groupID)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if !found {
		t.Error("Group not found")
		t.FailNow()
	}
	if group.Name != groupGet.Name {
		t.Error("The group obtained is different to the inserted group")
	}

	group.ID = groupID
	group.Name = "newName"
	groupsR.Update(group)
	groupGet, found, err = groupsR.Get(groupID)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if !found {
		t.Error("Group not found")
		t.FailNow()
	}
	if group.Name != groupGet.Name {
		t.Error("The user group was not updated")
	}
}

func TestGetAll(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroy(db, t)
	dbInit(db, t)

	groupsR := NewPostgresRepository(db)

	group1 := Group{
		Name: "userGroup1",
	}
	groupsR.Create(group1)
	group2 := Group{
		Name: "userGroup2",
	}
	groupsR.Create(group2)

	groups, _ := groupsR.GetAll()
	if len(groups) != 2 {
		t.Error("The Number of user groups is not as expected")
	}
}

func TestCreateGetMembership(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroy(db, t)
	dbInit(db, t)

	usersR := users.NewPostgresRepository(db)
	groupsR := NewPostgresRepository(db)

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

	group := Group{
		Name: "userGroup",
	}
	groupID, err := groupsR.Create(group)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	groupGet, found, err := groupsR.Get(groupID)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if !found {
		t.Error("group not found")
		t.FailNow()
	}
	if group.Name != groupGet.Name {
		t.Error("The group obtained is different to the inserted group")
	}

	membership := Membership{
		UserID:  userGet.ID,
		GroupID: groupGet.ID,
		Role:    1,
	}

	err = groupsR.CreateMembership(membership)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	membershipGet, found, err := groupsR.GetMembership(userID, groupID)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if !found {
		t.Error("membership not found")
		t.FailNow()
	}
	if membershipGet.UserID != membership.UserID || membershipGet.GroupID != membership.GroupID ||
		membershipGet.Role != membership.Role {
		t.Error("The group obtained is different to the expected group")
	}
}

func TestUpdateMembership(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroy(db, t)
	dbInit(db, t)

	usersR := users.NewPostgresRepository(db)
	groupsR := NewPostgresRepository(db)

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

	group := Group{
		Name: "userGroup",
	}
	groupID, err := groupsR.Create(group)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	groupGet, found, err := groupsR.Get(groupID)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if !found {
		t.Error("group not found")
		t.FailNow()
	}
	if group.Name != groupGet.Name {
		t.Error("The group obtained is different to the inserted group")
	}

	membership := Membership{
		UserID:  userGet.ID,
		GroupID: groupGet.ID,
		Role:    1,
	}

	groupsR.CreateMembership(membership)

	groups, err := groupsR.GetGroupsOfUser(userID)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if len(groups) != 1 {
		t.Error("The number of groups if not as expected")
		t.FailNow()
	}
	if groups[0].ID != groupGet.ID || groups[0].Name != groupGet.Name ||
		groups[0].UserRole != membership.Role {
		t.Error("The group obtained is different to the expected group")
	}

	membership.Role = 1
	groupsR.UpdateMembership(membership)
	groups, err = groupsR.GetGroupsOfUser(userID)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if len(groups) != 1 {
		t.Error("The number of groups if not as expected")
		t.FailNow()
	}
	if groups[0].ID != groupGet.ID || groups[0].Name != groupGet.Name ||
		groups[0].UserRole != membership.Role {
		t.Error("The group obtained is different to the expected group")
	}
}

func TestDeleteMembership(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroy(db, t)
	dbInit(db, t)

	usersR := users.NewPostgresRepository(db)
	groupsR := NewPostgresRepository(db)

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

	group := Group{
		Name: "userGroup",
	}
	groupID, err := groupsR.Create(group)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	groupGet, found, err := groupsR.Get(groupID)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if !found {
		t.Error("group not found")
		t.FailNow()
	}
	if group.Name != groupGet.Name {
		t.Error("The group obtained is different to the inserted group")
	}

	membership := Membership{
		UserID:  userGet.ID,
		GroupID: groupGet.ID,
		Role:    1,
	}

	groupsR.CreateMembership(membership)

	groups, err := groupsR.GetGroupsOfUser(userID)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if len(groups) != 1 {
		t.Error("The number of groups if not as expected")
	} else if groups[0].ID != groupGet.ID || groups[0].Name != groupGet.Name ||
		groups[0].UserRole != membership.Role {
		t.Error("The group obtained is different to the expected group")
	}

	groupsR.DeleteMembership(membership.UserID, membership.GroupID)
	groups, err = groupsR.GetGroupsOfUser(userID)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if len(groups) != 0 {
		t.Error("The number of groups if not as expected")
	}
}
