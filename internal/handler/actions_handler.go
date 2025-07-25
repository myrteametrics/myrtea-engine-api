package handler

import (
	"encoding/json"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/utils/httputil"
	"net/http"
	"sort"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/explainer/action"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/model"
	"go.uber.org/zap"
)

// GetActions godoc
//
//	@Id				GetActions
//
//	@Summary		Get all action definitions
//	@Description	Get all action definitions
//	@Tags			Actions
//	@Produce		json
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{array}	model.Action	"list of all actions"
//	@Failure		500	"internal server error"
//	@Router			/engine/actions [get]
func GetActions(w http.ResponseWriter, r *http.Request) {
	actions, err := action.R().GetAll()
	if err != nil {
		zap.L().Error("Error getting actions", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}

	actionsSlice := make([]model.Action, 0)
	for _, action := range actions {
		actionsSlice = append(actionsSlice, action)
	}

	sort.SliceStable(actionsSlice, func(i, j int) bool {
		return actionsSlice[i].ID < actionsSlice[j].ID
	})

	httputil.JSON(w, r, actionsSlice)
}

// GetAction godoc
//
//	@Id				GetAction
//
//	@Summary		Get a action definition
//	@Description	Get a action definition
//	@Tags			Actions
//	@Produce		json
//	@Param			id	path	string	true	"Action ID"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	model.Action	"action"
//	@Failure		400	"Status Bad Request"
//	@Router			/engine/actions/{id} [get]
func GetAction(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idAction, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("Error on parsing action id", zap.String("actionID", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	a, found, err := action.R().Get(idAction)
	if err != nil {
		zap.L().Error("Cannot get action", zap.Int64("actionid", idAction), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("Action does not exists", zap.Int64("actionid", idAction))
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFound, err)
		return
	}

	httputil.JSON(w, r, a)
}

// ValidateAction godoc
//
//	@Id				ValidateAction
//
//	@Summary		Validate a new action definition
//	@Description	Validate a new action definition
//	@Tags			Actions
//	@Accept			json
//	@Produce		json
//	@Param			action	body	model.Action	true	"Action definition (json)"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	model.Action	"action"
//	@Failure		400	"Status Bad Request"
//	@Failure		500	"Status"	internal	server	error"
//	@Router			/engine/actions/validate [post]
func ValidateAction(w http.ResponseWriter, r *http.Request) {

	var newAction model.Action
	err := json.NewDecoder(r.Body).Decode(&newAction)
	if err != nil {
		zap.L().Warn("Action json decoding", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDecodeJSONBody, err)
		return
	}

	if ok, err := newAction.IsValid(); !ok {
		zap.L().Warn("Action is not valid", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIResourceInvalid, err)
		return
	}

	httputil.JSON(w, r, newAction)
}

// PostAction godoc
//
//	@Id				PostAction
//
//	@Summary		Create a new action definition
//	@Description	Create a new action definition
//	@Tags			Actions
//	@Accept			json
//	@Produce		json
//	@Param			action	body	model.Action	true	"Action definition (json)"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	model.Action	"action"
//	@Failure		400	"Status Bad Request"
//	@Failure		500	"Status"	internal	server	error"
//	@Router			/engine/actions [post]
func PostAction(w http.ResponseWriter, r *http.Request) {

	var newAction model.Action
	err := json.NewDecoder(r.Body).Decode(&newAction)
	if err != nil {
		zap.L().Warn("Action json decoding", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDecodeJSONBody, err)
		return
	}

	if ok, err := newAction.IsValid(); !ok {
		zap.L().Warn("Action is not valid", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIResourceInvalid, err)
		return
	}

	newActionID, err := action.R().Create(nil, newAction)
	if err != nil {
		zap.L().Error("Error while creating the Action", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBInsertFailed, err)
		return
	}

	newActionGet, found, err := action.R().Get(newActionID)
	if err != nil {
		zap.L().Error("Cannot get action", zap.Int64("actionid", newActionID), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("Action does not exists after creation", zap.Int64("actionid", newActionID))
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFoundAfterInsert, err)
		return
	}

	httputil.JSON(w, r, newActionGet)
}

// PutAction godoc
//
//	@Id				PutAction
//
//	@Summary		Create or remplace a action definition
//	@Description	Create or remplace a action definition
//	@Tags			Actions
//	@Accept			json
//	@Produce		json
//	@Param			id		path	string			true	"Action ID"
//	@Param			action	body	model.Action	true	"Action definition (json)"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	model.Action	"action"
//	@Failure		400	"Status Bad Request"
//	@Failure		500	"Status"	internal	server	error"
//	@Router			/engine/actions/{id} [put]
func PutAction(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idAction, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("Error on parsing action id", zap.String("actionID", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	var newAction model.Action
	err = json.NewDecoder(r.Body).Decode(&newAction)
	if err != nil {
		zap.L().Warn("Action json decoding", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDecodeJSONBody, err)
		return
	}
	newAction.ID = idAction

	if ok, err := newAction.IsValid(); !ok {
		zap.L().Warn("Action is not valid", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIResourceInvalid, err)
		return
	}

	err = action.R().Update(nil, idAction, newAction)
	if err != nil {
		zap.L().Error("Error while updating the Action", zap.Int64("idAction", idAction), zap.Any("action", newAction), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBUpdateFailed, err)
		return
	}

	newActionGet, found, err := action.R().Get(idAction)
	if err != nil {
		zap.L().Error("Cannot get action", zap.Int64("actionid", idAction), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("Action does not exists after update", zap.Int64("actionid", idAction))
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFound, err)
		return
	}

	httputil.JSON(w, r, newActionGet)
}

// DeleteAction godoc
//
//	@Id				DeleteAction
//
//	@Summary		Delete a action definition
//	@Description	Delete a action definition
//	@Tags			Actions
//	@Produce		json
//	@Param			id	path	string	true	"Action ID"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	"Status OK"
//	@Failure		400	"Status Bad Request"
//	@Router			/engine/actions/{id} [delete]
func DeleteAction(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idAction, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("Error on parsing action id", zap.String("actionID", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	err = action.R().Delete(nil, idAction)
	if err != nil {
		zap.L().Error("Error while deleting the Action", zap.String("Action ID", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBDeleteFailed, err)
		return
	}

	httputil.OK(w, r)
}
