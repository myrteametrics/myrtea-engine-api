package export

import (
	"testing"

	"github.com/google/uuid"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/notifier/notification"
	"github.com/myrteametrics/myrtea-sdk/v5/expression"
)

func TestExportNotification(t *testing.T) {
	// init handler
	handler := notification.NewHandler(0)
	handler.RegisterNotificationType(notification.MockNotification{})
	handler.RegisterNotificationType(ExportNotification{})
	notification.ReplaceHandlerGlobals(handler)

	notif := ExportNotification{
		Export: WrapperItem{
			Id: uuid.New().String(),
		},
		Status: 1,
	}
	notif.Id = 1
	notif.IsRead = false

	bytes, err := notif.ToBytes()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if bytes == nil {
		t.Fatalf("Expected bytes, got nil")
	}

	t.Log(string(bytes))

	// find type and create new instance
	notifType, ok := notification.H().GetNotificationByType("ExportNotification")
	if !ok {
		t.Fatalf("ExportNotification type does not exist")
	}

	instance, err := notifType.NewInstance(1, bytes, false)
	if err != nil {
		t.Fatalf("ExportNotification couldn't be instanced")
	}
	bt, _ := instance.ToBytes()
	t.Log(string(bt))

	expression.AssertEqual(t, string(bytes), string(bt))
}

func TestExportNotification_Equals(t *testing.T) {
	id := uuid.New().String()
	exportNotification := ExportNotification{
		BaseNotification: notification.BaseNotification{
			Id:     1,
			Type:   "Test",
			IsRead: true,
		},
		Export: WrapperItem{
			Id: id,
		},
		Status: 1,
	}

	expression.AssertEqual(t, exportNotification.Equals(ExportNotification{
		BaseNotification: notification.BaseNotification{
			Id:     1,
			Type:   "Test",
			IsRead: true,
		},
		Status: 1,
		Export: WrapperItem{Id: id},
	}), true)

	expression.AssertEqual(t, exportNotification.Equals(ExportNotification{
		BaseNotification: notification.BaseNotification{
			Id:     2,
			Type:   "Test",
			IsRead: true,
		},
		Status: 1,
		Export: WrapperItem{Id: id},
	}), false)

	expression.AssertEqual(t, exportNotification.Equals(ExportNotification{
		BaseNotification: notification.BaseNotification{
			Id:     1,
			Type:   "Test",
			IsRead: true,
		},
		Status: 2,
		Export: WrapperItem{Id: id},
	}), false)

	expression.AssertEqual(t, exportNotification.Equals(ExportNotification{
		BaseNotification: notification.BaseNotification{
			Id:     1,
			Type:   "Test",
			IsRead: true,
		},
		Status: 1,
		Export: WrapperItem{Id: uuid.New().String()},
	}), false)

}

func TestExportNotification_SetId(t *testing.T) {
	notif, err := ExportNotification{}.NewInstance(1, []byte(`{}`), true)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	notif = notif.SetId(2)
	exportNotification, ok := notif.(ExportNotification)
	expression.AssertEqual(t, ok, true)
	expression.AssertEqual(t, exportNotification.Id, int64(2))
}
