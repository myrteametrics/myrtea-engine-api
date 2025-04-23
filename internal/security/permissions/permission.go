package permissions

import "github.com/google/uuid"

const (
	All          = "*"
	ActionList   = "list"
	ActionGet    = "get"
	ActionCreate = "create"
	ActionUpdate = "update"
	ActionDelete = "delete"
	ActionSearch = "search"
	ActionAccess = "access"

	TypeUser              = "user"
	TypePermission        = "permission"
	TypeRole              = "role"
	TypeFact              = "fact"
	TypeRule              = "rule"
	TypeSituation         = "situation"
	TypeSituationInstance = "situation_instance"
	TypeSituationFacts    = "situation_fact"
	TypeSituationRules    = "situation_rule"
	TypeSituationIssues   = "situation_issue"
	TypeScheduler         = "scheduler"
	TypeCalendar          = "calendar"
	TypeModel             = "model"
	TypeFrontend          = "frontend"
	TypeExport            = "export"
	TypeService           = "service"
	TypeConfig            = "config"
	TypeAPIKey            = "api_key"
)

type Permission struct {
	ID           uuid.UUID `json:"id"`
	ResourceType string    `json:"resourceType"`
	ResourceID   string    `json:"resourceId"`
	Action       string    `json:"action"`
}

// New returns a new Permission
func New(resourceType string, resourceID string, action string) Permission {
	return Permission{
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Action:       action,
	}
}

// ListMatchingPermissions returns a list of permissions matching the given permission
func ListMatchingPermissions(permissions []Permission, match Permission) []Permission {
	lst := make([]Permission, 0)
	for _, permission := range permissions {
		if !matchPermission(permission.ResourceType, match.ResourceType) {
			continue
		}
		if !matchPermission(permission.ResourceID, match.ResourceID) {
			continue
		}
		if !matchPermission(permission.Action, match.Action) {
			continue
		}
		lst = append(lst, permission)
	}
	return lst
}

// GetResourceIDs returns a list of resource IDs from a list of permissions
func GetResourceIDs(permissions []Permission) []string {
	resourceIDs := make([]string, 0)
	for _, permission := range permissions {
		resourceIDs = append(resourceIDs, permission.ResourceID)
	}
	return resourceIDs
}

// HasPermission checks if the user has the required permission
func matchPermission(permission string, required string) bool {
	if permission == All {
		return true
	}
	if required == All {
		return true
	}
	if permission == required {
		return true
	}
	return false
}

// HasPermission checks strictly if the user has the required permission
func matchPermissionStrict(permission string, required string) bool {
	if permission == All {
		return true
	}
	if permission == required {
		return true
	}
	return false
}

// HasPermission checks if the user has the required permission
func HasPermission(permissions []Permission, required Permission) bool {
	for _, permission := range permissions {
		if !matchPermissionStrict(permission.ResourceType, required.ResourceType) {
			continue
		}
		if !matchPermissionStrict(permission.ResourceID, required.ResourceID) {
			continue
		}
		if !matchPermissionStrict(permission.Action, required.Action) {
			continue
		}
		return true
	}
	return false
}

// HasPermissionAtLeastOne checks if the user has at least one of the required permissions
func HasPermissionAtLeastOne(permissions []Permission, requiredAtLeastOne []Permission) bool {
	for _, required := range requiredAtLeastOne {
		if HasPermission(permissions, required) {
			return true
		}
	}
	return false
}

// HasPermissionAll checks if the user has all the required permissions
func HasPermissionAll(permissions []Permission, requiredAll []Permission) bool {
	for _, required := range requiredAll {
		if !HasPermission(permissions, required) {
			return false
		}
	}
	return true
}
