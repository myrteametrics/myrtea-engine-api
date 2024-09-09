package handlers

import (
	"encoding/json"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/config/externalconfig"
	"net/http"
	"net/url"
	"sort"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/handlers/render"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/models"
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
// @Param id path string true "ExternalConfig ID"
// @Security Bearer
// @Success 200 {object} models.ExternalConfig "externalConfig"
// @Failure 400 "Status Bad Request"
// @Router /engine/externalconfigs/{id} [get]
func GetExternalConfig(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idExternalConfig, err := strconv.ParseInt(id, 10, 64)

	if err != nil {
		zap.L().Warn("Error on parsing external config id", zap.String("idExternalConfig", id), zap.Error(err))
		render.Error(w, r, render.ErrAPIParsingInteger, err)
		return
	}

	a, found, err := externalconfig.R().Get(idExternalConfig)
	if err != nil {
		zap.L().Error("Cannot get externalConfig", zap.String("externalConfigId", id), zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}

	if !found {
		zap.L().Warn("ExternalConfig does not exists", zap.String("externalConfigId", id))
		render.Error(w, r, render.ErrAPIDBResourceNotFound, err)
		return
	}

	render.JSON(w, r, a)
}

// GetExternalConfigByName godoc
// @Summary Get an externalConfig definition
// @Description Get an externalConfig definition
// @Tags ExternalConfigs
// @Produce json
// @Param name path string true "ExternalConfig Name (escaped html accepted)"
// @Security Bearer
// @Success 200 {object} models.ExternalConfig "externalConfig"
// @Failure 400 "Status Bad Request"
// @Router /engine/externalconfigs/name/{name} [get]
func GetExternalConfigByName(w http.ResponseWriter, r *http.Request) {
	escapedName := chi.URLParam(r, "name")
	name, err := url.QueryUnescape(escapedName)
	if err != nil {
		zap.L().Error("Cannot unescape external config name", zap.String("name", escapedName), zap.Error(err))
		render.Error(w, r, render.ErrAPIProcessError, err)
		return
	}

	a, found, err := externalconfig.R().GetByName(name)
	if err != nil {
		zap.L().Error("Cannot get externalConfig", zap.String("externalConfigname", name), zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}

	if !found {
		zap.L().Warn("ExternalConfig does not exists", zap.String("externalConfigname", name))
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

	id, err := externalconfig.R().Create(newExternalConfig)
	if err != nil {
		zap.L().Error("Error while creating the ExternalConfig", zap.Error(err))
		render.Error(w, r, render.ErrAPIDBInsertFailed, err)
		return
	}

	newExternalConfigGet, found, err := externalconfig.R().Get(id)
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
	id := chi.URLParam(r, "id")
	idExternalConfig, err := strconv.ParseInt(id, 10, 64)

	if err != nil {
		zap.L().Warn("Error on parsing external config id", zap.String("idExternalConfig", id), zap.Error(err))
		render.Error(w, r, render.ErrAPIParsingInteger, err)
		return
	}

	var newExternalConfig models.ExternalConfig
	err = json.NewDecoder(r.Body).Decode(&newExternalConfig)
	if err != nil {
		zap.L().Warn("ExternalConfig json decoding", zap.Error(err))
		render.Error(w, r, render.ErrAPIDecodeJSONBody, err)
		return
	}
	newExternalConfig.Id = idExternalConfig

	err = externalconfig.R().Update(idExternalConfig, newExternalConfig)
	if err != nil {
		zap.L().Error("Error while updating the ExternalConfig", zap.String("idExternalConfig", id), zap.Any("externalConfig", newExternalConfig), zap.Error(err))
		render.Error(w, r, render.ErrAPIDBUpdateFailed, err)
		return
	}

	newExternalConfigGet, found, err := externalconfig.R().Get(idExternalConfig)
	if err != nil {
		zap.L().Error("Cannot get externalConfig", zap.String("externalConfigId", id), zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("ExternalConfig does not exists after update", zap.String("externalConfigId", id))
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
	id := chi.URLParam(r, "id")
	idExternalConfig, err := strconv.ParseInt(id, 10, 64)

	if err != nil {
		zap.L().Warn("Error on parsing external config id", zap.String("idExternalConfig", id), zap.Error(err))
		render.Error(w, r, render.ErrAPIParsingInteger, err)
		return
	}

	err = externalconfig.R().Delete(idExternalConfig)

	if err != nil {
		zap.L().Error("Error while deleting the ExternalConfig", zap.String("ExternalConfig ID", id), zap.Error(err))
		render.Error(w, r, render.ErrAPIDBDeleteFailed, err)
		return
	}

	render.OK(w, r)
}

// GetAllOldVersions godoc
// @Summary Get all old versions of a specific externalConfig
// @Description Get all old versions of a specific externalConfig by id
// @Tags ExternalConfigs
// @Produce json
// @Security Bearer
// @Param id path int true "ExternalConfig ID"
// @Success 200 {array} models.ExternalConfig "list of all old versions of the externalConfig"
// @Failure 400 "bad request"
// @Failure 500 "internal server error"
// @Router /engine/externalconfigs/{id}/alloldversions [get]
func GetAllOldVersions(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idExternalConfig, err := strconv.ParseInt(id, 10, 64)

	if err != nil {
		zap.L().Warn("Error on parsing external config id", zap.String("idExternalConfig", id), zap.Error(err))
		render.Error(w, r, render.ErrAPIParsingInteger, err)
		return
	}

	oldVersions, err := externalconfig.R().GetAllOldVersions(idExternalConfig)
	if err != nil {
		zap.L().Error("Error getting old versions of externalConfig", zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}

	if oldVersions == nil {
		render.JSON(w, r, []models.ExternalConfig{})
		return
	}
	
	render.JSON(w, r, oldVersions)
}
