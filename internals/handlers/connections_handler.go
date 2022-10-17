package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/connector"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/handlers/render"
	"go.uber.org/zap"
)

// GetlastConnectorExecutionDateTime godoc
// @Summary Get the DateTime of the last connections readings
// @Description Gets the DateTime of the last connections readings.
// @Tags Admin
// @Produce json
// @Param id path string true "Connector ID"
// @Param successOnly query string false "true to ignore failed connector executions"
// @Param maxage query string false "maximum age of data (duration)"
// @Security Bearer
// @Success 200 {string} string "Status OK"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /engine/connector/{id}/executions/last [get]
func GetlastConnectorExecutionDateTime(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	successOnly, err := OptionnalQueryParamToBool(r, "successOnly", false)
	if err != nil {
		zap.L().Error("Parse input boolean", zap.Error(err), zap.String("successOnly", r.URL.Query().Get("successOnly")))
		render.Error(w, r, render.ErrAPIParsingInteger, err)
		return
	}

	maxAgeDays, err := OptionnalQueryParamToInt64(r, "maxage", 10)
	if err != nil {
		zap.L().Error("Parse input duration", zap.Error(err), zap.String("maxage", r.URL.Query().Get("maxage")))
		render.Error(w, r, render.ErrAPIParsingDuration, err)
		return
	}
	if maxAgeDays > 31 {
		maxAgeDays = 31
	}

	lastReading, err := connector.R().GetLastConnectionReading(id, successOnly, maxAgeDays)
	if err != nil {
		zap.L().Warn("Error reading last connections reading", zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}

	render.JSON(w, r, lastReading)
}
