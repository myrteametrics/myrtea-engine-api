package tests

import (
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/myrteametrics/myrtea-sdk/v5/postgres"
)

// DBClient returns a postgresql test client for integration tests
// It targets a specific hostname "postgres" in a Gitlab CI environnement
// or "localhost" by default
func DBClient(t *testing.T) *sqlx.DB {
	credentials := postgres.Credentials{
		URL:      "localhost",
		Port:     "5432",
		DbName:   "postgres",
		User:     "postgres",
		Password: "postgres",
	}
	if os.Getenv("GITLAB_CI") != "" {
		t.Log("Found GITLAB_CI environment variable")
		// credentials.URL = "gitlab-myrtea-tests-postgresql"
		// credentials.DbName = os.Getenv("POSTGRES_DB")
	}
	dbClient, err := postgres.DbConnection(credentials)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	return dbClient
}

// DBExec execute an sql query which can lead to an immediate failure of the unit test
func DBExec(dbClient *sqlx.DB, query string, t *testing.T, failNow bool) {
	_, err := dbClient.Exec(query)
	if err != nil {
		t.Error(err)
		if failNow {
			t.FailNow()
		}
	}
}
