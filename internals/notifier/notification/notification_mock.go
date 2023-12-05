package notification

import (
	"encoding/json"
	"time"
)

// MockNotification is an implementation of a notification main type
type MockNotification struct {
	BaseNotification
	CreationDate time.Time              `json:"creationDate"`
	Groups       []int64                `json:"groups"`
	Level        string                 `json:"level"`
	Title        string                 `json:"title"`
	SubTitle     string                 `json:"subtitle"`
	Description  string                 `json:"description"`
	Context      map[string]interface{} `json:"context,omitempty"`
}

// NewMockNotification renders a new MockNotification instance
func NewMockNotification(id int64, level string, title string, subTitle string, description string, creationDate time.Time,
	groups []int64, context map[string]interface{}) *MockNotification {

	return &MockNotification{
		BaseNotification: BaseNotification{
			Id:   id,
			Type: "MockNotification",
		},
		CreationDate: creationDate,
		Groups:       groups,
		Level:        level,
		Title:        title,
		SubTitle:     subTitle,
		Description:  description,
		Context:      context,
	}
}

// ToBytes convert a notification in a json byte slice to be sent though any required channel
func (n MockNotification) ToBytes() ([]byte, error) {
	b, err := json.Marshal(n)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// NewInstance returns a new instance of a MockNotification
func (n MockNotification) NewInstance(id int64, data []byte, isRead bool) (Notification, error) {
	var notification MockNotification
	err := json.Unmarshal(data, &notification)
	if err != nil {
		return nil, err
	}
	notification.Id = id
	notification.IsRead = isRead
	notification.Notification = notification
	return notification, nil
}

// Equals returns true if the two notifications are equals
func (n MockNotification) Equals(notification Notification) bool {
	notif, ok := notification.(MockNotification)
	if !ok {
		return ok
	}
	if !notif.BaseNotification.Equals(n.BaseNotification) {
		return false
	}
	if notif.CreationDate != n.CreationDate {
		return false
	}
	if notif.Level != n.Level {
		return false
	}
	if notif.Title != n.Title {
		return false
	}
	if notif.SubTitle != n.SubTitle {
		return false
	}
	if notif.Description != n.Description {
		return false
	}
	if notif.Context != nil && n.Context != nil {
		if len(notif.Context) != len(n.Context) {
			return false
		}
		for k, v := range notif.Context {
			if n.Context[k] != v {
				return false
			}
		}
	} else if notif.Context != nil || n.Context != nil {
		return false
	}
	if len(notif.Groups) != len(n.Groups) {
		return false
	}
	for i, v := range notif.Groups {
		if n.Groups[i] != v {
			return false
		}
	}
	return true
}

// SetId set the notification ID
func (n MockNotification) SetId(id int64) Notification {
	n.Id = id
	return n
}
