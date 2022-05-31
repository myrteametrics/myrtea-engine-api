package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/groups"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/handlers/render"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/models"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/users"
	"github.com/myrteametrics/myrtea-sdk/v4/security"
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
	user := r.Context().Value(models.ContextKeyUser)
	if user == nil {
		zap.L().Warn("No context user provided")
		render.Error(w, r, render.ErrAPIDBResourceNotFound, errors.New("No context user provided"))
		return
	}

	userWithGroups := user.(groups.UserWithGroups)
	render.JSON(w, r, userWithGroups)
}

// GetUsers godoc
// @Summary Get users
// @Description Gets a list of all users.
// @Tags Users
// @Produce json
// @Security Bearer
// @Success 200 {string} string "Status OK"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /admin/security/users [get]
func GetUsers(w http.ResponseWriter, r *http.Request) {
	users, err := users.R().GetAll()
	if err != nil {
		zap.L().Warn("GetUsers", zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}

	usersSlice := make([]security.User, 0)
	for _, user := range users {
		usersSlice = append(usersSlice, user)
	}

	sort.SliceStable(usersSlice, func(i, j int) bool {
		return usersSlice[i].ID < usersSlice[j].ID
	})

	render.JSON(w, r, usersSlice)
}

// GetUser godoc
// @Summary Get an user
// @Description Gets un user with the specified id.
// @Tags Users
// @Produce json
// @Param id path string true "user ID"
// @Security Bearer
// @Success 200 {string} string "status OK"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /admin/security/users/{id} [get]
func GetUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("GetUser.Getid", zap.Error(err))
		render.Error(w, r, render.ErrAPIParsingInteger, err)
	}

	user, found, err := users.R().Get(userID)
	if err != nil {
		zap.L().Error("Get user failed", zap.Int64("id", userID), zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("User not found", zap.Int64("id", userID))
		render.Error(w, r, render.ErrAPIDBResourceNotFound, err)
		return
	}

	userGroups, err := groups.R().GetGroupsOfUser(user.ID)
	if err != nil {
		zap.L().Error("Get user failed", zap.Int64("id", userID), zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}

	userWithGroups := groups.UserWithGroups{
		User:   user,
		Groups: userGroups,
	}

	render.JSON(w, r, userWithGroups)
}

// ValidateUser godoc
// @Summary Validate a new user definition
// @Description Validate a new user definition
// @Tags Users
// @Accept json
// @Produce json
// @Param user body interface{} true "user (json)"
// @Security Bearer
// @Success 200 {string} string "status OK"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /admin/security/users/validate [post]
func ValidateUser(w http.ResponseWriter, r *http.Request) {
	var user security.UserWithPassword
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		zap.L().Warn("ValidateUser decode json", zap.Error(err))
		render.Error(w, r, render.ErrAPIDecodeJSONBody, err)
		return
	}

	if ok, err := user.IsValid(); !ok {
		zap.L().Warn("Invalid User", zap.Error(err))
		render.Error(w, r, render.ErrAPIResourceInvalid, err)
		return
	}

	render.OK(w, r)
}

// PostUser godoc
// @Summary Create new user
// @Description Add an user to the users
// @Tags Users
// @Accept json
// @Produce json
// @Param user body interface{} true "user (json)"
// @Security Bearer
// @Success 200 {string} string "status OK"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /admin/security/users [post]
func PostUser(w http.ResponseWriter, r *http.Request) {
	var user security.UserWithPassword
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		zap.L().Warn("ValidateUser decode json", zap.Error(err))
		render.Error(w, r, render.ErrAPIDecodeJSONBody, err)
		return
	}

	if ok, err := user.IsValid(); !ok {
		zap.L().Warn("Invalid User", zap.Error(err))
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
		zap.L().Error("Get user failed", zap.Int64("id", userID), zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Error("User not found", zap.Int64("id", userID))
		render.Error(w, r, render.ErrAPIDBResourceNotFoundAfterInsert, fmt.Errorf("Resouce with id %d not found after creation", userID))
		return
	}

	render.JSON(w, r, newUser)
}

// PutUser godoc
// @Summary Update an user
// @Description Updates the user information concerning the user with id
// @Tags Users
// @Accept json
// @Produce json
// @Param id path string true "user ID"
// @Param user body interface{} true "user (json)"
// @Security Bearer
// @Success 200 {string} string "status OK"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /admin/security/users/{id} [put]
func PutUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("PutUser.GetId", zap.String("id", id), zap.Error(err))
		render.Error(w, r, render.ErrAPIParsingInteger, err)
	}

	var user security.User
	err = json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		zap.L().Warn("ValidateUser decode json", zap.Error(err))
		render.Error(w, r, render.ErrAPIDecodeJSONBody, err)
		return
	}
	user.ID = userID

	if ok, err := user.IsValid(); !ok {
		zap.L().Warn("Invalid User", zap.Error(err))
		render.Error(w, r, render.ErrAPIResourceInvalid, err)
		return
	}

	err = users.R().Update(user)
	if err != nil {
		zap.L().Error("PutUser.Update", zap.Error(err))
		render.Error(w, r, render.ErrAPIDBUpdateFailed, err)
		return
	}

	newUser, found, err := users.R().Get(userID)
	if err != nil {
		zap.L().Error("Get user failed", zap.Int64("id", userID), zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Error("User not found", zap.Int64("id", userID))
		render.Error(w, r, render.ErrAPIDBResourceNotFoundAfterInsert, fmt.Errorf("Resouce with id %d not found after update", userID))
		return
	}

	render.JSON(w, r, newUser)
}

// DeleteUser godoc
// @Summary Delete an user
// @Description Deletes an user from the users.
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
	userID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("DeleteUser.GetId", zap.Error(err))
		render.Error(w, r, render.ErrAPIParsingInteger, err)
	}

	err = users.R().Delete(userID)
	if err != nil {
		zap.L().Error("DeleteUser", zap.Error(err))
		render.Error(w, r, render.ErrAPIDBDeleteFailed, err)
		return
	}

	render.OK(w, r)
}
