package handlers

import (
	"net/http"

	"github.com/myrteametrics/myrtea-engine-api/v5/internals/handlers/render"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/models"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/notifier"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/security/users"
	"go.uber.org/zap"
)

// NotificationsWSRegister godoc
// @Summary Register a new client to the notifications system using WS
// @Description Register a new client to the notifications system using WS
// @Tags Notifications
// @Produce json
// @Param jwt query string false "Json Web Token"
// @Security Bearer
// @Success 200 "Status OK"
// @Failure 400 "Status Bad Request"
// @Router /engine/notifications/ws [get]
func NotificationsWSRegister(w http.ResponseWriter, r *http.Request) {

	zap.L().Info("New connection on /ws")

	_user := r.Context().Value(models.ContextKeyUser)
	if _user == nil {
		zap.L().Warn("No context user provided")
		return
	}
	user := _user.(users.UserWithPermissions)

	client, err := notifier.BuildWebsocketClient(w, r, &user)
	if err != nil {
		zap.L().Error("Build new WS Client", zap.Error(err))
		return
	}

	err = notifier.C().Register(client)
	if err != nil {
		zap.L().Error("Add new WS Client to manager", zap.Error(err))
		return
	}
	go client.Write()

	// go client.Read() // Disabled until proper usage
}

// NotificationsSSERegister godoc
// @Summary Register a new client to the notifications system using SSE
// @Description Register a new client to the notifications system using SSE
// @Tags Notifications
// @Produce json
// @Param jwt query string false "Json Web Token"
// @Security Bearer
// @Success 200 "Status OK"
// @Failure 400 "Status Bad Request"
// @Router /engine/notifications/sse [get]
func NotificationsSSERegister(w http.ResponseWriter, r *http.Request) {

	zap.L().Info("New connection on /sse")

	_user := r.Context().Value(models.ContextKeyUser)
	if _user == nil {
		zap.L().Warn("No context user provided")
		return
	}
	user := _user.(users.UserWithPermissions)

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*") // Might not stay there

	client, err := notifier.BuildSSEClient(w, &user)
	if err != nil {
		zap.L().Error("Build new SSE Client", zap.Error(err))
		return
	}

	err = notifier.C().Register(client)
	if err != nil {
		zap.L().Error("Add new SSE Client to manager", zap.Error(err))
		return
	}

	defer func() {
		err := notifier.C().Unregister(client)
		if err != nil {
			zap.L().Warn("Unregister notifier clien", zap.Error(err))
		}
	}()

	// Listen to connection close and un-register client
	notify := r.Context().Done()
	go func() {
		<-notify
		err := notifier.C().Unregister(client)
		if err != nil {
			zap.L().Warn("Unregister notifier clien", zap.Error(err))
		}
	}()

	client.Write()
}

// TriggerNotification godoc
// @Summary Send a notification
// @Description Generate a new NotifyTask with a default message for testing
// @Tags Notifications
// @Accept json
// @Param key query string true "Notifier cache key"
// @Param notification body interface{} true "Notify task definition (json)"
// @Security Bearer
// @Success 200 "Status OK"
// @Failure 400 "Status Bad Request"
// @Failure 500 "Status" internal server error"
// @Router /engine/notifications/trigger [post]
func TriggerNotification(w http.ResponseWriter, r *http.Request) {

	// cacheKey := r.URL.Query().Get("key")

	// var notif notification.MockNotification
	// err := json.NewDecoder(r.Body).Decode(&notif)
	// if err != nil {
	// 	zap.L().Warn("Notification json decode", zap.Error(err))
	// 	render.Error(w, r, render.ErrAPIDecodeJSONBody, err)
	// 	return
	// }
	// notif.CreationDate = time.Now().Truncate(1 * time.Millisecond).UTC()

	// notifier.C().SendToRoles(cacheKey, 1*time.Second, notif, notif.Groups)

	render.OK(w, r)
}
