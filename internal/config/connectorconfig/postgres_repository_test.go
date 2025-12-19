package connectorconfig

import (
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/model"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/tests"
	"github.com/myrteametrics/myrtea-sdk/v5/repositories/utils"
)

// SQL scripts for testing
const (
	// ConnectorsConfigDropTableV1 SQL statement for table drop
	ConnectorsConfigDropTableV1 string = `DROP TABLE IF EXISTS ` + table + `;`
	// ConnectorsConfigTableV1 SQL statement for the connectors_config table
	ConnectorsConfigTableV1 string = `CREATE TABLE IF NOT EXISTS ` + table + `
	(
		id            serial PRIMARY KEY NOT NULL,
		connector_id  varchar(100)       NOT NULL UNIQUE,
		name          varchar(100)       NOT NULL,
		current       text               NOT NULL,
		previous      text,
		last_modified timestamptz        NOT NULL DEFAULT current_timestamp
	);`
)

func dbInitRepo(dbClient *sqlx.DB, t *testing.T) {
	dbDestroyRepo(dbClient, t)

	_, err := dbClient.Exec(ConnectorsConfigTableV1)
	if err != nil {
		t.Error(err)
	}
}

func dbDestroyRepo(dbClient *sqlx.DB, t *testing.T) {
	_, err := dbClient.Exec(ConnectorsConfigDropTableV1)
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
		t.Error("ConnectorConfig Repository is nil")
	}
}

func TestPostgresReplaceGlobal(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	r := NewPostgresRepository(tests.DBClient(t))
	reverse := ReplaceGlobals(r)
	if R() == nil {
		t.Error("Global ConnectorConfig repository is nil")
	}
	reverse()
	if R() != nil {
		t.Error("Global ConnectorConfig repository is not nil after reverse")
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

	connectorConfigGet, found, err := r.Get(1)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if found {
		t.Error("found a connectorConfig from nowhere")
	}

	connectorConfig := model.ConnectorConfig{Name: "test_name", ConnectorId: "test_connector", Current: "test_current"}
	id, err := r.Create(nil, connectorConfig)
	if err != nil {
		t.Error(err)
	}

	connectorConfigGet, found, err = r.Get(id)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if !found {
		t.Error("ConnectorConfig doesn't exists after the creation")
		t.FailNow()
	}
	if id != connectorConfigGet.Id {
		t.Error("invalid ConnectorConfig ID")
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

	var err error

	connectorConfig := model.ConnectorConfig{Name: "test_name", ConnectorId: "test_connector", Current: "test_current"}
	id, err := r.Create(nil, connectorConfig)
	if err != nil {
		t.Error(err)
	}
	if id <= 0 {
		t.Error("invalid ConnectorConfig ID")
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

	connectorConfig1 := model.ConnectorConfig{Name: "test_name1", ConnectorId: "test_connector1", Current: "test_current1"}
	id1, err := r.Create(nil, connectorConfig1)
	if err != nil {
		t.Error(err)
	}
	if id1 <= 0 {
		t.Error("invalid ConnectorConfig ID")
	}

	connectorConfig2 := model.ConnectorConfig{Name: "test_name2", ConnectorId: "test_connector2", Current: "test_current2"}
	id2, err := r.Create(nil, connectorConfig2)
	if err != nil {
		t.Error(err)
	}
	if id2 <= 0 {
		t.Error("invalid ConnectorConfig ID")
	}

	if id1 == id2 {
		t.Error("IDs should be different")
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

	connectorConfig := model.ConnectorConfig{Name: "test_name", ConnectorId: "test_connector", Current: "test_current"}
	id, err := r.Create(nil, connectorConfig)
	if err != nil {
		t.Error(err)
	}

	connectorConfigGet, found, err := r.Get(id)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if !found {
		t.Error("ConnectorConfig doesn't exists after the creation")
		t.FailNow()
	}

	connectorConfigGet.Name = "test_name_updated"
	connectorConfigGet.Current = "test_current_updated"

	err = r.Update(nil, id, connectorConfigGet)
	if err != nil {
		t.Error(err)
	}

	connectorConfigUpdated, found, err := r.Get(id)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if !found {
		t.Error("ConnectorConfig doesn't exists after the update")
		t.FailNow()
	}

	if connectorConfigUpdated.Name != "test_name_updated" {
		t.Error("ConnectorConfig name not updated")
	}
	if connectorConfigUpdated.Current != "test_current_updated" {
		t.Error("ConnectorConfig current not updated")
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

	connectorConfig := model.ConnectorConfig{Name: "test_name", ConnectorId: "test_connector", Current: "test_current"}
	err = r.Update(nil, 1, connectorConfig)
	if err == nil {
		t.Error("Should not be able to update a non-existing ConnectorConfig")
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

	connectorConfig := model.ConnectorConfig{Name: "test_name", ConnectorId: "test_connector", Current: "test_current"}
	id, err := r.Create(nil, connectorConfig)
	if err != nil {
		t.Error(err)
	}

	_, found, err := r.Get(id)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if !found {
		t.Error("ConnectorConfig doesn't exists after the creation")
		t.FailNow()
	}

	err = r.Delete(nil, id)
	if err != nil {
		t.Error(err)
	}

	_, found, err = r.Get(id)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if found {
		t.Error("ConnectorConfig still exists after deletion")
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

	var err error

	err = r.Delete(nil, 1)
	if err == nil {
		t.Error("Should not be able to delete a non-existing ConnectorConfig")
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

	var err error

	connectorConfigs, err := r.GetAll()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if len(connectorConfigs) != 0 {
		t.Error("ConnectorConfigs should be empty")
	}

	connectorConfig1 := model.ConnectorConfig{Name: "test_name1", ConnectorId: "test_connector1", Current: "test_current1"}
	id1, err := r.Create(nil, connectorConfig1)
	if err != nil {
		t.Error(err)
	}

	connectorConfig2 := model.ConnectorConfig{Name: "test_name2", ConnectorId: "test_connector2", Current: "test_current2"}
	id2, err := r.Create(nil, connectorConfig2)
	if err != nil {
		t.Error(err)
	}

	connectorConfigs, err = r.GetAll()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if len(connectorConfigs) != 2 {
		t.Error("ConnectorConfigs should have 2 elements")
	}

	if _, ok := connectorConfigs[id1]; !ok {
		t.Error("ConnectorConfig 1 should be in the map")
	}
	if _, ok := connectorConfigs[id2]; !ok {
		t.Error("ConnectorConfig 2 should be in the map")
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

	var err error

	connectorConfigs, err := r.GetAll()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if len(connectorConfigs) != 0 {
		t.Error("ConnectorConfigs should be empty")
	}
}

func TestPostgresRefreshNextIdGen(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroyRepo(db, t)
	dbInitRepo(db, t)
	r := NewPostgresRepository(db).(*PostgresRepository)

	var err error

	connectorConfig := model.ConnectorConfig{Name: "test_name", ConnectorId: "test_connector", Current: "test_current"}
	id, err := r.Create(nil, connectorConfig)
	if err != nil {
		t.Error(err)
	}

	nextId, found, err := utils.RefreshNextIdGen(r.conn.DB, table)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if !found {
		t.Error("Next ID not found")
	}
	if nextId != id+1 {
		t.Errorf("Next ID should be %d, got %d", id+1, nextId)
	}
}
