package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"sort"

	"github.com/go-chi/chi/v5"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/security/permissions"
	roles2 "github.com/myrteametrics/myrtea-engine-api/v5/pkg/security/roles"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/utils/httputil"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// GetRoles godoc
//
//	@Id				GetRoles
//
//	@Summary		Get all user roles
//	@Description	Gets a list of all user roles.
//	@Tags			Roles
//	@Produce		json
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{array}		roles.Role	"list of roles"
//	@Failure		500	{object}	httputil.APIError	"Internal Server Error"
//	@Router			/admin/security/roles [get]
func GetRoles(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeRole, permissions.All, permissions.ActionList)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	rolesSlice, err := roles2.R().GetAll()
	if err != nil {
		zap.L().Error("GetRoles", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}

	sort.SliceStable(rolesSlice, func(i, j int) bool {
		return rolesSlice[i].Name < rolesSlice[j].Name
	})

	httputil.JSON(w, r, rolesSlice)
}

// GetRole godoc
//
//	@Id				GetRole
//
//	@Summary		Get an user role
//	@Description	Gets an user role with the specified id
//	@Tags			Roles
//	@Produce		json
//	@Param			id	path	string	true	"role ID"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	roles.Role			"role"
//	@Failure		400	{object}	httputil.APIError	"Bad Request"
//	@Failure		404	{object}	httputil.APIError	"Status Not Found"
//	@Failure		500	{object}	httputil.APIError	"Internal Server Error"
//	@Router			/admin/security/roles/{id} [get]
func GetRole(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	roleID, err := uuid.Parse(id)
	if err != nil {
		zap.L().Warn("Parse role id", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeRole, roleID.String(), permissions.ActionGet)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	role, found, err := roles2.R().Get(roleID)
	if err != nil {
		zap.L().Error("Cannot get role", zap.String("uuid", roleID.String()), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("Role not found", zap.String("uuid", roleID.String()))
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFound, err)
		return
	}

	httputil.JSON(w, r, role)
}

// ValidateRole godoc
//
//	@Id				ValidateRole
//
//	@Summary		Validate a new role definition
//	@Description	Validate a new role definition
//	@Tags			Roles
//	@Accept			json
//	@Produce		json
//	@Param			role	body	roles.Role	true	"role (json)"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	roles.Role	"role"
//	@Failure		400	{object}	httputil.APIError	"Bad Request"
//	@Failure		500	{object}	httputil.APIError	"Internal Server Error"
//	@Router			/admin/security/roles/validate [post]
func ValidateRole(w http.ResponseWriter, r *http.Request) {
	var newRole roles2.Role
	err := json.NewDecoder(r.Body).Decode(&newRole)
	if err != nil {
		zap.L().Warn("Role json decode", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDecodeJSONBody, err)
		return
	}

	// if ok, err := newRole.IsValid(); !ok {
	// 	zap.L().Warn("Role is not valid", zap.Error(err))
	// 	render.Error(w, r, render.ErrAPIResourceInvalid, err)
	// 	return
	// }

	httputil.JSON(w, r, newRole)
}

// PostRole godoc
//
//	@Id				PostRole
//
//	@Summary		Create a new role
//	@Description	Add an user role to the user roles
//	@Tags			Roles
//	@Accept			json
//	@Produce		json
//	@Param			role	body	roles.Role	true	"role (json)"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	roles.Role	"role"
//	@Failure		400	{object}	httputil.APIError	"Bad Request"
//	@Failure		500	{object}	httputil.APIError	"Internal Server Error"
//	@Router			/admin/security/roles [post]
func PostRole(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeRole, permissions.All, permissions.ActionCreate)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	var newRole roles2.Role
	err := json.NewDecoder(r.Body).Decode(&newRole)
	if err != nil {
		zap.L().Warn("Role json decode", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDecodeJSONBody, err)
		return
	}

	// if ok, err := newRole.IsValid(); !ok {
	// 	zap.L().Warn("Role is not valid", zap.Error(err))
	// 	render.Error(w, r, render.ErrAPIResourceInvalid, err)
	// 	return
	// }

	roleID, err := roles2.R().Create(newRole)
	if err != nil {
		zap.L().Error("PostRole.Create", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBInsertFailed, err)
		return
	}

	newRole, found, err := roles2.R().Get(roleID)
	if err != nil {
		zap.L().Error("Cannot get role", zap.String("uuid", roleID.String()), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Error("Role not found after creation", zap.String("uuid", roleID.String()))
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFoundAfterInsert, err)
		return
	}

	httputil.JSON(w, r, newRole)
}

// PutRole godoc
//
//	@Id				PutRole
//
//	@Summary		Update role
//	@Description	Updates the user role information concerning the user role with id
//	@Tags			Roles
//	@Accept			json
//	@Produce		json
//	@Param			id		path	string		true	"role ID"
//	@Param			role	body	roles.Role	true	"role (json)"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	roles.Role	"role"
//	@Failure		400	{object}	httputil.APIError	"Bad Request"
//	@Failure		500	{object}	httputil.APIError	"Internal Server Error"
//	@Router			/admin/security/roles/{id} [put]
func PutRole(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	roleID, err := uuid.Parse(id)
	if err != nil {
		zap.L().Warn("Parse role id", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeRole, roleID.String(), permissions.ActionUpdate)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	var newRole roles2.Role
	err = json.NewDecoder(r.Body).Decode(&newRole)
	if err != nil {
		zap.L().Warn("Role json decode", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDecodeJSONBody, err)
		return
	}
	newRole.ID = roleID

	// if ok, err := newRole.IsValid(); !ok {
	// 	zap.L().Warn("Role is not valid", zap.Error(err))
	// 	render.Error(w, r, render.ErrAPIResourceInvalid, err)
	// 	return
	// }

	err = roles2.R().Update(newRole)
	if err != nil {
		zap.L().Error("PutRole.Update", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBUpdateFailed, err)
		return
	}

	newRole, found, err := roles2.R().Get(roleID)
	if err != nil {
		zap.L().Error("Cannot get role", zap.String("uuid", roleID.String()), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Error("Role not found after creation", zap.String("uuid", roleID.String()))
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFoundAfterInsert, err)
		return
	}

	httputil.JSON(w, r, newRole)
}

// DeleteRole godoc
//
//	@Id				DeleteRole
//
//	@Summary		Delete role
//	@Description	Deletes an user role
//	@Tags			Roles
//	@Produce		json
//	@Param			id	path	string	true	"role ID"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{string}	string				"status OK"
//	@Failure		400	{object}	httputil.APIError	"Bad Request"
//	@Failure		500	{object}	httputil.APIError	"Internal Server Error"
//	@Router			/admin/security/roles/{id} [delete]
func DeleteRole(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	roleID, err := uuid.Parse(id)
	if err != nil {
		zap.L().Warn("Parse role id", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeRole, roleID.String(), permissions.ActionDelete)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	err = roles2.R().Delete(roleID)
	if err != nil {
		zap.L().Error("Cannot delete role", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBDeleteFailed, err)
		return
	}

	httputil.OK(w, r)
}

// SetRolePermissions godoc
//
//	@Id				SetRolePermissions
//
//	@Summary		Set permissions on a role
//	@Description	Set permissions on a role
//	@Tags			Roles
//	@Accept			json
//	@Produce		json
//	@Param			id		path	string		true	"role ID"
//	@Param			role	body	[]string	true	"List of permissions UUIDs"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	roles.Role	"role"
//	@Failure		400	{object}	httputil.APIError	"Bad Request"
//	@Failure		500	{object}	httputil.APIError	"Internal Server Error"
//	@Router			/admin/security/roles/{id}/permissions [put]
func SetRolePermissions(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	roleID, err := uuid.Parse(id)
	if err != nil {
		zap.L().Warn("Parse role id", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeRole, roleID.String(), permissions.ActionUpdate)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	var rawPermissionUUIDs []string
	err = json.NewDecoder(r.Body).Decode(&rawPermissionUUIDs)
	if err != nil {
		zap.L().Warn("Invalid UUID", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDecodeJSONBody, err)
		return
	}

	permissionUUIDs := make([]uuid.UUID, 0)
	for _, rawPermissionUUID := range rawPermissionUUIDs {
		permissionUUID, err := uuid.Parse(rawPermissionUUID)
		if err != nil {
			zap.L().Warn("Invalid UUID", zap.Error(err))
			httputil.Error(w, r, httputil.ErrAPIDecodeJSONBody, err)
			return
		}
		permissionUUIDs = append(permissionUUIDs, permissionUUID)
	}

	err = roles2.R().SetRolePermissions(roleID, permissionUUIDs)
	if err != nil {
		zap.L().Error("PutRole.Update", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBUpdateFailed, err)
		return
	}

	httputil.OK(w, r)
}
