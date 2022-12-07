package externalconfig

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/models"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/tests"
	"testing"
)

func dbInitRepo(dbClient *sqlx.DB, t *testing.T) {
	dbDestroyRepo(dbClient, t)

	_, err := dbClient.Exec(tests.ExternalGenericConfigV1)
	if err != nil {
		t.Error(err)
	}
}

func dbDestroyRepo(dbClient *sqlx.DB, t *testing.T) {
	_, err := dbClient.Exec(tests.ExternalGenericConfigDropTableV1)
	if err != nil {
		t.Error(err)
	}
}

func TestNewPostgresRepository(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	r := NewPostgresRepository(tests.DBClient(t))
	if r == nil {
		t.Error("situation Repository is nil")
	}
}

func TestPostgresReplaceGlobal(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	r := NewPostgresRepository(tests.DBClient(t))
	reverse := ReplaceGlobals(r)
	if R() == nil {
		t.Error("Global situation repository is nil")
	}
	reverse()
	if R() != nil {
		t.Error("Global situation repository is not nil after reverse")
	}
}

func TestPostgresGet(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroyRepo(db, t)
	dbInitRepo(db, t)
	r := NewPostgresRepository(db)

	var err error

	externalConfigGet, found, err := r.Get(1)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if found {
		t.Error("found a externalConfig from nowhere")
	}

	externalConfig := models.ExternalConfig{Name: "test_name", Data: "{\"test\": \"test\"}"}
	id, err := r.Create(externalConfig)
	if err != nil {
		t.Error(err)
	}

	externalConfigGet, found, err = r.Get(id)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if !found {
		t.Error("ExternalConfig doesn't exists after the creation")
		t.FailNow()
	}
	if id != externalConfigGet.Id {
		t.Error("invalid ExternalConfig ID")
	}
}

func TestPostgresGetByName(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroyRepo(db, t)
	dbInitRepo(db, t)
	r := NewPostgresRepository(db)

	var err error

	externalConfigGet, found, err := r.GetByName("test")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if found {
		t.Error("found a externalConfig from nowhere")
	}

	externalConfig := models.ExternalConfig{Name: "test", Data: "{\"test\": \"test\"}"}
	id, err := r.Create(externalConfig)
	if err != nil {
		t.Error(err)
	}

	externalConfigGet, found, err = r.GetByName("test")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if !found {
		t.Error("ExternalConfig doesn't exists after the creation")
		t.FailNow()
	}
	if id != externalConfigGet.Id {
		t.Error("invalid ExternalConfig ID")
	}
}

func TestPostgresCreate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroyRepo(db, t)
	dbInitRepo(db, t)
	r := NewPostgresRepository(db)

	externalConfig := models.ExternalConfig{Name: "test_name", Data: "{\"test\": \"test\"}"}
	id, err := r.Create(externalConfig)
	if err != nil {
		t.Error(err)
	}
	externalConfigGet, found, err := r.Get(id)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if !found {
		t.Error("externalConfig not found")
		t.FailNow()
	}
	if externalConfigGet.Id != id || externalConfigGet.Name != externalConfig.Name ||
		externalConfigGet.Data != externalConfig.Data {
		t.Error("invalid externalConfig value")
	}

}

func TestPostgresCreateMultiple(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroyRepo(db, t)
	dbInitRepo(db, t)
	r := NewPostgresRepository(db)

	var err error
	externalConfig := models.ExternalConfig{Name: "test1", Data: "{\"test1\": \"test1\"}"}
	id1, err := r.Create(externalConfig)
	if err != nil {
		t.Error(err)
	}
	externalConfigGet, found, err := r.Get(id1)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if !found {
		t.Error("externalConfig not found")
		t.FailNow()
	}
	if id1 != externalConfigGet.Id {
		t.Error("invalid ID")
	}

	externalConfig2 := models.ExternalConfig{Name: "test2", Data: "{\"test2\": \"test2\"}"}
	id2, err := r.Create(externalConfig2)
	if err != nil {
		t.Error(err)
	}
	externalConfig2Get, found, err := r.Get(id2)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if !found {
		t.Error("externalConfig not found")
		t.FailNow()
	}
	if externalConfig2Get.Id != id2 {
		t.Error("invalid ID")
	}
}

func TestPostgresUpdate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroyRepo(db, t)
	dbInitRepo(db, t)
	r := NewPostgresRepository(db)

	var err error

	externalConfig := models.ExternalConfig{Name: "test1", Data: "{\"test1\": \"test1\"}"}
	id, err := r.Create(externalConfig)
	if err != nil {
		t.Error(err)
	}
	externalConfig2 := models.ExternalConfig{Name: "test2", Data: "{\"test2\": \"test2\"}"}
	err = r.Update(id, externalConfig2)
	if err != nil {
		t.Error(err)
	}
	externalConfigGet, found, err := r.Get(id)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if !found {
		t.Error("externalConfig not found")
		t.FailNow()
	}
	if externalConfigGet.Name != "test2" || externalConfigGet.Data != "{\"test2\": \"test2\"}" {
		t.Error("Couldn't update the externalConfig")
	}
}

func TestPostgresUpdateNotExists(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroyRepo(db, t)
	dbInitRepo(db, t)
	r := NewPostgresRepository(db)

	var err error
	externalConfig := models.ExternalConfig{Name: "test1", Data: "{\"test1\": \"test1\"}"}
	err = r.Update(1, externalConfig)
	if err == nil {
		t.Error("updating a non-existing externalConfig should return an error")
	}
}

func TestPostgresDelete(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroyRepo(db, t)
	dbInitRepo(db, t)
	r := NewPostgresRepository(db)

	var err error

	externalConfig := models.ExternalConfig{Name: "test1", Data: "{\"test1\": \"test1\"}"}
	id, err := r.Create(externalConfig)
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
		t.FailNow()
	}
	if found {
		t.Error("ExternalConfig should not exists")
	}
}

func TestPostgresDeleteNotExists(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroyRepo(db, t)
	dbInitRepo(db, t)
	r := NewPostgresRepository(db)

	err := r.Delete(1)
	if err == nil {
		t.Error("Cannot delete a non-existing externalConfig")
	}

	_, found, err := r.Get(1)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if found {
		t.Error("externalConfig should not exists")
	}
}

func TestPostgresGetAll(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroyRepo(db, t)
	dbInitRepo(db, t)
	r := NewPostgresRepository(db)

	externalConfig := models.ExternalConfig{Name: "test1", Data: "{\"test1\": \"test1\"}"}
	e1ID, err := r.Create(externalConfig)
	if err != nil {
		t.Error(err)
	}
	externalConfig2 := models.ExternalConfig{Name: "test2", Data: "{\"test2\": \"test2\"}"}
	e2ID, err := r.Create(externalConfig2)
	if err != nil {
		t.Error(err)
	}

	externalConfigs, err := r.GetAll()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if len(externalConfigs) != 2 {
		t.Error("wrong externalConfigs count")
	}
	if externalConfigs[e1ID].Name != externalConfig.Name || externalConfigs[e1ID].Data != externalConfig.Data {
		t.Error("The situation " + fmt.Sprint(e1ID) + " is not as expected")
	}
	if externalConfigs[e2ID].Name != externalConfig2.Name || externalConfigs[e2ID].Data != externalConfig2.Data {
		t.Error("The situation " + fmt.Sprint(e2ID) + " is not as expected")
	}
}

func TestPostgresGetAllEmpty(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroyRepo(db, t)
	dbInitRepo(db, t)
	r := NewPostgresRepository(db)

	externalConfigs, err := r.GetAll()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if externalConfigs == nil {
		t.Error("externalConfigs should not be nil")
	}
	if len(externalConfigs) != 0 {
		t.Error("wrong externalConfigs count")
	}
}
