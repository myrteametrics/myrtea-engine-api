package notification

import (
	"github.com/google/uuid"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/export"
	"github.com/myrteametrics/myrtea-sdk/v4/expression"
	"reflect"
	"testing"
)

func TestBaseNotificationToBytes(t *testing.T) {
	notification := BaseNotification{
		Id:     1,
		Type:   "Test",
		IsRead: true,
	}

	bytes, err := notification.ToBytes()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if bytes == nil {
		t.Errorf("Expected bytes, got nil")
	}
}

func TestBaseNotificationNewInstance(t *testing.T) {
	data := []byte(`{"Notification":null,"Id":1,"Type":"Test","IsRead":true}`)
	notification, err := BaseNotification{}.NewInstance(1, data, true)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := &BaseNotification{
		Id:     1,
		Type:   "Test",
		IsRead: true,
	}

	expression.AssertEqual(t, reflect.DeepEqual(notification, expected), true)
}

func TestBaseNotificationNewInstanceWithInvalidData(t *testing.T) {
	data := []byte(`{"Notification":null,"Id":1,"Type":"Test","IsRead":"invalid"}`)
	_, err := BaseNotification{}.NewInstance(1, data, true)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

func TextExportNotification(t *testing.T) {
	// init handler
	ReplaceHandlerGlobals(NewHandler())

	notification := ExportNotification{
		Export: export.WrapperItem{
			Id: uuid.New().String(),
		},
		Status: 1,
	}
	notification.Id = 1
	notification.IsRead = false

	bytes, err := notification.ToBytes()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if bytes == nil {
		t.Errorf("Expected bytes, got nil")
	}

	t.Log(string(bytes))

	// find type and create new instance
	notifType, ok := H().notificationTypes["ExportNotification"]
	if !ok {
		t.Errorf("Notification type does not exist")
		t.FailNow()
	}

	instance, err := notifType.NewInstance(1, bytes, false)
	if err != nil {
		t.Errorf("Notification couldn't be instanced")
		t.FailNow()
	}
	bt, _ := instance.ToBytes()
	t.Log(string(bt))

	expression.AssertEqual(t, string(bytes), string(bt))
}
