package notification

import (
	"encoding/json"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/export"
)

type ExportNotification struct {
	BaseNotification
	Export export.WrapperItem `json:"export"`
	Status int                `json:"status"`
}

func NewExportNotification(id int64, export export.WrapperItem, status int) *ExportNotification {
	return &ExportNotification{
		BaseNotification: BaseNotification{
			Id:   id,
			Type: "ExportNotification",
		},
		Export: export,
		Status: status,
	}
}

func (e ExportNotification) ToBytes() ([]byte, error) {
	b, err := json.Marshal(e)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// NewInstance returns a new instance of a ExportNotification
func (e ExportNotification) NewInstance(id int64, data []byte, isRead bool) (Notification, error) {
	var notification ExportNotification
	err := json.Unmarshal(data, &notification)
	if err != nil {
		return nil, err
	}
	notification.Id = id
	notification.IsRead = isRead
	notification.Notification = notification
	return &notification, nil
}
