package notification

import (
	"github.com/myrteametrics/myrtea-sdk/v5/expression"
	"testing"
)

func TestNewHandler(t *testing.T) {
	handler := NewHandler(0)
	expression.AssertNotEqual(t, handler, nil, "NewHandler() should not return nil")
}

func TestHandler_RegisterNotificationType_AddsNewType(t *testing.T) {
	handler := NewHandler(0)
	notification := BaseNotification{}
	handler.RegisterNotificationType(notification)
	_, exists := handler.notificationTypes[getType(notification)]
	expression.AssertEqual(t, exists, true, "RegisterNotificationType() should add new type")
}

func TestHandler_RegisterNotificationType_OverwritesExistingType(t *testing.T) {
	handler := NewHandler(0)
	notification := BaseNotification{}
	handler.RegisterNotificationType(notification)
	notification2 := BaseNotification{} // Assuming this has the same type as the first one
	handler.RegisterNotificationType(notification2)
	expression.AssertEqual(t, handler.notificationTypes[getType(notification)], notification2, "RegisterNotificationType() should overwrite existing type")
}

func TestHandler_UnregisterNotificationType_RemovesExistingType(t *testing.T) {
	handler := NewHandler(0)
	notification := BaseNotification{}
	handler.RegisterNotificationType(notification)
	handler.UnregisterNotificationType(notification)
	_, exists := handler.notificationTypes[getType(notification)]
	expression.AssertEqual(t, exists, false, "UnregisterNotificationType() should remove existing type")
}

func TestHandler_UnregisterNotificationType_DoesNothingForNonExistingType(t *testing.T) {
	handler := NewHandler(0)
	notification := BaseNotification{}
	handler.UnregisterNotificationType(notification)
	_, exists := handler.notificationTypes[getType(notification)]
	expression.AssertEqual(t, exists, false, "UnregisterNotificationType() should do nothing for non-existing type")
}

func TestReplaceHandlerGlobals_ReplacesGlobalHandler(t *testing.T) {
	handler := NewHandler(0)
	prevHandler := H()
	undo := ReplaceHandlerGlobals(handler)
	expression.AssertEqual(t, H(), handler, "ReplaceHandlerGlobals() should replace global handler")
	undo()
	expression.AssertEqual(t, H(), prevHandler, "Undo function should restore previous global handler")
}
