package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/handlers/render"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/models"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/variablesconfig"
	"go.uber.org/zap"
)

// GetVariablesConfig godoc
// @Summary Get all Variables Config definitions
// @Description Get all VariableConfig definitions
// @Tags VariablesConfig
// @Produce json
// @Security Bearer
// @Success 200 {array} models.VariablesConfig "list of all Variables with it's config"
// @Failure 500 "internal server error"
// @Router /engine/variablesconfig [get]
func GetVariablesConfig(w http.ResponseWriter, r *http.Request) {
	VariablesConfig, err := variablesconfig.R().GetAll()
	if err != nil {
		zap.L().Error("Error getting VariableConfigs", zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}

	render.JSON(w, r, VariablesConfig)
}

// GetVariableConfig godoc
// @Summary Get an Variable Config definition
// @Description Get an Variable Config definition
// @Tags VariablesConfig
// @Produce json
// @Param id path string true "Variable Config ID"
// @Security Bearer
// @Success 200 {object} models.VariablesConfig "One Variable Config getted by id"
// @Failure 400 "Status Bad Request"
// @Router /engine/variablesconfig/{id} [get]
func GetVariableConfig(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idVariableConfig, err := strconv.ParseInt(id, 10, 64)

	if err != nil {
		zap.L().Warn("Error on parsing variable config id", zap.String("idVariableConfig", id), zap.Error(err))
		render.Error(w, r, render.ErrAPIParsingInteger, err)
		return
	}

	a, found, err := variablesconfig.R().Get(idVariableConfig)
	if err != nil {
		zap.L().Error("Cannot get Variable Config", zap.String("Variable Config Id", id), zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}

	if !found {
		zap.L().Warn("VariableConfig does not exists", zap.String("VariableConfigId", id))
		render.Error(w, r, render.ErrAPIDBResourceNotFound, err)
		return
	}

	render.JSON(w, r, a)
}

// GetVariableConfigByKey godoc
// @Summary Get an Variable Config definition
// @Description Get an Variable Config definition
// @Tags VariablesConfig
// @Produce json
// @Param key path string true "Variable Config key"
// @Security Bearer
// @Success 200 {object} models.VariablesConfig " One Variable Config getted by key"
// @Failure 400 "Status Bad Request"
// @Router /engine/variablesconfig/key/{key} [get]
func GetVariableConfigByKey(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")

	a, found, err := variablesconfig.R().GetByKey(key)
	if err != nil {
		zap.L().Error("Cannot get VariableConfig", zap.String("Variable Config key", key), zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}

	if !found {
		zap.L().Warn("VariableConfig does not exists", zap.String("Variable Config key", key))
		render.Error(w, r, render.ErrAPIDBResourceNotFound, err)
		return
	}

	render.JSON(w, r, a)
}

// PostVariableConfig godoc
// @Summary Create a new VariableConfig definition
// @Description Create a new VariableConfig definition
// @Tags VariablesConfig
// @Accept json
// @Produce json
// @Param VariableConfig body models.VariablesConfig true "VariableConfig definition (json)"
// @Security Bearer
// @Success 200 {object} models.VariablesConfig "Ppost VariableConfig"
// @Failure 400 "Status Bad Request"
// @Failure 500 "Status" internal server error"
// @Router /engine/variablesconfig [post]
func PostVariableConfig(w http.ResponseWriter, r *http.Request) {

	var newVariableConfig models.VariablesConfig
	err := json.NewDecoder(r.Body).Decode(&newVariableConfig)
	if err != nil {
		zap.L().Warn("Variable Config json decoding", zap.Error(err))
		render.Error(w, r, render.ErrAPIDecodeJSONBody, err)
		return
	}

	id, err := variablesconfig.R().Create(newVariableConfig)
	if err != nil {
		zap.L().Error("Error while creating the VariableConfig", zap.Error(err))
		render.Error(w, r, render.ErrAPIDBInsertFailed, err)
		return
	}

	newVariableConfigGet, found, err := variablesconfig.R().Get(id)
	if err != nil {
		zap.L().Error("Cannot get Variable Config", zap.String("VariableConfig key ", newVariableConfig.Key), zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("VariableConfig does not exists after creation", zap.String("VariableConfig key ", newVariableConfig.Key))
		render.Error(w, r, render.ErrAPIDBResourceNotFoundAfterInsert, err)
		return
	}

	render.JSON(w, r, newVariableConfigGet)
}

// PutVariableConfig godoc
// @Summary update an Variable Config definition
// @Description update an Variable Config definition
// @Tags VariablesConfig
// @Accept json
// @Produce json
// @Param id path string true "VariableConfig ID"
// @Param VariableConfig body models.VariablesConfig true "VariableConfig definition (json)"
// @Security Bearer
// @Success 200 {object} models.VariablesConfig "VariableConfig"
// @Failure 400 "Status Bad Request"
// @Failure 500 "Status" internal server error"
// @Router /engine/variablesconfig/{id} [put]
func PutVariableConfig(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idVariableConfig, err := strconv.ParseInt(id, 10, 64)

	if err != nil {
		zap.L().Warn("Error on parsing variable config id", zap.String("idVariableConfig", id), zap.Error(err))
		render.Error(w, r, render.ErrAPIParsingInteger, err)
		return
	}

	var newVariableConfig models.VariablesConfig
	err = json.NewDecoder(r.Body).Decode(&newVariableConfig)
	if err != nil {
		zap.L().Warn("VariableConfig json decoding", zap.Error(err))
		render.Error(w, r, render.ErrAPIDecodeJSONBody, err)
		return
	}
	newVariableConfig.Id = idVariableConfig

	err = variablesconfig.R().Update(idVariableConfig, newVariableConfig)
	if err != nil {
		zap.L().Error("Error while updating the Variable Config", zap.String("idVariableConfig", id), zap.Any("Variable Config", newVariableConfig), zap.Error(err))
		render.Error(w, r, render.ErrAPIDBUpdateFailed, err)
		return
	}

	newVariableConfigGet, found, err := variablesconfig.R().Get(idVariableConfig)
	if err != nil {
		zap.L().Error("Cannot get VariableConfig", zap.String("VariableConfigId", id), zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("VariableConfig does not exists after update", zap.String("VariableConfigId", id))
		render.Error(w, r, render.ErrAPIDBResourceNotFound, err)
		return
	}

	render.JSON(w, r, newVariableConfigGet)
}

// DeleteVariableConfig godoc
// @Summary Delete an Variable Config definition
// @Description Delete an Variable Config definition
// @Tags VariablesConfig
// @Produce json
// @Param id path string true "VariableConfig ID"
// @Security Bearer
// @Success 200 "Status OK"
// @Failure 400 "Status Bad Request"
// @Router /engine/variablesconfig/{id} [delete]
func DeleteVariableConfig(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idVariableConfig, err := strconv.ParseInt(id, 10, 64)

	if err != nil {
		zap.L().Warn("Error on parsing variable config id", zap.String("idVariableConfig", id), zap.Error(err))
		render.Error(w, r, render.ErrAPIParsingInteger, err)
		return
	}

	err = variablesconfig.R().Delete(idVariableConfig)

	if err != nil {
		zap.L().Error("Error while deleting the VariableConfig", zap.String("VariableConfig ID", id), zap.Error(err))
		render.Error(w, r, render.ErrAPIDBDeleteFailed, err)
		return
	}

	render.OK(w, r)
}
