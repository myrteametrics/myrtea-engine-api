package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/security/permissions"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/utils/httputil"
	"github.com/myrteametrics/myrtea-sdk/v5/repositories/externalconfig"
	"go.uber.org/zap"
)

const maxFolderDepth = 10

// PostExternalConfigFolder godoc
//
//	@Id				PostExternalConfigFolder
//
//	@Summary		Create a new external config folder
//	@Description	Create a new folder in the external config hierarchy. Maximum folder depth is 10.
//	@Tags			ExternalConfigs
//	@Accept			json
//	@Produce		json
//	@Param			folder	body	externalconfig.Folder	true	"Folder definition (json)"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	externalconfig.Folder	"folder"
//	@Failure		400	{object}	httputil.APIError		"Bad Request"
//	@Failure		403	{object}	httputil.APIError		"Forbidden"
//	@Failure		422	{object}	httputil.APIError		"Unprocessable Entity - max depth reached"
//	@Failure		500	{object}	httputil.APIError		"Internal Server Error"
//	@Router			/engine/externalconfigs/folders [post]
func PostExternalConfigFolder(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeConfig, permissions.All, permissions.ActionCreate)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	var folder externalconfig.Folder
	if err := json.NewDecoder(r.Body).Decode(&folder); err != nil {
		zap.L().Warn("ExternalConfigFolder json decoding", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDecodeJSONBody, err)
		return
	}

	// Check parent folder depth before creating
	if folder.ParentId != nil {
		depth, err := externalconfig.R().GetFolderParentCount(*folder.ParentId)
		if err != nil {
			zap.L().Error("Error getting folder parent count", zap.Error(err))
			httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
			return
		}
		// depth is the number of ancestors of the parent; new folder depth = depth + 1
		// we want total depth (from root) of new folder <= maxFolderDepth
		if depth+1 >= maxFolderDepth {
			zap.L().Warn("Max folder depth reached", zap.Int64p("ParentId", folder.ParentId), zap.Int("depth", depth))
			httputil.Error(w, r, httputil.ErrAPIProcessError, fmt.Errorf("maximum folder depth of %d has been reached", maxFolderDepth))
			return
		}
	}

	id, err := externalconfig.R().CreateFolder(folder)
	if err != nil {
		zap.L().Error("Error while creating the ExternalConfigFolder", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBInsertFailed, err)
		return
	}

	created, found, err := externalconfig.R().GetFolder(id)
	if err != nil {
		zap.L().Error("Cannot get ExternalConfigFolder after creation", zap.Int64("id", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("ExternalConfigFolder not found after creation", zap.Int64("id", id))
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFoundAfterInsert, err)
		return
	}

	httputil.JSON(w, r, created)
}

// GetExternalConfigFolder godoc
//
//	@Id				GetExternalConfigFolder
//
//	@Summary		Get an external config folder
//	@Description	Get an external config folder by ID
//	@Tags			ExternalConfigs
//	@Produce		json
//	@Param			id	path	int	true	"Folder ID"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	externalconfig.Folder	"folder"
//	@Failure		400	{object}	httputil.APIError		"Bad Request"
//	@Failure		403	{object}	httputil.APIError		"Forbidden"
//	@Failure		404	{object}	httputil.APIError		"Not Found"
//	@Failure		500	{object}	httputil.APIError		"Internal Server Error"
//	@Router			/engine/externalconfigs/folders/{id} [get]
func GetExternalConfigFolder(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeConfig, permissions.All, permissions.ActionGet)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		zap.L().Warn("Error parsing folder id", zap.String("id", chi.URLParam(r, "id")), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	folder, found, err := externalconfig.R().GetFolder(id)
	if err != nil {
		zap.L().Error("Cannot get ExternalConfigFolder", zap.Int64("id", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("ExternalConfigFolder not found", zap.Int64("id", id))
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFound, nil)
		return
	}

	httputil.JSON(w, r, folder)
}

// PutExternalConfigFolder godoc
//
//	@Id				PutExternalConfigFolder
//
//	@Summary		Update an external config folder
//	@Description	Update an existing external config folder
//	@Tags			ExternalConfigs
//	@Accept			json
//	@Produce		json
//	@Param			id		path	int						true	"Folder ID"
//	@Param			folder	body	externalconfig.Folder	true	"Folder definition (json)"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	externalconfig.Folder	"folder"
//	@Failure		400	{object}	httputil.APIError		"Bad Request"
//	@Failure		403	{object}	httputil.APIError		"Forbidden"
//	@Failure		404	{object}	httputil.APIError		"Not Found"
//	@Failure		500	{object}	httputil.APIError		"Internal Server Error"
//	@Router			/engine/externalconfigs/folders/{id} [put]
func PutExternalConfigFolder(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeConfig, permissions.All, permissions.ActionUpdate)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		zap.L().Warn("Error parsing folder id", zap.String("id", chi.URLParam(r, "id")), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	var folder externalconfig.Folder
	if err := json.NewDecoder(r.Body).Decode(&folder); err != nil {
		zap.L().Warn("ExternalConfigFolder json decoding", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDecodeJSONBody, err)
		return
	}
	folder.Id = id

	if err := externalconfig.R().UpdateFolder(id, folder); err != nil {
		zap.L().Error("Error while updating ExternalConfigFolder", zap.Int64("id", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBUpdateFailed, err)
		return
	}

	updated, found, err := externalconfig.R().GetFolder(id)
	if err != nil {
		zap.L().Error("Cannot get ExternalConfigFolder after update", zap.Int64("id", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("ExternalConfigFolder not found after update", zap.Int64("id", id))
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFound, nil)
		return
	}

	httputil.JSON(w, r, updated)
}

// DeleteExternalConfigFolder godoc
//
//	@Id				DeleteExternalConfigFolder
//
//	@Summary		Delete an external config folder
//	@Description	Delete an external config folder by ID. Child folders and configs are cascaded.
//	@Tags			ExternalConfigs
//	@Produce		json
//	@Param			id	path	int	true	"Folder ID"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	"Status OK"
//	@Failure		400	{object}	httputil.APIError	"Bad Request"
//	@Failure		403	{object}	httputil.APIError	"Forbidden"
//	@Failure		500	{object}	httputil.APIError	"Internal Server Error"
//	@Router			/engine/externalconfigs/folders/{id} [delete]
func DeleteExternalConfigFolder(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeConfig, permissions.All, permissions.ActionDelete)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		zap.L().Warn("Error parsing folder id", zap.String("id", chi.URLParam(r, "id")), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	if err := externalconfig.R().DeleteFolder(id); err != nil {
		zap.L().Error("Error while deleting ExternalConfigFolder", zap.Int64("id", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBDeleteFailed, err)
		return
	}

	httputil.OK(w, r)
}

// GetExternalConfigFolderHierarchy godoc
//
//	@Id				GetExternalConfigFolderHierarchy
//
//	@Summary		Get external config folder hierarchy
//	@Description	Get the full folder hierarchy tree, optionally from a given parent folder
//	@Tags			ExternalConfigs
//	@Produce		json
//	@Param			ParentId	query	int	false	"Parent folder ID (omit for root)"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{array}		externalconfig.FolderNode	"folder hierarchy"
//	@Failure		400	{object}	httputil.APIError			"Bad Request"
//	@Failure		403	{object}	httputil.APIError			"Forbidden"
//	@Failure		500	{object}	httputil.APIError			"Internal Server Error"
//	@Router			/engine/externalconfigs/folders/hierarchy [get]
func GetExternalConfigFolderHierarchy(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeConfig, permissions.All, permissions.ActionGet)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	var parentId *int64
	if parentIDStr := r.URL.Query().Get("ParentId"); parentIDStr != "" {
		pid, err := strconv.ParseInt(parentIDStr, 10, 64)
		if err != nil {
			zap.L().Warn("Error parsing ParentId", zap.String("ParentId", parentIDStr), zap.Error(err))
			httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
			return
		}
		parentId = &pid
	}

	hierarchy, err := externalconfig.R().GetFolderHierarchy(parentId)
	if err != nil {
		zap.L().Error("Error getting ExternalConfigFolder hierarchy", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}

	httputil.JSON(w, r, hierarchy)
}

// MoveExternalConfigToFolder godoc
//
//	@Id				MoveExternalConfigToFolder
//
//	@Summary		Move an external config to a folder
//	@Description	Move an external config to a different folder (use null folderId to move to root)
//	@Tags			ExternalConfigs
//	@Accept			json
//	@Produce		json
//	@Param			id		path	int						true	"ExternalConfig ID"
//	@Param			body	body	MoveConfigRequest		true	"Move request"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	"Status OK"
//	@Failure		400	{object}	httputil.APIError	"Bad Request"
//	@Failure		403	{object}	httputil.APIError	"Forbidden"
//	@Failure		500	{object}	httputil.APIError	"Internal Server Error"
//	@Router			/engine/externalconfigs/{id}/move [post]
func MoveExternalConfigToFolder(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeConfig, permissions.All, permissions.ActionUpdate)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		zap.L().Warn("Error parsing external config id", zap.String("id", chi.URLParam(r, "id")), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	var req MoveConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		zap.L().Warn("MoveExternalConfigToFolder json decoding", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDecodeJSONBody, err)
		return
	}

	if err := externalconfig.R().MoveConfig(id, req.FolderId); err != nil {
		zap.L().Error("Error while moving ExternalConfig", zap.Int64("id", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBUpdateFailed, err)
		return
	}

	httputil.OK(w, r)
}

// MoveExternalConfigFolder godoc
//
//	@Id				MoveExternalConfigFolder
//
//	@Summary		Move an external config folder
//	@Description	Move a folder to a new parent (use null newParentId to move to root). Returns error if move would create a circular reference.
//	@Tags			ExternalConfigs
//	@Accept			json
//	@Produce		json
//	@Param			id		path	int						true	"Folder ID"
//	@Param			body	body	MoveFolderRequest		true	"Move request"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	"Status OK"
//	@Failure		400	{object}	httputil.APIError	"Bad Request"
//	@Failure		403	{object}	httputil.APIError	"Forbidden"
//	@Failure		500	{object}	httputil.APIError	"Internal Server Error"
//	@Router			/engine/externalconfigs/folders/{id}/move [post]
func MoveExternalConfigFolder(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeConfig, permissions.All, permissions.ActionUpdate)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		zap.L().Warn("Error parsing folder id", zap.String("id", chi.URLParam(r, "id")), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	var req MoveFolderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		zap.L().Warn("MoveExternalConfigFolder json decoding", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDecodeJSONBody, err)
		return
	}

	if err := externalconfig.R().MoveFolder(id, req.ParentId); err != nil {
		zap.L().Error("Error while moving ExternalConfigFolder", zap.Int64("id", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBUpdateFailed, err)
		return
	}

	httputil.OK(w, r)
}

// MoveConfigRequest is the request body for moving an ExternalConfig to a folder
type MoveConfigRequest struct {
	FolderId *int64 `json:"folderId"`
}

// MoveFolderRequest is the request body for moving a folder
type MoveFolderRequest struct {
	ParentId *int64 `json:"ParentId"`
}
