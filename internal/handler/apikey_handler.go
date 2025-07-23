package handler

import (
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/security/apikey"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/security/permissions"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/utils/httputil"
	ttlcache "github.com/myrteametrics/myrtea-sdk/v5/cache"
	"go.uber.org/zap"
	"net/http"
	"sort"
	"time"
)

type ApikeyHandler struct {
	Cache *ttlcache.Cache
}

func NewApiKeyHandler(cacheDuration time.Duration) *ApikeyHandler {
	return &ApikeyHandler{
		Cache: ttlcache.NewCache(cacheDuration),
	}
}

// GetAPIKeys godoc
//
//	@Id				GetAPIKeys
//
//	@Summary		Get all API keys
//	@Description	Gets a list of all API keys
//	@Tags			APIKeys
//	@Produce		json
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{array}		apikey.APIKey	"list of API keys"
//	@Failure		500	{string}	string			"Internal Server Error"
//	@Router			/admin/security/apikey [get]
func (a *ApikeyHandler) GetAPIKeys(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeAPIKey, permissions.All, permissions.ActionList)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	apiKeys, err := apikey.R().GetAll(userCtx.Login)
	if err != nil {
		zap.L().Error("GetAPIKeys", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}

	sort.SliceStable(apiKeys, func(i, j int) bool {
		return apiKeys[i].Name < apiKeys[j].Name
	})

	httputil.JSON(w, r, apiKeys)
}

// GetAPIKey godoc
//
//	@Id				GetAPIKey
//
//	@Summary		Get an API key
//	@Description	Gets an API key with the specified id
//	@Tags			APIKeys
//	@Produce		json
//	@Param			id	path	string	true	"API key ID"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	apikey.APIKey	"API key"
//	@Failure		400	{string}	string			"Bad Request"
//	@Failure		404	{string}	string			"Not Found"
//	@Failure		500	{string}	string			"Internal Server Error"
//	@Router			/admin/security/apikey/{id} [get]
func (a *ApikeyHandler) GetAPIKey(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	keyID, err := uuid.Parse(id)
	if err != nil {
		zap.L().Warn("Parse API key id", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeAPIKey, keyID.String(), permissions.ActionGet)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	key, found, err := apikey.R().Get(keyID, userCtx.Login)
	if err != nil {
		zap.L().Error("Cannot get API key", zap.String("uuid", keyID.String()), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("API key not found", zap.String("uuid", keyID.String()))
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFound, err)
		return
	}

	httputil.JSON(w, r, key)
}

// ValidateAPIKey godoc
//
//	@Id				ValidateAPIKey
//
//	@Summary		Validate an API key
//	@Description	Validates an API key and returns its associated information
//	@Tags			APIKeys
//	@Accept			json
//	@Produce		json
//	@Param			X-API-Key	header		string			true	"API Key"
//	@Success		200			{object}	apikey.APIKey	"API key information"
//	@Failure		401			{string}	string			"Unauthorized - Invalid API key"
//	@Failure		500			{string}	string			"Internal Server Error"
//	@Router			/engine/security/apikey/validate [get]
func (a *ApikeyHandler) ValidateAPIKey(w http.ResponseWriter, r *http.Request) {
	apiKeyValue := r.Header.Get("X-API-Key")

	if apiKeyValue == "" {
		zap.L().Warn("No API key provided")
		httputil.Error(w, r, httputil.ErrAPIProcessError, errors.New("no API key provided"))
		return
	}

	authKey, found := a.Cache.Get(apiKeyValue)

	if found {
		httputil.JSON(w, r, authKey)
		return
	}

	key, valid, err := apikey.R().Validate(apiKeyValue)
	if err != nil {
		zap.L().Error("Error validating API key", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}

	if !valid {
		zap.L().Warn("Invalid API key")
		httputil.Error(w, r, httputil.ErrAPIProcessError, errors.New("invalid API key"))
		return
	}

	httputil.JSON(w, r, key)
}

// PostAPIKey godoc
//
//	@Id				PostAPIKey
//
//	@Summary		Create a new API key
//	@Description	Add an API key
//	@Tags			APIKeys
//	@Accept			json
//	@Produce		json
//	@Param			apikey	body	apikey.APIKey	true	"API key (json)"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	apikey.APIKey	"API key with value"
//	@Failure		400	{string}	string			"Bad Request"
//	@Failure		500	{string}	string			"Internal Server Error"
//	@Router			/admin/security/apikey [post]
func (a *ApikeyHandler) CreateAPIKey(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeAPIKey, permissions.All, permissions.ActionCreate)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	var key apikey.APIKey
	err := json.NewDecoder(r.Body).Decode(&key)
	if err != nil {
		zap.L().Warn("API key json decode", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDecodeJSONBody, err)
		return
	}

	key.CreatedBy = userCtx.Login
	if err := key.IsValidForCreate(); err != nil {
		zap.L().Warn("API key is not valid", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIResourceInvalid, err)
		return
	}

	key.CreatedBy = userCtx.Login

	apiKey, err := apikey.R().Create(key)
	if err != nil {
		zap.L().Error("PostAPIKey.Create", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBInsertFailed, err)
		return
	}

	httputil.JSON(w, r, apiKey)
}

// PutAPIKey godoc
//
//	@Id				PutAPIKey
//
//	@Summary		Update API key
//	@Description	Updates the API key information
//	@Tags			APIKeys
//	@Accept			json
//	@Produce		json
//	@Param			id		path	string			true	"API key ID"
//	@Param			apikey	body	apikey.APIKey	true	"API key (json)"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	apikey.APIKey	"API key"
//	@Failure		400	{string}	string			"Bad Request"
//	@Failure		500	{string}	string			"Internal Server Error"
//	@Router			/admin/security/apikey/{id} [put]
func (a *ApikeyHandler) PutAPIKey(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	keyID, err := uuid.Parse(id)
	if err != nil {
		zap.L().Warn("Parse API key id", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeAPIKey, keyID.String(), permissions.ActionUpdate)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	var key apikey.APIKey
	err = json.NewDecoder(r.Body).Decode(&key)
	if err != nil {
		zap.L().Warn("API key json decode", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDecodeJSONBody, err)
		return
	}
	key.ID = keyID
	key.CreatedBy = userCtx.Login

	if err := key.IsValidForUpdate(); err != nil {
		zap.L().Warn("API key is not valid", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIResourceInvalid, err)
		return
	}

	err = apikey.R().Update(key, userCtx.Login)
	if err != nil {
		zap.L().Error("PutAPIKey.Update", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBUpdateFailed, err)
		return
	}

	authKey, found, err := apikey.R().Get(keyID, userCtx.Login)
	if err != nil {
		zap.L().Error("Cannot get API key", zap.String("uuid", keyID.String()), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Error("API key not found after update", zap.String("uuid", keyID.String()))
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFoundAfterInsert, err)
		return
	}

	a.Cache.Cleanup()

	httputil.JSON(w, r, authKey)
}

// DeleteAPIKey godoc
//
//	@Id				DeleteAPIKey
//
//	@Summary		Delete API key
//	@Description	Deletes an API key
//	@Tags			APIKeys
//	@Produce		json
//	@Param			id	path	string	true	"API key ID"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{string}	string	"status OK"
//	@Failure		400	{string}	string	"Bad Request"
//	@Failure		500	{string}	string	"Internal Server Error"
//	@Router			/admin/security/apikey/{id} [delete]
func (a *ApikeyHandler) DeleteAPIKey(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	keyID, err := uuid.Parse(id)
	if err != nil {
		zap.L().Warn("Parse API key id", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeAPIKey, keyID.String(), permissions.ActionDelete)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	err = apikey.R().Delete(keyID, userCtx.Login)
	if err != nil {
		zap.L().Error("Cannot delete API key", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBDeleteFailed, err)
		return
	}

	a.Cache.Cleanup()

	httputil.OK(w, r)
}

// DeactivateAPIKey godoc
//
//	@Id				DeactivateAPIKey
//
//	@Summary		Deactivate API key
//	@Description	Deactivates an API key without deleting it
//	@Tags			APIKeys
//	@Produce		json
//	@Param			id	path	string	true	"API key ID"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{string}	string	"status OK"
//	@Failure		400	{string}	string	"Bad Request"
//	@Failure		500	{string}	string	"Internal Server Error"
//	@Router			/admin/security/apikey/{id}/deactivate [post]
func (a *ApikeyHandler) DeactivateAPIKey(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	keyID, err := uuid.Parse(id)
	if err != nil {
		zap.L().Warn("Parse API key id", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeAPIKey, keyID.String(), permissions.ActionUpdate)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	err = apikey.R().Deactivate(keyID, userCtx.Login)
	if err != nil {
		zap.L().Error("Cannot deactivate API key", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBUpdateFailed, err)
		return
	}

	a.Cache.Cleanup()

	httputil.OK(w, r)
}

// GetAPIKeysForRole godoc
//
//	@Id				GetAPIKeysForRole
//
//	@Summary		Get API keys for a specific role
//	@Description	Gets a list of all API keys associated with a specific role
//	@Tags			APIKeys
//	@Produce		json
//	@Param			roleId	path	string	true	"Role ID"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{array}		apikey.APIKey	"list of API keys"
//	@Failure		400	{string}	string			"Bad Request"
//	@Failure		500	{string}	string			"Internal Server Error"
//	@Router			/admin/security/roles/{roleId}/apikey [get]
func (a *ApikeyHandler) GetAPIKeysForRole(w http.ResponseWriter, r *http.Request) {
	roleID := chi.URLParam(r, "roleId")
	roleUUID, err := uuid.Parse(roleID)
	if err != nil {
		zap.L().Warn("Parse role id", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeAPIKey, permissions.All, permissions.ActionList)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	apiKeys, err := apikey.R().GetAllForRole(roleUUID, userCtx.Login)
	if err != nil {
		zap.L().Error("GetAPIKeysForRole", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}

	sort.SliceStable(apiKeys, func(i, j int) bool {
		return apiKeys[i].Name < apiKeys[j].Name
	})

	httputil.JSON(w, r, apiKeys)
}
