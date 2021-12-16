package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/connector"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/security/users"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/tests"
)

func dbConnectorInit(dbClient *sqlx.DB, t *testing.T) {
	dbConnectorDestroy(dbClient, t)
	tests.DBExec(dbClient, tests.ConnectorExecutionsLogsTableV1, t, true)
}

func dbConnectorDestroy(dbClient *sqlx.DB, t *testing.T) {
	tests.DBExec(dbClient, tests.ConnectorExecutionsLogsDropTableV1, t, true)
}

func TestGetlastConnectorExecutionDateTime(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	db := tests.DBClient(t)

	connector.ReplaceGlobals(connector.NewPostgresRepository(db))
	defer dbConnectorDestroy(db, t)
	dbConnectorInit(db, t)

	var ts time.Time

	for i := 1; i <= 5; i++ {
		ts = time.Now().UTC()
		logConnectorExecution(t, db, "connector_1", "connection_1", ts, true)
		logConnectorExecution(t, db, "connector_1", "connection_2", ts, true)
		logConnectorExecution(t, db, "connector_2", "connection_1", ts, true)
		logConnectorExecution(t, db, "connector_2", "connection_2", ts, true)
	}

	expectedMap := map[string]string{
		"connection_1": ts.Format("2006-01-02T15:04:05.999999Z"),
		"connection_2": ts.Format("2006-01-02T15:04:05.999999Z"),
	}
	expectedResult, _ := json.Marshal(expectedMap)

	ts1 := time.Now().UTC()
	logConnectorExecution(t, db, "connector_1", "connection_1", ts1, false)
	logConnectorExecution(t, db, "connector_1", "connection_2", ts1, false)
	logConnectorExecution(t, db, "connector_2", "connection_1", ts1, false)
	logConnectorExecution(t, db, "connector_2", "connection_2", ts1, false)

	rr := tests.BuildTestHandler(t, "GET", "/connector/connector_1/executions/last?successOnly=true", "", "/connector/{id}/executions/last", GetlastConnectorExecutionDateTime, users.UserWithPermissions{})
	tests.CheckTestHandler(t, rr, http.StatusOK, string(expectedResult)+"\n")
	t.Log(rr.Body.String())

	rr = tests.BuildTestHandler(t, "GET", "/connector/connector_2/executions/last?successOnly=true", "", "/connector/{id}/executions/last", GetlastConnectorExecutionDateTime, users.UserWithPermissions{})
	tests.CheckTestHandler(t, rr, http.StatusOK, string(expectedResult)+"\n")
	t.Log(rr.Body.String())

	expectedMap = map[string]string{
		"connection_1": ts1.Format("2006-01-02T15:04:05.999999Z"),
		"connection_2": ts1.Format("2006-01-02T15:04:05.999999Z"),
	}
	expectedResult, _ = json.Marshal(expectedMap)

	rr = tests.BuildTestHandler(t, "GET", "/connector/connector_1/executions/last", "", "/connector/{id}/executions/last", GetlastConnectorExecutionDateTime, users.UserWithPermissions{})
	tests.CheckTestHandler(t, rr, http.StatusOK, string(expectedResult)+"\n")
	t.Log(rr.Body.String())

	rr = tests.BuildTestHandler(t, "GET", "/connector/connector_2/executions/last?successOnly=false", "", "/connector/{id}/executions/last", GetlastConnectorExecutionDateTime, users.UserWithPermissions{})
	tests.CheckTestHandler(t, rr, http.StatusOK, string(expectedResult)+"\n")
	t.Log(rr.Body.String())
}

func logConnectorExecution(t *testing.T, dbClient *sqlx.DB, connectorID string, connectionName string, ts time.Time, success bool) {
	query := fmt.Sprintf("INSERT INTO connectors_executions_log_v1 (id, connector_id, name, ts, success) VALUES (DEFAULT, '%s', '%s', '%s', %t)",
		connectorID, connectionName, ts.Format("2006-01-02T15:04:05.999999Z"), success)
	tests.DBExec(dbClient, query, t, true)
}
