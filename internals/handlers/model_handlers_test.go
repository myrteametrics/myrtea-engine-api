package handlers

import (
	"encoding/json"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"net/http"
	"testing"

	model "github.com/myrteametrics/myrtea-engine-api/v5/internals/modeler"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/security/permissions"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/security/users"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/tests"
	"github.com/myrteametrics/myrtea-sdk/v5/modeler"
)

func initModelRepository(t *testing.T) []modeler.Model {
	mm := model.NewNativeMapRepository()
	model.ReplaceGlobals(mm)
	var fieldarray []modeler.Field

	fieldarray = append(fieldarray,
		&modeler.FieldLeaf{
			Name:     "test",
			Ftype:    modeler.String,
			Semantic: false,
			Synonyms: []string{"other"},
		},
		&modeler.FieldObject{
			Name:  "test_object",
			Ftype: modeler.Object,
			Fields: []modeler.Field{
				&modeler.FieldLeaf{
					Name:     "test_subfield",
					Ftype:    modeler.String,
					Semantic: false,
					Synonyms: []string{"other2"},
				},
			},
		})

	newModel := modeler.Model{
		Name:     "test",
		Synonyms: []string{"model1"},
		Fields:   fieldarray,
		ElasticsearchOptions: modeler.ElasticsearchOptions{
			Rollmode:                  "cron",
			Rollcron:                  "0 0 * * *",
			EnablePurge:               true,
			PurgeMaxConcurrentIndices: 30,
			PatchAliasMaxIndices:      2,
			AdvancedSettings:          types.IndexSettings{NumberOfReplicas: "2", NumberOfShards: "6"},
		},
		Source: "{}",
	}
	newModel2 := modeler.Model{
		Name:     "test_2",
		Synonyms: []string{"model2"},
		Fields:   fieldarray,
		ElasticsearchOptions: modeler.ElasticsearchOptions{
			Rollmode:                  "cron",
			Rollcron:                  "0 0 * * *",
			EnablePurge:               true,
			PurgeMaxConcurrentIndices: 30,
			PatchAliasMaxIndices:      2,
			AdvancedSettings:          types.IndexSettings{NumberOfReplicas: "2", NumberOfShards: "6"},
		},
		Source: "{}",
	}
	idModel1, err := model.R().Create(newModel)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	newModel.ID = idModel1

	idModel2, err := model.R().Create(newModel2)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	newModel2.ID = idModel2

	return []modeler.Model{newModel, newModel2}
}

func initModelEmptyRepository() {
	mm := model.NewNativeMapRepository()
	model.ReplaceGlobals(mm)
}

func TestGetModels(t *testing.T) {
	models := initModelRepository(t)
	user := users.UserWithPermissions{Permissions: []permissions.Permission{permissions.New(permissions.TypeModel, permissions.All, permissions.ActionList), permissions.New(permissions.TypeModel, permissions.All, permissions.ActionGet)}}
	rr := tests.BuildTestHandler(t, "GET", "/models", ``, "/models", GetModels, user)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	b, _ := json.Marshal(models)
	expected := string(b) + "\n"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}

func TestGetModelsEmpty(t *testing.T) {
	initModelEmptyRepository()
	user := users.UserWithPermissions{Permissions: []permissions.Permission{permissions.New(permissions.TypeModel, permissions.All, permissions.ActionList)}}
	rr := tests.BuildTestHandler(t, "GET", "/models", ``, "/models", GetModels, user)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	expected := `[]` + "\n"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}

func TestGetModel(t *testing.T) {
	initModelRepository(t)
	user := users.UserWithPermissions{Permissions: []permissions.Permission{permissions.New(permissions.TypeModel, permissions.All, permissions.ActionGet)}}
	rr := tests.BuildTestHandler(t, "GET", "/models/1", ``, "/models/{id}", GetModel, user)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expected := `{"id":1,"name":"test","synonyms":["model1"],"fields":[{"name":"test","type":"string","semantic":false,"synonyms":["other"]},{"name":"test_object","type":"object","keepObjectSeparation":false,"fields":[{"name":"test_subfield","type":"string","semantic":false,"synonyms":["other2"]}]}],"source":"{}","elasticsearchOptions":{"rollmode":"cron","rollcron":"0 0 * * *","enablePurge":true,"purgeMaxConcurrentIndices":30,"patchAliasMaxIndices":2,"advancedSettings":{"number_of_replicas":"2","number_of_shards":"6"}}}` + "\n"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}

func TestGetModelInvalidID(t *testing.T) {
	initModelRepository(t)
	user := users.UserWithPermissions{Permissions: []permissions.Permission{permissions.New(permissions.TypeModel, permissions.All, permissions.ActionGet)}}
	rr := tests.BuildTestHandler(t, "GET", "/models/not_an_id", ``, "/models/{id}", GetModel, user)
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
	expected := `{"requestID":"","status":400,"type":"ParsingError","code":1002,"message":"Failed to parse a query param of type 'integer'"}` + "\n"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), "")
	}
}

func TestPostModel(t *testing.T) {
	models := initModelRepository(t)

	m := models[0]
	m.Name = "test_3"
	b, _ := json.Marshal(m)

	user := users.UserWithPermissions{Permissions: []permissions.Permission{permissions.New(permissions.TypeModel, permissions.All, permissions.ActionCreate)}}
	rr := tests.BuildTestHandler(t, "POST", "/models", string(b), "/models", PostModel, user)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	model, found, err := model.R().GetByName("test_3")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if !found {
		t.Error("Model not created while it should")
		t.FailNow()
	}
	if model.Name != "test_3" {
		t.Error("invalid model name")
	}
}

func TestPutModel(t *testing.T) {
	sourceModels := initModelRepository(t)

	m := sourceModels[0]
	m.Name = "test"
	b, _ := json.Marshal(m)

	user := users.UserWithPermissions{Permissions: []permissions.Permission{permissions.New(permissions.TypeModel, permissions.All, permissions.ActionUpdate)}}
	rr := tests.BuildTestHandler(t, "PUT", "/models/1", string(b), "/models/{id}", PutModel, user)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	models, err := model.R().GetAll()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if _, found := models[1]; !found {
		t.Error("Model new_model should not be nil")
	}
	m2 := models[1]
	if m2.Name != "test" {
		t.Error("Model new_model has not been updated")
	}
}

func TestPutModelInvalidResource(t *testing.T) {
	initModelRepository(t)

	user := users.UserWithPermissions{Permissions: []permissions.Permission{permissions.New(permissions.TypeModel, permissions.All, permissions.ActionUpdate)}}
	rr := tests.BuildTestHandler(t, "PUT", "/models/99", `{"name":"test","fields":[{"name":"test","type":"string","semantic":false}],"source":"{}","rollmode":"test","rollcron":"test"}`, "/models/{id}", PutModel, user)
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
	expected := `{"requestID":"","status":400,"type":"ResourceError","code":2001,"message":"Provided resource definition can be parsed, but is invalid"}` + "\n"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), "")
	}
}

func TestPutModelInvalidBody(t *testing.T) {
	initModelRepository(t)

	user := users.UserWithPermissions{Permissions: []permissions.Permission{permissions.New(permissions.TypeModel, permissions.All, permissions.ActionUpdate)}}
	rr := tests.BuildTestHandler(t, "PUT", "/models/1", `Not a json string`, "/models/{id}", PutModel, user)
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}

	expected := `{"requestID":"","status":400,"type":"ParsingError","code":1000,"message":"Failed to parse the JSON body provided in the request"}` + "\n"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}

func TestDeleteModel(t *testing.T) {
	initModelRepository(t)

	user := users.UserWithPermissions{Permissions: []permissions.Permission{permissions.New(permissions.TypeModel, permissions.All, permissions.ActionDelete)}}
	rr := tests.BuildTestHandler(t, "DELETE", "/models/1", ``, "/models/{id}", DeleteModel, user)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expected := ""
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), "")
	}
}

func TestDeleteModelInvalidID(t *testing.T) {
	initModelRepository(t)

	user := users.UserWithPermissions{Permissions: []permissions.Permission{permissions.New(permissions.TypeModel, permissions.All, permissions.ActionDelete)}}
	rr := tests.BuildTestHandler(t, "DELETE", "/models/not_an_id", ``, "/models/{id}", DeleteModel, user)
	t.Log(rr)
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}

	expected := `{"requestID":"","status":400,"type":"ParsingError","code":1002,"message":"Failed to parse a query param of type 'integer'"}` + "\n"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}
