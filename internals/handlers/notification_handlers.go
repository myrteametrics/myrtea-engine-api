package handlers

import (
	"github.com/go-chi/chi/v5"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/dbutils"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/handlers/render"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/models"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/notifier/notification"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/security/users"
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

	// generate rando mock notifications for testing 1 to 15
	//for i := 2; i < 17; i++ {
	//	notifications = append(notifications, notification.NewMockNotification(int64(i), "OK", "MockNotification", "Toodododo", "You must do something lol", time.Now().AddDate(0, 0, -i), []int64{1}, map[string]interface{}{"issueId": 1}))
	//}

	//notifications = append(notifications, notification.NewMockNotification(1, "OK", "MockNotification", "Toodododo", "You must do something lol", time.Now(), []int64{1}, map[string]interface{}{"issueId": 1}))
	//notifications = append(notifications, export.NewExportNotification(2, export.WrapperItem{Id: "test"}, 1))
	//notifications = append(notifications, notification.NewBaseNotification(3, false, true))
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
