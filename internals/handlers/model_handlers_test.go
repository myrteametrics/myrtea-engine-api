package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	model "github.com/myrteametrics/myrtea-engine-api/v4/internals/modeler"
	"github.com/myrteametrics/myrtea-sdk/v4/modeler"
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
			AdvancedSettings:          map[string]interface{}{"number_of_replica": 2, "number_of_shard": 6},
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
			AdvancedSettings:          map[string]interface{}{"number_of_replica": 2, "number_of_shard": 6},
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

	req, err := http.NewRequest("GET", "/models", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	r := chi.NewRouter()
	r.Get("/models", GetModels)
	r.ServeHTTP(rr, req)

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

	req, err := http.NewRequest("GET", "/models", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	r := chi.NewRouter()
	r.Get("/models", GetModels)
	r.ServeHTTP(rr, req)

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

	req, err := http.NewRequest("GET", "/models/1", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	r := chi.NewRouter()
	r.Get("/models/{id}", GetModel)
	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expected := `{"id":1,"name":"test","synonyms":["model1"],"fields":[{"name":"test","type":"string","semantic":false,"synonyms":["other"]},{"name":"test_object","type":"object","keepObjectSeparation":false,"fields":[{"name":"test_subfield","type":"string","semantic":false,"synonyms":["other2"]}]}],"source":"{}","elasticsearchOptions":{"rollmode":"cron","rollcron":"0 0 * * *","enablePurge":true,"purgeMaxConcurrentIndices":30,"patchAliasMaxIndices":2,"advancedSettings":{"number_of_replica":2,"number_of_shard":6}}}` + "\n"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}

func TestGetModelInvalidID(t *testing.T) {
	initModelRepository(t)

	req, err := http.NewRequest("GET", "/models/not_an_id", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	r := chi.NewRouter()
	r.Get("/models/{id}", GetModel)
	r.ServeHTTP(rr, req)

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
	req, err := http.NewRequest("POST", "/models", bytes.NewBuffer(b))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	r := chi.NewRouter()
	r.Post("/models", PostModel)
	r.ServeHTTP(rr, req)

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
	req, err := http.NewRequest("PUT", "/models/1", bytes.NewBuffer(b))
	if err != nil {
		t.Fatal(err)
	}

	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	r := chi.NewRouter()
	r.Put("/models/{id}", PutModel)
	r.ServeHTTP(rr, req)

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

	req, err := http.NewRequest("PUT", "/models/99", bytes.NewBuffer(
		[]byte(`{"name":"test","fields":[{"name":"test","type":"string","semantic":false}],"source":"{}","rollmode":"test","rollcron":"test"}`)))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	r := chi.NewRouter()
	r.Put("/models/{id}", PutModel)
	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
	expected := `{"requestID":"","status":400,"type":"RessourceError","code":2001,"message":"Provided resource definition can be parsed, but is invalid"}` + "\n"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), "")
	}
}

func TestPutModelInvalidBody(t *testing.T) {
	initModelRepository(t)

	req, err := http.NewRequest("PUT", "/models/1", bytes.NewBuffer([]byte(`Not a json string`)))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	r := chi.NewRouter()
	r.Put("/models/{id}", PutModel)
	r.ServeHTTP(rr, req)

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

	req, err := http.NewRequest("DELETE", "/models/1", nil)

	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	r := chi.NewRouter()
	r.Delete("/models/{id}", DeleteModel)
	r.ServeHTTP(rr, req)

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

	req, err := http.NewRequest("DELETE", "/models/not_an_id", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	r := chi.NewRouter()
	r.Delete("/models/{id}", DeleteModel)
	r.ServeHTTP(rr, req)

	t.Log(rr)
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}

	expected := `{"requestID":"","status":400,"type":"ParsingError","code":1002,"message":"Failed to parse a query param of type 'integer'"}` + "\n"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}
