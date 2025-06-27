package historyconfig

import (
	"errors"
	"time"
)

// ConfigHistory represents a history entry for configuration changes
type ConfigHistory struct {
	ID         int64  `json:"id"`                    // timestamp in milliseconds
	Commentary string `json:"commentary"`            // optional comment about the change
	Type       string `json:"type" db:"update_type"` // type of configuration change
	User       string `json:"user" db:"update_user"` // user who made the change
	Config     string `json:"config"`                // configuration content, can be very large (1-2MB+)
}

// NewConfigHistory creates a new ConfigHistory instance with the current timestamp as ID
func NewConfigHistory(commentary string, changeType string, user string, config string) ConfigHistory {
	return ConfigHistory{
		ID:         time.Now().UnixNano() / int64(time.Millisecond),
		Commentary: commentary,
		Type:       changeType,
		User:       user,
		Config:     config,
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
