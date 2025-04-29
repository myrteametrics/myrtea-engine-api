package handlers

import (
	"encoding/json"
	"errors"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/security/permissions"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/utils/httputil"
	"github.com/myrteametrics/myrtea-sdk/v5/repositories/externalconfig"
	"net/http"
	"net/url"
	"sort"
	"strconv"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// GetExternalConfigs godoc
//
//	@Summary		Get all externalConfig definitions
//	@Description	Get all externalConfig definitions
//	@Tags			ExternalConfigs
//	@Produce		json
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{array}	externalconfig.ExternalConfig	"list of all externalConfigs"
//	@Failure		500	"internal server error"
//	@Router			/engine/externalconfigs [get]
func GetExternalConfigs(w http.ResponseWriter, r *http.Request) {

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeConfig, permissions.All, permissions.ActionList)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	externalConfigs, err := externalconfig.R().GetAll()
	if err != nil {
		zap.L().Error("Error getting externalConfigs", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}

	externalConfigsSlice := make([]externalconfig.ExternalConfig, 0)
	for _, externalConfig := range externalConfigs {
		externalConfigsSlice = append(externalConfigsSlice, externalConfig)
	}

	sort.SliceStable(externalConfigsSlice, func(i, j int) bool {
		return externalConfigsSlice[i].Name < externalConfigsSlice[j].Name
	})

	httputil.JSON(w, r, externalConfigsSlice)
}

// GetExternalConfig godoc
//
//	@Summary		Get an externalConfig definition
//	@Description	Get an externalConfig definition
//	@Tags			ExternalConfigs
//	@Produce		json
//	@Param			id	path	string	true	"ExternalConfig ID"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	externalconfig.ExternalConfig	"externalConfig"
//	@Failure		400	"Status Bad Request"
//	@Router			/engine/externalconfigs/{id} [get]
func GetExternalConfig(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")
	idExternalConfig, err := strconv.ParseInt(id, 10, 64)

	if err != nil {
		zap.L().Warn("Error on parsing external config id", zap.String("idExternalConfig", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	a, found, err := externalconfig.R().Get(idExternalConfig)
	if err != nil {
		zap.L().Error("Cannot get externalConfig", zap.String("externalConfigId", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}

	if !found {
		zap.L().Warn("ExternalConfig does not exists", zap.String("externalConfigId", id))
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFound, err)
		return
	}

	httputil.JSON(w, r, a)
}

// GetExternalConfigByName godoc
//
//	@Summary		Get an externalConfig definition
//	@Description	Get an externalConfig definition
//	@Tags			ExternalConfigs
//	@Produce		json
//	@Param			name	path	string	true	"ExternalConfig Name (escaped html accepted)"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	externalconfig.ExternalConfig	"externalConfig"
//	@Failure		400	"Status Bad Request"
//	@Router			/engine/externalconfigs/name/{name} [get]
func GetExternalConfigByName(w http.ResponseWriter, r *http.Request) {

	escapedName := chi.URLParam(r, "name")
	name, err := url.QueryUnescape(escapedName)
	if err != nil {
		zap.L().Error("Cannot unescape external config name", zap.String("name", escapedName), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIProcessError, err)
		return
	}

	a, found, err := externalconfig.R().GetByName(name)
	if err != nil {
		zap.L().Error("Cannot get externalConfig", zap.String("externalConfigname", name), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}

	if !found {
		zap.L().Warn("ExternalConfig does not exists", zap.String("externalConfigname", name))
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFound, err)
		return
	}

	httputil.JSON(w, r, a)
}

// PostExternalConfig godoc
//
//	@Summary		Create a new externalConfig definition
//	@Description	Create a new externalConfig definition
//	@Tags			ExternalConfigs
//	@Accept			json
//	@Produce		json
//	@Param			externalConfig	body	externalconfig.ExternalConfig	true	"ExternalConfig definition (json)"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	externalconfig.ExternalConfig	"externalConfig"
//	@Failure		400	"Status Bad Request"
//	@Failure		500	"Status"	internal	server	error"
//	@Router			/engine/externalconfigs [post]
func PostExternalConfig(w http.ResponseWriter, r *http.Request) {

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeConfig, permissions.All, permissions.ActionCreate)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	var newExternalConfig externalconfig.ExternalConfig
	err := json.NewDecoder(r.Body).Decode(&newExternalConfig)
	if err != nil {
		zap.L().Warn("ExternalConfig json decoding", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDecodeJSONBody, err)
		return
	}

	id, err := externalconfig.R().Create(newExternalConfig)
	if err != nil {
		zap.L().Error("Error while creating the ExternalConfig", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBInsertFailed, err)
		return
	}

	newExternalConfigGet, found, err := externalconfig.R().Get(id)
	if err != nil {
		zap.L().Error("Cannot get externalConfig", zap.String("externalConfigname", newExternalConfig.Name), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("ExternalConfig does not exists after creation", zap.String("externalConfigname", newExternalConfig.Name))
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFoundAfterInsert, err)
		return
	}

	httputil.JSON(w, r, newExternalConfigGet)
}

// PutExternalConfig godoc
//
//	@Summary		Create or remplace an externalConfig definition
//	@Description	Create or remplace an externalConfig definition
//	@Tags			ExternalConfigs
//	@Accept			json
//	@Produce		json
//	@Param			name			path	string					true	"ExternalConfig ID"
//	@Param			externalConfig	body	externalconfig.ExternalConfig	true	"ExternalConfig definition (json)"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	externalconfig.ExternalConfig	"externalConfig"
//	@Failure		400	"Status Bad Request"
//	@Failure		500	"Status"	internal	server	error"
//	@Router			/engine/externalconfigs/{name} [put]
func PutExternalConfig(w http.ResponseWriter, r *http.Request) {

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeConfig, permissions.All, permissions.ActionUpdate)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	id := chi.URLParam(r, "id")
	idExternalConfig, err := strconv.ParseInt(id, 10, 64)

	if err != nil {
		zap.L().Warn("Error on parsing external config id", zap.String("idExternalConfig", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	var newExternalConfig externalconfig.ExternalConfig
	err = json.NewDecoder(r.Body).Decode(&newExternalConfig)
	if err != nil {
		zap.L().Warn("ExternalConfig json decoding", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDecodeJSONBody, err)
		return
	}
	newExternalConfig.Id = idExternalConfig

	err = externalconfig.R().Update(idExternalConfig, newExternalConfig)
	if err != nil {
		zap.L().Error("Error while updating the ExternalConfig", zap.String("idExternalConfig", id), zap.Any("externalConfig", newExternalConfig), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBUpdateFailed, err)
		return
	}

	newExternalConfigGet, found, err := externalconfig.R().Get(idExternalConfig)
	if err != nil {
		zap.L().Error("Cannot get externalConfig", zap.String("externalConfigId", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("ExternalConfig does not exists after update", zap.String("externalConfigId", id))
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFound, err)
		return
	}

	httputil.JSON(w, r, newExternalConfigGet)
}

// DeleteExternalConfig godoc
//
//	@Summary		Delete an externalConfig definition
//	@Description	Delete an externalConfig definition
//	@Tags			ExternalConfigs
//	@Produce		json
//	@Param			name	path	string	true	"ExternalConfig ID"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	"Status OK"
//	@Failure		400	"Status Bad Request"
//	@Router			/engine/externalconfigs/{name} [delete]
func DeleteExternalConfig(w http.ResponseWriter, r *http.Request) {

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeConfig, permissions.All, permissions.ActionDelete)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	id := chi.URLParam(r, "id")
	idExternalConfig, err := strconv.ParseInt(id, 10, 64)

	if err != nil {
		zap.L().Warn("Error on parsing external config id", zap.String("idExternalConfig", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	err = externalconfig.R().Delete(idExternalConfig)

	if err != nil {
		zap.L().Error("Error while deleting the ExternalConfig", zap.String("ExternalConfig ID", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBDeleteFailed, err)
		return
	}

	httputil.OK(w, r)
}

// GetAllOldVersions godoc
//
//	@Summary		Get all old versions of a specific externalConfig
//	@Description	Get all old versions of a specific externalConfig by id
//	@Tags			ExternalConfigs
//	@Produce		json
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Param			id	path	int						true	"ExternalConfig ID"
//	@Success		200	{array}	externalconfig.ExternalConfig	"list of all old versions of the externalConfig"
//	@Failure		400	"bad request"
//	@Failure		500	"internal server error"
//	@Router			/engine/externalconfigs/{id}/alloldversions [get]
func GetAllOldVersions(w http.ResponseWriter, r *http.Request) {

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeConfig, permissions.All, permissions.ActionGet)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	id := chi.URLParam(r, "id")
	idExternalConfig, err := strconv.ParseInt(id, 10, 64)

	if err != nil {
		zap.L().Warn("Error on parsing external config id", zap.String("idExternalConfig", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	oldVersions, err := externalconfig.R().GetAllOldVersions(idExternalConfig)
	if err != nil {
		zap.L().Error("Error getting old versions of externalConfig", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}

	if oldVersions == nil {
		httputil.JSON(w, r, []externalconfig.ExternalConfig{})
		return
	}
	httputil.JSON(w, r, oldVersions)
}
