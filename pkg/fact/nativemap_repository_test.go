package fact

import (
	"testing"

	"github.com/myrteametrics/myrtea-sdk/v5/engine"
)

func TestNew(t *testing.T) {
	r := NewNativeMapRepository()
	if r == nil {
		t.Error("Fact Repository is nil")
	}
}

func TestReplaceGlobal(t *testing.T) {
	r := NewNativeMapRepository()
	reverse := ReplaceGlobals(r)
	if R() == nil {
		t.Error("Global fact repository is nil")
	}
	reverse()
	if R() != nil {
		t.Error("Global fact repository is not nil after reverse")
	}
}

func TestCreate(t *testing.T) {
	var err error
	r := NewNativeMapRepository()
	fact := engine.Fact{Name: "test_name", Comment: "test comment"}
	id, err := r.Create(fact)
	if err != nil {
		t.Error(err)
	}
	if id != 1 {
		t.Error("invalid generated fact id")
	}

	factGet, found, err := r.Get(id)
	if err != nil {
		t.Error(err)
	}
	if !found {
		t.Error("fact not found")
	}
	if id != factGet.ID {
		t.Error("invalid fact ID")
	}
	if fact.Comment != factGet.Comment {
		t.Error("invalid fact Comment")
	}
}

func TestGet(t *testing.T) {
	var err error
	r := NewNativeMapRepository()

	factGet, found, err := r.Get(1)
	if err != nil {
		t.Error(err)
	}
	if found {
		t.Error("found a fact from nowhere")
	}

	fact := engine.Fact{Name: "test_name", Comment: "test comment"}
	id, err := r.Create(fact)
	if err != nil {
		t.Error(err)
	}
	factGet, found, err = r.Get(id)
	if err != nil {
		t.Error(err)
	}
	if !found {
		t.Error("fact not found")
	}
	if id != factGet.ID {
		t.Error("invalid fact ID")
	}
	if fact.Comment != factGet.Comment {
		t.Error("invalid fact Comment")
	}
}

func TestUpdate(t *testing.T) {
	var err error
	r := NewNativeMapRepository()
	id, err := r.Create(engine.Fact{Name: "test_name", Comment: "test comment"})
	if err != nil {
		t.Error(err)
	}

	// Update existing
	fact := engine.Fact{Name: "test_name", Comment: "test comment 2"}
	err = r.Update(id, fact)
	if err != nil {
		t.Error(err)
	}
	factGet, found, err := r.Get(id)
	if err != nil {
		t.Error(err)
	}
	if !found {
		t.Error("fact not found")
	}
	if id != factGet.ID {
		t.Error("invalid fact ID")
	}
	if fact.Comment != factGet.Comment {
		t.Error("invalid fact Comment")
	}
}

func TestUpdateNotExists(t *testing.T) {
	var err error
	r := NewNativeMapRepository()
	id, err := r.Create(engine.Fact{Name: "test_name", Comment: "test comment"})
	if err != nil {
		t.Error(err)
	}

	fact := engine.Fact{Name: "test_name_2", Comment: "test comment 2"}
	err = r.Update(id+1, fact)
	if err == nil {
		t.Error("Fact doesn't exists and cannot be updated")
	}
	_, found, err := r.Get(id + 1)
	if err != nil {
		t.Error(err)
	}
	if found {
		t.Error("Fact should not have been created")
	}
}

func TestDelete(t *testing.T) {
	var err error
	r := NewNativeMapRepository()
	id, err := r.Create(engine.Fact{Name: "test_name", Comment: "test comment"})
	if err != nil {
		t.Error(err)
	}
	id2, err := r.Create(engine.Fact{Name: "test_name2", Comment: "test comment 2"})
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
		t.Error("fact has not been deleted")
	}
	_, found, err = r.Get(id2)
	if err != nil {
		t.Error(err)
	}
	if !found {
		t.Error("fact2 has been deleted while it should not")
	}
}

func TestDeleteNotExists(t *testing.T) {
	var err error
	r := NewNativeMapRepository()
	id, err := r.Create(engine.Fact{Name: "test_name", Comment: "test comment"})
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
		t.Error("fact has been deleted while it should not")
	}
}

func TestGetAll(t *testing.T) {
	var err error
	r := NewNativeMapRepository()
	id, err := r.Create(engine.Fact{Name: "test_name", Comment: "test comment"})
	if err != nil {
		t.Error(err)
	}
	id2, err := r.Create(engine.Fact{Name: "test_name2", Comment: "test comment 2"})
	if err != nil {
		t.Error(err)
	}

	facts, err := r.GetAll()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if facts == nil {
		t.Error("facts is nil")
		t.FailNow()
	}
	if len(facts) != 2 {
		t.Error("facts doesn't contains 2 elements")
	}
	if _, found := facts[id]; !found {
		t.Error("facts doesn't contains element 'test_id'")
	}
	if _, found := facts[id2]; !found {
		t.Error("facts doesn't contains element 'test_id2'")
	}
	fact1 := facts[id]
	if fact1.ID != id || fact1.Comment != "test comment" {
		t.Error("fact 'test_id' has been modified")
	}
	fact2 := facts[id2]
	if fact2.ID != id2 || fact2.Comment != "test comment 2" {
		t.Error("fact 'test_id2' has been modified")
	}
}
