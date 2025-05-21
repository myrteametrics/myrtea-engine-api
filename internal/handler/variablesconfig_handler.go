package handler

import (
	"encoding/json"
	"errors"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/security/permissions"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/utils/httputil"
	"github.com/myrteametrics/myrtea-sdk/v5/handlers/render"
	"github.com/myrteametrics/myrtea-sdk/v5/repositories/variablesconfig"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// GetVariablesConfig godoc
//
//	@Summary		Get all Variables Config definitions
//	@Description	Get all VariableConfig definitions
//	@Tags			VariablesConfig
//	@Produce		json
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{array}	variablesconfig.VariablesConfig	"list of all Variables with it's config"
//	@Failure		500	"internal server error"
//	@Router			/engine/variablesconfig [get]
func GetVariablesConfig(w http.ResponseWriter, r *http.Request) {

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeConfig, permissions.All, permissions.ActionList)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	VariablesConfig, err := variablesconfig.R().GetAll()
	if err != nil {
		zap.L().Error("Error getting VariableConfigs", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}

	httputil.JSON(w, r, VariablesConfig)
}

// GetVariablesConfigByScope godoc
// @Summary Get all Variables Config definitions by scope
// @Description Get all VariableConfig definitions that match with scope value or scope 'global' by default
// @Tags VariablesConfig
// @Produce json
// @Security Bearer
// @Success 200 {array} variablesconfig.VariablesConfig "list of all Variables with it's config"
// @Failure 500 "internal server error"
// @Router /engine/variablesconfig/{scope} [get]
func GetVariablesConfigByScope(w http.ResponseWriter, r *http.Request) {

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeConfig, permissions.All, permissions.ActionList)) {
		render.Error(w, r, render.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	scope := chi.URLParam(r, "scope")
	VariablesConfig, err := variablesconfig.R().GetAllByScope(scope)
	if err != nil {
		zap.L().Error("Error getting VariableConfigs filtered by scope", zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}

	render.JSON(w, r, VariablesConfig)
}

// GetVariableConfig godoc
//
//	@Summary		Get an Variable Config definition
//	@Description	Get an Variable Config definition
//	@Tags			VariablesConfig
//	@Produce		json
//	@Param			id	path	string	true	"Variable Config ID"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	variablesconfig.VariablesConfig	"One Variable Config getted by id"
//	@Failure		400	"Status Bad Request"
//	@Router			/engine/variablesconfig/{id} [get]
func GetVariableConfig(w http.ResponseWriter, r *http.Request) {

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeConfig, permissions.All, permissions.ActionGet)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	id := chi.URLParam(r, "id")
	idVariableConfig, err := strconv.ParseInt(id, 10, 64)

	if err != nil {
		zap.L().Warn("Error on parsing variable config id", zap.String("idVariableConfig", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	a, found, err := variablesconfig.R().Get(idVariableConfig)
	if err != nil {
		zap.L().Error("Cannot get Variable Config", zap.String("Variable Config Id", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}

	if !found {
		zap.L().Warn("VariableConfig does not exists", zap.String("VariableConfigId", id))
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFound, err)
		return
	}

	httputil.JSON(w, r, a)
}

// GetVariableConfigByKey godoc
//
//	@Summary		Get an Variable Config definition
//	@Description	Get an Variable Config definition
//	@Tags			VariablesConfig
//	@Produce		json
//	@Param			key	path	string	true	"Variable Config key"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	variablesconfig.VariablesConfig	" One Variable Config getted by key"
//	@Failure		400	"Status Bad Request"
//	@Router			/engine/variablesconfig/key/{key} [get]
func GetVariableConfigByKey(w http.ResponseWriter, r *http.Request) {

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeConfig, permissions.All, permissions.ActionGet)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	key := chi.URLParam(r, "key")

	a, found, err := variablesconfig.R().GetByKey(key)
	if err != nil {
		zap.L().Error("Cannot get VariableConfig", zap.String("Variable Config key", key), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}

	if !found {
		zap.L().Warn("VariableConfig does not exists", zap.String("Variable Config key", key))
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFound, err)
		return
	}

	httputil.JSON(w, r, a)
}

// PostVariableConfig godoc
//
//	@Summary		Create a new VariableConfig definition
//	@Description	Create a new VariableConfig definition
//	@Tags			VariablesConfig
//	@Accept			json
//	@Produce		json
//	@Param			VariableConfig	body	variablesconfig.VariablesConfig	true	"VariableConfig definition (json)"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	variablesconfig.VariablesConfig	"Ppost VariableConfig"
//	@Failure		400	"Status Bad Request"
//	@Failure		500	"Status"	internal	server	error"
//	@Router			/engine/variablesconfig [post]
func PostVariableConfig(w http.ResponseWriter, r *http.Request) {

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeConfig, permissions.All, permissions.ActionCreate)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	var newVariableConfig variablesconfig.VariablesConfig
	err := json.NewDecoder(r.Body).Decode(&newVariableConfig)
	if err != nil {
		zap.L().Warn("Variable Config json decoding", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDecodeJSONBody, err)
		return
	}

	id, err := variablesconfig.R().Create(newVariableConfig)
	if err != nil {
		zap.L().Error("Error while creating the VariableConfig", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBInsertFailed, err)
		return
	}

	newVariableConfigGet, found, err := variablesconfig.R().Get(id)
	if err != nil {
		zap.L().Error("Cannot get Variable Config", zap.String("VariableConfig key ", newVariableConfig.Key), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("VariableConfig does not exists after creation", zap.String("VariableConfig key ", newVariableConfig.Key))
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFoundAfterInsert, err)
		return
	}

	httputil.JSON(w, r, newVariableConfigGet)
}

// PutVariableConfig godoc
//
//	@Summary		update an Variable Config definition
//	@Description	update an Variable Config definition
//	@Tags			VariablesConfig
//	@Accept			json
//	@Produce		json
//	@Param			id				path	string					true	"VariableConfig ID"
//	@Param			VariableConfig	body	variablesconfig.VariablesConfig	true	"VariableConfig definition (json)"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	variablesconfig.VariablesConfig	"VariableConfig"
//	@Failure		400	"Status Bad Request"
//	@Failure		500	"Status"	internal	server	error"
//	@Router			/engine/variablesconfig/{id} [put]
func PutVariableConfig(w http.ResponseWriter, r *http.Request) {

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeConfig, permissions.All, permissions.ActionUpdate)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	id := chi.URLParam(r, "id")
	idVariableConfig, err := strconv.ParseInt(id, 10, 64)

	if err != nil {
		zap.L().Warn("Error on parsing variable config id", zap.String("idVariableConfig", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	var newVariableConfig variablesconfig.VariablesConfig
	err = json.NewDecoder(r.Body).Decode(&newVariableConfig)
	if err != nil {
		zap.L().Warn("VariableConfig json decoding", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDecodeJSONBody, err)
		return
	}
	newVariableConfig.Id = idVariableConfig

	err = variablesconfig.R().Update(idVariableConfig, newVariableConfig)
	if err != nil {
		zap.L().Error("Error while updating the Variable Config", zap.String("idVariableConfig", id), zap.Any("Variable Config", newVariableConfig), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBUpdateFailed, err)
		return
	}

	newVariableConfigGet, found, err := variablesconfig.R().Get(idVariableConfig)
	if err != nil {
		zap.L().Error("Cannot get VariableConfig", zap.String("VariableConfigId", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("VariableConfig does not exists after update", zap.String("VariableConfigId", id))
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFound, err)
		return
	}

	httputil.JSON(w, r, newVariableConfigGet)
}

// DeleteVariableConfig godoc
//
//	@Summary		Delete an Variable Config definition
//	@Description	Delete an Variable Config definition
//	@Tags			VariablesConfig
//	@Produce		json
//	@Param			id	path	string	true	"VariableConfig ID"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	"Status OK"
//	@Failure		400	"Status Bad Request"
//	@Router			/engine/variablesconfig/{id} [delete]
func DeleteVariableConfig(w http.ResponseWriter, r *http.Request) {

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeConfig, permissions.All, permissions.ActionUpdate)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	id := chi.URLParam(r, "id")
	idVariableConfig, err := strconv.ParseInt(id, 10, 64)

	if err != nil {
		zap.L().Warn("Error on parsing variable config id", zap.String("idVariableConfig", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	err = variablesconfig.R().Delete(idVariableConfig)

	if err != nil {
		zap.L().Error("Error while deleting the VariableConfig", zap.String("VariableConfig ID", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBDeleteFailed, err)
		return
	}

	httputil.OK(w, r)
}
