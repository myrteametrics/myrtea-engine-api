package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"sort"

	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/security/permissions"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/security/users"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/utils/httputil"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// GetUserSelf godoc
//
//	@Id				GetUserSelf
//
//	@Summary		Get an user
//	@Description	Gets un user with the specified id.
//	@Tags			Users
//	@Produce		json
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	users.UserWithPermissions	"status OK"
//	@Failure		400	{object}	httputil.APIError			"Bad Request"
//	@Failure		500	{object}	httputil.APIError			"Internal Server Error"
//	@Router			/engine/security/myself [get]
func GetUserSelf(w http.ResponseWriter, r *http.Request) {
	userCtx, found := GetUserFromContext(r)
	if !found {
		zap.L().Warn("No context user provided")
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFound, errors.New("no context user provided"))
		return
	}

	httputil.JSON(w, r, userCtx)
}

// GetUsers godoc
//
//	@Id				GetUsers
//
//	@Summary		Get all user users
//	@Description	Gets a list of all user users.
//	@Tags			Users
//	@Produce		json
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{array}		users.User			"list of users"
//	@Failure		500	{object}	httputil.APIError	"Internal Server Error"
//	@Router			/admin/security/users [get]
func GetUsers(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeUser, permissions.All, permissions.ActionList)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	usersSlice, err := users.R().GetAll()
	if err != nil {
		zap.L().Error("GetUsers", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}

	sort.SliceStable(usersSlice, func(i, j int) bool {
		return usersSlice[i].LastName < usersSlice[j].LastName
	})

	httputil.JSON(w, r, usersSlice)
}

// GetUser godoc
//
//	@Id				GetUser
//
//	@Summary		Get an user user
//	@Description	Gets an user user with the specified id
//	@Tags			Users
//	@Produce		json
//	@Param			id	path	string	true	"user ID"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	users.User			"user"
//	@Failure		400	{object}	httputil.APIError	"Bad Request"
//	@Failure		404	{object}	httputil.APIError	"Status Not Found"
//	@Failure		500	{object}	httputil.APIError	"Internal Server Error"
//	@Router			/admin/security/users/{id} [get]
func GetUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userID, err := uuid.Parse(id)
	if err != nil {
		zap.L().Warn("Parse user id", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeUser, userID.String(), permissions.ActionGet)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	user, found, err := users.R().Get(userID)
	if err != nil {
		zap.L().Error("Cannot get user", zap.String("uuid", userID.String()), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("User not found", zap.String("uuid", userID.String()))
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFound, err)
		return
	}

	httputil.JSON(w, r, user)
}

// ValidateUser godoc
//
//	@Id				ValidateUser
//
//	@Summary		Validate a new user definition
//	@Description	Validate a new user definition
//	@Tags			Users
//	@Accept			json
//	@Produce		json
//	@Param			user	body	users.User	true	"user (json)"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	users.User			"user"
//	@Failure		400	{object}	httputil.APIError	"Bad Request"
//	@Failure		500	{object}	httputil.APIError	"Internal Server Error"
//	@Router			/admin/security/users/validate [post]
func ValidateUser(w http.ResponseWriter, r *http.Request) {
	var newUser users.UserWithPassword
	err := json.NewDecoder(r.Body).Decode(&newUser)
	if err != nil {
		zap.L().Warn("User json decode", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDecodeJSONBody, err)
		return
	}

	if ok, err := newUser.IsValid(); !ok {
		zap.L().Warn("User is not valid", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIResourceInvalid, err)
		return
	}

	httputil.JSON(w, r, newUser)
}

// PostUser godoc
//
//	@Id				PostUser
//
//	@Summary		Create a new user
//	@Description	Add an user user to the user users
//	@Tags			Users
//	@Accept			json
//	@Produce		json
//	@Param			user	body	users.User	true	"user (json)"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	users.User			"user"
//	@Failure		400	{object}	httputil.APIError	"Bad Request"
//	@Failure		500	{object}	httputil.APIError	"Internal Server Error"
//	@Router			/admin/security/users [post]
func PostUser(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeUser, permissions.All, permissions.ActionCreate)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	var user users.UserWithPassword
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		zap.L().Warn("User json decode", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDecodeJSONBody, err)
		return
	}

	if ok, err := user.IsValid(); !ok {
		zap.L().Warn("User is not valid", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIResourceInvalid, err)
		return
	}

	userID, err := users.R().Create(user)
	if err != nil {
		zap.L().Error("PostUser.Create", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBInsertFailed, err)
		return
	}

	newUser, found, err := users.R().Get(userID)
	if err != nil {
		zap.L().Error("Cannot get user", zap.String("uuid", userID.String()), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Error("User not found after creation", zap.String("uuid", userID.String()))
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFoundAfterInsert, err)
		return
	}

	httputil.JSON(w, r, newUser)
}

// PutUser godoc
//
//	@Id				PutUser
//
//	@Summary		Update user
//	@Description	Updates the user user information concerning the user user with id
//	@Tags			Users
//	@Accept			json
//	@Produce		json
//	@Param			id		path	string		true	"user ID"
//	@Param			user	body	users.User	true	"user (json)"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	users.User			"user"
//	@Failure		400	{object}	httputil.APIError	"Bad Request"
//	@Failure		500	{object}	httputil.APIError	"Internal Server Error"
//	@Router			/admin/security/users/{id} [put]
func PutUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userID, err := uuid.Parse(id)
	if err != nil {
		zap.L().Warn("Parse user id", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeUser, userID.String(), permissions.ActionUpdate)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	var user users.User
	err = json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		zap.L().Warn("User json decode", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDecodeJSONBody, err)
		return
	}
	user.ID = userID

	if ok, err := user.IsValid(); !ok {
		zap.L().Warn("User is not valid", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIResourceInvalid, err)
		return
	}

	err = users.R().Update(user)
	if err != nil {
		zap.L().Error("PutUser.Update", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBUpdateFailed, err)
		return
	}

	user, found, err := users.R().Get(userID)
	if err != nil {
		zap.L().Error("Cannot get user", zap.String("uuid", userID.String()), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Error("User not found after creation", zap.String("uuid", userID.String()))
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFoundAfterInsert, err)
		return
	}

	httputil.JSON(w, r, user)
}

// DeleteUser godoc
//
//	@Id				DeleteUser
//
//	@Summary		Delete user
//	@Description	Deletes an user user
//	@Tags			Users
//	@Produce		json
//	@Param			id	path	string	true	"user ID"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{string}	string				"status OK"
//	@Failure		400	{object}	httputil.APIError	"Bad Request"
//	@Failure		500	{object}	httputil.APIError	"Internal Server Error"
//	@Router			/admin/security/users/{id} [delete]
func DeleteUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userID, err := uuid.Parse(id)
	if err != nil {
		zap.L().Warn("Parse user id", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeUser, userID.String(), permissions.ActionDelete)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	err = users.R().Delete(userID)
	if err != nil {
		zap.L().Error("Cannot delete user", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBDeleteFailed, err)
		return
	}

	httputil.OK(w, r)
}

// SetUserRoles godoc
//
//	@Id				SetUserRoles
//
//	@Summary		Set roles on a user
//	@Description	Set roles on a user
//	@Tags			Users
//	@Accept			json
//	@Produce		json
//	@Param			id		path	string		true	"user ID"
//	@Param			user	body	[]string	true	"List of roles UUIDs"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	users.User			"user"
//	@Failure		400	{object}	httputil.APIError	"Bad Request"
//	@Failure		500	{object}	httputil.APIError	"Internal Server Error"
//	@Router			/admin/security/users/{id}/roles [put]
func SetUserRoles(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userID, err := uuid.Parse(id)
	if err != nil {
		zap.L().Warn("Parse user id", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeUser, userID.String(), permissions.ActionUpdate)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	var rawRoleUUIDs []string
	err = json.NewDecoder(r.Body).Decode(&rawRoleUUIDs)
	if err != nil {
		zap.L().Warn("Invalid UUID", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDecodeJSONBody, err)
		return
	}

	roleUUIDs := make([]uuid.UUID, 0)

	for _, rawRoleUUID := range rawRoleUUIDs {
		roleUUID, err := uuid.Parse(rawRoleUUID)
		if err != nil {
			zap.L().Warn("Invalid UUID", zap.Error(err))
			httputil.Error(w, r, httputil.ErrAPIDecodeJSONBody, err)
			return
		}
		roleUUIDs = append(roleUUIDs, roleUUID)
	}

	err = users.R().SetUserRoles(userID, roleUUIDs)
	if err != nil {
		zap.L().Error("PutUser.Update", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBUpdateFailed, err)
		return
	}

	httputil.OK(w, r)
}
