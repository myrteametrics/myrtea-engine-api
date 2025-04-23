package situation

import (
	"testing"

	"github.com/myrteametrics/myrtea-engine-api/v5/internal/fact"
	"github.com/myrteametrics/myrtea-sdk/v5/engine"
)

func TestBuildSituationsFromFileNoFile(t *testing.T) {
	_, errs := BuildSituationsFromFile("testdata", "not_a_file")
	if len(errs) == 0 {
		t.Error("file doesn't exists but no error returned")
	}
}

func TestBuildSituationsFromFileValid(t *testing.T) {
	fr := fact.NewNativeMapRepository()
	fact.ReplaceGlobals(fr)

	fr.Create(engine.Fact{Name: "fact_test_1"})
	fr.Create(engine.Fact{Name: "fact_test_2"})
	fr.Create(engine.Fact{Name: "fact_test_3"})

	_, errs := BuildSituationsFromFile("testdata", "situations")
	if errs != nil {
		t.Error(errs)
	}
}

func TestBuildSituationsFactNotFound(t *testing.T) {
	fr := fact.NewNativeMapRepository()
	fact.ReplaceGlobals(fr)

	_, errs := BuildSituationsFromFile("testdata", "situations")
	if errs == nil || len(errs) == 0 {
		t.Error("File should not be loadable (missing facts)")
	}
}
