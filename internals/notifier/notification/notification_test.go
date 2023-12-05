package notification

import (
	"github.com/google/uuid"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/export"
	"github.com/myrteametrics/myrtea-sdk/v4/expression"
	"testing"
	"time"
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
	s := BaseNotification{
		Id:     1,
		Type:   "Test",
		IsRead: true,
	}
	se, e := s.ToBytes()
	if e == nil {
		t.Log(string(se))
	}

	data := []byte(`{"id":1,"type":"Test","isRead":true}`)
	notification, err := BaseNotification{}.NewInstance(1, data, true)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := BaseNotification{
		Id:     1,
		Type:   "Test",
		IsRead: true,
	}

	expression.AssertEqual(t, expected.Equals(notification), true)
}

func TestBaseNotificationNewInstanceWithInvalidData(t *testing.T) {
	data := []byte(`{"id":1,"type":"Test","isRead":"invalid"}`)
	_, err := BaseNotification{}.NewInstance(1, data, true)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

func TestExportNotification(t *testing.T) {
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

func TestBaseNotification_Equals(t *testing.T) {
	notif := BaseNotification{
		Id:     1,
		Type:   "Test",
		IsRead: true,
	}

	expression.AssertEqual(t, notif.Equals(BaseNotification{
		Id:     1,
		Type:   "Test",
		IsRead: true,
	}), true)

	expression.AssertEqual(t, notif.Equals(BaseNotification{
		Id:     2,
		Type:   "Test",
		IsRead: true,
	}), false)

	expression.AssertEqual(t, notif.Equals(BaseNotification{
		Id:     1,
		Type:   "Test2",
		IsRead: true,
	}), false)

	expression.AssertEqual(t, notif.Equals(BaseNotification{
		Id:     1,
		Type:   "Test",
		IsRead: false,
	}), false)
}

func TestMockNotification_Equals(t *testing.T) {
	baseNotification := BaseNotification{
		Id:     1,
		Type:   "Test",
		IsRead: true,
	}
	now := time.Now()
	notif := MockNotification{
		BaseNotification: baseNotification,
		CreationDate:     now,
		Level:            "info",
		Title:            "title",
		SubTitle:         "subTitle",
		Description:      "description",
		Context:          map[string]interface{}{"test": "test"},
		Groups:           []int64{1, 2},
	}

	expression.AssertEqual(t, notif.Equals(MockNotification{
		BaseNotification: baseNotification,
		CreationDate:     now,
		Level:            "info",
		Title:            "title",
		SubTitle:         "subTitle",
		Description:      "description",
		Context:          map[string]interface{}{"test": "test"},
		Groups:           []int64{1, 2},
	}), true)

	expression.AssertEqual(t, notif.Equals(MockNotification{
		BaseNotification: BaseNotification{
			Id:     2,
			Type:   "Test",
			IsRead: true,
		},
		CreationDate: now,
		Level:        "info",
		Title:        "title",
		SubTitle:     "subTitle",
		Description:  "description",
		Context:      map[string]interface{}{"test": "test"},
		Groups:       []int64{1, 2},
	}), false)

	expression.AssertEqual(t, notif.Equals(MockNotification{
		BaseNotification: baseNotification,
		CreationDate:     time.Now().AddDate(1, 0, 0),
		Level:            "info",
		Title:            "title",
		SubTitle:         "subTitle",
		Description:      "description",
		Context:          map[string]interface{}{"test": "test"},
		Groups:           []int64{1, 2},
	}), false)

	expression.AssertEqual(t, notif.Equals(MockNotification{
		BaseNotification: baseNotification,
		CreationDate:     now,
		Level:            "infos",
		Title:            "title",
		SubTitle:         "subTitle",
		Description:      "description",
		Context:          map[string]interface{}{"test": "test"},
		Groups:           []int64{1, 2},
	}), false)

	expression.AssertEqual(t, notif.Equals(MockNotification{
		BaseNotification: baseNotification,
		CreationDate:     now,
		Level:            "info",
		Title:            "titles",
		SubTitle:         "subTitle",
		Description:      "description",
		Context:          map[string]interface{}{"test": "test"},
		Groups:           []int64{1, 2},
	}), false)

	expression.AssertEqual(t, notif.Equals(MockNotification{
		BaseNotification: baseNotification,
		CreationDate:     now,
		Level:            "info",
		Title:            "title",
		SubTitle:         "subTitles",
		Description:      "description",
		Context:          map[string]interface{}{"test": "test"},
		Groups:           []int64{1, 2},
	}), false)
	expression.AssertEqual(t, notif.Equals(MockNotification{
		BaseNotification: baseNotification,
		CreationDate:     now,
		Level:            "info",
		Title:            "title",
		SubTitle:         "subTitle",
		Description:      "descriptions",
		Context:          map[string]interface{}{"test": "test"},
		Groups:           []int64{1, 2},
	}), false)

	expression.AssertEqual(t, notif.Equals(MockNotification{
		BaseNotification: baseNotification,
		CreationDate:     now,
		Level:            "info",
		Title:            "title",
		SubTitle:         "subTitle",
		Description:      "description",
		Context:          map[string]interface{}{"tests": "test"},
		Groups:           []int64{1, 2},
	}), false)

	expression.AssertEqual(t, notif.Equals(MockNotification{
		BaseNotification: baseNotification,
		CreationDate:     now,
		Level:            "info",
		Title:            "title",
		SubTitle:         "subTitle",
		Description:      "description",
		Context:          map[string]interface{}{"test": "test"},
		Groups:           []int64{1, 2, 3},
	}), false)

}

func TestExportNotification_Equals(t *testing.T) {
	id := uuid.New().String()
	exportNotification := ExportNotification{
		BaseNotification: BaseNotification{
			Id:     1,
			Type:   "Test",
			IsRead: true,
		},
		Export: export.WrapperItem{
			Id: id,
		},
		Status: 1,
	}

	expression.AssertEqual(t, exportNotification.Equals(ExportNotification{
		BaseNotification: BaseNotification{
			Id:     1,
			Type:   "Test",
			IsRead: true,
		},
		Status: 1,
		Export: export.WrapperItem{Id: id},
	}), true)

	expression.AssertEqual(t, exportNotification.Equals(ExportNotification{
		BaseNotification: BaseNotification{
			Id:     2,
			Type:   "Test",
			IsRead: true,
		},
		Status: 1,
		Export: export.WrapperItem{Id: id},
	}), false)

	expression.AssertEqual(t, exportNotification.Equals(ExportNotification{
		BaseNotification: BaseNotification{
			Id:     1,
			Type:   "Test",
			IsRead: true,
		},
		Status: 2,
		Export: export.WrapperItem{Id: id},
	}), false)

	expression.AssertEqual(t, exportNotification.Equals(ExportNotification{
		BaseNotification: BaseNotification{
			Id:     1,
			Type:   "Test",
			IsRead: true,
		},
		Status: 1,
		Export: export.WrapperItem{Id: uuid.New().String()},
	}), false)

}
