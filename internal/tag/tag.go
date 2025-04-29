package tag

import (
	"errors"
	"time"
)

// Tag is a struct used to represent a tag (that either applies to a situation or a template instance)
type Tag struct {
	Id          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Color       string    `json:"color"`
	UpdatedAt   time.Time `json:"updatedAt"`
	CreatedAt   time.Time `json:"createdAt"`
}

// IsValid checks if a tag definition is valid and has no missing mandatory fields
func (t Tag) IsValid() (bool, error) {
	if t.Name == "" {
		return false, errors.New("missing name")
	}
	if len(t.Name) > 32 {
		return false, errors.New("invalid name")
	}
	if t.Color == "" {
		return false, errors.New("missing color")
	}
	// color must be a hex color (#RRGGBB regex)
	if len(t.Color) != 7 || t.Color[0] != '#' {
		return false, errors.New("invalid color")
	}

	return true, nil
}
