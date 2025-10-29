package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"sort"
	"strconv"

	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/security/permissions"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/utils/httputil"

	"github.com/go-chi/chi/v5"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/config/esconfig"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/model"
	"go.uber.org/zap"
)

// GetElasticSearchConfigs godoc
//
//	@Id				GetElasticSearchConfigs
//
//	@Summary		Get all elasticSearchConfig definitions
//	@Description	Get all elasticSearchConfig definitions
//	@Tags			ElasticSearchConfigs
//	@Produce		json
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{array}	model.ElasticSearchConfig	"list of all elasticSearchConfigs"
//	@Failure		500	"internal server error"
//	@Router			/engine/esconfigs [get]
func GetElasticSearchConfigs(w http.ResponseWriter, r *http.Request) {

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeConfig, permissions.All, permissions.ActionList)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	elasticSearchConfigs, err := esconfig.R().GetAll()
	if err != nil {
		zap.L().Error("Error getting elasticSearchConfigs", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}

	elasticSearchConfigsSlice := make([]model.ElasticSearchConfig, 0)
	for _, elasticSearchConfig := range elasticSearchConfigs {
		// Password is already excluded from GetAll in the repository
		elasticSearchConfigsSlice = append(elasticSearchConfigsSlice, elasticSearchConfig)
	}

	sort.SliceStable(elasticSearchConfigsSlice, func(i, j int) bool {
		return elasticSearchConfigsSlice[i].Name < elasticSearchConfigsSlice[j].Name
	})

	httputil.JSON(w, r, elasticSearchConfigsSlice)
}

// GetElasticSearchConfig godoc
//
//	@Id				GetElasticSearchConfig
//
//	@Summary		Get an elasticSearchConfig definition
//	@Description	Get an elasticSearchConfig definition
//	@Tags			ElasticSearchConfigs
//	@Produce		json
//	@Param			id	path	string	true	"ElasticSearchConfig ID"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	model.ElasticSearchConfig	"elasticSearchConfig"
//	@Failure		400	"Status Bad Request"
//	@Router			/engine/esconfigs/{id} [get]
func GetElasticSearchConfig(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeConfig, permissions.All, permissions.ActionGet)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	id := chi.URLParam(r, "id")
	idElasticSearchConfig, err := strconv.ParseInt(id, 10, 64)

	if err != nil {
		zap.L().Warn("Error on parsing external config id", zap.String("idElasticSearchConfig", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	a, found, err := esconfig.R().Get(idElasticSearchConfig)
	if err != nil {
		zap.L().Error("Cannot get elasticSearchConfig", zap.String("elasticSearchConfigId", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}

	if !found {
		zap.L().Warn("ElasticSearchConfig does not exists", zap.String("elasticSearchConfigId", id))
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFound, err)
		return
	}

	// Password is returned with hashed value from Get - keep it as is for security
	httputil.JSON(w, r, a)
}

// GetElasticSearchConfigByName godoc
//
//	@Id				GetElasticSearchConfigByName
//
//	@Summary		Get an elasticSearchConfig definition
//	@Description	Get an elasticSearchConfig definition
//	@Tags			ElasticSearchConfigs
//	@Produce		json
//	@Param			name	path	string	true	"ElasticSearchConfig Name (escaped html accepted)"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	model.ElasticSearchConfig	"elasticSearchConfig"
//	@Failure		400	"Status Bad Request"
//	@Router			/engine/esconfigs/name/{name} [get]
func GetElasticSearchConfigByName(w http.ResponseWriter, r *http.Request) {

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeConfig, permissions.All, permissions.ActionGet)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	escapedName := chi.URLParam(r, "name")
	name, err := url.QueryUnescape(escapedName)
	if err != nil {
		zap.L().Error("Cannot unescape external config name", zap.String("name", escapedName), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIProcessError, err)
		return
	}

	a, found, err := esconfig.R().GetByName(name)
	if err != nil {
		zap.L().Error("Cannot get elasticSearchConfig", zap.String("elasticSearchConfigName", name), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}

	if !found {
		zap.L().Warn("ElasticSearchConfig does not exists", zap.String("elasticSearchConfigName", name))
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFound, err)
		return
	}

	httputil.JSON(w, r, a)
}

// GetDefaultElasticSearchConfig godoc
//
//	@Id				GetDefaultElasticSearchConfig
//
//	@Summary		Get the default elasticSearchConfig definition
//	@Description	Get the default elasticSearchConfig definition
//	@Tags			ElasticSearchConfigs
//	@Produce		json
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	model.ElasticSearchConfig	"elasticSearchConfig"
//	@Failure		400	"Status Bad Request"
//	@Router			/engine/esconfigs/default [get]
func GetDefaultElasticSearchConfig(w http.ResponseWriter, r *http.Request) {

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeConfig, permissions.All, permissions.ActionGet)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	a, found, err := esconfig.R().GetDefault()
	if err != nil {
		zap.L().Error("Cannot get default elasticSearchConfig", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}

	if !found {
		zap.L().Warn("Default ElasticSearchConfig does not exists")
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFound, err)
		return
	}

	httputil.JSON(w, r, a)
}

// PostElasticSearchConfig godoc
//
//	@Id				PostElasticSearchConfig
//
//	@Summary		Create a new elasticSearchConfig definition
//	@Description	Create a new elasticSearchConfig definition
//	@Tags			ElasticSearchConfigs
//	@Accept			json
//	@Produce		json
//	@Param			elasticSearchConfig	body	model.ElasticSearchConfig	true	"ElasticSearchConfig definition (json)"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	model.ElasticSearchConfig	"elasticSearchConfig"
//	@Failure		400	"Status Bad Request"
//	@Failure		500	"Status"	internal	server	error"
//	@Router			/engine/esconfigs [post]
func PostElasticSearchConfig(w http.ResponseWriter, r *http.Request) {

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeConfig, permissions.All, permissions.ActionCreate)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	var newElasticSearchConfig model.ElasticSearchConfig
	err := json.NewDecoder(r.Body).Decode(&newElasticSearchConfig)
	if err != nil {
		zap.L().Warn("ElasticSearchConfig json decoding", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDecodeJSONBody, err)
		return
	}

	id, err := esconfig.R().Create(newElasticSearchConfig)
	if err != nil {
		zap.L().Error("Error while creating the ElasticSearchConfig", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBInsertFailed, err)
		return
	}

	newElasticSearchConfigGet, found, err := esconfig.R().Get(id)
	if err != nil {
		zap.L().Error("Cannot get elasticSearchConfig", zap.String("elasticSearchConfigName", newElasticSearchConfig.Name), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("ElasticSearchConfig does not exists after creation", zap.String("elasticSearchConfigName", newElasticSearchConfig.Name))
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFoundAfterInsert, err)
		return
	}

	httputil.JSON(w, r, newElasticSearchConfigGet)
}

// PutElasticSearchConfig godoc
//
//	@Id				PutElasticSearchConfig
//
//	@Summary		Create or remplace an elasticSearchConfig definition
//	@Description	Create or remplace an elasticSearchConfig definition
//	@Tags			ElasticSearchConfigs
//	@Accept			json
//	@Produce		json
//	@Param			id					path	string						true	"ElasticSearchConfig ID"
//	@Param			elasticSearchConfig	body	model.ElasticSearchConfig	true	"ElasticSearchConfig definition (json)"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	model.ElasticSearchConfig	"elasticSearchConfig"
//	@Failure		400	"Status Bad Request"
//	@Failure		500	"Status"	internal	server	error"
//	@Router			/engine/esconfigs/{id} [put]
func PutElasticSearchConfig(w http.ResponseWriter, r *http.Request) {

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeConfig, permissions.All, permissions.ActionUpdate)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	id := chi.URLParam(r, "id")
	idElasticSearchConfig, err := strconv.ParseInt(id, 10, 64)

	if err != nil {
		zap.L().Warn("Error on parsing external config id", zap.String("idElasticSearchConfig", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	var newElasticSearchConfig model.ElasticSearchConfig
	err = json.NewDecoder(r.Body).Decode(&newElasticSearchConfig)
	if err != nil {
		zap.L().Warn("ElasticSearchConfig json decoding", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDecodeJSONBody, err)
		return
	}
	newElasticSearchConfig.Id = idElasticSearchConfig

	err = esconfig.R().Update(idElasticSearchConfig, newElasticSearchConfig)
	if err != nil {
		zap.L().Error("Error while updating the ElasticSearchConfig", zap.String("idElasticSearchConfig", id), zap.Any("elasticSearchConfig", newElasticSearchConfig), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBUpdateFailed, err)
		return
	}

	newEsConfig, found, err := esconfig.R().Get(idElasticSearchConfig)
	if err != nil {
		zap.L().Error("Cannot get elasticSearchConfig", zap.String("elasticSearchConfigId", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("ElasticSearchConfig does not exists after update", zap.String("elasticSearchConfigId", id))
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFound, err)
		return
	}

	httputil.JSON(w, r, newEsConfig)
}

// DeleteElasticSearchConfig godoc
//
//	@Id				DeleteElasticSearchConfig
//
//	@Summary		Delete an elasticSearchConfig definition
//	@Description	Delete an elasticSearchConfig definition
//	@Tags			ElasticSearchConfigs
//	@Produce		json
//	@Param			id	path	string	true	"ElasticSearchConfig ID"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	"Status OK"
//	@Failure		400	"Status Bad Request"
//	@Router			/engine/esconfigs/{id} [delete]
func DeleteElasticSearchConfig(w http.ResponseWriter, r *http.Request) {

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeConfig, permissions.All, permissions.ActionDelete)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	id := chi.URLParam(r, "id")
	idElasticSearchConfig, err := strconv.ParseInt(id, 10, 64)

	if err != nil {
		zap.L().Warn("Error on parsing external config id", zap.String("idElasticSearchConfig", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	err = esconfig.R().Delete(idElasticSearchConfig)
	if err != nil {
		zap.L().Error("Error while deleting the ElasticSearchConfig", zap.String("ElasticSearchConfig ID", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBDeleteFailed, err)
		return
	}

	httputil.OK(w, r)
}
