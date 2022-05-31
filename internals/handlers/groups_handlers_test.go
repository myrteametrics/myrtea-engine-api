package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/groups"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/tests"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/users"
	"github.com/myrteametrics/myrtea-sdk/v4/security"
)

func groupsDBInit(dbClient *sqlx.DB, t *testing.T) {
	groupsDBDestroy(dbClient, t)
	tests.DBExec(dbClient, `CREATE EXTENSION IF NOT EXISTS pgcrypto;`, t, true)
	tests.DBExec(dbClient, tests.UsersTableV1, t, true)
	tests.DBExec(dbClient, tests.GroupsTableV1, t, true)
	tests.DBExec(dbClient, tests.UserMembershipsTableV1, t, true)
}

func groupsDBDestroy(dbClient *sqlx.DB, t *testing.T) {
	tests.DBExec(dbClient, tests.UserMembershipsDropTableV1, t, false)
	tests.DBExec(dbClient, tests.GroupsDropTableV1, t, false)
	tests.DBExec(dbClient, tests.UsersDropTableV1, t, false)
}

func TestGetGroups(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer groupsDBDestroy(db, t)
	groupsDBInit(db, t)

	groupsR := groups.NewPostgresRepository(db)
	groups.ReplaceGlobals(groupsR)

	group1 := groups.Group{
		Name: "group1",
	}
	group2 := groups.Group{
		Name: "group2",
	}
	group1.ID, _ = groupsR.Create(group1)
	group2.ID, _ = groupsR.Create(group2)

	req, err := http.NewRequest("GET", "/groups", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	r := chi.NewRouter()
	r.Get("/groups", GetGroups)
	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	mapGroups := []groups.Group{group1, group2}
	usersData, _ := json.Marshal(mapGroups)
	expected := string(usersData) + "\n"

	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}

func TestGetGroup(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer groupsDBDestroy(db, t)
	groupsDBInit(db, t)

	groupsR := groups.NewPostgresRepository(db)
	groups.ReplaceGlobals(groupsR)

	group := groups.Group{
		Name: "group",
	}
	group.ID, _ = groupsR.Create(group)

	req, err := http.NewRequest("GET", "/groups/"+strconv.FormatInt(group.ID, 10), nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	r := chi.NewRouter()
	r.Get("/groups/{id}", GetGroup)
	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	userData, _ := json.Marshal(group)
	expected := string(userData) + "\n"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}

func TestPostGroup(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer groupsDBDestroy(db, t)
	groupsDBInit(db, t)

	groupsR := groups.NewPostgresRepository(db)
	groups.ReplaceGlobals(groupsR)

	group := groups.Group{
		Name: "group",
	}
	groupData, _ := json.Marshal(group)

	req, err := http.NewRequest("POST", "/groups", bytes.NewBuffer(groupData))
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	r := chi.NewRouter()
	r.Post("/groups", PostGroup)
	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	groupGet, found, err := groups.R().Get(1)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if !found {
		t.Error("User not found")
		t.FailNow()
	}
	if groupGet.Name != group.Name {
		t.Errorf("The user group was not inserted correctly")
	}
}

func TestDeleteGroup(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer groupsDBDestroy(db, t)
	groupsDBInit(db, t)

	groupsR := groups.NewPostgresRepository(db)
	groups.ReplaceGlobals(groupsR)

	group := groups.Group{
		Name: "group",
	}
	group.ID, _ = groupsR.Create(group)

	req, err := http.NewRequest("DELETE", "/groups/"+strconv.FormatInt(group.ID, 10), nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	r := chi.NewRouter()
	r.Delete("/groups/{id}", DeleteGroup)
	r.ServeHTTP(rr, req)

	_, found, err := groupsR.Get(group.ID)
	if err != nil {
		t.Error("Error was unexpected")
		t.FailNow()
	}
	if found {
		t.Error("Group found while it should be deleted")
	}
}

func TestPutGroup(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer groupsDBDestroy(db, t)
	groupsDBInit(db, t)

	groupsR := groups.NewPostgresRepository(db)
	groups.ReplaceGlobals(groupsR)

	group := groups.Group{
		Name: "group",
	}
	group.ID, _ = groupsR.Create(group)

	group.Name = "newName"

	groupData, _ := json.Marshal(group)
	req, err := http.NewRequest("PUT", "/groups/"+strconv.FormatInt(group.ID, 10), bytes.NewBuffer(groupData))
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	r := chi.NewRouter()
	r.Put("/groups/{id}", PutGroup)
	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	groupGet, found, err := groups.R().Get(1)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if !found {
		t.Error("User not found")
		t.FailNow()
	}
	if groupGet.Name != group.Name {
		t.Errorf("The user group was not updated correctly")
	}
}

func TestPutMembership(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer groupsDBDestroy(db, t)
	groupsDBInit(db, t)

	groupsR := groups.NewPostgresRepository(db)
	groups.ReplaceGlobals(groupsR)

	usersR := users.NewPostgresRepository(db)
	users.ReplaceGlobals(usersR)

	group := groups.Group{
		Name: "group",
	}
	group.ID, _ = groupsR.Create(group)

	user := security.UserWithPassword{
		User: security.User{
			Login: "admin",
			Role:  1,
		},
		Password: "pass1",
	}
	user.ID, _ = usersR.Create(user)

	membership := groups.Membership{
		UserID:  user.ID,
		GroupID: group.ID,
		Role:    1,
	}

	membershipData, _ := json.Marshal(membership)
	req, err := http.NewRequest("PUT", "/groups/"+strconv.FormatInt(group.ID, 10)+"/users/"+strconv.FormatInt(user.ID, 10), bytes.NewBuffer(membershipData))
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	r := chi.NewRouter()
	r.Put("/groups/{groupid}/users/{userid}", PutMembership)
	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	groupsOfUser, err := groups.R().GetGroupsOfUser(user.ID)
	if err != nil {
		t.Errorf("unexpected error: " + err.Error())
	} else if groupsOfUser[0].Name != group.Name || groupsOfUser[0].UserRole != membership.Role {
		t.Errorf("The user membership was not created or updated correctly")
	}

	membership.Role = 1

	membershipData, _ = json.Marshal(membership)
	req, err = http.NewRequest("PUT", "/groups/"+strconv.FormatInt(group.ID, 10)+"/users/"+strconv.FormatInt(user.ID, 10), bytes.NewBuffer(membershipData))
	if err != nil {
		t.Fatal(err)
	}
	rr = httptest.NewRecorder()
	r = chi.NewRouter()
	r.Put("/groups/{groupid}/users/{userid}", PutMembership)
	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	groupsOfUser, err = groups.R().GetGroupsOfUser(user.ID)
	if err != nil {
		t.Errorf("unexpected error: " + err.Error())
	} else if groupsOfUser[0].Name != group.Name || groupsOfUser[0].UserRole != membership.Role {
		t.Errorf("The user membership was not created or updated correctly")
	}
}

func TestDeleteMembership(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer groupsDBDestroy(db, t)
	groupsDBInit(db, t)

	groupsR := groups.NewPostgresRepository(db)
	groups.ReplaceGlobals(groupsR)

	usersR := users.NewPostgresRepository(db)
	users.ReplaceGlobals(usersR)

	group := groups.Group{
		Name: "group",
	}
	group.ID, _ = groupsR.Create(group)

	user := security.UserWithPassword{
		User: security.User{
			Login: "admin",
			Role:  1,
		},
		Password: "pass1",
	}
	user.ID, _ = usersR.Create(user)

	membership := groups.Membership{
		UserID:  user.ID,
		GroupID: group.ID,
		Role:    1,
	}
	err := groupsR.CreateMembership(membership)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	groupsOfUser, err := groups.R().GetGroupsOfUser(user.ID)
	if err != nil {
		t.Errorf("unexpected error: " + err.Error())
	} else if groupsOfUser[0].Name != group.Name || groupsOfUser[0].UserRole != membership.Role {
		t.Errorf("The user membership was not created correctly")
	}

	req, err := http.NewRequest("DELETE", "/groups/"+strconv.FormatInt(group.ID, 10)+"/users/"+strconv.FormatInt(user.ID, 10), nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	r := chi.NewRouter()
	r.Delete("/groups/{groupid}/users/{userid}", DeleteMembership)
	r.ServeHTTP(rr, req)

	groupsOfUser, err = groups.R().GetGroupsOfUser(user.ID)
	if err != nil {
		t.Errorf("unexpected error: " + err.Error())
	} else if len(groupsOfUser) > 0 {
		t.Errorf("The user membership was not deleted")
	}

}

func TestGetUsersOfGroup(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer groupsDBDestroy(db, t)
	groupsDBInit(db, t)

	groupsR := groups.NewPostgresRepository(db)
	groups.ReplaceGlobals(groupsR)

	usersR := users.NewPostgresRepository(db)
	users.ReplaceGlobals(usersR)

	group := groups.Group{
		Name: "group",
	}
	group.ID, _ = groupsR.Create(group)

	user := security.UserWithPassword{
		User: security.User{
			Login:     "admin",
			Role:      1,
			Created:   time.Now().Truncate(1 * time.Millisecond).UTC(),
			LastName:  "lastName1",
			FirstName: "firstName1",
			Email:     "user1@myrtea.com",
			Phone:     "0123456789",
		},
		Password: "pass1",
	}
	user.ID, _ = usersR.Create(user)

	membership := groups.Membership{
		UserID:  user.ID,
		GroupID: group.ID,
		Role:    1,
	}
	err := groupsR.CreateMembership(membership)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	mapUsersOfGroup := []users.UserOfGroup{
		{
			User:        user.User,
			RoleInGroup: membership.Role,
		},
	}
	usersData, _ := json.Marshal(mapUsersOfGroup)
	expected := string(usersData) + "\n"

	req, err := http.NewRequest("GET", "/groups/"+strconv.FormatInt(group.ID, 10)+"/users", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	r := chi.NewRouter()
	r.Get("/groups/{groupid}/users", GetUsersOfGroup)
	r.ServeHTTP(rr, req)

	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}

}
