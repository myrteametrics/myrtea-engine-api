package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/myrteametrics/myrtea-engine-api/v5/internals/handlers/render"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/history"
	"go.uber.org/zap"
)

// @Summary Get Today's Fact Result by Criteria
// @Description Fetches the result of a historical fact based on provided criteria for today's date.
// @Tags Facts_history
// @Accept json
// @Produce json
// @Param ParamGetFactHistory body interface{} true "JSON payload containing criteria for fetching today's history fact result."
// @Security Bearer
// @Success 200 "Successfully fetched result"
// @Failure 400 "Status Bad Request"
// @Failure 500 "Status" internal server error"
// @Router /engine/history/facts/today/result [post]
func GetFactResultForTodayByCriteria(w http.ResponseWriter, r *http.Request) {

	var param history.ParamGetFactHistory
	err := json.NewDecoder(r.Body).Decode(&param)
	if err != nil {
		zap.L().Warn("Get Fact History json decode", zap.Error(err))
		render.Error(w, r, render.ErrAPIDecodeJSONBody, err)
		return
	}

	if ok, err := param.IsValid(); !ok {
		zap.L().Warn("parameter of Get Fact History  json is invalid", zap.Error(err))
		render.Error(w, r, render.ErrAPIResourceInvalid, err)
		return
	}

	resulat, err := history.S().HistoryFactsQuerier.GetTodaysFactResultByParameters(param)

	render.JSON(w, r, resulat)
}
