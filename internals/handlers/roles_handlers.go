package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"sort"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/handlers/render"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/security/permissions"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/security/roles"
	"go.uber.org/zap"
)

// GetRoles godoc
// @Summary Get all user roles
// @Description Gets a list of all user roles.
// @Tags Roles
// @Produce json
// @Security Bearer
// @Success 200 {array} roles.Role "list of roles"
// @Failure 500 {string} string "Internal Server Error"
// @Router /admin/security/roles [get]
func GetRoles(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeRole, permissions.All, permissions.ActionList)) {
		render.Error(w, r, render.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	rolesSlice, err := roles.R().GetAll()
	if err != nil {
		zap.L().Error("GetRoles", zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}

	sort.SliceStable(rolesSlice, func(i, j int) bool {
		return rolesSlice[i].Name < rolesSlice[j].Name
	})

	render.JSON(w, r, rolesSlice)
}

// GetRole godoc
// @Summary Get an user role
// @Description Gets an user role with the specified id
// @Tags Roles
// @Produce json
// @Param id path string true "role ID"
// @Security Bearer
// @Success 200 {object} roles.Role "role"
// @Failure 400 {string} string "Bad Request"
// @Failure 404 {string} string "Not Found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /admin/security/roles/{id} [get]
func GetRole(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	roleID, err := uuid.Parse(id)
	if err != nil {
		zap.L().Warn("Parse role id", zap.Error(err))
		render.Error(w, r, render.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeRole, roleID.String(), permissions.ActionGet)) {
		render.Error(w, r, render.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	role, found, err := roles.R().Get(roleID)
	if err != nil {
		zap.L().Error("Cannot get role", zap.String("uuid", roleID.String()), zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("Role not found", zap.String("uuid", roleID.String()))
		render.Error(w, r, render.ErrAPIDBResourceNotFound, err)
		return
	}

	render.JSON(w, r, role)
}

// ValidateRole godoc
// @Summary Validate a new role definition
// @Description Validate a new role definition
// @Tags Roles
// @Accept json
// @Produce json
// @Param role body roles.Role true "role (json)"
// @Security Bearer
// @Success 200 {object} roles.Role "role"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /admin/security/roles/validate [post]
func ValidateRole(w http.ResponseWriter, r *http.Request) {
	var newRole roles.Role
	err := json.NewDecoder(r.Body).Decode(&newRole)
	if err != nil {
		zap.L().Warn("Role json decode", zap.Error(err))
		render.Error(w, r, render.ErrAPIDecodeJSONBody, err)
		return
	}

	// if ok, err := newRole.IsValid(); !ok {
	// 	zap.L().Warn("Role is not valid", zap.Error(err))
	// 	render.Error(w, r, render.ErrAPIResourceInvalid, err)
	// 	return
	// }

	render.JSON(w, r, newRole)
}

// PostRole godoc
// @Summary Create a new role
// @Description Add an user role to the user roles
// @Tags Roles
// @Accept json
// @Produce json
// @Param role body roles.Role true "role (json)"
// @Security Bearer
// @Success 200 {object} roles.Role "role"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /admin/security/roles [post]
func PostRole(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeRole, permissions.All, permissions.ActionCreate)) {
		render.Error(w, r, render.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	var newRole roles.Role
	err := json.NewDecoder(r.Body).Decode(&newRole)
	if err != nil {
		zap.L().Warn("Role json decode", zap.Error(err))
		render.Error(w, r, render.ErrAPIDecodeJSONBody, err)
		return
	}

	// if ok, err := newRole.IsValid(); !ok {
	// 	zap.L().Warn("Role is not valid", zap.Error(err))
	// 	render.Error(w, r, render.ErrAPIResourceInvalid, err)
	// 	return
	// }

	roleID, err := roles.R().Create(newRole)
	if err != nil {
		zap.L().Error("PostRole.Create", zap.Error(err))
		render.Error(w, r, render.ErrAPIDBInsertFailed, err)
		return
	}

	newRole, found, err := roles.R().Get(roleID)
	if err != nil {
		zap.L().Error("Cannot get role", zap.String("uuid", roleID.String()), zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Error("Role not found after creation", zap.String("uuid", roleID.String()))
		render.Error(w, r, render.ErrAPIDBResourceNotFoundAfterInsert, err)
		return
	}

	render.JSON(w, r, newRole)
}

// PutRole godoc
// @Summary Update role
// @Description Updates the user role information concerning the user role with id
// @Tags Roles
// @Accept json
// @Produce json
// @Param id path string true "role ID"
// @Param role body roles.Role true "role (json)"
// @Security Bearer
// @Success 200 {object} roles.Role "role"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /admin/security/roles/{id} [put]
func PutRole(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	roleID, err := uuid.Parse(id)
	if err != nil {
		zap.L().Warn("Parse role id", zap.Error(err))
		render.Error(w, r, render.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeRole, roleID.String(), permissions.ActionUpdate)) {
		render.Error(w, r, render.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	var newRole roles.Role
	err = json.NewDecoder(r.Body).Decode(&newRole)
	if err != nil {
		zap.L().Warn("Role json decode", zap.Error(err))
		render.Error(w, r, render.ErrAPIDecodeJSONBody, err)
		return
	}
	newRole.ID = roleID

	// if ok, err := newRole.IsValid(); !ok {
	// 	zap.L().Warn("Role is not valid", zap.Error(err))
	// 	render.Error(w, r, render.ErrAPIResourceInvalid, err)
	// 	return
	// }

	err = roles.R().Update(newRole)
	if err != nil {
		zap.L().Error("PutRole.Update", zap.Error(err))
		render.Error(w, r, render.ErrAPIDBUpdateFailed, err)
		return
	}

	newRole, found, err := roles.R().Get(roleID)
	if err != nil {
		zap.L().Error("Cannot get role", zap.String("uuid", roleID.String()), zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Error("Role not found after creation", zap.String("uuid", roleID.String()))
		render.Error(w, r, render.ErrAPIDBResourceNotFoundAfterInsert, err)
		return
	}

	render.JSON(w, r, newRole)
}

// DeleteRole godoc
// @Summary Delete role
// @Description Deletes an user role
// @Tags Roles
// @Produce json
// @Param id path string true "role ID"
// @Security Bearer
// @Success 200 {string} string "status OK"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /admin/security/roles/{id} [delete]
func DeleteRole(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	roleID, err := uuid.Parse(id)
	if err != nil {
		zap.L().Warn("Parse role id", zap.Error(err))
		render.Error(w, r, render.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeRole, roleID.String(), permissions.ActionDelete)) {
		render.Error(w, r, render.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	err = roles.R().Delete(roleID)
	if err != nil {
		zap.L().Error("Cannot delete role", zap.Error(err))
		render.Error(w, r, render.ErrAPIDBDeleteFailed, err)
		return
	}

	render.OK(w, r)
}

// SetRolePermissions godoc
// @Summary Set permissions on a role
// @Description Set permissions on a role
// @Tags Roles
// @Accept json
// @Produce json
// @Param id path string true "role ID"
// @Param role body []string true "List of permissions UUIDs"
// @Security Bearer
// @Success 200 {object} roles.Role "role"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /admin/security/roles/{id}/permissions [put]
func SetRolePermissions(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	roleID, err := uuid.Parse(id)
	if err != nil {
		zap.L().Warn("Parse role id", zap.Error(err))
		render.Error(w, r, render.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeRole, roleID.String(), permissions.ActionUpdate)) {
		render.Error(w, r, render.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	var rawPermissionUUIDs []string
	err = json.NewDecoder(r.Body).Decode(&rawPermissionUUIDs)
	if err != nil {
		zap.L().Warn("Invalid UUID", zap.Error(err))
		render.Error(w, r, render.ErrAPIDecodeJSONBody, err)
		return
	}

	permissionUUIDs := make([]uuid.UUID, 0)
	for _, rawPermissionUUID := range rawPermissionUUIDs {
		permissionUUID, err := uuid.Parse(rawPermissionUUID)
		if err != nil {
			zap.L().Warn("Invalid UUID", zap.Error(err))
			render.Error(w, r, render.ErrAPIDecodeJSONBody, err)
			return
		}
		permissionUUIDs = append(permissionUUIDs, permissionUUID)
	}

	err = roles.R().SetRolePermissions(roleID, permissionUUIDs)
	if err != nil {
		zap.L().Error("PutRole.Update", zap.Error(err))
		render.Error(w, r, render.ErrAPIDBUpdateFailed, err)
		return
	}

	render.OK(w, r)
}
