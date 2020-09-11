package groups

import (
	"errors"

	"github.com/myrteametrics/myrtea-sdk/v4/security"
)

// Group user group
type Group struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// IsValid checks if a group definition is valid and has no missing mandatory fields
func (group *Group) IsValid() (bool, error) {
	if group.Name == "" {
		return false, errors.New("Missing Name")
	}
	return true, nil
}

// Membership relation between users and groups
type Membership struct {
	UserID  int64 `json:"userId"`
	GroupID int64 `json:"groupId"`
	Role    int64 `json:"role"`
}

// IsValid checks if a group definition is valid and has no missing mandatory fields
func (membership *Membership) IsValid() (bool, error) {
	if membership.UserID == 0 {
		return false, errors.New("Missing UserID (or 0 value)")
	}
	if membership.GroupID == 0 {
		return false, errors.New("Missing GroupID (or 0 value)")
	}
	if membership.Role == 0 {
		return false, errors.New("Missing Role (or 0 value)")
	}
	return true, nil
}

// UserWithGroups embed the standard User struct with his groups
type UserWithGroups struct {
	security.User
	Groups []GroupOfUser `json:"groups"`
}

// GroupOfUser group of a user
type GroupOfUser struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	UserRole int64  `json:"groupRole" db:"role"`
}
