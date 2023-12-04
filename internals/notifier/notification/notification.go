package notification

import (
	"encoding/json"
)

// Notification is a general interface for all notifications types
type Notification interface {
	ToBytes() ([]byte, error)
	NewInstance(id int64, data []byte, isRead bool) (Notification, error)
}

// BaseNotification data structure represents a basic notification and her current state
type BaseNotification struct {
	Notification `json:"-"`
	Id           int64  `json:"id"`
	IsRead       bool   `json:"isRead"`
	Type         string `json:"type"`
}

// NewInstance returns a new instance of a BaseNotification
func (n BaseNotification) NewInstance(id int64, data []byte, isRead bool) (Notification, error) {
	var notification BaseNotification
	err := json.Unmarshal(data, &notification)
	if err != nil {
		return nil, err
	}
	notification.Id = id
	notification.IsRead = isRead
	notification.Notification = notification
	return &notification, nil
}

func (n BaseNotification) ToBytes() ([]byte, error) {
	b, err := json.Marshal(n)
	if err != nil {
		return nil, err
	}
	return b, nil
}
