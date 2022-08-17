package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/myrteametrics/myrtea-engine-api/v4/internals/handlers/render"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/history"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/security/permissions"
	"go.uber.org/zap"
)

func baseSearchOptions(w http.ResponseWriter, r *http.Request) (history.GetHistorySituationsOptions, render.APIError, error) {

	situationID, err := QueryParamToOptionnalInt64(r, "situationid", -1)
	if err != nil {
		zap.L().Warn("Error on parsing situationid", zap.String("situationID", r.URL.Query().Get("situationid")), zap.Error(err))
		return history.GetHistorySituationsOptions{}, render.ErrAPIParsingInteger, err
	}

	situationInstanceID, err := QueryParamToOptionnalInt64(r, "situationinstanceid", -1)
	if err != nil {
		zap.L().Warn("Error on parsing situationinstanceid", zap.String("situationInstanceID", r.URL.Query().Get("situationinstanceid")), zap.Error(err))
		return history.GetHistorySituationsOptions{}, render.ErrAPIParsingInteger, err
	}

	maxDate, err := QueryParamToOptionnalTime(r, "maxdate", time.Time{})
	if err != nil {
		zap.L().Warn("Parse input maxdate", zap.Error(err), zap.String("maxdate", r.URL.Query().Get("maxdate")))
		return history.GetHistorySituationsOptions{}, render.ErrAPIParsingDateTime, err
	}

	minDate, err := QueryParamToOptionnalTime(r, "mindate", time.Time{})
	if err != nil {
		zap.L().Warn("Parse input mindate", zap.Error(err), zap.String("mindate", r.URL.Query().Get("mindate")))
		return history.GetHistorySituationsOptions{}, render.ErrAPIParsingDateTime, err
	}

	if !maxDate.IsZero() && minDate.IsZero() {
		minDate = maxDate.Add(-1 * 60 * 24 * time.Hour)
	}

	options := history.GetHistorySituationsOptions{
		SituationID:         situationID,
		SituationInstanceID: situationInstanceID,
		FromTS:              minDate,
		ToTS:                maxDate,
	}

	return options, render.APIError{}, nil
}

// Search godoc
// @Summary query situation history data
// @Description query situation history data
// @Tags Search
// @Accept json
// @Produce json
// @Param situationid 			query	int		true
// @Param situationinstanceid 	query	int		true
// @Param maxdate				query 	string 	true	"time.Time"
// @Param mindate				query 	string 	true	"time.Time"
// @Security Bearer
// @Success 200 {array} search.QueryResult "query result"
// @Failure 500 "internal server error"
// @Router /engine/search/last [get]
func SearchLast(w http.ResponseWriter, r *http.Request) {

	options, apiError, err := baseSearchOptions(w, r)
	if err != nil {
		render.Error(w, r, apiError, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeSituation, strconv.FormatInt(options.SituationID, 10), permissions.ActionSearch)) {
		render.Error(w, r, render.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	historySituations, err := history.S().GetHistorySituationsIdsLast(options)
	if err != nil {
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}

	historyFacts, historySituationFacts, err := history.S().GetHistoryFactsFromSituation(historySituations)
	if err != nil {
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}

	result := history.ExtractHistoryDataSearch(historySituations, historySituationFacts, historyFacts)

	render.JSON(w, r, result)
}

// Search godoc
// @Summary query situation history data
// @Description query situation history data
// @Tags Search
// @Accept json
// @Produce json
// @Param situationid 			query	int		true
// @Param situationinstanceid 	query	int		true
// @Param maxdate				query 	string 	true	"time.Time"
// @Param mindate				query 	string 	true	"time.Time"
// @Param interval				query 	string 	true	"year | quarter | month | week | day | hour | minute"
// @Security Bearer
// @Success 200 {array} search.QueryResult "query result"
// @Failure 500 "internal server error"
// @Router /engine/search/last/byinterval [get]
func SearchLastByInterval(w http.ResponseWriter, r *http.Request) {

	options, apiError, err := baseSearchOptions(w, r)
	if err != nil {
		render.Error(w, r, apiError, err)
		return
	}

	interval := r.URL.Query().Get("interval")
	if interval != "year" && interval != "quarter" && interval != "month" && interval != "week" && interval != "day" && interval != "hour" && interval != "minute" {
		zap.L().Warn("Error on parsing interval", zap.String("interval", interval), zap.Error(err))
		render.Error(w, r, render.ErrAPIParsingDuration, fmt.Errorf("interval %s is not supported", interval))
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeSituation, strconv.FormatInt(options.SituationID, 10), permissions.ActionSearch)) {
		render.Error(w, r, render.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	historySituations, err := history.S().GetHistorySituationsIdsByStandardInterval(options, interval)
	if err != nil {
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}

	historyFacts, historySituationFacts, err := history.S().GetHistoryFactsFromSituation(historySituations)
	if err != nil {
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}

	result := history.ExtractHistoryDataSearch(historySituations, historySituationFacts, historyFacts)

	render.JSON(w, r, result)
}

// Search godoc
// @Summary query situation history data
// @Description query situation history data
// @Tags Search
// @Accept json
// @Produce json
// @Param situationid 			query	int		true
// @Param situationinstanceid 	query	int		true
// @Param maxdate				query 	string 	true	"time.Time"
// @Param mindate				query 	string 	true	"time.Time"
// @Param referencedate			query 	string 	true	"time.Time"
// @Param interval				query 	string 	true	"time.Duration"
// @Security Bearer
// @Success 200 {array} search.QueryResult "query result"
// @Failure 500 "internal server error"
// @Router /engine/search/last/bycustominterval [get]
func SearchLastByCustomInterval(w http.ResponseWriter, r *http.Request) {

	options, apiError, err := baseSearchOptions(w, r)
	if err != nil {
		render.Error(w, r, apiError, err)
		return
	}

	interval, err := time.ParseDuration(r.URL.Query().Get("interval"))
	if err != nil {
		zap.L().Warn("Error on parsing interval", zap.String("interval", r.URL.Query().Get("interval")), zap.Error(err))
		render.Error(w, r, render.ErrAPIParsingDuration, fmt.Errorf("interval %s is not supported", r.URL.Query().Get("interval")))
		return
	}
	if interval < time.Minute {
		zap.L().Warn("Too small interval", zap.Duration("interval", interval))
		render.Error(w, r, render.ErrAPIParsingDuration, fmt.Errorf("interval %s is too small (<1min)", interval))
		return
	}

	referenceDate, err := QueryParamToTime(r, "referencedate")
	if err != nil {
		zap.L().Warn("Parse input mindate", zap.Error(err), zap.String("mindate", r.URL.Query().Get("mindate")))
		render.Error(w, r, render.ErrAPIParsingDateTime, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeSituation, strconv.FormatInt(options.SituationID, 10), permissions.ActionSearch)) {
		render.Error(w, r, render.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	historySituations, err := history.S().GetHistorySituationsIdsByCustomInterval(options, referenceDate, interval)
	if err != nil {
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}

	historyFacts, historySituationFacts, err := history.S().GetHistoryFactsFromSituation(historySituations)
	if err != nil {
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}

	result := history.ExtractHistoryDataSearch(historySituations, historySituationFacts, historyFacts)

	render.JSON(w, r, result)
}
