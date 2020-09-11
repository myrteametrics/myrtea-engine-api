package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/myrteametrics/myrtea-engine-api/v4/internals/handlers/render"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/search"
	"go.uber.org/zap"
)

// Search godoc
// @Summary query situation history data
// @Description query situation history data
// @Tags Search
// @Accept json
// @Produce json
// @Param query body search.Query true "query (json)"
// @Security Bearer
// @Success 200 {array} search.QueryResult "query result"
// @Failure 500 "internal server error"
// @Router /engine/search [post]
func Search(w http.ResponseWriter, r *http.Request) {
	var query search.Query
	err := json.NewDecoder(r.Body).Decode(&query)
	if err != nil {
		zap.L().Warn("Query json decode", zap.Error(err))
		render.Error(w, r, render.ErrAPIDecodeJSONBody, err)
		return
	}

	result, err := query.Execute()
	if err != nil {
		switch err.(type) {
		case search.ErrResourceNotFound:
			render.Error(w, r, render.ErrAPIDBResourceNotFound, err)
		case search.ErrDatabase:
			render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		default:
			render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		}
		return
	}

	render.JSON(w, r, result)
}
