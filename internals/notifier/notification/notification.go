package notification

import "encoding/json"

// BaseNotification data structure represents a basic notification and her current state
type BaseNotification struct {
	Notification
	Id     int64
	Type   string
	IsRead bool
}

// Notification is a general interface for all notifications types
type Notification interface {
	ToBytes() ([]byte, error)
	NewInstance(id int64, data []byte, isRead bool) (Notification, error)
}

func (n BaseNotification) ToBytes() ([]byte, error) {
	//TODO:
	return nil, nil
}

func (n BaseNotification) NewInstance(id int64, data []byte, isRead bool) (Notification, error) {
	var notification BaseNotification
	err := json.Unmarshal(data, &notification)
	if err != nil {
		return nil, err
	}
	notification.Id = id
	notification.IsRead = isRead
	return &notification, nil
}
