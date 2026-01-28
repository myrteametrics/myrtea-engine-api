package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/functionalsituation"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/security/permissions"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/situation"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/utils/httputil"
	"go.uber.org/zap"
)

// InstanceIDPayload represents the payload for adding an instance
type InstanceIDPayload struct {
	InstanceID int64                  `json:"instanceId"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

// SituationIDPayload represents the payload for adding a situation
type SituationIDPayload struct {
	SituationID int64                  `json:"situationId"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
}

// GetFunctionalSituations godoc
//
//	@Id				GetFunctionalSituations
//	@Summary		Get all functional situations
//	@Description	Get all functional situations with optional filtering
//	@Tags			FunctionalSituations
//	@Produce		json
//	@Param			parentId	query	int	false	"Filter by parent ID (use -1 for roots)"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{array}		functionalsituation.FunctionalSituation
//	@Failure		400	{object}	httputil.APIError	"Bad Request"
//	@Failure		403	{object}	httputil.APIError	"Forbidden"
//	@Failure		500	{object}	httputil.APIError	"Internal Server Error"
//	@Router			/engine/functionalsituations [get]
func GetFunctionalSituations(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeFunctionalSituation, permissions.All, permissions.ActionList)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	// Check if filtering by parentId
	parentIDStr := r.URL.Query().Get("parentId")
	if parentIDStr != "" {
		parentID, err := strconv.ParseInt(parentIDStr, 10, 64)
		if err != nil {
			zap.L().Warn("Error parsing parentId", zap.String("parentId", parentIDStr), zap.Error(err))
			httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
			return
		}

		var fsList []functionalsituation.FunctionalSituation
		if parentID == -1 {
			// Get roots
			fsList, err = functionalsituation.R().GetRoots()
		} else {
			// Get children
			fsList, err = functionalsituation.R().GetChildren(parentID)
		}

		if err != nil {
			zap.L().Error("Error getting functional situations", zap.Error(err))
			httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
			return
		}

		httputil.JSON(w, r, fsList)
		return
	}

	// Get all
	fsList, err := functionalsituation.R().GetAll()
	if err != nil {
		zap.L().Error("Error getting functional situations", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}

	httputil.JSON(w, r, EnsureSlice(fsList))
}

// CreateFunctionalSituation godoc
//
//	@Id				CreateFunctionalSituation
//	@Summary		Create a new functional situation
//	@Description	Create a new functional situation
//	@Tags			FunctionalSituations
//	@Accept			json
//	@Produce		json
//	@Param			functionalSituation	body	functionalsituation.FunctionalSituationCreate	true	"Functional Situation definition (json)"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	functionalsituation.FunctionalSituation
//	@Failure		400	{object}	httputil.APIError	"Bad Request"
//	@Failure		403	{object}	httputil.APIError	"Forbidden"
//	@Failure		500	{object}	httputil.APIError	"Internal Server Error"
//	@Router			/engine/functionalsituations [post]
func CreateFunctionalSituation(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeFunctionalSituation, permissions.All, permissions.ActionCreate)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	var fsCreate functionalsituation.FunctionalSituationCreate
	err := json.NewDecoder(r.Body).Decode(&fsCreate)
	if err != nil {
		zap.L().Warn("FunctionalSituation json decoding", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDecodeJSONBody, err)
		return
	}

	if ok, err := fsCreate.IsValid(); !ok {
		zap.L().Warn("FunctionalSituation is not valid", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIResourceInvalid, err)
		return
	}

	// Convert create payload to FunctionalSituation
	fs := functionalsituation.FunctionalSituation{
		Name:        fsCreate.Name,
		Description: fsCreate.Description,
		ParentID:    fsCreate.ParentID,
		Color:       fsCreate.Color,
		Icon:        fsCreate.Icon,
		Parameters:  fsCreate.Parameters,
	}

	id, err := functionalsituation.R().Create(fs, userCtx.User.Login)
	if err != nil {
		zap.L().Error("Error creating functional situation", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBInsertFailed, err)
		return
	}

	created, found, err := functionalsituation.R().Get(id)
	if err != nil {
		zap.L().Error("Error getting created functional situation", zap.Int64("id", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Error("Functional situation not found after creation", zap.Int64("id", id))
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFoundAfterInsert, errors.New("resource not found after insert"))
		return
	}

	httputil.JSON(w, r, created)
}

// GetFunctionalSituationTree godoc
//
//	@Id				GetFunctionalSituationTree
//	@Summary		Get functional situation hierarchy tree
//	@Description	Get the complete hierarchical tree of functional situations
//	@Tags			FunctionalSituations
//	@Produce		json
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{array}		functionalsituation.FunctionalSituation
//	@Failure		403	{object}	httputil.APIError	"Forbidden"
//	@Failure		500	{object}	httputil.APIError	"Internal Server Error"
//	@Router			/engine/functionalsituations/tree [get]
func GetFunctionalSituationTree(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeFunctionalSituation, permissions.All, permissions.ActionList)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	tree, err := functionalsituation.R().GetTree()
	if err != nil {
		zap.L().Error("Error getting functional situation tree", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}

	httputil.JSON(w, r, EnsureSlice(tree))
}

// GetFunctionalSituationEnrichedTree godoc
//
//	@Id				GetFunctionalSituationEnrichedTree
//	@Summary		Get enriched functional situation tree
//	@Description	Get the complete hierarchical tree with all template instances and situations included
//	@Tags			FunctionalSituations
//	@Produce		json
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{array}		functionalsituation.FunctionalSituationTreeNode
//	@Failure		403	{object}	httputil.APIError	"Forbidden"
//	@Failure		500	{object}	httputil.APIError	"Internal Server Error"
//	@Router			/engine/functionalsituations/tree/enriched [get]
func GetFunctionalSituationEnrichedTree(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeFunctionalSituation, permissions.All, permissions.ActionList)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	tree, err := functionalsituation.R().GetEnrichedTree()
	if err != nil {
		zap.L().Error("Error getting functional situation enriched tree", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}

	httputil.JSON(w, r, EnsureSlice(tree))
}

// GetFunctionalSituationOverview godoc
//
//	@Id				GetFunctionalSituationOverview
//	@Summary		Get functional situation overview
//	@Description	Get an overview of all functional situations with aggregated counts and status
//	@Tags			FunctionalSituations
//	@Produce		json
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{array}		functionalsituation.FunctionalSituationOverview
//	@Failure		403	{object}	httputil.APIError	"Forbidden"
//	@Failure		500	{object}	httputil.APIError	"Internal Server Error"
//	@Router			/engine/functionalsituations/overview [get]
func GetFunctionalSituationOverview(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeFunctionalSituation, permissions.All, permissions.ActionList)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	overview, err := functionalsituation.R().GetOverview()
	if err != nil {
		zap.L().Error("Error getting functional situation overview", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}

	httputil.JSON(w, r, overview)
}

// GetFunctionalSituation godoc
//
//	@Id				GetFunctionalSituation
//	@Summary		Get a functional situation
//	@Description	Get a functional situation by its ID
//	@Tags			FunctionalSituations
//	@Produce		json
//	@Param			id	path	int	true	"Functional Situation ID"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	functionalsituation.FunctionalSituation
//	@Failure		400	{object}	httputil.APIError	"Bad Request"
//	@Failure		403	{object}	httputil.APIError	"Forbidden"
//	@Failure		404	{object}	httputil.APIError	"Not Found"
//	@Failure		500	{object}	httputil.APIError	"Internal Server Error"
//	@Router			/engine/functionalsituations/{id} [get]
func GetFunctionalSituation(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	fsID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("Error parsing functional situation id", zap.String("id", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeFunctionalSituation, strconv.FormatInt(fsID, 10), permissions.ActionGet)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	fs, found, err := functionalsituation.R().Get(fsID)
	if err != nil {
		zap.L().Error("Error getting functional situation", zap.Int64("id", fsID), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("Functional situation not found", zap.Int64("id", fsID))
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFound, errors.New("functional situation not found"))
		return
	}

	httputil.JSON(w, r, fs)
}

// UpdateFunctionalSituation godoc
//
//	@Id				UpdateFunctionalSituation
//	@Summary		Update a functional situation
//	@Description	Update a functional situation by its ID
//	@Tags			FunctionalSituations
//	@Accept			json
//	@Produce		json
//	@Param			id					path	int												true	"Functional Situation ID"
//	@Param			functionalSituation	body	functionalsituation.FunctionalSituationUpdate	true	"Functional Situation update payload (json)"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	functionalsituation.FunctionalSituation
//	@Failure		400	{object}	httputil.APIError	"Bad Request"
//	@Failure		403	{object}	httputil.APIError	"Forbidden"
//	@Failure		404	{object}	httputil.APIError	"Not Found"
//	@Failure		500	{object}	httputil.APIError	"Internal Server Error"
//	@Router			/engine/functionalsituations/{id} [put]
func UpdateFunctionalSituation(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	fsID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("Error parsing functional situation id", zap.String("id", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeFunctionalSituation, strconv.FormatInt(fsID, 10), permissions.ActionUpdate)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	// Check if exists
	_, found, err := functionalsituation.R().Get(fsID)
	if err != nil {
		zap.L().Error("Error getting functional situation", zap.Int64("id", fsID), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("Functional situation not found", zap.Int64("id", fsID))
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFound, errors.New("functional situation not found"))
		return
	}

	var fsUpdate functionalsituation.FunctionalSituationUpdate
	err = json.NewDecoder(r.Body).Decode(&fsUpdate)
	if err != nil {
		zap.L().Warn("FunctionalSituation json decoding", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDecodeJSONBody, err)
		return
	}

	err = functionalsituation.R().Update(fsID, fsUpdate, userCtx.User.Login)
	if err != nil {
		zap.L().Error("Error updating functional situation", zap.Int64("id", fsID), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBUpdateFailed, err)
		return
	}

	updated, found, err := functionalsituation.R().Get(fsID)
	if err != nil {
		zap.L().Error("Error getting updated functional situation", zap.Int64("id", fsID), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Error("Functional situation not found after update", zap.Int64("id", fsID))
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFoundAfterInsert, errors.New("resource not found after update"))
		return
	}

	httputil.JSON(w, r, updated)
}

// DeleteFunctionalSituation godoc
//
//	@Id				DeleteFunctionalSituation
//	@Summary		Delete a functional situation
//	@Description	Delete a functional situation by its ID
//	@Tags			FunctionalSituations
//	@Produce		json
//	@Param			id	path	int	true	"Functional Situation ID"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	object	"Deleted"
//	@Failure		400	{object}	httputil.APIError	"Bad Request"
//	@Failure		403	{object}	httputil.APIError	"Forbidden"
//	@Failure		404	{object}	httputil.APIError	"Not Found"
//	@Failure		500	{object}	httputil.APIError	"Internal Server Error"
//	@Router			/engine/functionalsituations/{id} [delete]
func DeleteFunctionalSituation(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	fsID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("Error parsing functional situation id", zap.String("id", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeFunctionalSituation, strconv.FormatInt(fsID, 10), permissions.ActionDelete)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	// Check if exists
	_, found, err := functionalsituation.R().Get(fsID)
	if err != nil {
		zap.L().Error("Error getting functional situation", zap.Int64("id", fsID), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("Functional situation not found", zap.Int64("id", fsID))
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFound, errors.New("functional situation not found"))
		return
	}

	err = functionalsituation.R().Delete(fsID)
	if err != nil {
		zap.L().Error("Error deleting functional situation", zap.Int64("id", fsID), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBDeleteFailed, err)
		return
	}

	httputil.JSON(w, r, map[string]interface{}{"deleted": true, "id": fsID})
}

// GetFSInstances godoc
//
//	@Id				GetFSInstances
//	@Summary		Get template instances for a functional situation
//	@Description	Get all template instances associated with a functional situation
//	@Tags			FunctionalSituations
//	@Produce		json
//	@Param			id	path	int	true	"Functional Situation ID"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{array}		situation.TemplateInstance
//	@Failure		400	{object}	httputil.APIError	"Bad Request"
//	@Failure		403	{object}	httputil.APIError	"Forbidden"
//	@Failure		404	{object}	httputil.APIError	"Not Found"
//	@Failure		500	{object}	httputil.APIError	"Internal Server Error"
//	@Router			/engine/functionalsituations/{id}/instances [get]
func GetFSInstances(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	fsID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("Error parsing functional situation id", zap.String("id", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeFunctionalSituationInstance, strconv.FormatInt(fsID, 10), permissions.ActionList)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	// Check if functional situation exists
	_, found, err := functionalsituation.R().Get(fsID)
	if err != nil {
		zap.L().Error("Error getting functional situation", zap.Int64("id", fsID), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("Functional situation not found", zap.Int64("id", fsID))
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFound, errors.New("functional situation not found"))
		return
	}

	instanceIDs, err := functionalsituation.R().GetTemplateInstances(fsID)
	if err != nil {
		zap.L().Error("Error getting template instances", zap.Int64("fsID", fsID), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}

	// Get full template instance objects
	instancesMap, err := situation.R().GetAllTemplateInstancesByIDs(instanceIDs, gvalParsingEnabled(r.URL.Query()))
	if err != nil {
		zap.L().Error("Error getting template instance details", zap.Int64("fsID", fsID), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}

	// Get all instance references with parameters in one call
	instanceRefs, err := functionalsituation.R().GetTemplateInstancesWithParameters(fsID)
	if err != nil {
		zap.L().Error("Error getting instance parameters", zap.Int64("fsID", fsID), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}

	// Convert map to slice maintaining order and including parameters
	instances := make([]functionalsituation.TemplateInstanceWithParameters, 0, len(instancesMap))
	for _, instanceID := range instanceIDs {
		if instance, ok := instancesMap[instanceID]; ok {
			instanceWithParams := functionalsituation.TemplateInstanceWithParameters{
				Instance:   instance,
				Parameters: instanceRefs[instanceID],
			}
			instances = append(instances, instanceWithParams)
		}
	}

	httputil.JSON(w, r, EnsureSlice(instances))
}

// AddFSInstance godoc
//
//	@Id				AddFSInstance
//	@Summary		Add a template instance to a functional situation
//	@Description	Associate a template instance with a functional situation
//	@Tags			FunctionalSituations
//	@Accept			json
//	@Produce		json
//	@Param			id		path	int					true	"Functional Situation ID"
//	@Param			payload	body	InstanceIDPayload	true	"Instance ID payload"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	object	"Added"
//	@Failure		400	{object}	httputil.APIError	"Bad Request"
//	@Failure		403	{object}	httputil.APIError	"Forbidden"
//	@Failure		404	{object}	httputil.APIError	"Not Found"
//	@Failure		500	{object}	httputil.APIError	"Internal Server Error"
//	@Router			/engine/functionalsituations/{id}/instances [post]
func AddFSInstance(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	fsID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("Error parsing functional situation id", zap.String("id", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeFunctionalSituationInstance, strconv.FormatInt(fsID, 10), permissions.ActionCreate)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	// Check if functional situation exists
	_, found, err := functionalsituation.R().Get(fsID)
	if err != nil {
		zap.L().Error("Error getting functional situation", zap.Int64("id", fsID), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("Functional situation not found", zap.Int64("id", fsID))
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFound, errors.New("functional situation not found"))
		return
	}

	var payload InstanceIDPayload
	err = json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		zap.L().Warn("Instance payload json decoding", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDecodeJSONBody, err)
		return
	}

	if payload.InstanceID <= 0 {
		zap.L().Warn("Invalid instance ID", zap.Int64("instanceId", payload.InstanceID))
		httputil.Error(w, r, httputil.ErrAPIResourceInvalid, errors.New("instanceId must be positive"))
		return
	}

	err = functionalsituation.R().AddTemplateInstance(fsID, payload.InstanceID, payload.Parameters, userCtx.User.Login)
	if err != nil {
		zap.L().Error("Error adding template instance", zap.Int64("fsID", fsID), zap.Int64("instanceID", payload.InstanceID), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBInsertFailed, err)
		return
	}

	httputil.JSON(w, r, map[string]interface{}{"added": true, "fsId": fsID, "instanceId": payload.InstanceID})
}

// RemoveFSInstance godoc
//
//	@Id				RemoveFSInstance
//	@Summary		Remove a template instance from a functional situation
//	@Description	Remove the association between a template instance and a functional situation
//	@Tags			FunctionalSituations
//	@Produce		json
//	@Param			id			path	int	true	"Functional Situation ID"
//	@Param			instanceId	path	int	true	"Template Instance ID"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	object	"Removed"
//	@Failure		400	{object}	httputil.APIError	"Bad Request"
//	@Failure		403	{object}	httputil.APIError	"Forbidden"
//	@Failure		404	{object}	httputil.APIError	"Not Found"
//	@Failure		500	{object}	httputil.APIError	"Internal Server Error"
//	@Router			/engine/functionalsituations/{id}/instances/{instanceId} [delete]
func RemoveFSInstance(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	fsID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("Error parsing functional situation id", zap.String("id", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	instanceIDStr := chi.URLParam(r, "instanceId")
	instanceID, err := strconv.ParseInt(instanceIDStr, 10, 64)
	if err != nil {
		zap.L().Warn("Error parsing instance id", zap.String("instanceId", instanceIDStr), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeFunctionalSituationInstance, strconv.FormatInt(fsID, 10), permissions.ActionDelete)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	// Check if functional situation exists
	_, found, err := functionalsituation.R().Get(fsID)
	if err != nil {
		zap.L().Error("Error getting functional situation", zap.Int64("id", fsID), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("Functional situation not found", zap.Int64("id", fsID))
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFound, errors.New("functional situation not found"))
		return
	}

	err = functionalsituation.R().RemoveTemplateInstance(fsID, instanceID)
	if err != nil {
		zap.L().Error("Error removing template instance", zap.Int64("fsID", fsID), zap.Int64("instanceID", instanceID), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBDeleteFailed, err)
		return
	}

	httputil.JSON(w, r, map[string]interface{}{"removed": true, "fsId": fsID, "instanceId": instanceID})
}

// GetFSSituations godoc
//
//	@Id				GetFSSituations
//	@Summary		Get situations for a functional situation
//	@Description	Get all situations associated with a functional situation
//	@Tags			FunctionalSituations
//	@Produce		json
//	@Param			id	path	int	true	"Functional Situation ID"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{array}		functionalsituation.SituationWithParameters
//	@Failure		400	{object}	httputil.APIError	"Bad Request"
//	@Failure		403	{object}	httputil.APIError	"Forbidden"
//	@Failure		404	{object}	httputil.APIError	"Not Found"
//	@Failure		500	{object}	httputil.APIError	"Internal Server Error"
//	@Router			/engine/functionalsituations/{id}/situations [get]
func GetFSSituations(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	fsID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("Error parsing functional situation id", zap.String("id", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeFunctionalSituationContent, strconv.FormatInt(fsID, 10), permissions.ActionList)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	// Check if functional situation exists
	_, found, err := functionalsituation.R().Get(fsID)
	if err != nil {
		zap.L().Error("Error getting functional situation", zap.Int64("id", fsID), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("Functional situation not found", zap.Int64("id", fsID))
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFound, errors.New("functional situation not found"))
		return
	}

	situationIDs, err := functionalsituation.R().GetSituations(fsID)
	if err != nil {
		zap.L().Error("Error getting situations", zap.Int64("fsID", fsID), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}

	// Get full situation objects
	situationsMap, err := situation.R().GetAllByIDs(situationIDs, gvalParsingEnabled(r.URL.Query()))
	if err != nil {
		zap.L().Error("Error getting situation details", zap.Int64("fsID", fsID), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}

	// Get all situation references with parameters in one call
	situationRefs, err := functionalsituation.R().GetSituationsWithParameters(fsID)
	if err != nil {
		zap.L().Error("Error getting situation parameters", zap.Int64("fsID", fsID), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}

	// Convert map to slice maintaining order and including parameters
	situations := make([]functionalsituation.SituationWithParameters, 0, len(situationsMap))
	for _, situationID := range situationIDs {
		if sit, ok := situationsMap[situationID]; ok {
			situationWithParams := functionalsituation.SituationWithParameters{
				Situation:  sit,
				Parameters: situationRefs[situationID],
			}
			situations = append(situations, situationWithParams)
		}
	}

	httputil.JSON(w, r, EnsureSlice(situations))
}

// AddFSSituation godoc
//
//	@Id				AddFSSituation
//	@Summary		Add a situation to a functional situation
//	@Description	Associate a situation with a functional situation
//	@Tags			FunctionalSituations
//	@Accept			json
//	@Produce		json
//	@Param			id		path	int					true	"Functional Situation ID"
//	@Param			payload	body	SituationIDPayload	true	"Situation ID payload"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	object	"Added"
//	@Failure		400	{object}	httputil.APIError	"Bad Request"
//	@Failure		403	{object}	httputil.APIError	"Forbidden"
//	@Failure		404	{object}	httputil.APIError	"Not Found"
//	@Failure		500	{object}	httputil.APIError	"Internal Server Error"
//	@Router			/engine/functionalsituations/{id}/situations [post]
func AddFSSituation(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	fsID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("Error parsing functional situation id", zap.String("id", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeFunctionalSituationContent, strconv.FormatInt(fsID, 10), permissions.ActionCreate)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	// Check if functional situation exists
	_, found, err := functionalsituation.R().Get(fsID)
	if err != nil {
		zap.L().Error("Error getting functional situation", zap.Int64("id", fsID), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("Functional situation not found", zap.Int64("id", fsID))
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFound, errors.New("functional situation not found"))
		return
	}

	var payload SituationIDPayload
	err = json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		zap.L().Warn("Situation payload json decoding", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDecodeJSONBody, err)
		return
	}

	if payload.SituationID <= 0 {
		zap.L().Warn("Invalid situation ID", zap.Int64("situationId", payload.SituationID))
		httputil.Error(w, r, httputil.ErrAPIResourceInvalid, errors.New("situationId must be positive"))
		return
	}

	err = functionalsituation.R().AddSituation(fsID, payload.SituationID, payload.Parameters, userCtx.User.Login)
	if err != nil {
		zap.L().Error("Error adding situation", zap.Int64("fsID", fsID), zap.Int64("situationID", payload.SituationID), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBInsertFailed, err)
		return
	}

	httputil.JSON(w, r, map[string]interface{}{"added": true, "fsId": fsID, "situationId": payload.SituationID})
}

// RemoveFSSituation godoc
//
//	@Id				RemoveFSSituation
//	@Summary		Remove a situation from a functional situation
//	@Description	Remove the association between a situation and a functional situation
//	@Tags			FunctionalSituations
//	@Produce		json
//	@Param			id			path	int	true	"Functional Situation ID"
//	@Param			situationId	path	int	true	"Situation ID"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	object	"Removed"
//	@Failure		400	{object}	httputil.APIError	"Bad Request"
//	@Failure		403	{object}	httputil.APIError	"Forbidden"
//	@Failure		404	{object}	httputil.APIError	"Not Found"
//	@Failure		500	{object}	httputil.APIError	"Internal Server Error"
//	@Router			/engine/functionalsituations/{id}/situations/{situationId} [delete]
func RemoveFSSituation(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	fsID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("Error parsing functional situation id", zap.String("id", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	situationIDStr := chi.URLParam(r, "situationId")
	situationID, err := strconv.ParseInt(situationIDStr, 10, 64)
	if err != nil {
		zap.L().Warn("Error parsing situation id", zap.String("situationId", situationIDStr), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeFunctionalSituationContent, strconv.FormatInt(fsID, 10), permissions.ActionDelete)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	// Check if functional situation exists
	_, found, err := functionalsituation.R().Get(fsID)
	if err != nil {
		zap.L().Error("Error getting functional situation", zap.Int64("id", fsID), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("Functional situation not found", zap.Int64("id", fsID))
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFound, errors.New("functional situation not found"))
		return
	}

	err = functionalsituation.R().RemoveSituation(fsID, situationID)
	if err != nil {
		zap.L().Error("Error removing situation", zap.Int64("fsID", fsID), zap.Int64("situationID", situationID), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBDeleteFailed, err)
		return
	}

	httputil.JSON(w, r, map[string]interface{}{"removed": true, "fsId": fsID, "situationId": situationID})
}

// GetInstanceReferenceParameters godoc
//
//	@Id				GetInstanceReferenceParameters
//	@Summary		Get parameters for a template instance reference
//	@Description	Get the parameters associated with a template instance reference
//	@Tags			FunctionalSituations
//	@Produce		json
//	@Param			instanceId	path	int	true	"Template Instance ID"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	functionalsituation.InstanceReference
//	@Failure		400	{object}	httputil.APIError	"Bad Request"
//	@Failure		403	{object}	httputil.APIError	"Forbidden"
//	@Failure		404	{object}	httputil.APIError	"Not Found"
//	@Failure		500	{object}	httputil.APIError	"Internal Server Error"
//	@Router			/engine/functionalsituations/instances/{instanceId}/parameters [get]
func GetInstanceReferenceParameters(w http.ResponseWriter, r *http.Request) {
	instanceIDStr := chi.URLParam(r, "instanceId")
	instanceID, err := strconv.ParseInt(instanceIDStr, 10, 64)
	if err != nil {
		zap.L().Warn("Error parsing instance id", zap.String("instanceId", instanceIDStr), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeFunctionalSituationInstance, strconv.FormatInt(instanceID, 10), permissions.ActionGet)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	ref, found, err := functionalsituation.R().GetInstanceReference(instanceID)
	if err != nil {
		zap.L().Error("Error getting instance reference", zap.Int64("instanceId", instanceID), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("Instance reference not found", zap.Int64("instanceId", instanceID))
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFound, errors.New("instance reference not found"))
		return
	}

	httputil.JSON(w, r, ref)
}

// UpdateInstanceReferenceParameters godoc
//
//	@Id				UpdateInstanceReferenceParameters
//	@Summary		Update parameters for a template instance reference
//	@Description	Update the parameters associated with a template instance reference
//	@Tags			FunctionalSituations
//	@Accept			json
//	@Produce		json
//	@Param			instanceId	path	int						true	"Template Instance ID"
//	@Param			parameters	body	map[string]interface{}	true	"Parameters (json)"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	functionalsituation.InstanceReference
//	@Failure		400	{object}	httputil.APIError	"Bad Request"
//	@Failure		403	{object}	httputil.APIError	"Forbidden"
//	@Failure		404	{object}	httputil.APIError	"Not Found"
//	@Failure		500	{object}	httputil.APIError	"Internal Server Error"
//	@Router			/engine/functionalsituations/instances/{instanceId}/parameters [put]
func UpdateInstanceReferenceParameters(w http.ResponseWriter, r *http.Request) {
	instanceIDStr := chi.URLParam(r, "instanceId")
	instanceID, err := strconv.ParseInt(instanceIDStr, 10, 64)
	if err != nil {
		zap.L().Warn("Error parsing instance id", zap.String("instanceId", instanceIDStr), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeFunctionalSituationInstance, strconv.FormatInt(instanceID, 10), permissions.ActionUpdate)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	// Check if instance reference exists
	_, found, err := functionalsituation.R().GetInstanceReference(instanceID)
	if err != nil {
		zap.L().Error("Error getting instance reference", zap.Int64("instanceId", instanceID), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("Instance reference not found", zap.Int64("instanceId", instanceID))
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFound, errors.New("instance reference not found"))
		return
	}

	var parameters map[string]interface{}
	err = json.NewDecoder(r.Body).Decode(&parameters)
	if err != nil {
		zap.L().Warn("Parameters json decoding", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDecodeJSONBody, err)
		return
	}

	err = functionalsituation.R().UpdateInstanceReferenceParameters(instanceID, parameters)
	if err != nil {
		zap.L().Error("Error updating instance reference parameters", zap.Int64("instanceId", instanceID), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBUpdateFailed, err)
		return
	}

	updated, found, err := functionalsituation.R().GetInstanceReference(instanceID)
	if err != nil {
		zap.L().Error("Error getting updated instance reference", zap.Int64("instanceId", instanceID), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Error("Instance reference not found after update", zap.Int64("instanceId", instanceID))
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFoundAfterInsert, errors.New("resource not found after update"))
		return
	}

	httputil.JSON(w, r, updated)
}

// GetSituationReferenceParameters godoc
//
//	@Id				GetSituationReferenceParameters
//	@Summary		Get parameters for a situation reference
//	@Description	Get the parameters associated with a situation reference
//	@Tags			FunctionalSituations
//	@Produce		json
//	@Param			situationId	path	int	true	"Situation ID"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	functionalsituation.SituationReference
//	@Failure		400	{object}	httputil.APIError	"Bad Request"
//	@Failure		403	{object}	httputil.APIError	"Forbidden"
//	@Failure		404	{object}	httputil.APIError	"Not Found"
//	@Failure		500	{object}	httputil.APIError	"Internal Server Error"
//	@Router			/engine/functionalsituations/situations/{situationId}/parameters [get]
func GetSituationReferenceParameters(w http.ResponseWriter, r *http.Request) {
	situationIDStr := chi.URLParam(r, "situationId")
	situationID, err := strconv.ParseInt(situationIDStr, 10, 64)
	if err != nil {
		zap.L().Warn("Error parsing situation id", zap.String("situationId", situationIDStr), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeFunctionalSituationContent, strconv.FormatInt(situationID, 10), permissions.ActionGet)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	ref, found, err := functionalsituation.R().GetSituationReference(situationID)
	if err != nil {
		zap.L().Error("Error getting situation reference", zap.Int64("situationId", situationID), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("Situation reference not found", zap.Int64("situationId", situationID))
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFound, errors.New("situation reference not found"))
		return
	}

	httputil.JSON(w, r, ref)
}

// UpdateSituationReferenceParameters godoc
//
//	@Id				UpdateSituationReferenceParameters
//	@Summary		Update parameters for a situation reference
//	@Description	Update the parameters associated with a situation reference
//	@Tags			FunctionalSituations
//	@Accept			json
//	@Produce		json
//	@Param			situationId	path	int						true	"Situation ID"
//	@Param			parameters	body	map[string]interface{}	true	"Parameters (json)"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	functionalsituation.SituationReference
//	@Failure		400	{object}	httputil.APIError	"Bad Request"
//	@Failure		403	{object}	httputil.APIError	"Forbidden"
//	@Failure		404	{object}	httputil.APIError	"Not Found"
//	@Failure		500	{object}	httputil.APIError	"Internal Server Error"
//	@Router			/engine/functionalsituations/situations/{situationId}/parameters [put]
func UpdateSituationReferenceParameters(w http.ResponseWriter, r *http.Request) {
	situationIDStr := chi.URLParam(r, "situationId")
	situationID, err := strconv.ParseInt(situationIDStr, 10, 64)
	if err != nil {
		zap.L().Warn("Error parsing situation id", zap.String("situationId", situationIDStr), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeFunctionalSituationContent, strconv.FormatInt(situationID, 10), permissions.ActionUpdate)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	// Check if situation reference exists
	_, found, err := functionalsituation.R().GetSituationReference(situationID)
	if err != nil {
		zap.L().Error("Error getting situation reference", zap.Int64("situationId", situationID), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("Situation reference not found", zap.Int64("situationId", situationID))
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFound, errors.New("situation reference not found"))
		return
	}

	var parameters map[string]interface{}
	err = json.NewDecoder(r.Body).Decode(&parameters)
	if err != nil {
		zap.L().Warn("Parameters json decoding", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDecodeJSONBody, err)
		return
	}

	// Validate parameters
	if ok, err := functionalsituation.ValidateParameters(parameters); !ok {
		zap.L().Warn("Invalid parameters", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIResourceInvalid, err)
		return
	}

	err = functionalsituation.R().UpdateSituationReferenceParameters(situationID, parameters)
	if err != nil {
		zap.L().Error("Error updating situation reference parameters", zap.Int64("situationId", situationID), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBUpdateFailed, err)
		return
	}

	updated, found, err := functionalsituation.R().GetSituationReference(situationID)
	if err != nil {
		zap.L().Error("Error getting updated situation reference", zap.Int64("situationId", situationID), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Error("Situation reference not found after update", zap.Int64("situationId", situationID))
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFoundAfterInsert, errors.New("resource not found after update"))
		return
	}

	httputil.JSON(w, r, updated)
}
