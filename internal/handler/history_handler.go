package handler

import (
	"encoding/json"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/utils/httputil"
	"net/http"

	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/history"
	"go.uber.org/zap"
)

// @Summary		Get Today's Fact Result by Criteria
// @Description	Fetches the result of a historical fact based on provided criteria for today's date.
// @Tags			Facts_history
// @Accept			json
// @Produce		json
// @Param			factHistory	body	history.ParamGetFactHistory	true	"JSON payload containing criteria for fetching today's history fact result."
// @Security		Bearer
// @Security		ApiKeyAuth
// @Success		200	{object}	history.FactResult	"Successfully fetched result"
// @Failure		400	"Status Bad Request"
// @Failure		500	"Status"	internal	server	error"
// @Router			/engine/history/facts/today/result [post]
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

//	@Summary		Get Fact Result by Date Criteria
//	@Description	Fetches the result of a historical fact based on provided criteria within specified date range.
//
// The dates should be in the format "2006-01-02 15:04:05".
//
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
