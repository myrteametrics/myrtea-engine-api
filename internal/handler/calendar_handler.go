package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"strconv"

	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/calendar"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/security/permissions"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/utils/httputil"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// GetCalendars godoc
//
//	@Id				GetCalendars
//
//	@Summary		Get all calendars
//	@Description	Get all calendars
//	@Tags			Calendars
//	@Produce		json
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{array}		calendar.Calendar	"list of calendars"
//	@Failure		500	{object}	httputil.APIError	"Internal Server Error"
//	@Router			/engine/calendars [get]
func GetCalendars(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeCalendar, permissions.All, permissions.ActionList)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	calendars, err := calendar.R().GetAll()
	if err != nil {
		zap.L().Error("Cannot retrieve issues", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}

	calendarsSlice := make([]calendar.Calendar, 0)
	for _, calend := range calendars {
		calendarsSlice = append(calendarsSlice, calend)
	}

	sort.SliceStable(calendarsSlice, func(i, j int) bool {
		return calendarsSlice[i].ID < calendarsSlice[j].ID
	})

	httputil.JSON(w, r, calendarsSlice)
}

// GetCalendar godoc
//
//	@Id				GetCalendar
//
//	@Summary		Get a Calendar
//	@Description	Get an calendar
//	@Tags			Calendars
//	@Produce		json
//	@Param			id	path	int	true	"Calendar ID"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	calendar.Calendar	"calendar"
//	@Failure		400	{object}	httputil.APIError	"Bad Request"
//	@Router			/engine/calendars/{id} [get]
func GetCalendar(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idCalendar, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("Error on parsing calendar id", zap.String("calendarID", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeCalendar, strconv.FormatInt(idCalendar, 10), permissions.ActionGet)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	calendar, found, err := calendar.R().Get(idCalendar)
	if err != nil {
		zap.L().Error("Cannot retrieve calendar", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("calendar does not exists", zap.String("calendarID", id))
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFound, err)
		return
	}

	httputil.JSON(w, r, calendar)
}

// GetResolvedCalendar godoc
//
//	@Id				GetResolvedCalendar
//
//	@Summary		Get a resolved Calendar
//	@Description	Get a resolved Calendar
//	@Tags			Calendars
//	@Produce		json
//	@Param			id	path	int	true	"Calendar ID"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	calendar.Calendar	"calendar"
//	@Failure		400	{object}	httputil.APIError	"Bad Request"
//	@Router			/engine/calendars/resolved/{id} [get]
func GetResolvedCalendar(w http.ResponseWriter, r *http.Request) {
	calendar.CBase().Update()

	id := chi.URLParam(r, "id")
	idCalendar, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("Error on parsing calendar id", zap.String("calendarID", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeCalendar, strconv.FormatInt(idCalendar, 10), permissions.ActionGet)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	calendar, found, err := calendar.CBase().GetResolved(idCalendar)
	if err != nil {
		zap.L().Error("Cannot retrieve calendar", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("calendar does not exists", zap.String("calendarID", id))
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFound, err)
		return
	}

	httputil.JSON(w, r, calendar)
}

// IsInCalendarPeriod godoc
//
//	@Id				IsInCalendarPeriod
//
//	@Summary		Determines wether a timestamp is within a valid calendar period
//	@Description	Determines wether a timestamp is within a valid calendar period
//	@Tags			Calendars
//	@Produce		json
//	@Param			id		path	int		true	"Calendar ID"
//	@Param			time	query	string	true	"Timestamp to be found within a calendar period"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	calendar.InPeriodContains	"InPeriodContains"
//	@Failure		400	{object}	httputil.APIError			"Bad Request"
//	@Router			/engine/calendars/{id}/contains [get]
func IsInCalendarPeriod(w http.ResponseWriter, r *http.Request) {
	calendar.CBase().Update()

	id := chi.URLParam(r, "id")
	idCalendar, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("Error on parsing calendar id", zap.String("calendarID", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeCalendar, strconv.FormatInt(idCalendar, 10), permissions.ActionGet)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	t, err := ParseTime(r.URL.Query().Get("time"))
	if err != nil {
		zap.L().Error("Parse input time", zap.Error(err), zap.String("rawTime", r.URL.Query().Get("time")))
		httputil.Error(w, r, httputil.ErrAPIParsingDateTime, err)
		return
	}

	found, valid, err := calendar.CBase().InPeriodFromCalendarID(idCalendar, t)
	if err != nil {
		zap.L().Error("Cannot retrieve the period with the date", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Error("Calendar not found", zap.Int64("id", idCalendar))
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFound, err)
		return
	}

	httputil.JSON(w, r, calendar.InPeriodContains{Contains: valid})
}

// PostCalendar godoc
//
//	@Id				PostCalendar
//
//	@Summary		Creates a Calendar
//	@Description	Creates a Calendar
//	@Tags			Calendars
//	@Accept			json
//	@Produce		json
//	@Param			calendar	body	interface{}	true	"Calendar (json)"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	calendar.Calendar	"calendar"
//	@Failure		400	{object}	httputil.APIError	"Bad Request"
//	@Failure		500	{object}	httputil.APIError	"Internal Server Error"
//	@Router			/engine/calendars [post]
func PostCalendar(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeCalendar, permissions.All, permissions.ActionCreate)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	var newCalendar calendar.Calendar
	err := json.NewDecoder(r.Body).Decode(&newCalendar)
	if err != nil {
		zap.L().Warn("Invalid calendar json defined", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDecodeJSONBody, err)
		return
	}

	calendarID, err := calendar.R().Create(newCalendar)
	if err != nil {
		zap.L().Error("Error while creating the calendar", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBInsertFailed, err)
		return
	}

	newCalendar, found, err := calendar.R().Get(calendarID)
	if err != nil {
		zap.L().Error("Get calendar failed", zap.Int64("id", calendarID), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Error("Calendar not found", zap.Int64("id", calendarID))
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFoundAfterInsert, err)
		return
	}

	httputil.JSON(w, r, newCalendar)
}

// PutCalendar godoc
//
//	@Id				PutCalendar
//
//	@Summary		Update a calendar
//	@Description	Updates the calendar
//	@Tags			Calendars
//	@Accept			json
//	@Produce		json
//	@Param			id		path	int			true	"Calendar ID"
//	@Param			user	body	interface{}	true	"calendar (json)"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	calendar.Calendar	"calendar"
//	@Failure		400	{string}	string				"Bad Request"
//	@Failure		500	{string}	string				"Internal Server Error"
//	@Router			/engine/calendars/{id} [put]
func PutCalendar(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idCalendar, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("PutCalendar.GetId", zap.String("id", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeCalendar, strconv.FormatInt(idCalendar, 10), permissions.ActionUpdate)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	var calend calendar.Calendar
	err = json.NewDecoder(r.Body).Decode(&calend)
	if err != nil {
		zap.L().Warn("calendar decode json", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDecodeJSONBody, err)
		return
	}
	calend.ID = idCalendar

	err = calendar.R().Update(calend)
	if err != nil {
		zap.L().Error("PutUser.Update", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBUpdateFailed, err)
		return
	}

	newCalendar, found, err := calendar.R().Get(idCalendar)
	if err != nil {
		zap.L().Error("Get calendar failed", zap.Int64("id", idCalendar), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Error("Calendar not found", zap.Int64("id", idCalendar))
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFoundAfterInsert, err)
		return
	}

	httputil.JSON(w, r, newCalendar)
}

// DeleteCalendar godoc
//
//	@Id				DeleteCalendar
//
//	@Summary		Delete calendar
//	@Description	Delete calendar
//	@Tags			Calendars
//	@Produce		json
//	@Param			id	path	int	true	"Calendar ID"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	"Status OK"
//	@Failure		400	{object}	httputil.APIError	"Bad Request"
//	@Failure		500	{object}	httputil.APIError	"Internal Server Error"
//	@Router			/engine/calendars/{id} [delete]
func DeleteCalendar(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idCalendar, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("Error on parsing calendar id", zap.String("CalendarID", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeCalendar, strconv.FormatInt(idCalendar, 10), permissions.ActionDelete)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	err = calendar.R().Delete(idCalendar)
	if err != nil {
		zap.L().Error("Delete calendar", zap.String("ID", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBDeleteFailed, err)
		return
	}

	httputil.OK(w, r)
}
