package config_history

import (
	"errors"
	"time"
)

// ConfigHistory represents a history entry for configuration changes
type ConfigHistory struct {
	ID         int64  `json:"id"`         // timestamp in milliseconds
	Commentary string `json:"commentary"` // optional comment about the change
	Type       string `json:"type"`       // type of configuration change
	User       string `json:"user"`       // user who made the change
}

// NewConfigHistory creates a new ConfigHistory instance with the current timestamp as ID
func NewConfigHistory(commentary string, changeType string, user string) ConfigHistory {
	return ConfigHistory{
		ID:         time.Now().UnixNano() / int64(time.Millisecond),
		Commentary: commentary,
		Type:       changeType,
		User:       user,
	}
}

// IsValid checks if a ConfigHistory has valid required fields
func (ch *ConfigHistory) IsValid() (bool, error) {
	if ch.ID <= 0 {
		return false, errors.New("missing or invalid ID")
	}
	if ch.Type == "" {
		return false, errors.New("missing Type")
	}
	if ch.User == "" {
		return false, errors.New("missing User")
	}
	return true, nil
}
