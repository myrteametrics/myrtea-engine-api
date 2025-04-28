package handlers

import (
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"net/http"
	"sort"

	"github.com/google/uuid"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/handlers/render"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/security/permissions"
	"go.uber.org/zap"
)

// GetPermissions godoc
//
//	@Summary		Get all user permissions
//	@Description	Gets a list of all user permissions.
//	@Tags			Permissions
//	@Produce		json
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{array}		permissions.Permission	"list of permissions"
//	@Failure		500	{string}	string					"Internal Server Error"
//	@Router			/admin/security/permissions [get]
func GetPermissions(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypePermission, permissions.All, permissions.ActionList)) {
		render.Error(w, r, render.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	permissionsSlice, err := permissions.R().GetAll()
	if err != nil {
		zap.L().Error("GetPermissions", zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}

	sort.SliceStable(permissionsSlice, func(i, j int) bool {
		return permissionsSlice[i].ResourceType < permissionsSlice[j].ResourceType
	})

	render.JSON(w, r, permissionsSlice)
}

// GetPermission godoc
//
//	@Summary		Get an user permission
//	@Description	Gets an user permission with the specified id
//	@Tags			Permissions
//	@Produce		json
//	@Param			id	path	string	true	"permission ID"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	permissions.Permission	"permission"
//	@Failure		400	{string}	string					"Bad Request"
//	@Failure		404	{string}	string					"Not Found"
//	@Failure		500	{string}	string					"Internal Server Error"
//	@Router			/admin/security/permissions/{id} [get]
func GetPermission(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	permissionID, err := uuid.Parse(id)
	if err != nil {
		zap.L().Warn("Parse role id", zap.Error(err))
		render.Error(w, r, render.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypePermission, permissionID.String(), permissions.ActionGet)) {
		render.Error(w, r, render.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	permission, found, err := permissions.R().Get(permissionID)
	if err != nil {
		zap.L().Error("Cannot get permission", zap.String("uuid", permissionID.String()), zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("Permission not found", zap.String("uuid", permissionID.String()))
		render.Error(w, r, render.ErrAPIDBResourceNotFound, err)
		return
	}

	render.JSON(w, r, permission)
}

// ValidatePermission godoc
//
//	@Summary		Validate a new permission definition
//	@Description	Validate a new permission definition
//	@Tags			Permissions
//	@Accept			json
//	@Produce		json
//	@Param			permission	body	permissions.Permission	true	"permission (json)"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	permissions.Permission	"permission"
//	@Failure		400	{string}	string					"Bad Request"
//	@Failure		500	{string}	string					"Internal Server Error"
//	@Router			/admin/security/permissions/validate [post]
func ValidatePermission(w http.ResponseWriter, r *http.Request) {
	var newPermission permissions.Permission
	err := json.NewDecoder(r.Body).Decode(&newPermission)
	if err != nil {
		zap.L().Warn("Permission json decode", zap.Error(err))
		render.Error(w, r, render.ErrAPIDecodeJSONBody, err)
		return
	}

	// if ok, err := newPermission.IsValid(); !ok {
	// 	zap.L().Warn("Permission is not valid", zap.Error(err))
	// 	render.Error(w, r, render.ErrAPIResourceInvalid, err)
	// 	return
	// }

	render.JSON(w, r, newPermission)
}

// PostPermission godoc
//
//	@Summary		Create a new permission
//	@Description	Add an user permission to the user permissions
//	@Tags			Permissions
//	@Accept			json
//	@Produce		json
//	@Param			permission	body	permissions.Permission	true	"permission (json)"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	permissions.Permission	"permission"
//	@Failure		400	{string}	string					"Bad Request"
//	@Failure		500	{string}	string					"Internal Server Error"
//	@Router			/admin/security/permissions [post]
func PostPermission(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypePermission, permissions.All, permissions.ActionList)) {
		render.Error(w, r, render.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	var newPermission permissions.Permission
	err := json.NewDecoder(r.Body).Decode(&newPermission)
	if err != nil {
		zap.L().Warn("Permission json decode", zap.Error(err))
		render.Error(w, r, render.ErrAPIDecodeJSONBody, err)
		return
	}

	// if ok, err := newPermission.IsValid(); !ok {
	// 	zap.L().Warn("Permission is not valid", zap.Error(err))
	// 	render.Error(w, r, render.ErrAPIResourceInvalid, err)
	// 	return
	// }

	permissionID, err := permissions.R().Create(newPermission)
	if err != nil {
		zap.L().Error("PostPermission.Create", zap.Error(err))
		render.Error(w, r, render.ErrAPIDBInsertFailed, err)
		return
	}

	newPermission, found, err := permissions.R().Get(permissionID)
	if err != nil {
		zap.L().Error("Cannot get permission", zap.String("uuid", permissionID.String()), zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Error("Permission not found after creation", zap.String("uuid", permissionID.String()))
		render.Error(w, r, render.ErrAPIDBResourceNotFoundAfterInsert, err)
		return
	}

	render.JSON(w, r, newPermission)
}

// PutPermission godoc
//
//	@Summary		Update permission
//	@Description	Updates the user permission information concerning the user permission with id
//	@Tags			Permissions
//	@Accept			json
//	@Produce		json
//	@Param			id			path	string					true	"permission ID"
//	@Param			permission	body	permissions.Permission	true	"permission (json)"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	permissions.Permission	"permission"
//	@Failure		400	{string}	string					"Bad Request"
//	@Failure		500	{string}	string					"Internal Server Error"
//	@Router			/admin/security/permissions/{id} [put]
func PutPermission(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	permissionID, err := uuid.Parse(id)
	if err != nil {
		zap.L().Warn("Parse role id", zap.Error(err))
		render.Error(w, r, render.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypePermission, permissionID.String(), permissions.ActionCreate)) {
		render.Error(w, r, render.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	var newPermission permissions.Permission
	err = json.NewDecoder(r.Body).Decode(&newPermission)
	if err != nil {
		zap.L().Warn("Permission json decode", zap.Error(err))
		render.Error(w, r, render.ErrAPIDecodeJSONBody, err)
		return
	}
	newPermission.ID = permissionID

	// if ok, err := newPermission.IsValid(); !ok {
	// 	zap.L().Warn("Permission is not valid", zap.Error(err))
	// 	render.Error(w, r, render.ErrAPIResourceInvalid, err)
	// 	return
	// }

	err = permissions.R().Update(newPermission)
	if err != nil {
		zap.L().Error("PutPermission.Update", zap.Error(err))
		render.Error(w, r, render.ErrAPIDBUpdateFailed, err)
		return
	}

	newPermission, found, err := permissions.R().Get(permissionID)
	if err != nil {
		zap.L().Error("Cannot get permission", zap.String("uuid", permissionID.String()), zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Error("Permission not found after creation", zap.String("uuid", permissionID.String()))
		render.Error(w, r, render.ErrAPIDBResourceNotFoundAfterInsert, err)
		return
	}

	render.JSON(w, r, newPermission)
}

// DeletePermission godoc
//
//	@Summary		Delete permission
//	@Description	Deletes an user permission
//	@Tags			Permissions
//	@Produce		json
//	@Param			id	path	string	true	"permission ID"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{string}	string	"status OK"
//	@Failure		400	{string}	string	"Bad Request"
//	@Failure		500	{string}	string	"Internal Server Error"
//	@Router			/admin/security/permissions/{id} [delete]
func DeletePermission(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	permissionID, err := uuid.Parse(id)
	if err != nil {
		zap.L().Warn("Parse role id", zap.Error(err))
		render.Error(w, r, render.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypePermission, permissionID.String(), permissions.ActionDelete)) {
		render.Error(w, r, render.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	err = permissions.R().Delete(permissionID)
	if err != nil {
		zap.L().Error("Cannot delete permission", zap.Error(err))
		render.Error(w, r, render.ErrAPIDBDeleteFailed, err)
		return
	}

	render.OK(w, r)
}
