package handlers

import (
	"encoding/json"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/config/connectorconfig"
	"net/http"
	"sort"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/handlers/render"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/models"
	"go.uber.org/zap"
)

// GetConnectorConfigs godoc
// @Summary Get all connectorConfig definitions
// @Description Get all connectorConfig definitions
// @Tags ConnectorConfigs
// @Produce json
// @Security Bearer
// @Security ApiKeyAuth
// @Success 200 {array} models.ConnectorConfig "list of all connectorConfigs"
// @Failure 500 "internal server error"
// @Router /engine/connectorconfigs [get]
func GetConnectorConfigs(w http.ResponseWriter, r *http.Request) {
	connectorConfigs, err := connectorconfig.R().GetAll()
	if err != nil {
		zap.L().Error("Error getting connectorConfigs", zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}

	connectorConfigsSlice := make([]models.ConnectorConfig, 0)
	for _, connectorConfig := range connectorConfigs {
		connectorConfigsSlice = append(connectorConfigsSlice, connectorConfig)
	}

	sort.SliceStable(connectorConfigsSlice, func(i, j int) bool {
		return connectorConfigsSlice[i].Name < connectorConfigsSlice[j].Name
	})

	render.JSON(w, r, connectorConfigsSlice)
}

// GetConnectorConfig godoc
// @Summary Get an connectorConfig definition
// @Description Get an connectorConfig definition
// @Tags ConnectorConfigs
// @Produce json
// @Param id path string true "ConnectorConfig ID"
// @Security Bearer
// @Security ApiKeyAuth
// @Success 200 {object} models.ConnectorConfig "connectorConfig"
// @Failure 400 "Status Bad Request"
// @Router /engine/connectorconfigs/{id} [get]
func GetConnectorConfig(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idConnectorConfig, err := strconv.ParseInt(id, 10, 64)

	if err != nil {
		zap.L().Warn("Error on parsing connector config id", zap.String("idConnectorConfig", id), zap.Error(err))
		render.Error(w, r, render.ErrAPIParsingInteger, err)
		return
	}

	a, found, err := connectorconfig.R().Get(idConnectorConfig)
	if err != nil {
		zap.L().Error("Cannot get connectorConfig", zap.String("externalConfigId", id), zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}

	if !found {
		zap.L().Warn("ConnectorConfig does not exists", zap.String("externalConfigId", id))
		render.Error(w, r, render.ErrAPIDBResourceNotFound, err)
		return
	}

	render.JSON(w, r, a)
}

// PostConnectorConfig godoc
// @Summary Create a new connectorConfig definition
// @Description Create a new connectorConfig definition
// @Tags ConnectorConfigs
// @Accept json
// @Produce json
// @Param connectorConfig body models.ConnectorConfig true "ConnectorConfig definition (json)"
// @Security Bearer
// @Security ApiKeyAuth
// @Success 200 {object} models.ConnectorConfig "connectorConfig"
// @Failure 400 "Status Bad Request"
// @Failure 500 "Status" internal server error"
// @Router /engine/connectorconfigs [post]
func PostConnectorConfig(w http.ResponseWriter, r *http.Request) {
	var newExternalConfig models.ConnectorConfig
	err := json.NewDecoder(r.Body).Decode(&newExternalConfig)
	if err != nil {
		zap.L().Warn("ConnectorConfig json decoding", zap.Error(err))
		render.Error(w, r, render.ErrAPIDecodeJSONBody, err)
		return
	}

	id, err := connectorconfig.R().Create(nil, newExternalConfig)
	if err != nil {
		zap.L().Error("Error while creating the ConnectorConfig", zap.Error(err))
		render.Error(w, r, render.ErrAPIDBInsertFailed, err)
		return
	}

	newExternalConfigGet, found, err := connectorconfig.R().Get(id)
	if err != nil {
		zap.L().Error("Cannot get connectorConfig", zap.String("externalConfigname", newExternalConfig.Name), zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("ConnectorConfig does not exists after creation", zap.String("externalConfigname", newExternalConfig.Name))
		render.Error(w, r, render.ErrAPIDBResourceNotFoundAfterInsert, err)
		return
	}

	render.JSON(w, r, newExternalConfigGet)
}

// PutConnectorConfig godoc
// @Summary Create or remplace an connectorConfig definition
// @Description Create or remplace an connectorConfig definition
// @Tags ConnectorConfigs
// @Accept json
// @Produce json
// @Param name path string true "ConnectorConfig ID"
// @Param connectorConfig body models.ConnectorConfig true "ConnectorConfig definition (json)"
// @Security Bearer
// @Security ApiKeyAuth
// @Success 200 {object} models.ConnectorConfig "connectorConfig"
// @Failure 400 "Status Bad Request"
// @Failure 500 "Status" internal server error"
// @Router /engine/connectorconfigs/{name} [put]
func PutConnectorConfig(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idConnectorConfig, err := strconv.ParseInt(id, 10, 64)

	if err != nil {
		zap.L().Warn("Error on parsing connector config id", zap.String("idConnectorConfig", id), zap.Error(err))
		render.Error(w, r, render.ErrAPIParsingInteger, err)
		return
	}

	var newExternalConfig models.ConnectorConfig
	err = json.NewDecoder(r.Body).Decode(&newExternalConfig)
	if err != nil {
		zap.L().Warn("ConnectorConfig json decoding", zap.Error(err))
		render.Error(w, r, render.ErrAPIDecodeJSONBody, err)
		return
	}
	newExternalConfig.Id = idConnectorConfig

	err = connectorconfig.R().Update(nil, idConnectorConfig, newExternalConfig)
	if err != nil {
		zap.L().Error("Error while updating the ConnectorConfig", zap.String("idConnectorConfig", id), zap.Any("connectorConfig", newExternalConfig), zap.Error(err))
		render.Error(w, r, render.ErrAPIDBUpdateFailed, err)
		return
	}

	newExternalConfigGet, found, err := connectorconfig.R().Get(idConnectorConfig)
	if err != nil {
		zap.L().Error("Cannot get connectorConfig", zap.String("externalConfigId", id), zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("ConnectorConfig does not exists after update", zap.String("externalConfigId", id))
		render.Error(w, r, render.ErrAPIDBResourceNotFound, err)
		return
	}

	render.JSON(w, r, newExternalConfigGet)
}

// DeleteConnectorConfig godoc
// @Summary Delete an connectorConfig definition
// @Description Delete an connectorConfig definition
// @Tags ConnectorConfigs
// @Produce json
// @Param name path string true "ConnectorConfig ID"
// @Security Bearer
// @Security ApiKeyAuth
// @Success 200 "Status OK"
// @Failure 400 "Status Bad Request"
// @Router /engine/connectorconfigs/{name} [delete]
func DeleteConnectorConfig(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idConnectorConfig, err := strconv.ParseInt(id, 10, 64)

	if err != nil {
		zap.L().Warn("Error on parsing connector config id", zap.String("idConnectorConfig", id), zap.Error(err))
		render.Error(w, r, render.ErrAPIParsingInteger, err)
		return
	}

	err = connectorconfig.R().Delete(nil, idConnectorConfig)

	if err != nil {
		zap.L().Error("Error while deleting the ConnectorConfig", zap.String("ConnectorConfig ID", id), zap.Error(err))
		render.Error(w, r, render.ErrAPIDBDeleteFailed, err)
		return
	}

	render.OK(w, r)
}
