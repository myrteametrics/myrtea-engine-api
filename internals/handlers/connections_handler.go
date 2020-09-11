package handlers

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/connector"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/handlers/render"
	"go.uber.org/zap"
)

// GetlastConnectorExecutionDateTime godoc
// @Summary Get the DateTime of the last connections readings
// @Description Gets the DateTime of the last connections readings.
// @Tags Admin
// @Produce json
// @Param id path string true "Connector ID"
// @Param successOnly query string false "true to ignore failed connector executions"
// @Security Bearer
// @Success 200 {string} string "Status OK"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /engine/connector/{id}/executions/last [get]
func GetlastConnectorExecutionDateTime(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	successOnly, _ := strconv.ParseBool(r.URL.Query().Get("successOnly"))
	lastReading, err := connector.R().GetLastConnectionReading(id, successOnly)
	if err != nil {
		zap.L().Warn("Error reading last connections reading", zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}

	render.JSON(w, r, lastReading)
}
