package handlers

import (
	"encoding/json"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/config/esconfig"
	"net/http"
	"net/url"
	"sort"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/handlers/render"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/models"
	"go.uber.org/zap"
)

// GetElasticSearchConfigs godoc
// @Summary Get all elasticSearchConfig definitions
// @Description Get all elasticSearchConfig definitions
// @Tags ElasticSearchConfigs
// @Produce json
// @Security Bearer
// @Success 200 {array} models.ElasticSearchConfig "list of all elasticSearchConfigs"
// @Failure 500 "internal server error"
// @Router /engine/esconfigs [get]
func GetElasticSearchConfigs(w http.ResponseWriter, r *http.Request) {
	elasticSearchConfigs, err := esconfig.R().GetAll()
	if err != nil {
		zap.L().Error("Error getting elasticSearchConfigs", zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}

	elasticSearchConfigsSlice := make([]models.ElasticSearchConfig, 0)
	for _, elasticSearchConfig := range elasticSearchConfigs {
		elasticSearchConfigsSlice = append(elasticSearchConfigsSlice, elasticSearchConfig)
	}

	sort.SliceStable(elasticSearchConfigsSlice, func(i, j int) bool {
		return elasticSearchConfigsSlice[i].Name < elasticSearchConfigsSlice[j].Name
	})

	render.JSON(w, r, elasticSearchConfigsSlice)
}

// GetElasticSearchConfig godoc
// @Summary Get an elasticSearchConfig definition
// @Description Get an elasticSearchConfig definition
// @Tags ElasticSearchConfigs
// @Produce json
// @Param id path string true "ElasticSearchConfig ID"
// @Security Bearer
// @Success 200 {object} models.ElasticSearchConfig "elasticSearchConfig"
// @Failure 400 "Status Bad Request"
// @Router /engine/esconfigs/{id} [get]
func GetElasticSearchConfig(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idElasticSearchConfig, err := strconv.ParseInt(id, 10, 64)

	if err != nil {
		zap.L().Warn("Error on parsing external config id", zap.String("idElasticSearchConfig", id), zap.Error(err))
		render.Error(w, r, render.ErrAPIParsingInteger, err)
		return
	}

	a, found, err := esconfig.R().Get(idElasticSearchConfig)
	if err != nil {
		zap.L().Error("Cannot get elasticSearchConfig", zap.String("elasticSearchConfigId", id), zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}

	if !found {
		zap.L().Warn("ElasticSearchConfig does not exists", zap.String("elasticSearchConfigId", id))
		render.Error(w, r, render.ErrAPIDBResourceNotFound, err)
		return
	}

	render.JSON(w, r, a)
}

// GetElasticSearchConfigByName godoc
// @Summary Get an elasticSearchConfig definition
// @Description Get an elasticSearchConfig definition
// @Tags ElasticSearchConfigs
// @Produce json
// @Param name path string true "ElasticSearchConfig Name (escaped html accepted)"
// @Security Bearer
// @Success 200 {object} models.ElasticSearchConfig "elasticSearchConfig"
// @Failure 400 "Status Bad Request"
// @Router /engine/esconfigs/name/{name} [get]
func GetElasticSearchConfigByName(w http.ResponseWriter, r *http.Request) {
	escapedName := chi.URLParam(r, "name")
	name, err := url.QueryUnescape(escapedName)
	if err != nil {
		zap.L().Error("Cannot unescape external config name", zap.String("name", escapedName), zap.Error(err))
		render.Error(w, r, render.ErrAPIProcessError, err)
		return
	}

	a, found, err := esconfig.R().GetByName(name)
	if err != nil {
		zap.L().Error("Cannot get elasticSearchConfig", zap.String("elasticSearchConfigName", name), zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}

	if !found {
		zap.L().Warn("ElasticSearchConfig does not exists", zap.String("elasticSearchConfigName", name))
		render.Error(w, r, render.ErrAPIDBResourceNotFound, err)
		return
	}

	render.JSON(w, r, a)
}

// GetDefaultElasticSearchConfig godoc
// @Summary Get the default elasticSearchConfig definition
// @Description Get the default elasticSearchConfig definition
// @Tags ElasticSearchConfigs
// @Produce json
// @Security Bearer
// @Success 200 {object} models.ElasticSearchConfig "elasticSearchConfig"
// @Failure 400 "Status Bad Request"
// @Router /engine/esconfigs/default [get]
func GetDefaultElasticSearchConfig(w http.ResponseWriter, r *http.Request) {
	a, found, err := esconfig.R().GetDefault()
	if err != nil {
		zap.L().Error("Cannot get default elasticSearchConfig", zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}

	if !found {
		zap.L().Warn("Default ElasticSearchConfig does not exists")
		render.Error(w, r, render.ErrAPIDBResourceNotFound, err)
		return
	}

	render.JSON(w, r, a)
}

// PostElasticSearchConfig godoc
// @Summary Create a new elasticSearchConfig definition
// @Description Create a new elasticSearchConfig definition
// @Tags ElasticSearchConfigs
// @Accept json
// @Produce json
// @Param elasticSearchConfig body models.ElasticSearchConfig true "ElasticSearchConfig definition (json)"
// @Security Bearer
// @Success 200 {object} models.ElasticSearchConfig "elasticSearchConfig"
// @Failure 400 "Status Bad Request"
// @Failure 500 "Status" internal server error"
// @Router /engine/esconfigs [post]
func PostElasticSearchConfig(w http.ResponseWriter, r *http.Request) {

	var newElasticSearchConfig models.ElasticSearchConfig
	err := json.NewDecoder(r.Body).Decode(&newElasticSearchConfig)
	if err != nil {
		zap.L().Warn("ElasticSearchConfig json decoding", zap.Error(err))
		render.Error(w, r, render.ErrAPIDecodeJSONBody, err)
		return
	}

	id, err := esconfig.R().Create(newElasticSearchConfig)
	if err != nil {
		zap.L().Error("Error while creating the ElasticSearchConfig", zap.Error(err))
		render.Error(w, r, render.ErrAPIDBInsertFailed, err)
		return
	}

	newElasticSearchConfigGet, found, err := esconfig.R().Get(id)
	if err != nil {
		zap.L().Error("Cannot get elasticSearchConfig", zap.String("elasticSearchConfigName", newElasticSearchConfig.Name), zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("ElasticSearchConfig does not exists after creation", zap.String("elasticSearchConfigName", newElasticSearchConfig.Name))
		render.Error(w, r, render.ErrAPIDBResourceNotFoundAfterInsert, err)
		return
	}

	render.JSON(w, r, newElasticSearchConfigGet)
}

// PutElasticSearchConfig godoc
// @Summary Create or remplace an elasticSearchConfig definition
// @Description Create or remplace an elasticSearchConfig definition
// @Tags ElasticSearchConfigs
// @Accept json
// @Produce json
// @Param name path string true "ElasticSearchConfig ID"
// @Param elasticSearchConfig body models.ElasticSearchConfig true "ElasticSearchConfig definition (json)"
// @Security Bearer
// @Success 200 {object} models.ElasticSearchConfig "elasticSearchConfig"
// @Failure 400 "Status Bad Request"
// @Failure 500 "Status" internal server error"
// @Router /engine/esconfigs/{name} [put]
func PutElasticSearchConfig(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idElasticSearchConfig, err := strconv.ParseInt(id, 10, 64)

	if err != nil {
		zap.L().Warn("Error on parsing external config id", zap.String("idElasticSearchConfig", id), zap.Error(err))
		render.Error(w, r, render.ErrAPIParsingInteger, err)
		return
	}

	var newElasticSearchConfig models.ElasticSearchConfig
	err = json.NewDecoder(r.Body).Decode(&newElasticSearchConfig)
	if err != nil {
		zap.L().Warn("ElasticSearchConfig json decoding", zap.Error(err))
		render.Error(w, r, render.ErrAPIDecodeJSONBody, err)
		return
	}
	newElasticSearchConfig.Id = idElasticSearchConfig

	err = esconfig.R().Update(idElasticSearchConfig, newElasticSearchConfig)
	if err != nil {
		zap.L().Error("Error while updating the ElasticSearchConfig", zap.String("idElasticSearchConfig", id), zap.Any("elasticSearchConfig", newElasticSearchConfig), zap.Error(err))
		render.Error(w, r, render.ErrAPIDBUpdateFailed, err)
		return
	}

	newEsConfig, found, err := esconfig.R().Get(idElasticSearchConfig)
	if err != nil {
		zap.L().Error("Cannot get elasticSearchConfig", zap.String("elasticSearchConfigId", id), zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("ElasticSearchConfig does not exists after update", zap.String("elasticSearchConfigId", id))
		render.Error(w, r, render.ErrAPIDBResourceNotFound, err)
		return
	}

	render.JSON(w, r, newEsConfig)
}

// DeleteElasticSearchConfig godoc
// @Summary Delete an elasticSearchConfig definition
// @Description Delete an elasticSearchConfig definition
// @Tags ElasticSearchConfigs
// @Produce json
// @Param name path string true "ElasticSearchConfig ID"
// @Security Bearer
// @Success 200 "Status OK"
// @Failure 400 "Status Bad Request"
// @Router /engine/esconfigs/{name} [delete]
func DeleteElasticSearchConfig(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idElasticSearchConfig, err := strconv.ParseInt(id, 10, 64)

	if err != nil {
		zap.L().Warn("Error on parsing external config id", zap.String("idElasticSearchConfig", id), zap.Error(err))
		render.Error(w, r, render.ErrAPIParsingInteger, err)
		return
	}

	err = esconfig.R().Delete(idElasticSearchConfig)

	if err != nil {
		zap.L().Error("Error while deleting the ElasticSearchConfig", zap.String("ElasticSearchConfig ID", id), zap.Error(err))
		render.Error(w, r, render.ErrAPIDBDeleteFailed, err)
		return
	}

	render.OK(w, r)
}
