package handlers

import (
	"net/http"
	"testing"

	"github.com/myrteametrics/myrtea-engine-api/v4/internals/search"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/security/users"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/situation"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/tests"
)

func TestSearch(t *testing.T) {
	t.Skip("Development test")
	db := tests.DBClient(t)

	situationR := situation.NewPostgresRepository(db)
	situation.ReplaceGlobals(situationR)

	searchR := search.NewPostgresRepository(db)
	search.ReplaceGlobals(searchR)

	q := `{
		"situationId": 2,
		"situationInstanceId": 0,
		"range": "10s",
		"start": "2020-04-29T19:04:45+02:00",
		"downSampling": {
			"granularity": "2s",
			"operation": "avg"
		}
	}`

	rr := tests.BuildTestHandler(t, "POST", "/search", q, "/search", Search, users.UserWithPermissions{})
	tests.CheckTestHandler(t, rr, http.StatusOK, `{}`)
}
