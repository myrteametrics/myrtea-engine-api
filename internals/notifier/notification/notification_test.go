package notification

import (
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

func TestBaseNotification_SetId(t *testing.T) {
	notif, err := BaseNotification{}.NewInstance(1, []byte(`{}`), true)
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	notif = notif.SetId(2)
	baseNotification, ok := notif.(BaseNotification)
	expression.AssertEqual(t, ok, true)
	expression.AssertEqual(t, baseNotification.Id, int64(2))
}

func TestMockNotification_SetId(t *testing.T) {
	notif, err := MockNotification{}.NewInstance(1, []byte(`{}`), true)
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	notif = notif.SetId(2)
	mockNotification, ok := notif.(MockNotification)
	expression.AssertEqual(t, ok, true)
	expression.AssertEqual(t, mockNotification.Id, int64(2))
}
