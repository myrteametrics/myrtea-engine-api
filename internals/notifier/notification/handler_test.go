package notification

import (
	"github.com/myrteametrics/myrtea-sdk/v4/expression"
	"testing"
)

func TestNewHandler(t *testing.T) {
	handler := NewHandler()
	expression.AssertNotEqual(t, handler, nil, "NewHandler() should not return nil")
	expression.AssertNotEqual(t, len(handler.notificationTypes), 0, "NewHandler() should not return an empty notificationTypes")
}

func TestHandler_RegisterNotificationType(t *testing.T) {
}

func TestHandler_RegisterNotificationTypes(t *testing.T) {

}

func TestHandler_UnregisterNotificationType(t *testing.T) {

}
