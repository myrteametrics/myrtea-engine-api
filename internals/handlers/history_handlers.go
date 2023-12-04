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
// @Description Example Request:
// @Description <pre>{
// @Description   "factID": 123,
// @Description   "situationId": 456,
// @Description   "situationInstanceId": 789
// @Description }</pre>
// @Tags Facts_history
// @Accept json
// @Produce json
// @Param history.ParamGetFactHistory body interface{} true "JSON payload containing criteria for fetching today's history fact result."
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

	if err := param.IsValid(); err != nil {
		zap.L().Warn("parameter of Get Fact History  json is invalid", zap.Error(err))
		render.Error(w, r, render.ErrAPIResourceInvalid, err)
		return
	}

	resulat, err := history.S().HistoryFactsQuerier.GetTodaysFactResultByParameters(param)

	if err != nil {
		zap.L().Warn("error getting fact history by date", zap.Error(err))
		render.Error(w, r, render.ErrAPIProcessError, err)
		return
	}

	render.JSON(w, r, resulat)
}

// @Summary Get Fact Result by Date Criteria
// @Description Fetches the result of a historical fact based on provided criteria within specified date range.
// @Description Example Request:
// @Description <pre>{
// @Description   "factID": 42,
// @Description   "situationId": 29,
// @Description   "situationInstanceId": 572,
// @Description   "startDate": "2023-01-01 00:00:00",
// @Description   "endDate": "2023-12-30 00:00:00"
// @Description }</pre>
// The dates should be in the format "2006-01-02 15:04:05".
// @Tags Facts_history
// @Accept json
// @Produce json
// @Param history.ParamGetFactHistoryByDate body interface{} true "JSON payload containing criteria and date range for fetching history fact result, with dates in '2006-01-02 15:04:05' format."
// @Security Bearer
// @Success 200 "Successfully fetched result"
// @Failure 400 "Status Bad Request"
// @Failure 500 "Status Internal Server Error"
// @Router /engine/history/facts/date/result [post]
func GetFactResultByDateCriteria(w http.ResponseWriter, r *http.Request) {

	var param history.ParamGetFactHistoryByDate
	err := json.NewDecoder(r.Body).Decode(&param)
	if err != nil {
		zap.L().Warn("Get Fact History By Date json decode", zap.Error(err))
		render.Error(w, r, render.ErrAPIDecodeJSONBody, err)
		return
	}

	zap.L().Error("ParamGetFactHistoryByDate", zap.Any("", param))

	if err := param.IsValid(); err != nil {
		zap.L().Warn("parameter of Get Fact History By Date json is invalid", zap.Error(err))
		render.Error(w, r, render.ErrAPIResourceInvalid, err)
		return
	}

	result, err := history.S().HistoryFactsQuerier.GetFactResultByDate(param)
	if err != nil {
		zap.L().Warn("error getting fact history by date", zap.Error(err))
		render.Error(w, r, render.ErrAPIProcessError, err)
		return
	}

	render.JSON(w, r, result)
}
