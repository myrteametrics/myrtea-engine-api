package handlers

import (
	"encoding/json"
	"net/http"
	"sort"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/groups"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/handlers/render"
	"go.uber.org/zap"
)

// GetGroups godoc
// @Summary Get all user groups
// @Description Gets a list of all user groups.
// @Tags Groups
// @Produce json
// @Security Bearer
// @Success 200 {array} groups.Group "list of groups"
// @Failure 500 {string} string "Internal Server Error"
// @Router /admin/security/groups [get]
func GetGroups(w http.ResponseWriter, r *http.Request) {
	groupsMap, err := groups.R().GetAll()
	if err != nil {
		zap.L().Error("GetGroups", zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}

	groupsSlice := make([]groups.Group, 0)
	for _, action := range groupsMap {
		groupsSlice = append(groupsSlice, action)
	}

	sort.SliceStable(groupsSlice, func(i, j int) bool {
		return groupsSlice[i].ID < groupsSlice[j].ID
	})

	render.JSON(w, r, groupsSlice)
}

// GetGroup godoc
// @Summary Get an user group
// @Description Gets an user group with the specified id
// @Tags Groups
// @Produce json
// @Param id path string true "group ID"
// @Security Bearer
// @Success 200 {object} groups.Group "group"
// @Failure 400 {string} string "Bad Request"
// @Failure 404 {string} string "Not Found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /admin/security/groups/{id} [get]
func GetGroup(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	groupID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("Parse group id", zap.Error(err))
		render.Error(w, r, render.ErrAPIParsingInteger, err)
		return
	}

	group, found, err := groups.R().Get(groupID)
	if err != nil {
		zap.L().Error("Cannot get group", zap.Int64("id", groupID), zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("Group not found", zap.Int64("id", groupID))
		render.Error(w, r, render.ErrAPIDBResourceNotFound, err)
		return
	}

	render.JSON(w, r, group)
}

// ValidateGroup godoc
// @Summary Validate a new group definition
// @Description Validate a new group definition
// @Tags Groups
// @Accept json
// @Produce json
// @Param group body groups.Group true "group (json)"
// @Security Bearer
// @Success 200 {object} groups.Group "group"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /admin/security/groups/validate [post]
func ValidateGroup(w http.ResponseWriter, r *http.Request) {
	var newGroup groups.Group
	err := json.NewDecoder(r.Body).Decode(&newGroup)
	if err != nil {
		zap.L().Warn("Group json decode", zap.Error(err))
		render.Error(w, r, render.ErrAPIDecodeJSONBody, err)
		return
	}

	if ok, err := newGroup.IsValid(); !ok {
		zap.L().Warn("Group is not valid", zap.Error(err))
		render.Error(w, r, render.ErrAPIResourceInvalid, err)
		return
	}

	render.JSON(w, r, newGroup)
}

// PostGroup godoc
// @Summary Create a new group
// @Description Add an user group to the user groups
// @Tags Groups
// @Accept json
// @Produce json
// @Param group body groups.Group true "group (json)"
// @Security Bearer
// @Success 200 {object} groups.Group "group"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /admin/security/groups [post]
func PostGroup(w http.ResponseWriter, r *http.Request) {
	var newGroup groups.Group
	err := json.NewDecoder(r.Body).Decode(&newGroup)
	if err != nil {
		zap.L().Warn("Group json decode", zap.Error(err))
		render.Error(w, r, render.ErrAPIDecodeJSONBody, err)
		return
	}

	if ok, err := newGroup.IsValid(); !ok {
		zap.L().Warn("Group is not valid", zap.Error(err))
		render.Error(w, r, render.ErrAPIResourceInvalid, err)
		return
	}

	groupID, err := groups.R().Create(newGroup)
	if err != nil {
		zap.L().Error("PostGroup.Create", zap.Error(err))
		render.Error(w, r, render.ErrAPIDBInsertFailed, err)
		return
	}

	newGroup, found, err := groups.R().Get(groupID)
	if err != nil {
		zap.L().Error("Cannot get group", zap.Int64("id", groupID), zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Error("Group not found after creation", zap.Int64("id", groupID))
		render.Error(w, r, render.ErrAPIDBResourceNotFoundAfterInsert, err)
		return
	}

	render.JSON(w, r, newGroup)
}

// PutGroup godoc
// @Summary Update group
// @Description Updates the user group information concerning the user group with id
// @Tags Groups
// @Accept json
// @Produce json
// @Param id path string true "group ID"
// @Param group body groups.Group true "group (json)"
// @Security Bearer
// @Success 200 {object} groups.Group "group"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /admin/security/groups/{id} [put]
func PutGroup(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	groupID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("DeleteGroup.GetId", zap.Error(err))
		render.Error(w, r, render.ErrAPIParsingInteger, err)
		return
	}

	var newGroup groups.Group
	err = json.NewDecoder(r.Body).Decode(&newGroup)
	if err != nil {
		zap.L().Warn("Group json decode", zap.Error(err))
		render.Error(w, r, render.ErrAPIDecodeJSONBody, err)
		return
	}
	newGroup.ID = groupID

	if ok, err := newGroup.IsValid(); !ok {
		zap.L().Warn("Group is not valid", zap.Error(err))
		render.Error(w, r, render.ErrAPIResourceInvalid, err)
		return
	}

	err = groups.R().Update(newGroup)
	if err != nil {
		zap.L().Error("PutGroup.Update", zap.Error(err))
		render.Error(w, r, render.ErrAPIDBUpdateFailed, err)
		return
	}

	newGroup, found, err := groups.R().Get(groupID)
	if err != nil {
		zap.L().Error("Cannot get group", zap.Int64("id", groupID), zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Error("Group not found after creation", zap.Int64("id", groupID))
		render.Error(w, r, render.ErrAPIDBResourceNotFoundAfterInsert, err)
		return
	}

	render.JSON(w, r, newGroup)
}

// DeleteGroup godoc
// @Summary Delete group
// @Description Deletes an user group
// @Tags Groups
// @Produce json
// @Param id path string true "group ID"
// @Security Bearer
// @Success 200 {string} string "status OK"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /admin/security/groups/{id} [delete]
func DeleteGroup(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	groupID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("DeleteGroup.GetId", zap.Error(err))
		render.Error(w, r, render.ErrAPIParsingInteger, err)
		return
	}

	err = groups.R().Delete(groupID)
	if err != nil {
		zap.L().Error("Cannot delete group", zap.Error(err))
		render.Error(w, r, render.ErrAPIDBDeleteFailed, err)
		return
	}

	render.OK(w, r)
}
