package users

import "github.com/myrteametrics/myrtea-sdk/v4/security"

// UserOfGroup user in a group
type UserOfGroup struct {
	security.User
	RoleInGroup int64 `json:"groupRole" db:"role"`
}
