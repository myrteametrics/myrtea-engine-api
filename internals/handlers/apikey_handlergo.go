package handlers

import (
	"encoding/json"
	"errors"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/security/apikey"
	"net/http"
	"sort"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/handlers/render"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/security/permissions"
	"go.uber.org/zap"
)

// GetAPIKeys godoc
// @Summary Get all API keys
// @Description Gets a list of all API keys
// @Tags APIKeys
// @Produce json
// @Security Bearer
// @Success 200 {array} apikey.APIKey "list of API keys"
// @Failure 500 {string} string "Internal Server Error"
// @Router /admin/security/apikeys [get]
func GetAPIKeys(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeAPIKey, permissions.All, permissions.ActionList)) {
		render.Error(w, r, render.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	apiKeys, err := apikey.R().GetAll()
	if err != nil {
		zap.L().Error("GetAPIKeys", zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}

	sort.SliceStable(apiKeys, func(i, j int) bool {
		return apiKeys[i].Name < apiKeys[j].Name
	})

	render.JSON(w, r, apiKeys)
}

// GetAPIKey godoc
// @Summary Get an API key
// @Description Gets an API key with the specified id
// @Tags APIKeys
// @Produce json
// @Param id path string true "API key ID"
// @Security Bearer
// @Success 200 {object} apikey.APIKey "API key"
// @Failure 400 {string} string "Bad Request"
// @Failure 404 {string} string "Not Found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /admin/security/apikey/{id} [get]
func GetAPIKey(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	keyID, err := uuid.Parse(id)
	if err != nil {
		zap.L().Warn("Parse API key id", zap.Error(err))
		render.Error(w, r, render.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeAPIKey, keyID.String(), permissions.ActionGet)) {
		render.Error(w, r, render.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	key, found, err := apikey.R().Get(keyID)
	if err != nil {
		zap.L().Error("Cannot get API key", zap.String("uuid", keyID.String()), zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("API key not found", zap.String("uuid", keyID.String()))
		render.Error(w, r, render.ErrAPIDBResourceNotFound, err)
		return
	}

	render.JSON(w, r, key)
}

// ValidateAPIKey godoc
// @Summary Validate an API key
// @Description Validates an API key and returns its associated information
// @Tags APIKeys
// @Accept json
// @Produce json
// @Param X-API-Key header string true "API Key"
// @Success 200 {object} apikey.APIKey "API key information"
// @Failure 401 {string} string "Unauthorized - Invalid API key"
// @Failure 500 {string} string "Internal Server Error"
// @Router /engine/security/apikey/validate [get]
func ValidateAPIKey(w http.ResponseWriter, r *http.Request) {
	apiKeyValue := r.Header.Get("X-API-Key")

	if apiKeyValue == "" {
		zap.L().Warn("No API key provided")
		render.Error(w, r, render.ErrAPIProcessError, errors.New("no API key provided"))
		return
	}

	key, valid, err := apikey.R().Validate(apiKeyValue)
	if err != nil {
		zap.L().Error("Error validating API key", zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}

	if !valid {
		zap.L().Warn("Invalid API key")
		render.Error(w, r, render.ErrAPIProcessError, errors.New("invalid API key"))
		return
	}

	render.JSON(w, r, key)
}

// PostAPIKey godoc
// @Summary Create a new API key
// @Description Add an API key
// @Tags APIKeys
// @Accept json
// @Produce json
// @Param apikey body apikey.APIKey true "API key (json)"
// @Security Bearer
// @Success 200 {object} apikey.APIKeyWithValue "API key with value"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /admin/security/apikeys [post]
func CreateAPIKey(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeAPIKey, permissions.All, permissions.ActionCreate)) {
		render.Error(w, r, render.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	var key apikey.APIKey
	err := json.NewDecoder(r.Body).Decode(&key)
	if err != nil {
		zap.L().Warn("API key json decode", zap.Error(err))
		render.Error(w, r, render.ErrAPIDecodeJSONBody, err)
		return
	}

	key.CreatedBy = userCtx.Login
	if err := key.IsValidForCreate(); err != nil {
		zap.L().Warn("API key is not valid", zap.Error(err))
		render.Error(w, r, render.ErrAPIResourceInvalid, err)
		return
	}

	// Set the creator to the current user
	key.CreatedBy = userCtx.Login

	// Create the API key
	keyWithValue, err := apikey.R().Create(key)
	if err != nil {
		zap.L().Error("PostAPIKey.Create", zap.Error(err))
		render.Error(w, r, render.ErrAPIDBInsertFailed, err)
		return
	}

	render.JSON(w, r, keyWithValue)
}

// PutAPIKey godoc
// @Summary Update API key
// @Description Updates the API key information
// @Tags APIKeys
// @Accept json
// @Produce json
// @Param id path string true "API key ID"
// @Param apikey body apikey.APIKey true "API key (json)"
// @Security Bearer
// @Success 200 {object} apikey.APIKey "API key"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /admin/security/apikey/{id} [put]
func PutAPIKey(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	keyID, err := uuid.Parse(id)
	if err != nil {
		zap.L().Warn("Parse API key id", zap.Error(err))
		render.Error(w, r, render.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeAPIKey, keyID.String(), permissions.ActionUpdate)) {
		render.Error(w, r, render.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	var key apikey.APIKey
	err = json.NewDecoder(r.Body).Decode(&key)
	if err != nil {
		zap.L().Warn("API key json decode", zap.Error(err))
		render.Error(w, r, render.ErrAPIDecodeJSONBody, err)
		return
	}
	key.ID = keyID
	key.CreatedBy = userCtx.Login

	if err := key.IsValidForUpdate(); err != nil {
		zap.L().Warn("API key is not valid", zap.Error(err))
		render.Error(w, r, render.ErrAPIResourceInvalid, err)
		return
	}

	err = apikey.R().Update(key)
	if err != nil {
		zap.L().Error("PutAPIKey.Update", zap.Error(err))
		render.Error(w, r, render.ErrAPIDBUpdateFailed, err)
		return
	}

	key, found, err := apikey.R().Get(keyID)
	if err != nil {
		zap.L().Error("Cannot get API key", zap.String("uuid", keyID.String()), zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Error("API key not found after update", zap.String("uuid", keyID.String()))
		render.Error(w, r, render.ErrAPIDBResourceNotFoundAfterInsert, err)
		return
	}

	render.JSON(w, r, key)
}

// DeleteAPIKey godoc
// @Summary Delete API key
// @Description Deletes an API key
// @Tags APIKeys
// @Produce json
// @Param id path string true "API key ID"
// @Security Bearer
// @Success 200 {string} string "status OK"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /admin/security/apikey/{id} [delete]
func DeleteAPIKey(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	keyID, err := uuid.Parse(id)
	if err != nil {
		zap.L().Warn("Parse API key id", zap.Error(err))
		render.Error(w, r, render.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeAPIKey, keyID.String(), permissions.ActionDelete)) {
		render.Error(w, r, render.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	err = apikey.R().Delete(keyID)
	if err != nil {
		zap.L().Error("Cannot delete API key", zap.Error(err))
		render.Error(w, r, render.ErrAPIDBDeleteFailed, err)
		return
	}

	render.OK(w, r)
}

// DeactivateAPIKey godoc
// @Summary Deactivate API key
// @Description Deactivates an API key without deleting it
// @Tags APIKeys
// @Produce json
// @Param id path string true "API key ID"
// @Security Bearer
// @Success 200 {string} string "status OK"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /admin/security/apikey/{id}/deactivate [post]
func DeactivateAPIKey(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	keyID, err := uuid.Parse(id)
	if err != nil {
		zap.L().Warn("Parse API key id", zap.Error(err))
		render.Error(w, r, render.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeAPIKey, keyID.String(), permissions.ActionUpdate)) {
		render.Error(w, r, render.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	err = apikey.R().Deactivate(keyID)
	if err != nil {
		zap.L().Error("Cannot deactivate API key", zap.Error(err))
		render.Error(w, r, render.ErrAPIDBUpdateFailed, err)
		return
	}

	render.OK(w, r)
}

// GetAPIKeysForRole godoc
// @Summary Get API keys for a specific role
// @Description Gets a list of all API keys associated with a specific role
// @Tags APIKeys
// @Produce json
// @Param roleId path string true "Role ID"
// @Security Bearer
// @Success 200 {array} apikey.APIKey "list of API keys"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /admin/security/roles/{roleId}/apikeys [get]
func GetAPIKeysForRole(w http.ResponseWriter, r *http.Request) {
	roleID := chi.URLParam(r, "roleId")
	roleUUID, err := uuid.Parse(roleID)
	if err != nil {
		zap.L().Warn("Parse role id", zap.Error(err))
		render.Error(w, r, render.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeAPIKey, permissions.All, permissions.ActionList)) {
		render.Error(w, r, render.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	apiKeys, err := apikey.R().GetAllForRole(roleUUID)
	if err != nil {
		zap.L().Error("GetAPIKeysForRole", zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}

	sort.SliceStable(apiKeys, func(i, j int) bool {
		return apiKeys[i].Name < apiKeys[j].Name
	})

	render.JSON(w, r, apiKeys)
}
