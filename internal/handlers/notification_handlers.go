package handlers

import (
	"github.com/go-chi/chi/v5"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/handlers/render"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/models"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/notifier/notification"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/security/users"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/utils/dbutils"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

// GetNotifications godoc
// @Summary Get all notifications
// @Description Get all notifications of the authentified user
// @Tags Notifications
// @Produce json
// @Param maxage query string false "Notification maximum age (use duration format, ex: 48h)"
// @Param nhit query int false "Hit per page"
// @Param offset query int false "Offset number for pagination"
// @Security Bearer
// @Security ApiKeyAuth
// @Failure 500 "internal server error"
// @Router /engine/notifications [get]
func GetNotifications(w http.ResponseWriter, r *http.Request) {

	// FIXME: DON'T FORGET TO FIX THIS !
	maxAge, err := ParseDuration(r.URL.Query().Get("maxage"))
	if err != nil {
		zap.L().Warn("Parse duration input maxage", zap.Error(err), zap.String("rawMaxAge", r.URL.Query().Get("maxage")))
		render.Error(w, r, render.ErrAPIParsingDuration, err)
		return
	}

	nhit, err := ParseInt(r.URL.Query().Get("nhit"))
	if err != nil {
		zap.L().Warn("Parse input nhit", zap.Error(err), zap.String("rawNhit", r.URL.Query().Get("nhit")))
		render.Error(w, r, render.ErrAPIParsingInteger, err)
		return
	}

	offset, err := ParseInt(r.URL.Query().Get("offset"))
	if err != nil {
		zap.L().Warn("Parse input offset", zap.Error(err), zap.String("rawPage", r.URL.Query().Get("offset")))
		render.Error(w, r, render.ErrAPIParsingInteger, err)
		return
	}

	_user := r.Context().Value(models.ContextKeyUser)
	if _user == nil {
		zap.L().Warn("No context user provided")
		return
	}
	user := _user.(users.UserWithPermissions)

	queryOptional := dbutils.DBQueryOptionnal{
		Limit:  nhit,
		Offset: offset,
		MaxAge: maxAge,
	}
	notifications, err := notification.R().GetAll(queryOptional, user.Login)
	if err != nil {
		zap.L().Error("Error getting notifications", zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}

	render.JSON(w, r, notifications)
}

// UpdateRead godoc
// @Summary Update the "read" status of the notification
// @Description Mark a notification as "read"
// @Tags Notifications
// @Produce json
// @Param id path int false "notification ID"
// @Param status query boolean false "notification's read property given value"
// @Security Bearer
// @Security ApiKeyAuth
// @Success 200 "Status OK"
// @Failure 500 "internal server error"
// @Failure 400 "Status bad request"
// @Router /engine/notifications/{id}/read [put]
func UpdateRead(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idNotif, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("Error while Parsing notification id", zap.String("notification", id), zap.Error(err))
		render.Error(w, r, render.ErrAPIParsingInteger, err)
		return
	}
	_status := r.URL.Query().Get("status")
	status := false

	_user := r.Context().Value(models.ContextKeyUser)
	if _user == nil {
		zap.L().Warn("No context user provided")
		return
	}
	user := _user.(users.UserWithPermissions)

	if _status == "true" {
		status = true
	}

	err = notification.R().UpdateRead(idNotif, status, user.Login)
	if err != nil {
		zap.L().Error("Error while updating notifications", zap.Error(err))
		render.Error(w, r, render.ErrAPIDBUpdateFailed, err)
		return
	}

	render.OK(w, r)
}
