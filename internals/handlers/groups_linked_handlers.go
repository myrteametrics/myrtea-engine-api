package handlers

import (
	"encoding/json"
	"net/http"
	"sort"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/groups"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/handlers/render"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/users"
	"go.uber.org/zap"
)

// GetUsersOfGroup godoc
// @Summary Get all users of a group
// @Description Gets a list of all users of a specific group.
// @Tags Group Memberships
// @Produce json
// @Param groupid path string true "user group ID"
// @Security Bearer
// @Success 200 {string} string "Status OK"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /admin/security/groups/{groupid}/users [get]
func GetUsersOfGroup(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "groupid")
	groupID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("GetUsersOfGroup.GetGroupId", zap.Error(err))
		render.Error(w, r, render.ErrAPIParsingInteger, err)
		return
	}

	usersMap, err := users.R().GetUsersOfGroup(groupID)
	if err != nil {
		zap.L().Error("GetUsersOfGroup", zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}

	usersSlice := make([]users.UserOfGroup, 0)
	for _, user := range usersMap {
		usersSlice = append(usersSlice, user)
	}

	sort.SliceStable(usersSlice, func(i, j int) bool {
		return usersSlice[i].ID < usersSlice[j].ID
	})

	render.JSON(w, r, usersSlice)
}

// PutMembership godoc
// @Summary Create membership
// @Description Updates the user membership information concerning the user id and group id
// @Tags Group Memberships
// @Accept json
// @Produce json
// @Param groupid path string true "user group ID"
// @Param userid path string true "user ID"
// @Param membership body groups.Membership true "membership (json)"
// @Security Bearer
// @Success 200 {string} string "status OK"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /admin/security/groups/{groupid}/users/{userid} [put]
func PutMembership(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "groupid")
	groupID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("DeleteMembership.GetGroupId", zap.Error(err))
		render.Error(w, r, render.ErrAPIParsingInteger, err)
		return
	}

	id = chi.URLParam(r, "userid")
	userID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("DeleteMembership.GetGroupId", zap.Error(err))
		render.Error(w, r, render.ErrAPIParsingInteger, err)
		return
	}

	var newMembership groups.Membership
	err = json.NewDecoder(r.Body).Decode(&newMembership)
	if err != nil {
		zap.L().Warn("Membership json decoding", zap.Error(err))
		render.Error(w, r, render.ErrAPIDecodeJSONBody, err)
		return
	}
	newMembership.GroupID = groupID
	newMembership.UserID = userID

	if ok, err := newMembership.IsValid(); !ok {
		zap.L().Warn("Membership is not valid", zap.Error(err))
		render.Error(w, r, render.ErrAPIResourceInvalid, err)
		return
	}

	_, found, err := groups.R().GetMembership(newMembership.UserID, newMembership.GroupID)
	if err != nil {
		zap.L().Error("GetMembership", zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		err = groups.R().CreateMembership(newMembership)
		if err != nil {
			zap.L().Error("PutMembership", zap.Error(err))
			render.Error(w, r, render.ErrAPIDBInsertFailed, err)
			return
		}
	} else {
		err = groups.R().UpdateMembership(newMembership)
		if err != nil {
			zap.L().Error("PutMembership", zap.Error(err))
			render.Error(w, r, render.ErrAPIDBUpdateFailed, err)
			return
		}
	}

	render.OK(w, r)
}

// DeleteMembership godoc
// @Summary Delete membership
// @Description Deletes an user membership from the repository.
// @Tags Group Memberships
// @Produce json
// @Param groupid path string true "user group ID"
// @Param userid path string true "user ID"
// @Security Bearer
// @Success 200 {string} string "status OK"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /admin/security/groups/{groupid}/users/{userid} [delete]
func DeleteMembership(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "groupid")
	groupID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("DeleteMembership.GetGroupId", zap.Error(err))
		render.Error(w, r, render.ErrAPIParsingInteger, err)
		return
	}

	id = chi.URLParam(r, "userid")
	userID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("DeleteMembership.GetGroupId", zap.Error(err))
		render.Error(w, r, render.ErrAPIParsingInteger, err)
		return
	}

	err = groups.R().DeleteMembership(userID, groupID)
	if err != nil {
		zap.L().Error("DeleteMembership", zap.Error(err))
		render.Error(w, r, render.ErrAPIDBDeleteFailed, err)
		return
	}

	render.OK(w, r)
}
