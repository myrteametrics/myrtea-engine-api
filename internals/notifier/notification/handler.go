package notification

import (
	"context"
	"go.uber.org/zap"
	"sync"
	"time"
)

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
	notificationTypes    map[string]Notification
	notificationLifetime time.Duration
}

// NewHandler returns a pointer to a new instance of Handler
func NewHandler(notificationLifetime time.Duration) *Handler {
	handler := &Handler{
		notificationTypes:    make(map[string]Notification),
		notificationLifetime: notificationLifetime,
	}
	handler.RegisterNotificationTypes()

	// useless to start cleaner if lifetime is less than 0
	if notificationLifetime > 0 {
		go handler.startCleaner(context.Background())
	} else {
		zap.L().Info("Notification cleaner will not be started", zap.Duration("notificationLifetime", notificationLifetime))
	}

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

// RegisterNotificationTypes register all notification types
func (h *Handler) RegisterNotificationTypes() {
	h.RegisterNotificationType(BaseNotification{})
	h.RegisterNotificationType(ExportNotification{})
}

// startCleaner start a ticker to clean expired notifications in database every 24 hours
func (h *Handler) startCleaner(context context.Context) {
	cleanRate := time.Hour * 24
	zap.L().Info("Starting notification cleaner", zap.Duration("cleanRate", cleanRate), zap.Duration("notificationLifetime", h.notificationLifetime))
	ticker := time.NewTicker(cleanRate)
	defer ticker.Stop()
	for {
		select {
		case <-context.Done():
			return
		case <-ticker.C:
			affectedRows, err := R().CleanExpired(h.notificationLifetime)
			if err != nil {
				zap.L().Error("Error while cleaning expired notifications", zap.Error(err))
			} else {
				zap.L().Debug("Cleaned expired notifications", zap.Int64("affectedRows", affectedRows))
			}
		}
	}
}
