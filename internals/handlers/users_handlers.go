package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"sort"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/handlers/render"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/security/permissions"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/security/users"
	"go.uber.org/zap"
)

// GetUserSelf godoc
// @Summary Get an user
// @Description Gets un user with the specified id.
// @Tags Users
// @Produce json
// @Security Bearer
// @Success 200 {string} string "status OK"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /engine/security/myself [get]
func GetUserSelf(w http.ResponseWriter, r *http.Request) {
	userCtx, found := GetUserFromContext(r)
	if !found {
		zap.L().Warn("No context user provided")
		render.Error(w, r, render.ErrAPIDBResourceNotFound, errors.New("no context user provided"))
		return
	}

	render.JSON(w, r, userCtx)
}

// GetUsers godoc
// @Summary Get all user users
// @Description Gets a list of all user users.
// @Tags Users
// @Produce json
// @Security Bearer
// @Success 200 {array} users.User "list of users"
// @Failure 500 {string} string "Internal Server Error"
// @Router /admin/security/users [get]
func GetUsers(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeUser, permissions.All, permissions.ActionList)) {
		render.Error(w, r, render.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	usersSlice, err := users.R().GetAll()
	if err != nil {
		zap.L().Error("GetUsers", zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}

	sort.SliceStable(usersSlice, func(i, j int) bool {
		return usersSlice[i].LastName < usersSlice[j].LastName
	})

	render.JSON(w, r, usersSlice)
}

// GetUser godoc
// @Summary Get an user user
// @Description Gets an user user with the specified id
// @Tags Users
// @Produce json
// @Param id path string true "user ID"
// @Security Bearer
// @Success 200 {object} users.User "user"
// @Failure 400 {string} string "Bad Request"
// @Failure 404 {string} string "Not Found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /admin/security/users/{id} [get]
func GetUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userID, err := uuid.Parse(id)
	if err != nil {
		zap.L().Warn("Parse user id", zap.Error(err))
		render.Error(w, r, render.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeUser, userID.String(), permissions.ActionGet)) {
		render.Error(w, r, render.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	user, found, err := users.R().Get(userID)
	if err != nil {
		zap.L().Error("Cannot get user", zap.String("uuid", userID.String()), zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("User not found", zap.String("uuid", userID.String()))
		render.Error(w, r, render.ErrAPIDBResourceNotFound, err)
		return
	}

	render.JSON(w, r, user)
}

// ValidateUser godoc
// @Summary Validate a new user definition
// @Description Validate a new user definition
// @Tags Users
// @Accept json
// @Produce json
// @Param user body users.User true "user (json)"
// @Security Bearer
// @Success 200 {object} users.User "user"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /admin/security/users/validate [post]
func ValidateUser(w http.ResponseWriter, r *http.Request) {
	var newUser users.UserWithPassword
	err := json.NewDecoder(r.Body).Decode(&newUser)
	if err != nil {
		zap.L().Warn("User json decode", zap.Error(err))
		render.Error(w, r, render.ErrAPIDecodeJSONBody, err)
		return
	}

	if ok, err := newUser.IsValid(); !ok {
		zap.L().Warn("User is not valid", zap.Error(err))
		render.Error(w, r, render.ErrAPIResourceInvalid, err)
		return
	}

	render.JSON(w, r, newUser)
}

// PostUser godoc
// @Summary Create a new user
// @Description Add an user user to the user users
// @Tags Users
// @Accept json
// @Produce json
// @Param user body users.User true "user (json)"
// @Security Bearer
// @Success 200 {object} users.User "user"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /admin/security/users [post]
func PostUser(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeUser, permissions.All, permissions.ActionCreate)) {
		render.Error(w, r, render.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	var user users.UserWithPassword
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		zap.L().Warn("User json decode", zap.Error(err))
		render.Error(w, r, render.ErrAPIDecodeJSONBody, err)
		return
	}

	if ok, err := user.IsValid(); !ok {
		zap.L().Warn("User is not valid", zap.Error(err))
		render.Error(w, r, render.ErrAPIResourceInvalid, err)
		return
	}

	userID, err := users.R().Create(user)
	if err != nil {
		zap.L().Error("PostUser.Create", zap.Error(err))
		render.Error(w, r, render.ErrAPIDBInsertFailed, err)
		return
	}

	newUser, found, err := users.R().Get(userID)
	if err != nil {
		zap.L().Error("Cannot get user", zap.String("uuid", userID.String()), zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Error("User not found after creation", zap.String("uuid", userID.String()))
		render.Error(w, r, render.ErrAPIDBResourceNotFoundAfterInsert, err)
		return
	}

	render.JSON(w, r, newUser)
}

// PutUser godoc
// @Summary Update user
// @Description Updates the user user information concerning the user user with id
// @Tags Users
// @Accept json
// @Produce json
// @Param id path string true "user ID"
// @Param user body users.User true "user (json)"
// @Security Bearer
// @Success 200 {object} users.User "user"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /admin/security/users/{id} [put]
func PutUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userID, err := uuid.Parse(id)
	if err != nil {
		zap.L().Warn("Parse user id", zap.Error(err))
		render.Error(w, r, render.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeUser, userID.String(), permissions.ActionUpdate)) {
		render.Error(w, r, render.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	var user users.User
	err = json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		zap.L().Warn("User json decode", zap.Error(err))
		render.Error(w, r, render.ErrAPIDecodeJSONBody, err)
		return
	}
	user.ID = userID

	if ok, err := user.IsValid(); !ok {
		zap.L().Warn("User is not valid", zap.Error(err))
		render.Error(w, r, render.ErrAPIResourceInvalid, err)
		return
	}

	err = users.R().Update(user)
	if err != nil {
		zap.L().Error("PutUser.Update", zap.Error(err))
		render.Error(w, r, render.ErrAPIDBUpdateFailed, err)
		return
	}

	user, found, err := users.R().Get(userID)
	if err != nil {
		zap.L().Error("Cannot get user", zap.String("uuid", userID.String()), zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Error("User not found after creation", zap.String("uuid", userID.String()))
		render.Error(w, r, render.ErrAPIDBResourceNotFoundAfterInsert, err)
		return
	}

	render.JSON(w, r, user)
}

// DeleteUser godoc
// @Summary Delete user
// @Description Deletes an user user
// @Tags Users
// @Produce json
// @Param id path string true "user ID"
// @Security Bearer
// @Success 200 {string} string "status OK"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /admin/security/users/{id} [delete]
func DeleteUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userID, err := uuid.Parse(id)
	if err != nil {
		zap.L().Warn("Parse user id", zap.Error(err))
		render.Error(w, r, render.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeUser, userID.String(), permissions.ActionDelete)) {
		render.Error(w, r, render.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	err = users.R().Delete(userID)
	if err != nil {
		zap.L().Error("Cannot delete user", zap.Error(err))
		render.Error(w, r, render.ErrAPIDBDeleteFailed, err)
		return
	}

	render.OK(w, r)
}

// SetUserRoles godoc
// @Summary Set roles on a user
// @Description Set roles on a user
// @Tags Users
// @Accept json
// @Produce json
// @Param id path string true "user ID"
// @Param user body []string true "List of roles UUIDs"
// @Security Bearer
// @Success 200 {object} users.User "user"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /admin/security/users/{id}/roles [put]
func SetUserRoles(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userID, err := uuid.Parse(id)
	if err != nil {
		zap.L().Warn("Parse user id", zap.Error(err))
		render.Error(w, r, render.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeUser, userID.String(), permissions.ActionUpdate)) {
		render.Error(w, r, render.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	var rawRoleUUIDs []string
	err = json.NewDecoder(r.Body).Decode(&rawRoleUUIDs)
	if err != nil {
		zap.L().Warn("Invalid UUID", zap.Error(err))
		render.Error(w, r, render.ErrAPIDecodeJSONBody, err)
		return
	}

	roleUUIDs := make([]uuid.UUID, 0)

	for _, rawRoleUUID := range rawRoleUUIDs {
		roleUUID, err := uuid.Parse(rawRoleUUID)
		if err != nil {
			zap.L().Warn("Invalid UUID", zap.Error(err))
			render.Error(w, r, render.ErrAPIDecodeJSONBody, err)
			return
		}
		roleUUIDs = append(roleUUIDs, roleUUID)
	}

	err = users.R().SetUserRoles(userID, roleUUIDs)
	if err != nil {
		zap.L().Error("PutUser.Update", zap.Error(err))
		render.Error(w, r, render.ErrAPIDBUpdateFailed, err)
		return
	}

	render.OK(w, r)
}
