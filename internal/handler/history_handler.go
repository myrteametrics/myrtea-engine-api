package handler

import (
	"encoding/json"
	"net/http"

	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/utils/httputil"

	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/history"
	"go.uber.org/zap"
)

// GetFactResultForTodayByCriteria
//
//	@Id				GetFactResultForTodayByCriteria
//
//	@Summary		Get Today's Fact Result by Criteria
//	@Description	Fetches the result of a historical fact based on provided criteria for today's date.
//	@Tags			Facts_history
//	@Accept			json
//	@Produce		json
//	@Param			factHistory	body	history.ParamGetFactHistory	true	"JSON payload containing criteria for fetching today's history fact result."
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	history.FactResult	"Successfully fetched result"
//	@Failure		400	"Status Bad Request"
//	@Failure		500	{object}	httputil.APIError	"Internal Server Error"
//	@Router			/engine/history/facts/today/result [post]
func GetFactResultForTodayByCriteria(w http.ResponseWriter, r *http.Request) {

	var param history.ParamGetFactHistory
	err := json.NewDecoder(r.Body).Decode(&param)
	if err != nil {
		zap.L().Warn("Get Fact History json decode", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDecodeJSONBody, err)
		return
	}

	if err := param.IsValid(); err != nil {
		zap.L().Warn("parameter of Get Fact History  json is invalid", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIResourceInvalid, err)
		return
	}

	resulat, err := history.S().HistoryFactsQuerier.GetTodaysFactResultByParameters(param)

	if err != nil {
		zap.L().Warn("error getting fact history by date", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIProcessError, err)
		return
	}

	httputil.JSON(w, r, resulat)
}

// GetFactResultByDateCriteria
//
//	@Id				GetFactResultByDateCriteria
//
//	@Summary		Get Fact Result by Date Criteria
//
//	@Description	Fetches the result of a historical fact based on provided criteria within specified date range. The dates should be in the format "2006-01-02 15:04:05".
//	@Tags			Facts_history
//	@Accept			json
//	@Produce		json
//	@Param			factHistory	body	history.ParamGetFactHistoryByDate	true	"JSON payload containing criteria and date range for fetching history fact result, with dates in '2006-01-02 15:04:05' format."
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	history.FactResult	"Successfully fetched result"
//	@Failure		400	"Status Bad Request"
//	@Failure		500	"Status Internal Server Error"
//	@Router			/engine/history/facts/date/result [post]
func GetFactResultByDateCriteria(w http.ResponseWriter, r *http.Request) {

	var param history.ParamGetFactHistoryByDate
	err := json.NewDecoder(r.Body).Decode(&param)
	if err != nil {
		zap.L().Warn("Get Fact History By Date json decode", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDecodeJSONBody, err)
		return
	}

	if err := param.IsValid(); err != nil {
		zap.L().Warn("parameter of Get Fact History By Date json is invalid", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIResourceInvalid, err)
		return
	}

	result, err := history.S().HistoryFactsQuerier.GetFactResultByDate(param)
	if err != nil {
		zap.L().Warn("error getting fact history by date", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIProcessError, err)
		return
	}

	httputil.JSON(w, r, result)
}

// GetFactExprResultForTodayByCriteria
//
//	@Id				GetFactExprResultForTodayByCriteria
//
//	@Summary		Get Today's Fact Expression Result by Criteria
//	@Description	Fetches the result of a historical fact expression based on provided criteria for today's date.
//	@Tags			situation_history
//	@Accept			json
//	@Produce		json
//	@Param			factExprHistory	body	history.ParamGetFactExprHistory	true	"JSON payload containing criteria for fetching today's history fact expression result."
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	history.FactExprResult	"Successfully fetched result"
//	@Failure		400	"Status Bad Request"
//	@Failure		500	{object}	httputil.APIError	"Internal Server Error"
//	@Router			/engine/history/factexpr/today/result [post]
func GetFactExprResultForTodayByCriteria(w http.ResponseWriter, r *http.Request) {

	var param history.ParamGetFactExprHistory
	err := json.NewDecoder(r.Body).Decode(&param)
	if err != nil {
		zap.L().Warn("Get Fact expression History json decode", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDecodeJSONBody, err)
		return
	}

	if err := param.IsValid(); err != nil {
		zap.L().Warn("parameter of Get Fact expression History  json is invalid", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIResourceInvalid, err)
		return
	}

	resulat, err := history.S().HistorySituationsQuerier.GetTodaysFactExprResultByParameters(param)

	if err != nil {
		zap.L().Warn("error getting fact expression history by date", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIProcessError, err)
		return
	}

	httputil.JSON(w, r, resulat)
}

// GetFactExprResultByDateCriteria
//
//	@Id				GetFactExprResultByDateCriteria
//
//	@Summary		Get Fact expression Result by Date Criteria
//	@Description	Fetches the result of a historical fact expression based on provided criteria within specified date range.
//
// The dates should be in the format "2006-01-02 15:04:05".
//
//	@Tags			situation_history
//	@Accept			json
//	@Produce		json
//	@Param			factExprHistory	body	history.ParamGetFactExprHistoryByDate	true	"JSON payload containing criteria and date range for fetching history fact expression result, with dates in '2006-01-02 15:04:05' format."
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	history.FactExprResult	"Successfully fetched result"
//	@Failure		400	{object}	httputil.APIError		"Bad Request"
//	@Failure		500	{object}	httputil.APIError		"Internal Server Error"
//	@Router			/engine/history/factexpr/date/result [post]
func GetFactExprResultByDateCriteria(w http.ResponseWriter, r *http.Request) {

	var param history.ParamGetFactExprHistoryByDate
	err := json.NewDecoder(r.Body).Decode(&param)
	if err != nil {
		zap.L().Warn("Get Fact expression History By Date json decode", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDecodeJSONBody, err)
		return
	}

	if err := param.IsValid(); err != nil {
		zap.L().Warn("parameter of Get Fact expression History By Date json is invalid", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIResourceInvalid, err)
		return
	}

	result, err := history.S().HistorySituationsQuerier.GetFactExprResultByDate(param)
	if err != nil {
		zap.L().Warn("error getting fact expression history by date", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIProcessError, err)
		return
	}

	httputil.JSON(w, r, result)
}
