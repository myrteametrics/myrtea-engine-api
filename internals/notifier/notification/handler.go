package notification

import "sync"

var (
	_globalHandlerMu sync.RWMutex
	_globalHandler   *Handler
)

// H is used to access the global notification handler singleton
func H() *Handler {
	_globalHandlerMu.RLock()
	defer _globalHandlerMu.RUnlock()
	return _globalHandler
}

// ReplaceHandlerGlobals affects a new repository to the global notification handler singleton
func ReplaceHandlerGlobals(handler *Handler) func() {
	_globalHandlerMu.Lock()
	defer _globalHandlerMu.Unlock()

	prev := _globalHandler
	_globalHandler = handler
	return func() { ReplaceHandlerGlobals(prev) }
}

type Handler struct {
	notificationTypes map[string]Notification
}

func NewHandler() *Handler {
	handler := &Handler{
		notificationTypes: make(map[string]Notification),
	}
	handler.RegisterNotificationTypes()
	return handler
}

// RegisterNotificationType register a new notification type
func (h *Handler) RegisterNotificationType(notification Notification) {
	h.notificationTypes[getType(notification)] = notification
}

// UnregisterNotificationType unregister a notification type
func (h *Handler) UnregisterNotificationType(notification Notification) {
	delete(h.notificationTypes, getType(notification))
}

func (h *Handler) RegisterNotificationTypes() {
	h.RegisterNotificationType(BaseNotification{})
	h.RegisterNotificationType(ExportNotification{})
}
