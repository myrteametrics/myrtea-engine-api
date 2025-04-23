package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/calendar"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/handlers/render"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/security/permissions"
	"go.uber.org/zap"
)

// GetCalendars godoc
// @Summary Get all calendars
// @Description Get all calendars
// @Tags Calendars
// @Produce json
// @Security Bearer
// @Security ApiKeyAuth
// @Success 200 {array} calendar.Calendar "list of calendars"
// @Failure 500 "internal server error"
// @Router /engine/calendars [get]
func GetCalendars(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeCalendar, permissions.All, permissions.ActionList)) {
		render.Error(w, r, render.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	calendars, err := calendar.R().GetAll()
	if err != nil {
		zap.L().Error("Cannot retrieve issues", zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}

	calendarsSlice := make([]calendar.Calendar, 0)
	for _, calend := range calendars {
		calendarsSlice = append(calendarsSlice, calend)
	}

	sort.SliceStable(calendarsSlice, func(i, j int) bool {
		return calendarsSlice[i].ID < calendarsSlice[j].ID
	})

	render.JSON(w, r, calendarsSlice)
}

// GetCalendar godoc
// @Summary Get a Calendar
// @Description Get an calendar
// @Tags Calendars
// @Produce json
// @Param id path string true "Calendar ID"
// @Security Bearer
// @Security ApiKeyAuth
// @Success 200 {object} calendar.Calendar "calendar"
// @Failure 400 "Status Bad Request"
// @Router /engine/calendars/{id} [get]
func GetCalendar(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idCalendar, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("Error on parsing calendar id", zap.String("calendarID", id), zap.Error(err))
		render.Error(w, r, render.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeCalendar, strconv.FormatInt(idCalendar, 10), permissions.ActionGet)) {
		render.Error(w, r, render.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	calendar, found, err := calendar.R().Get(idCalendar)
	if err != nil {
		zap.L().Error("Cannot retrieve calendar", zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("calendar does not exists", zap.String("calendarID", id))
		render.Error(w, r, render.ErrAPIDBResourceNotFound, err)
		return
	}

	render.JSON(w, r, calendar)
}

// GetResolvedCalendar godoc
// @Summary Get a resolved Calendar
// @Description Get a resolved Calendar
// @Tags Calendars
// @Produce json
// @Param id path string true "Calendar ID"
// @Security Bearer
// @Security ApiKeyAuth
// @Success 200 {object} calendar.Calendar "calendar"
// @Failure 400 "Status Bad Request"
// @Router /engine/calendars/resolved/{id} [get]
func GetResolvedCalendar(w http.ResponseWriter, r *http.Request) {
	calendar.CBase().Update()

	id := chi.URLParam(r, "id")
	idCalendar, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("Error on parsing calendar id", zap.String("calendarID", id), zap.Error(err))
		render.Error(w, r, render.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeCalendar, strconv.FormatInt(idCalendar, 10), permissions.ActionGet)) {
		render.Error(w, r, render.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	calendar, found, err := calendar.CBase().GetResolved(idCalendar)
	if err != nil {
		zap.L().Error("Cannot retrieve calendar", zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("calendar does not exists", zap.String("calendarID", id))
		render.Error(w, r, render.ErrAPIDBResourceNotFound, err)
		return
	}

	render.JSON(w, r, calendar)
}

// IsInCalendarPeriod godoc
// @Summary Determines wether a timestamp is within a valid calendar period
// @Description Determines wether a timestamp is within a valid calendar period
// @Tags Calendars
// @Produce json
// @Param id path string true "Calendar ID"
// @Param time query string true "Timestamp to be found within a calendar period"
// @Security Bearer
// @Security ApiKeyAuth
// @Success 200 {object} calendar.InPeriodContains "InPeriodContains"
// @Failure 400 "Status Bad Request"
// @Router /engine/calendars/{id}/contains [get]
func IsInCalendarPeriod(w http.ResponseWriter, r *http.Request) {
	calendar.CBase().Update()

	id := chi.URLParam(r, "id")
	idCalendar, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("Error on parsing calendar id", zap.String("calendarID", id), zap.Error(err))
		render.Error(w, r, render.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeCalendar, strconv.FormatInt(idCalendar, 10), permissions.ActionGet)) {
		render.Error(w, r, render.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	t, err := ParseTime(r.URL.Query().Get("time"))
	if err != nil {
		zap.L().Error("Parse input time", zap.Error(err), zap.String("rawTime", r.URL.Query().Get("time")))
		render.Error(w, r, render.ErrAPIParsingDateTime, err)
		return
	}

	found, valid, err := calendar.CBase().InPeriodFromCalendarID(idCalendar, t)
	if err != nil {
		zap.L().Error("Cannot retrieve the period with the date", zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Error("Calendar not found", zap.Int64("id", idCalendar))
		render.Error(w, r, render.ErrAPIDBResourceNotFound, err)
		return
	}

	render.JSON(w, r, calendar.InPeriodContains{Contains: valid})
}

// PostCalendar godoc
// @Summary Creates a Calendar
// @Description Creates a Calendar
// @Tags Calendars
// @Accept json
// @Produce json
// @Param calendar body interface{} true "Calendar (json)"
// @Security Bearer
// @Security ApiKeyAuth
// @Success 200 {object} calendar.Calendar "calendar"
// @Failure 400 "Status Bad Request"
// @Failure 500 "Status" internal server error"
// @Router /engine/calendars [post]
func PostCalendar(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeCalendar, permissions.All, permissions.ActionCreate)) {
		render.Error(w, r, render.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	var newCalendar calendar.Calendar
	err := json.NewDecoder(r.Body).Decode(&newCalendar)
	if err != nil {
		zap.L().Warn("Invalid calendar json defined", zap.Error(err))
		render.Error(w, r, render.ErrAPIDecodeJSONBody, err)
		return
	}

	calendarID, err := calendar.R().Create(newCalendar)
	if err != nil {
		zap.L().Error("Error while creating the calendar", zap.Error(err))
		render.Error(w, r, render.ErrAPIDBInsertFailed, err)
		return
	}

	newCalendar, found, err := calendar.R().Get(calendarID)
	if err != nil {
		zap.L().Error("Get calendar failed", zap.Int64("id", calendarID), zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Error("Calendar not found", zap.Int64("id", calendarID))
		render.Error(w, r, render.ErrAPIDBResourceNotFoundAfterInsert, err)
		return
	}

	render.JSON(w, r, newCalendar)
}

// PutCalendar godoc
// @Summary Update a calendar
// @Description Updates the calendar
// @Tags Calendars
// @Accept json
// @Produce json
// @Param id path string true "Calendar ID"
// @Param user body interface{} true "calendar (json)"
// @Security Bearer
// @Security ApiKeyAuth
// @Success 200 {object} calendar.Calendar "calendar"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /engine/calendars/{id} [put]
func PutCalendar(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idCalendar, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("PutCalendar.GetId", zap.String("id", id), zap.Error(err))
		render.Error(w, r, render.ErrAPIParsingInteger, err)
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeCalendar, strconv.FormatInt(idCalendar, 10), permissions.ActionUpdate)) {
		render.Error(w, r, render.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	var calend calendar.Calendar
	err = json.NewDecoder(r.Body).Decode(&calend)
	if err != nil {
		zap.L().Warn("calendar decode json", zap.Error(err))
		render.Error(w, r, render.ErrAPIDecodeJSONBody, err)
		return
	}
	calend.ID = idCalendar

	err = calendar.R().Update(calend)
	if err != nil {
		zap.L().Error("PutUser.Update", zap.Error(err))
		render.Error(w, r, render.ErrAPIDBUpdateFailed, err)
		return
	}

	newCalendar, found, err := calendar.R().Get(idCalendar)
	if err != nil {
		zap.L().Error("Get calendar failed", zap.Int64("id", idCalendar), zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Error("Calendar not found", zap.Int64("id", idCalendar))
		render.Error(w, r, render.ErrAPIDBResourceNotFoundAfterInsert, err)
		return
	}

	render.JSON(w, r, newCalendar)
}

// DeleteCalendar godoc
// @Summary Delete calendar
// @Description Delete calendar
// @Tags Calendars
// @Produce json
// @Param id path string true "Calendar ID"
// @Security Bearer
// @Security ApiKeyAuth
// @Success 200 "Status OK"
// @Failure 400	"Status Bad Request"
// @Failure 500	"Status Internal Server Error"
// @Router /engine/calendars/{id} [delete]
func DeleteCalendar(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idCalendar, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("Error on parsing calendar id", zap.String("CalendarID", id), zap.Error(err))
		render.Error(w, r, render.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeCalendar, strconv.FormatInt(idCalendar, 10), permissions.ActionDelete)) {
		render.Error(w, r, render.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	err = calendar.R().Delete(idCalendar)
	if err != nil {
		zap.L().Error("Delete calendar", zap.String("ID", id), zap.Error(err))
		render.Error(w, r, render.ErrAPIDBDeleteFailed, err)
		return
	}

	render.OK(w, r)
}
