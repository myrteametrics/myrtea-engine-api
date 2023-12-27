package export

import (
	"encoding/json"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/notifier/notification"
	"reflect"
)

const (
	ExportNotificationStarted  = 0
	ExportNotificationArchived = 1 // happens when
	ExportNotificationDeleted  = 2 // happens when the export is deleted from archive
)

type ExportNotification struct {
	notification.BaseNotification
	Export WrapperItem `json:"export"`
	Status int         `json:"status"`
}

func NewExportNotification(id int64, export WrapperItem, status int) *ExportNotification {
	return &ExportNotification{
		BaseNotification: notification.BaseNotification{
			Id:         id,
			Type:       "ExportNotification",
			Persistent: false,
		},
		Export: export,
		Status: status,
	}
}

// ToBytes convert a notification in a json byte slice to be sent through any required channel
func (e ExportNotification) ToBytes() ([]byte, error) {
	b, err := json.Marshal(e)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// NewInstance returns a new instance of a ExportNotification
func (e ExportNotification) NewInstance(id int64, data []byte, isRead bool) (notification.Notification, error) {
	var notif ExportNotification
	err := json.Unmarshal(data, &notif)
	if err != nil {
		return nil, err
	}
	notif.Id = id
	notif.IsRead = isRead
	notif.Notification = notif
	return notif, nil
}

// Equals returns true if the two notifications are equals
func (e ExportNotification) Equals(notification notification.Notification) bool {
	notif, ok := notification.(ExportNotification)
	if !ok {
		return ok
	}
	if !notif.BaseNotification.Equals(e.BaseNotification) {
		return false
	}
	if !reflect.DeepEqual(notif.Export, e.Export) {
		return false
	}
	if notif.Status != e.Status {
		return false
	}
	return true
}

// SetId set the notification ID
func (e ExportNotification) SetId(id int64) notification.Notification {
	e.Id = id
	return e
}

// SetPersistent sets whether the notification is persistent (saved to a database)
func (e ExportNotification) SetPersistent(persistent bool) notification.Notification {
	e.Persistent = persistent
	return e
}

// IsPersistent returns whether the notification is persistent (saved to a database)
func (e ExportNotification) IsPersistent() bool {
	return e.Persistent
}
