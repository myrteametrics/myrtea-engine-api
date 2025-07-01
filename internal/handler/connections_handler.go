package handler

import (
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/utils/httputil"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/connector"
	"go.uber.org/zap"
)

// GetLastConnectorExecutionDateTime godoc
//
//	@Summary		Get the DateTime of the last connections readings
//	@Description	Gets the DateTime of the last connections readings.
//	@Tags			Admin
//	@Produce		json
//	@Param			id			path	string	true	"Connector ID"
//	@Param			successOnly	query	string	false	"true to ignore failed connector executions"
//	@Param			maxage		query	string	false	"maximum age of data (duration)"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{string}	string	"Status OK"
//	@Failure		400	{string}	string	"Bad Request"
//	@Failure		500	{string}	string	"Internal Server Error"
//	@Router			/engine/connector/{id}/executions/last [get]
func GetLastConnectorExecutionDateTime(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	successOnly, err := QueryParamToOptionalBool(r, "successOnly", false)
	if err != nil {
		zap.L().Error("Parse input boolean", zap.Error(err), zap.String("successOnly", r.URL.Query().Get("successOnly")))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	maxAgeDays, err := QueryParamToOptionalInt64(r, "maxage", 10)
	if err != nil {
		zap.L().Error("Parse input duration", zap.Error(err), zap.String("maxage", r.URL.Query().Get("maxage")))
		httputil.Error(w, r, httputil.ErrAPIParsingDuration, err)
		return
	}
	if maxAgeDays > 31 {
		maxAgeDays = 31
	}

	lastReading, err := connector.R().GetLastConnectionReading(id, successOnly, maxAgeDays)
	if err != nil {
		zap.L().Warn("Error reading last connections reading", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}

	httputil.JSON(w, r, lastReading)
}
