package notification

import (
	"encoding/json"
)

// Notification is a general interface for all notifications types
type Notification interface {
	ToBytes() ([]byte, error)
	NewInstance(id int64, data []byte, isRead bool) (Notification, error)
	Equals(notification Notification) bool
	SetId(id int64) Notification
	SetPersistent(persistent bool) Notification
	IsPersistent() bool
}

// BaseNotification data structure represents a basic notification and her current state
type BaseNotification struct {
	Notification `json:"-"`
	Id           int64  `json:"id"`
	IsRead       bool   `json:"isRead"`
	Type         string `json:"type"`
	Persistent   bool   `json:"persistent"` // is notification saved in db or not ?
}

// NewBaseNotification returns a new instance of a BaseNotification
func NewBaseNotification(id int64, isRead bool, persistent bool) BaseNotification {
	return BaseNotification{
		Id:         id,
		IsRead:     isRead,
		Persistent: persistent,
		Type:       "BaseNotification",
	}
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
	return notification, nil
}

// ToBytes convert a notification in a json byte slice to be sent through any required channel
func (n BaseNotification) ToBytes() ([]byte, error) {
	b, err := json.Marshal(n)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// Equals returns true if the two notifications are equals
func (n BaseNotification) Equals(notification Notification) bool {
	notif, ok := notification.(BaseNotification)
	if !ok {
		return ok
	}
	if n.Id != notif.Id {
		return false
	}
	if n.IsRead != notif.IsRead {
		return false
	}
	if n.Type != notif.Type {
		return false
	}
	return true
}

// SetId set the notification ID
func (n BaseNotification) SetId(id int64) Notification {
	n.Id = id
	return n
}

// SetPersistent sets whether the notification is persistent (saved to a database)
func (n BaseNotification) SetPersistent(persistent bool) Notification {
	n.Persistent = persistent
	return n
}

// IsPersistent returns whether the notification is persistent (saved to a database)
func (n BaseNotification) IsPersistent() bool {
	return n.Persistent
}
