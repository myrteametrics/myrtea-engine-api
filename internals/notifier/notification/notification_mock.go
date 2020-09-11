package notification

import (
	"encoding/json"
	"time"
)

// MockNotification is an implementation of a notification main type
type MockNotification struct {
	ID           int64                  `json:"id"`
	Type         string                 `json:"type"`
	CreationDate time.Time              `json:"creationDate"`
	Groups       []int64                `json:"groups"`
	Level        string                 `json:"level"`
	Title        string                 `json:"title"`
	SubTitle     string                 `json:"subtitle"`
	Description  string                 `json:"description"`
	Context      map[string]interface{} `json:"context,omitempty"`
}

// NewMockNotification renders a new MockNotification instance
func NewMockNotification(level string, title string, subTitle string, description string, creationDate time.Time,
	groups []int64, context map[string]interface{}) *MockNotification {

	return &MockNotification{
		Type:         "mock",
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
