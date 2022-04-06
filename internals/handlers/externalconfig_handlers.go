package handlers

import (
	"encoding/json"
	"net/http"
	"sort"

	"github.com/go-chi/chi"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/externalconfig"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/handlers/render"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/models"
	"go.uber.org/zap"
)

// GetExternalConfigs godoc
// @Summary Get all externalConfig definitions
// @Description Get all externalConfig definitions
// @Tags ExternalConfigs
// @Produce json
// @Security Bearer
// @Success 200 {array} models.ExternalConfig "list of all externalConfigs"
// @Failure 500 "internal server error"
// @Router /engine/externalconfigs [get]
func GetExternalConfigs(w http.ResponseWriter, r *http.Request) {
	externalConfigs, err := externalconfig.R().GetAll()
	if err != nil {
		zap.L().Error("Error getting externalConfigs", zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}

	externalConfigsSlice := make([]models.ExternalConfig, 0)
	for _, externalConfig := range externalConfigs {
		externalConfigsSlice = append(externalConfigsSlice, externalConfig)
	}

	sort.SliceStable(externalConfigsSlice, func(i, j int) bool {
		return externalConfigsSlice[i].Name < externalConfigsSlice[j].Name
	})

	render.JSON(w, r, externalConfigsSlice)
}

// GetExternalConfig godoc
// @Summary Get an externalConfig definition
// @Description Get an externalConfig definition
// @Tags ExternalConfigs
// @Produce json
// @Param name path string true "ExternalConfig ID"
// @Security Bearer
// @Success 200 {object} models.ExternalConfig "externalConfig"
// @Failure 400 "Status Bad Request"
// @Router /engine/externalconfigs/{name} [get]
func GetExternalConfig(w http.ResponseWriter, r *http.Request) {
	nameExternalConfig := chi.URLParam(r, "name")
	a, found, err := externalconfig.R().Get(nameExternalConfig)
	if err != nil {
		zap.L().Error("Cannot get externalConfig", zap.String("externalConfigname", nameExternalConfig), zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("ExternalConfig does not exists", zap.String("externalConfigname", nameExternalConfig))
		render.Error(w, r, render.ErrAPIDBResourceNotFound, err)
		return
	}

	render.JSON(w, r, a)
}

// PostExternalConfig godoc
// @Summary Create a new externalConfig definition
// @Description Create a new externalConfig definition
// @Tags ExternalConfigs
// @Accept json
// @Produce json
// @Param externalConfig body models.ExternalConfig true "ExternalConfig definition (json)"
// @Security Bearer
// @Success 200 {object} models.ExternalConfig "externalConfig"
// @Failure 400 "Status Bad Request"
// @Failure 500 "Status" internal server error"
// @Router /engine/externalconfigs [post]
func PostExternalConfig(w http.ResponseWriter, r *http.Request) {

	var newExternalConfig models.ExternalConfig
	err := json.NewDecoder(r.Body).Decode(&newExternalConfig)
	if err != nil {
		zap.L().Warn("ExternalConfig json decoding", zap.Error(err))
		render.Error(w, r, render.ErrAPIDecodeJSONBody, err)
		return
	}

	err = externalconfig.R().Create(nil, newExternalConfig)
	if err != nil {
		zap.L().Error("Error while creating the ExternalConfig", zap.Error(err))
		render.Error(w, r, render.ErrAPIDBInsertFailed, err)
		return
	}

	newExternalConfigGet, found, err := externalconfig.R().Get(newExternalConfig.Name)
	if err != nil {
		zap.L().Error("Cannot get externalConfig", zap.String("externalConfigname", newExternalConfig.Name), zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("ExternalConfig does not exists after creation", zap.String("externalConfigname", newExternalConfig.Name))
		render.Error(w, r, render.ErrAPIDBResourceNotFoundAfterInsert, err)
		return
	}

	render.JSON(w, r, newExternalConfigGet)
}

// PutExternalConfig godoc
// @Summary Create or remplace an externalConfig definition
// @Description Create or remplace an externalConfig definition
// @Tags ExternalConfigs
// @Accept json
// @Produce json
// @Param name path string true "ExternalConfig ID"
// @Param externalConfig body models.ExternalConfig true "ExternalConfig definition (json)"
// @Security Bearer
// @Success 200 {object} models.ExternalConfig "externalConfig"
// @Failure 400 "Status Bad Request"
// @Failure 500 "Status" internal server error"
// @Router /engine/externalconfigs/{name} [put]
func PutExternalConfig(w http.ResponseWriter, r *http.Request) {
	nameExternalConfig := chi.URLParam(r, "name")

	var newExternalConfig models.ExternalConfig
	err := json.NewDecoder(r.Body).Decode(&newExternalConfig)
	if err != nil {
		zap.L().Warn("ExternalConfig json decoding", zap.Error(err))
		render.Error(w, r, render.ErrAPIDecodeJSONBody, err)
		return
	}
	newExternalConfig.Name = nameExternalConfig

	err = externalconfig.R().Update(nil, nameExternalConfig, newExternalConfig)
	if err != nil {
		zap.L().Error("Error while updating the ExternalConfig", zap.String("nameExternalConfig", nameExternalConfig), zap.Any("externalConfig", newExternalConfig), zap.Error(err))
		render.Error(w, r, render.ErrAPIDBUpdateFailed, err)
		return
	}

	newExternalConfigGet, found, err := externalconfig.R().Get(nameExternalConfig)
	if err != nil {
		zap.L().Error("Cannot get externalConfig", zap.String("externalConfigname", nameExternalConfig), zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("ExternalConfig does not exists after update", zap.String("externalConfigname", nameExternalConfig))
		render.Error(w, r, render.ErrAPIDBResourceNotFound, err)
		return
	}

	render.JSON(w, r, newExternalConfigGet)
}

// DeleteExternalConfig godoc
// @Summary Delete an externalConfig definition
// @Description Delete an externalConfig definition
// @Tags ExternalConfigs
// @Produce json
// @Param name path string true "ExternalConfig ID"
// @Security Bearer
// @Success 200 "Status OK"
// @Failure 400 "Status Bad Request"
// @Router /engine/externalconfigs/{name} [delete]
func DeleteExternalConfig(w http.ResponseWriter, r *http.Request) {
	nameExternalConfig := chi.URLParam(r, "name")
	err := externalconfig.R().Delete(nil, nameExternalConfig)
	if err != nil {
		zap.L().Error("Error while deleting the ExternalConfig", zap.String("ExternalConfig ID", nameExternalConfig), zap.Error(err))
		render.Error(w, r, render.ErrAPIDBDeleteFailed, err)
		return
	}

	render.OK(w, r)
}
