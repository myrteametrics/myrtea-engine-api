package modeler

import (
	"testing"

	"github.com/myrteametrics/myrtea-sdk/v5/modeler"
)

func TestNew(t *testing.T) {
	r := NewNativeMapRepository()
	if r == nil {
		t.Error("Model Repository is nil")
	}
}

func TestReplaceGlobal(t *testing.T) {
	r := NewNativeMapRepository()
	reverse := ReplaceGlobals(r)
	if R() == nil {
		t.Error("Global model repository is nil")
	}
	reverse()
	if R() != nil {
		t.Error("Global model repository is not nil after reverse")
	}
}

func TestCreate(t *testing.T) {
	var err error
	r := NewNativeMapRepository()
	model := modeler.Model{Name: "test_name", Synonyms: []string{"model_synonym"}}
	id, err := r.Create(model)
	if err != nil {
		t.Error(err)
	}
	if id != 1 {
		t.Error("invalid generated model id")
	}

	modelGet, found, err := r.Get(id)
	if err != nil {
		t.Error(err)
	}
	if !found {
		t.Error("model not found")
	}
	if id != modelGet.ID {
		t.Error("invalid model ID")
	}
	if model.Synonyms[0] != modelGet.Synonyms[0] {
		t.Error("invalid model Comment")
	}
}

func TestGet(t *testing.T) {
	var err error
	r := NewNativeMapRepository()

	modelGet, found, err := r.Get(1)
	if err != nil {
		t.Error(err)
	}
	if found {
		t.Error("found a model from nowhere")
	}

	model := modeler.Model{Name: "test_name", Synonyms: []string{"model_synonym"}}
	id, err := r.Create(model)
	if err != nil {
		t.Error(err)
	}
	modelGet, found, err = r.Get(id)
	if err != nil {
		t.Error(err)
	}
	if !found {
		t.Error("model not found")
	}
	if id != modelGet.ID {
		t.Error("invalid model ID")
	}
	if model.Synonyms[0] != modelGet.Synonyms[0] {
		t.Error("invalid model Comment")
	}
}

func TestUpdate(t *testing.T) {
	var err error
	r := NewNativeMapRepository()
	id, err := r.Create(modeler.Model{Name: "test_name", Synonyms: []string{"model_synonym"}})
	if err != nil {
		t.Error(err)
	}

	// Update existing
	model := modeler.Model{Name: "test_name", Synonyms: []string{"model_synonym"}}
	err = r.Update(id, model)
	if err != nil {
		t.Error(err)
	}
	modelGet, found, err := r.Get(id)
	if err != nil {
		t.Error(err)
	}
	if !found {
		t.Error("model not found")
	}
	if id != modelGet.ID {
		t.Error("invalid model ID")
	}
	if model.Synonyms[0] != modelGet.Synonyms[0] {
		t.Error("invalid model Comment")
	}
}

func TestUpdateNotExists(t *testing.T) {
	var err error
	r := NewNativeMapRepository()
	id, err := r.Create(modeler.Model{Name: "test_name", Synonyms: []string{"model_synonym"}})
	if err != nil {
		t.Error(err)
	}

	model := modeler.Model{Name: "test_name_2", Synonyms: []string{"model_synonym"}}
	err = r.Update(id+1, model)
	if err == nil {
		t.Error("Model doesn't exists and cannot be updated")
	}
	_, found, err := r.Get(id + 1)
	if err != nil {
		t.Error(err)
	}
	if found {
		t.Error("Model should not have been created")
	}
}

func TestDelete(t *testing.T) {
	var err error
	r := NewNativeMapRepository()
	id, err := r.Create(modeler.Model{Name: "test_name", Synonyms: []string{"model_synonym"}})
	if err != nil {
		t.Error(err)
	}
	id2, err := r.Create(modeler.Model{Name: "test_name2", Synonyms: []string{"model_synonym"}})
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
	}
	if found {
		t.Error("model has not been deleted")
	}
	_, found, err = r.Get(id2)
	if err != nil {
		t.Error(err)
	}
	if !found {
		t.Error("model2 has been deleted while it should not")
	}
}

func TestDeleteNotExists(t *testing.T) {
	var err error
	r := NewNativeMapRepository()
	id, err := r.Create(modeler.Model{Name: "test_name", Synonyms: []string{"model_synonym"}})
	if err != nil {
		t.Error(err)
	}

	err = r.Delete(id + 1)
	if err == nil {
		t.Error("Cannot delete a non-existing object")
	}
	_, found, err := r.Get(id)
	if err != nil {
		t.Error(err)
	}
	if !found {
		t.Error("model has been deleted while it should not")
	}
}

func TestGetAll(t *testing.T) {
	var err error
	r := NewNativeMapRepository()
	id, err := r.Create(modeler.Model{Name: "test_name", Synonyms: []string{"model_synonym"}})
	if err != nil {
		t.Error(err)
	}
	id2, err := r.Create(modeler.Model{Name: "test_name2", Synonyms: []string{"model_synonym"}})
	if err != nil {
		t.Error(err)
	}

	models, err := r.GetAll()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if models == nil {
		t.Error("models is nil")
		t.FailNow()
	}
	if len(models) != 2 {
		t.Error("models doesn't contains 2 elements")
	}
	if _, found := models[id]; !found {
		t.Error("models doesn't contains element 'test_id'")
	}
	if _, found := models[id2]; !found {
		t.Error("models doesn't contains element 'test_id2'")
	}
	model1 := models[id]
	if model1.ID != id || model1.Synonyms[0] != "model_synonym" {
		t.Error("model 'test_id' has been modified")
	}
	model2 := models[id2]
	if model2.ID != id2 || model2.Synonyms[0] != "model_synonym" {
		t.Error("model 'test_id2' has been modified")
	}
}
