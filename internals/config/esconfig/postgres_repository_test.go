package esconfig

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/models"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/tests"
	"testing"
)

func dbInitRepo(dbClient *sqlx.DB, t *testing.T) {
	dbDestroyRepo(dbClient, t)

	_, err := dbClient.Exec(tests.ElasticSearchConfigV1)
	if err != nil {
		t.Error(err)
	}
}

func dbDestroyRepo(dbClient *sqlx.DB, t *testing.T) {
	_, err := dbClient.Exec(tests.ElasticSearchConfigDropTableV1)
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

	esConfigGet, found, err := r.Get(1)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if found {
		t.Error("found a elasticSearchConfig from nowhere")
	}

	elasticSearchConfig := models.ElasticSearchConfig{Name: "test_name", URLs: []string{"http://localhost:9200"}, Default: false}
	id, err := r.Create(elasticSearchConfig)
	if err != nil {
		t.Error(err)
	}

	esConfigGet, found, err = r.Get(id)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if !found {
		t.Error("ElasticSearchConfig doesn't exists after the creation")
		t.FailNow()
	}
	if id != esConfigGet.Id {
		t.Error("invalid ElasticSearchConfig ID")
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

	esConfigGet, found, err := r.GetByName("test_name")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if found {
		t.Error("found a elasticSearchConfig from nowhere")
	}

	elasticSearchConfig := models.ElasticSearchConfig{Name: "test_name", URLs: []string{"http://localhost:9200"}, Default: false}
	id, err := r.Create(elasticSearchConfig)
	if err != nil {
		t.Error(err)
	}

	esConfigGet, found, err = r.GetByName("test_name")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if !found {
		t.Error("ElasticSearchConfig doesn't exists after the creation")
		t.FailNow()
	}
	if id != esConfigGet.Id {
		t.Error("invalid ElasticSearchConfig ID")
	}
}

func TestPostgresGetDefault(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroyRepo(db, t)
	dbInitRepo(db, t)
	r := NewPostgresRepository(db)

	esConfigGet, found, err := r.GetDefault()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if found {
		t.Error("found a elasticSearchConfig from nowhere")
	}

	elasticSearchConfig := models.ElasticSearchConfig{Name: "test_name", URLs: []string{"http://localhost:9200"}, Default: true}
	id, err := r.Create(elasticSearchConfig)
	if err != nil {
		t.Error(err)
	}

	esConfigGet, found, err = r.GetDefault()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if !found {
		t.Error("ElasticSearchConfig doesn't exists after the creation")
		t.FailNow()
	}
	if id != esConfigGet.Id {
		t.Error("invalid ElasticSearchConfig ID")
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

	elasticSearchConfig := models.ElasticSearchConfig{Name: "test_name", URLs: []string{"http://localhost:9200"}, Default: false}
	id, err := r.Create(elasticSearchConfig)
	if err != nil {
		t.Error(err)
	}
	esConfigGet, found, err := r.Get(id)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if !found {
		t.Error("elasticSearchConfig not found")
		t.FailNow()
	}
	if esConfigGet.Id != id || esConfigGet.Name != elasticSearchConfig.Name ||
		esConfigGet.Default != elasticSearchConfig.Default ||
		len(esConfigGet.URLs) != len(elasticSearchConfig.URLs) ||
		esConfigGet.URLs[0] != elasticSearchConfig.URLs[0] {
		t.Error("invalid elasticSearchConfig value")
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
	elasticSearchConfig := models.ElasticSearchConfig{Name: "test_name", URLs: []string{"http://localhost:9200"}, Default: false}
	id1, err := r.Create(elasticSearchConfig)
	if err != nil {
		t.Error(err)
	}
	esConfigGet, found, err := r.Get(id1)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if !found {
		t.Error("elasticSearchConfig not found")
		t.FailNow()
	}
	if id1 != esConfigGet.Id {
		t.Error("invalid ID")
	}

	elasticSearchConfig2 := models.ElasticSearchConfig{Name: "test_name2", URLs: []string{"http://localhost:9201"}, Default: false}
	id2, err := r.Create(elasticSearchConfig2)
	if err != nil {
		t.Error(err)
	}
	elasticSearchConfig2Get, found, err := r.Get(id2)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if !found {
		t.Error("elasticSearchConfig not found")
		t.FailNow()
	}
	if elasticSearchConfig2Get.Id != id2 {
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

	elasticSearchConfig := models.ElasticSearchConfig{Name: "test_name", URLs: []string{"http://localhost:9200"}, Default: false}
	id, err := r.Create(elasticSearchConfig)
	if err != nil {
		t.Error(err)
	}
	elasticSearchConfig2 := models.ElasticSearchConfig{Name: "test_name2", URLs: []string{"http://localhost:9200"}, Default: false}
	err = r.Update(id, elasticSearchConfig2)
	if err != nil {
		t.Error(err)
	}
	esConfigGet, found, err := r.Get(id)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if !found {
		t.Error("elasticSearchConfig not found")
		t.FailNow()
	}
	if esConfigGet.Name != "test_name2" || esConfigGet.Default != elasticSearchConfig2.Default ||
		len(esConfigGet.URLs) != len(elasticSearchConfig2.URLs) ||
		esConfigGet.URLs[0] != elasticSearchConfig2.URLs[0] {
		t.Error("Couldn't update the elasticSearchConfig")
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
	elasticSearchConfig := models.ElasticSearchConfig{Name: "test_name", URLs: []string{"http://localhost:9200"}, Default: false}
	err = r.Update(1, elasticSearchConfig)
	if err == nil {
		t.Error("updating a non-existing elasticSearchConfig should return an error")
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

	elasticSearchConfig := models.ElasticSearchConfig{Name: "test_name", URLs: []string{"http://localhost:9200"}, Default: false}
	id, err := r.Create(elasticSearchConfig)
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
		t.Error("ElasticSearchConfig should not exists")
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
		t.Error("Cannot delete a non-existing elasticSearchConfig")
	}

	_, found, err := r.Get(1)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if found {
		t.Error("elasticSearchConfig should not exists")
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

	elasticSearchConfig := models.ElasticSearchConfig{Name: "test1", URLs: []string{"http://localhost:9200"}, Default: false}
	e1ID, err := r.Create(elasticSearchConfig)
	if err != nil {
		t.Error(err)
	}
	elasticSearchConfig2 := models.ElasticSearchConfig{Name: "test2", URLs: []string{"http://localhost:9201"}, Default: true}
	e2ID, err := r.Create(elasticSearchConfig2)
	if err != nil {
		t.Error(err)
	}

	elasticSearchConfigs, err := r.GetAll()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if len(elasticSearchConfigs) != 2 {
		t.Error("wrong elasticSearchConfigs count")
	}
	if elasticSearchConfigs[e1ID].Name != elasticSearchConfig.Name ||
		elasticSearchConfigs[e1ID].Default != elasticSearchConfig.Default ||
		len(elasticSearchConfigs[e1ID].URLs) != len(elasticSearchConfig.URLs) ||
		elasticSearchConfigs[e1ID].URLs[0] != elasticSearchConfig.URLs[0] {
		t.Error("The elasticSearchConfig " + fmt.Sprint(e1ID) + " is not as expected")
	}
	if elasticSearchConfigs[e2ID].Name != elasticSearchConfig2.Name ||
		elasticSearchConfigs[e2ID].Default != elasticSearchConfig2.Default ||
		len(elasticSearchConfigs[e2ID].URLs) != len(elasticSearchConfig2.URLs) ||
		elasticSearchConfigs[e2ID].URLs[0] != elasticSearchConfig2.URLs[0] {
		t.Error("The elasticSearchConfig " + fmt.Sprint(e2ID) + " is not as expected")
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

	elasticSearchConfigs, err := r.GetAll()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if elasticSearchConfigs == nil {
		t.Error("elasticSearchConfigs should not be nil")
	}
	if len(elasticSearchConfigs) != 0 {
		t.Error("wrong elasticSearchConfigs count")
	}
}
