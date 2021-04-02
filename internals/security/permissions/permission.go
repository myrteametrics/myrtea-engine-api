package permissions

import "github.com/google/uuid"

const (
	PermissionAll          = "*"
	PermissionActionList   = "list"
	PermissionActionGet    = "get"
	PermissionActionCreate = "create"
	PermissionActionUpdate = "update"
	PermissionActionDelete = "delete"
)

type Permission struct {
	ID           uuid.UUID `json:"id"`
	ResourceType string    `json:"resourceType"`
	ResourceID   string    `json:"resourceID"`
	Action       string    `json:"action"`
}

func ListPermissions(permissions []Permission, resourceType string, resourceID string, action string) []Permission {
	lst := make([]Permission, 0)
	for _, permission := range permissions {
		if !matchPermission(permission.ResourceType, resourceType) {
			continue
		}
		if !matchPermission(permission.ResourceID, resourceID) {
			continue
		}
		if !matchPermission(permission.Action, action) {
			continue
		}
		lst = append(lst, permission)
	}
	return lst
}

func matchPermission(permission string, required string) bool {
	if permission == PermissionAll {
		return true
	}
	if required == PermissionAll {
		return true
	}
	if permission == required {
		return true
	}
	return false
}

func HasPermission(permissions []Permission, required Permission) bool {
	for _, permission := range permissions {
		if !matchResourceType(permission, required) {
			continue
		}
		if !matchResourceID(permission, required) {
			continue
		}
		if !matchResourceAction(permission, required) {
			continue
		}
		return true
	}
	return false
}

func matchResourceType(permission Permission, required Permission) bool {
	return matchPermission(permission.ResourceType, required.ResourceType)
}

func matchResourceID(permission Permission, required Permission) bool {
	return matchPermission(permission.ResourceID, required.ResourceID)
}

func matchResourceAction(permission Permission, required Permission) bool {
	return matchPermission(permission.Action, required.Action)
}
