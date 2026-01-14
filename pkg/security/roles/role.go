package roles

import (
	uuid "github.com/google/uuid"
)

type Role struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	HomePage string    `json:"home_page,omitempty"`
}
