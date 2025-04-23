package variablesconfig

import (
	"fmt"
	"strings"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/models"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/tests"
)

func dbInitRepo(dbClient *sqlx.DB, t *testing.T) {
	dbDestroyRepo(dbClient, t)

	_, err := dbClient.Exec(tests.VariablesConfigV1)
	if err != nil {
		t.Error(err)
	}
}

func dbDestroyRepo(dbClient *sqlx.DB, t *testing.T) {
	_, err := dbClient.Exec(tests.VariablesConfigV1DropTable)
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

	variableConfigGet, found, err := r.Get(1)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if found {
		t.Error("found a variable Config from nowhere")
	}

	variablesconfig := models.VariablesConfig{Key: "test_name", Value: "test_value"}
	id, err := r.Create(variablesconfig)
	if err != nil {
		t.Error(err)
	}

	variableConfigGet, found, err = r.Get(id)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if !found {
		t.Error("variablesconfig doesn't exists after the creation")
		t.FailNow()
	}
	if id != variableConfigGet.Id {
		t.Error("invalid variablesconfig ID")
	}
}

func TestPostgresGetByKey(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroyRepo(db, t)
	dbInitRepo(db, t)
	r := NewPostgresRepository(db)

	var err error

	variableConfigGet, found, err := r.GetByKey("test")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if found {
		t.Error("found a variableConfig from nowhere")
	}

	variableConfig := models.VariablesConfig{Key: "test_key", Value: "test_value"}
	id, err := r.Create(variableConfig)
	if err != nil {
		t.Error(err)
	}

	variableConfigGet, found, err = r.GetByKey("test_key")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if !found {
		t.Error("variableConfig doesn't exists after the creation")
		t.FailNow()
	}
	if id != variableConfigGet.Id {
		t.Error("invalid variableConfig ID")
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

	variableConfig := models.VariablesConfig{Key: "test_key", Value: "test_value"}
	id, err := r.Create(variableConfig)
	if err != nil {
		t.Error(err)
	}
	variableConfigGet, found, err := r.Get(id)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if !found {
		t.Error("variable Config not found")
		t.FailNow()
	}
	if variableConfigGet.Id != id || variableConfigGet.Key != variableConfig.Key ||
		variableConfigGet.Value != variableConfig.Value {
		t.Error("invalid variable Config value")
	}

}

func TestPostgresCreateDuplicate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroyRepo(db, t)
	dbInitRepo(db, t)
	r := NewPostgresRepository(db)

	variableConfig := models.VariablesConfig{Key: "test_key", Value: "test_value"}
	_, err := r.Create(variableConfig)
	if err != nil {
		t.Error(err)
	}

	variableConfig2 := models.VariablesConfig{Key: "test_key", Value: "test_value2"}
	_, err = r.Create(variableConfig2)
	if err != nil {
		if !strings.Contains(err.Error(), "pq: duplicate key value violates unique constraint") {
			t.Error(err)
		}
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
	variableConfig := models.VariablesConfig{Key: "test1key", Value: "test_Value"}
	id1, err := r.Create(variableConfig)
	if err != nil {
		t.Error(err)
	}
	variableConfigGet, found, err := r.Get(id1)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if !found {
		t.Error("variable Config not found")
		t.FailNow()
	}
	if id1 != variableConfigGet.Id {
		t.Error("invalid ID")
	}

	variableConfig2 := models.VariablesConfig{Key: "test2key", Value: "test_value"}
	id2, err := r.Create(variableConfig2)
	if err != nil {
		t.Error(err)
	}
	variableConfig2Get, found, err := r.Get(id2)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if !found {
		t.Error("variableConfig not found")
		t.FailNow()
	}
	if variableConfig2Get.Id != id2 {
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

	variableConfig := models.VariablesConfig{Key: "test1key", Value: "test1_value"}
	id, err := r.Create(variableConfig)
	if err != nil {
		t.Error(err)
	}
	variableConfig2 := models.VariablesConfig{Key: "test2key", Value: "test2_value"}
	err = r.Update(id, variableConfig2)
	if err != nil {
		t.Error(err)
	}
	variableConfigGet, found, err := r.Get(id)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if !found {
		t.Error("variable Config not found")
		t.FailNow()
	}
	if variableConfigGet.Key != "test2key" || variableConfigGet.Value != "test2_value" {
		t.Error("Couldn't update the variable Config")
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
	variableConfig := models.VariablesConfig{Key: "test2key", Value: "test_value"}
	err = r.Update(1, variableConfig)
	if err == nil {
		t.Error("updating a non-existing variable Config should return an error")
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

	variableConfig := models.VariablesConfig{Key: "test2key", Value: "test_value"}
	id, err := r.Create(variableConfig)
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
		t.Error("variable Config should not exists")
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
		t.Error("Cannot delete a non-existing variable Config")
	}

	_, found, err := r.Get(1)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if found {
		t.Error("variable Config should not exists")
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

	variableConfig := models.VariablesConfig{Key: "test1key", Value: "test1_value"}
	e1ID, err := r.Create(variableConfig)
	if err != nil {
		t.Error(err)
	}
	variableConfig2 := models.VariablesConfig{Key: "test2key", Value: "test2_value"}
	e2ID, err := r.Create(variableConfig2)
	if err != nil {
		t.Error(err)
	}

	variableConfigs, err := r.GetAll()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if len(variableConfigs) != 2 {
		t.Error("wrong variableConfigs count")
	}
	if variableConfigs[e1ID].Key != variableConfig.Key || variableConfigs[e1ID].Value != variableConfig.Value {
		t.Error("The Variable " + fmt.Sprint(e1ID) + " is not as expected")
	}
	if variableConfigs[e2ID].Key != variableConfig2.Key || variableConfigs[e2ID].Value != variableConfig2.Value {
		t.Error("The Variable " + fmt.Sprint(e2ID) + " is not as expected")
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

	variableConfigs, err := r.GetAll()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if variableConfigs == nil {
		t.Error("variable Configs should not be nil")
	}
	if len(variableConfigs) != 0 {
		t.Error("wrong variable Configs count")
	}
}

func TestPostgresGetAllAsMap(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroyRepo(db, t)
	dbInitRepo(db, t)
	r := NewPostgresRepository(db)

	variableConfig := models.VariablesConfig{Key: "test1key", Value: "test1_value"}
	_, err := r.Create(variableConfig)
	if err != nil {
		t.Error(err)
	}
	variableConfig2 := models.VariablesConfig{Key: "test2key", Value: "test2_value"}
	_, err = r.Create(variableConfig2)
	if err != nil {
		t.Error(err)
	}

	variableConfigs, err := r.GetAllAsMap()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if len(variableConfigs) != 2 {
		t.Error("wrong variableConfigs count")
	}
	if variableConfigs[variableConfig.Key] != variableConfig.Value {
		t.Error("The Variable " + fmt.Sprint(variableConfig.Key) + " is not as expected")
	}
	if variableConfigs[variableConfig2.Key] != variableConfig2.Value {
		t.Error("The Variable " + fmt.Sprint(variableConfig2.Key) + " is not as expected")
	}
}
