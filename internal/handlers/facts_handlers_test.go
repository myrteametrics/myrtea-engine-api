package handlers

import (
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/security/users"
	"net/http"
	"testing"

	"github.com/myrteametrics/myrtea-engine-api/v5/internal/fact"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/security/permissions"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/tests"
	"github.com/myrteametrics/myrtea-sdk/v5/engine"
)

func initGlobalRepository() {
	fm := fact.NewNativeMapRepository()
	fact.ReplaceGlobals(fm)
	_, err := fact.R().Create(engine.Fact{Name: "test1", Model: "model1"})
	if err != nil {
		return
	}
	_, err = fact.R().Create(engine.Fact{Name: "test2", Model: "model2"})
	if err != nil {
		return
	}
}

func initGlobalEmptyRepository() {
	fm := fact.NewNativeMapRepository()
	fact.ReplaceGlobals(fm)
}

func TestGetFacts(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping DB test in short mode")
	}
	initGlobalRepository()
	user := users.UserWithPermissions{Permissions: []permissions.Permission{permissions.New(permissions.TypeFact, "*", permissions.ActionList), permissions.New(permissions.TypeFact, "*", permissions.ActionGet)}}
	rr := tests.BuildTestHandler(t, "GET", "/facts", ``, "/facts", GetFacts, user)
	tests.CheckTestHandler(t, rr,
		http.StatusOK,
		`[{"id":1,"name":"test1","isObject":false,"model":"model1","comment":"","isTemplate":false},{"id":2,"name":"test2","isObject":false,"model":"model2","comment":"","isTemplate":false}]`+"\n",
	)
}

func TestGetFactsEmpty(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping DB test in short mode")
	}
	initGlobalEmptyRepository()
	user := users.UserWithPermissions{Permissions: []permissions.Permission{permissions.New(permissions.TypeFact, "*", permissions.ActionList), permissions.New(permissions.TypeFact, "*", permissions.ActionGet)}}
	rr := tests.BuildTestHandler(t, "GET", "/facts", ``, "/facts", GetFacts, user)
	tests.CheckTestHandler(t, rr, http.StatusOK, `[]`+"\n")
}

func TestGetFact(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping DB test in short mode")
	}
	initGlobalRepository()
	user := users.UserWithPermissions{Permissions: []permissions.Permission{permissions.New(permissions.TypeFact, "1", permissions.ActionGet)}}
	rr := tests.BuildTestHandler(t, "GET", "/facts/1", ``, "/facts/{id}", GetFact, user)
	tests.CheckTestHandler(t, rr, http.StatusOK, `{"id":1,"name":"test1","isObject":false,"model":"model1","comment":"","isTemplate":false}`+"\n")
}

// func TestGetFactByName(t *testing.T) {
// 	initGlobalRepository()
// 	user := users.UserWithPermissions{Permissions: []permissions.Permission{permissions.New(permissions.TypeFact, "1", permissions.ActionGet)}}
// 	rr := tests.BuildTestHandler(t, "GET", "/fact/test1?byName=true", ``, "/fact/{id}", GetFact, user)
// 	tests.CheckTestHandler(t, rr, http.StatusOK, `{"id":1,"name":"test1","isObject":false,"model":"model1","comment":"","isTemplate":false}`+"\n")
// }

func TestGetFactInvalidID(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping DB test in short mode")
	}
	initGlobalRepository()
	user := users.UserWithPermissions{Permissions: []permissions.Permission{permissions.New(permissions.TypeFact, "*", permissions.ActionGet)}}
	rr := tests.BuildTestHandler(t, "GET", "/fact/not_an_id", ``, "/fact/{id}", GetFact, user)
	tests.CheckTestHandler(t, rr, http.StatusBadRequest, `{"requestID":"","status":400,"type":"ParsingError","code":1002,"message":"Failed to parse a query param of type 'integer'"}`+"\n")
}

func TestGetFactNotExistingID(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping DB test in short mode")
	}
	initGlobalRepository()
	user := users.UserWithPermissions{Permissions: []permissions.Permission{permissions.New(permissions.TypeFact, "*", permissions.ActionGet)}}
	rr := tests.BuildTestHandler(t, "GET", "/fact/999", ``, "/fact/{id}", GetFact, user)
	tests.CheckTestHandler(t, rr, http.StatusNotFound, `{"requestID":"","status":404,"type":"RessourceError","code":3000,"message":"Ressource not found"}`+"\n")
}

// func TestGetFactNotExistingName(t *testing.T) {
// 	initGlobalRepository()
// 	user := users.UserWithPermissions{Permissions: []permissions.Permission{permissions.New(permissions.TypeFact, "1", permissions.ActionGet)}}
// 	rr := tests.BuildTestHandler(t, "GET", "/fact/not_an_id?byName=true", ``, "/fact/{id}", GetFact, user)
// 	tests.CheckTestHandler(t, rr, http.StatusNotFound, `{"requestID":"","status":404,"type":"RessourceError","code":3000,"message":"Ressource not found"}`+"\n")
// }

func TestPostFact(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping DB test in short mode")
	}
	initGlobalRepository()
	user := users.UserWithPermissions{Permissions: []permissions.Permission{permissions.New(permissions.TypeFact, "*", permissions.ActionCreate)}}
	rr := tests.BuildTestHandler(t, "POST", "/fact", `{"name":"test3","model":"newmodel","intent":{"operator":"count","term":"myterm"}}`, "/fact", PostFact, user)
	tests.CheckTestHandler(t, rr, http.StatusOK, `{"id":3,"name":"test3","isObject":false,"model":"newmodel","intent":{"operator":"count","term":"myterm"},"comment":"","isTemplate":false}`+"\n")

	facts, err := fact.R().GetAll()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if _, exists := facts[3]; !exists {
		t.Error("Fact test3 should exists")
	}
}

func TestPutFact(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping DB test in short mode")
	}
	initGlobalRepository()
	user := users.UserWithPermissions{Permissions: []permissions.Permission{permissions.New(permissions.TypeFact, "1", permissions.ActionUpdate)}}
	rr := tests.BuildTestHandler(t, "PUT", "/fact/1", `{"name":"test1","model":"newmodel","intent":{"operator":"count","term":"myterm"}}`, "/fact/{id}", PutFact, user)
	tests.CheckTestHandler(t, rr, http.StatusOK, `{"id":1,"name":"test1","isObject":false,"model":"newmodel","intent":{"operator":"count","term":"myterm"},"comment":"","isTemplate":false}`+"\n")

	facts, err := fact.R().GetAll()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if _, exists := facts[1]; !exists {
		t.Error("Fact test1 should exists")
	}
	f := facts[1]
	if f.Model != "newmodel" {
		t.Error("Fact test1 has not been updated")
	}
}

func TestPutFactInvalidBody(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping DB test in short mode")
	}
	initGlobalRepository()
	user := users.UserWithPermissions{Permissions: []permissions.Permission{permissions.New(permissions.TypeFact, "1", permissions.ActionUpdate)}}
	rr := tests.BuildTestHandler(t, "PUT", "/fact/1", `Not a json string`, "/fact/{id}", PutFact, user)
	tests.CheckTestHandler(t, rr, http.StatusBadRequest, `{"requestID":"","status":400,"type":"ParsingError","code":1000,"message":"Failed to parse the JSON body provided in the request"}`+"\n")

	facts, err := fact.R().GetAll()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if _, exists := facts[1]; !exists {
		t.Error("Fact test1 should exists")
		t.FailNow()
	}
	f := facts[1]
	if f.Model != "model1" {
		t.Error("Fact test1 should not have been updated")
	}
}

func TestDeleteFact(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping DB test in short mode")
	}
	initGlobalRepository()
	user := users.UserWithPermissions{Permissions: []permissions.Permission{permissions.New(permissions.TypeFact, "1", permissions.ActionDelete)}}
	rr := tests.BuildTestHandler(t, "DELETE", "/fact/1", ``, "/fact/{id}", DeleteFact, user)
	tests.CheckTestHandler(t, rr, http.StatusOK, ``)

	facts, err := fact.R().GetAll()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if _, exists := facts[1]; exists {
		t.Error("Fact test1 should have been deleted")
	}
	if _, exists := facts[2]; !exists {
		t.Error("Fact test2 should not have been deleted")
	}
}

func TestDeleteFactInvalidID(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping DB test in short mode")
	}
	initGlobalRepository()
	user := users.UserWithPermissions{Permissions: []permissions.Permission{permissions.New(permissions.TypeFact, "1", permissions.ActionDelete)}}
	rr := tests.BuildTestHandler(t, "DELETE", "/fact/not_an_id", ``, "/fact/{id}", DeleteFact, user)
	tests.CheckTestHandler(t, rr, http.StatusBadRequest, `{"requestID":"","status":400,"type":"ParsingError","code":1002,"message":"Failed to parse a query param of type 'integer'"}`+"\n")

	facts, err := fact.R().GetAll()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if _, exists := facts[1]; !exists {
		t.Error("Fact test1 should not have been deleted")
	}
	if _, exists := facts[2]; !exists {
		t.Error("Fact test2 should not have been deleted")
	}
}
