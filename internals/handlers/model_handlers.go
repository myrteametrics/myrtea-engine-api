package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/handlers/render"
	model "github.com/myrteametrics/myrtea-engine-api/v5/internals/modeler"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/security/permissions"
	"github.com/myrteametrics/myrtea-sdk/v4/modeler"
	"go.uber.org/zap"
)

// GetModels godoc
// @Summary Get all model definitions
// @Description Get all model definitions
// @Tags Models
// @Produce json
// @Security Bearer
// @Success 200 "Status OK"
// @Failure 500 "internal server error"
// @Router /engine/models [get]
func GetModels(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeModel, permissions.All, permissions.ActionList)) {
		render.Error(w, r, render.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	var models map[int64]modeler.Model
	var err error
	if userCtx.HasPermission(permissions.New(permissions.TypeModel, permissions.All, permissions.ActionGet)) {
		models, err = model.R().GetAll()
	} else {
		resourceIDs := userCtx.GetMatchingResourceIDsInt64(permissions.New(permissions.TypeModel, permissions.All, permissions.ActionGet))
		models, err = model.R().GetAllByIDs(resourceIDs)
	}
	if err != nil {
		zap.L().Error("Error getting models", zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}

	modelsSlice := make([]modeler.Model, 0)
	for _, action := range models {
		modelsSlice = append(modelsSlice, action)
	}

	sort.SliceStable(modelsSlice, func(i, j int) bool {
		return modelsSlice[i].ID < modelsSlice[j].ID
	})

	render.JSON(w, r, modelsSlice)
}

// GetModel godoc
// @Summary Get a model definition
// @Description Get a model definition
// @Tags Models
// @Produce json
// @Param id path string true "Model ID"
// @Param byName query string false "Find model by it's name"
// @Security Bearer
// @Success 200 "Status OK"
// @Failure 400 "Status Bad Request"
// @Router /engine/models/{id} [get]
func GetModel(w http.ResponseWriter, r *http.Request) {
	byName := false
	_byName := r.URL.Query().Get("byName")
	if _byName == "true" {
		byName = true
	}

	id := chi.URLParam(r, "id")
	var m modeler.Model
	var found bool
	var err error
	if byName {
		m, found, err = model.R().GetByName(id)
		if err != nil {
			zap.L().Error("Error while fetching model", zap.String("modelid", id), zap.Error(err))
			render.Error(w, r, render.ErrAPIDBSelectFailed, err)
			return
		}
		if !found {
			zap.L().Error("Model does not exists", zap.String("modelid", id))
			render.Error(w, r, render.ErrAPIDBResourceNotFound, err)
			return
		}
	} else {
		idModel, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			zap.L().Error("Error on parsing model id", zap.String("modelID", id))
			render.Error(w, r, render.ErrAPIParsingInteger, err)
			return
		}
		m, found, err = model.R().Get(idModel)
		if err != nil {
			zap.L().Error("Error while fetching model", zap.String("modelid", id), zap.Error(err))
			render.Error(w, r, render.ErrAPIDBSelectFailed, err)
			return
		}
		if !found {
			zap.L().Error("Model does not exists", zap.String("modelid", id))
			render.Error(w, r, render.ErrAPIDBResourceNotFound, err)
			return
		}
	}

	// Might be a security Issue (because we lookup for the fact ID / Name before any control)
	// Should be better to just remove the "lookup by name" feature (which is not used anymore, and has no sense in this API)
	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeModel, strconv.FormatInt(m.ID, 10), permissions.ActionGet)) {
		render.Error(w, r, render.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	render.JSON(w, r, m)
}

// ValidateModel godoc
// @Summary Validate a new model definition
// @Description Validate a new model definition
// @Tags Models
// @Accept json
// @Produce json
// @Param model body interface{} true "Model definition (json)"
// @Security Bearer
// @Success 200 "Status OK"
// @Failure 400 "Status Bad Request"
// @Failure 500 "Status" internal server error"
// @Router /engine/models/validate [post]
func ValidateModel(w http.ResponseWriter, r *http.Request) {
	var newModel modeler.Model
	err := json.NewDecoder(r.Body).Decode(&newModel)
	if err != nil {
		zap.L().Warn("Model json decode", zap.Error(err))
		render.Error(w, r, render.ErrAPIDecodeJSONBody, err)
		return
	}

	if ok, err := newModel.IsValid(); !ok {
		zap.L().Error("Model is invalid", zap.Error(err))
		render.Error(w, r, render.ErrAPIResourceInvalid, err)
		return
	}

	render.JSON(w, r, newModel)
}

// PostModel godoc
// @Summary Create a new model definition
// @Description Create a new model definition
// @Tags Models
// @Accept json
// @Produce json
// @Param model body interface{} true "Model definition (json)"
// @Security Bearer
// @Success 200 "Status OK"
// @Failure 400 "Status Bad Request"
// @Failure 500 "Status" internal server error"
// @Router /engine/models [post]
func PostModel(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeModel, permissions.All, permissions.ActionCreate)) {
		render.Error(w, r, render.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	var newModel modeler.Model
	err := json.NewDecoder(r.Body).Decode(&newModel)
	if err != nil {
		zap.L().Warn("Model json decode", zap.Error(err))
		render.Error(w, r, render.ErrAPIDecodeJSONBody, err)
		return
	}

	if ok, err := newModel.IsValid(); !ok {
		zap.L().Error("Model is invalid", zap.Error(err))
		render.Error(w, r, render.ErrAPIResourceInvalid, err)
		return
	}

	idModel, err := model.R().Create(newModel)
	if err != nil {
		zap.L().Error("Error while creating the Model", zap.Error(err))
		render.Error(w, r, render.ErrAPIDBInsertFailed, err)
		return
	}

	newModelGet, found, err := model.R().Get(idModel)
	if err != nil {
		zap.L().Error("Get Model by ID", zap.Int64("modelID", idModel), zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Error("Model does not exists", zap.Int64("modelID", idModel))
		render.Error(w, r, render.ErrAPIDBResourceNotFoundAfterInsert, err)
		return
	}

	render.JSON(w, r, newModelGet)
}

// PutModel godoc
// @Summary Create or remplace a model definition
// @Description Create or remplace a model definition
// @Tags Models
// @Accept json
// @Produce json
// @Param id path string true "Model ID"
// @Param model body interface{} true "Model definition (json)"
// @Security Bearer
// @Success 200 "Status OK"
// @Failure 400 "Status Bad Request"
// @Failure 500 "Status" internal server error"
// @Router /engine/models/{id} [put]
func PutModel(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idModel, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Error("Error on parsing model id", zap.String("modelID", id))
		render.Error(w, r, render.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeModel, strconv.FormatInt(idModel, 10), permissions.ActionUpdate)) {
		render.Error(w, r, render.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	var newModel modeler.Model
	err = json.NewDecoder(r.Body).Decode(&newModel)
	if err != nil {
		zap.L().Warn("Model json decode", zap.Error(err))
		render.Error(w, r, render.ErrAPIDecodeJSONBody, err)
		return
	}
	newModel.ID = idModel

	if ok, err := newModel.IsValid(); !ok {
		zap.L().Error("Model is invalid", zap.Error(err))
		render.Error(w, r, render.ErrAPIResourceInvalid, err)
		return
	}

	err = model.R().Update(idModel, newModel)
	if err != nil {
		zap.L().Error("Error while updating the Model:", zap.Int64("idModel", idModel), zap.Error(err))
		render.Error(w, r, render.ErrAPIDBUpdateFailed, err)
		return
	}

	newModelGet, found, err := model.R().Get(idModel)
	if err != nil {
		zap.L().Error("Get Model by ID", zap.Int64("modelID", idModel), zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Error("Model does not exists", zap.Int64("modelID", idModel))
		render.Error(w, r, render.ErrAPIDBResourceNotFoundAfterInsert, err)
		return
	}

	render.JSON(w, r, newModelGet)
}

// DeleteModel godoc
// @Summary Delete a model definition
// @Description Delete a model definition
// @Tags Models
// @Produce json
// @Param id path string true "Model ID"
// @Security Bearer
// @Success 200 "Status OK"
// @Failure 400 "Status Bad Request"
// @Router /engine/models/{id} [delete]
func DeleteModel(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idModel, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Error("Error on parsing model id", zap.String("modelID", id))
		render.Error(w, r, render.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeModel, strconv.FormatInt(idModel, 10), permissions.ActionDelete)) {
		render.Error(w, r, render.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	err = model.R().Delete(idModel)
	if err != nil {
		zap.L().Error("Error while deleting the Model:", zap.String("Model ID:", id), zap.Error(err))
		render.Error(w, r, render.ErrAPIDBDeleteFailed, err)
		return
	}

	//render.OK(w, r)
}
