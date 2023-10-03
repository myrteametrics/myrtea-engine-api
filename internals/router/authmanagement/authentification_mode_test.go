package authmanagement

import (
	"sync"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/myrteametrics/myrtea-sdk/v4/postgres"
)

func initialiseDB() (*sqlx.DB, error) {

    credentials := postgres.Credentials{
		URL:      "localhost",
		Port:     "5432",
		DbName:   "postgres",
		User:     "postgres",
		Password: "postgres",
	}

	db, err := postgres.DbConnection(credentials)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func TestSetModeIntegration(t *testing.T) {

	db, err := initialiseDB()

	if err != nil {
		t.Fatalf("Error initializing DB: %s",err)
	}

	defer db.Close()

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS AuthenticationMode (
		id SERIAL PRIMARY KEY,
		mode VARCHAR(10) CHECK (mode IN ('BASIC', 'SAML', 'OIDC'))
    );`)	
	if err != nil {
		t.Fatalf("unable to create table: %v", err)
	}

	querier := AuthenticationModeQuerier{
		Conn: db,
		Builder: AuthenticationModeBuilder{},
	}
	err = querier.SetMode("SAML")

	if err != nil {
		t.Errorf("error was not expected while setting mode: %s", err)
	}

	mode, err := querier.Query(querier.Builder.SelectMode())

	if err != nil {
		t.Errorf("error was not expected while selecting mode: %s", err)
	}

	if (mode.Mode != "SAML") {
        t.Errorf("Authentification mod SAML expected but get %s",mode.Mode)
	}

}

func TestConcurrentAccess(t *testing.T) {
	db, err := initialiseDB()

	if err != nil {
		t.Fatalf("Error initializing DB: %s",err)
	} 
	querier := New(db)
	restore := ReplaceGlobals(querier)
	defer restore()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			q := S()
			q.SetMode("BASIC")

		}()
	}
	wg.Wait()
}


