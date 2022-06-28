package handlers

import (
	"errors"
	"net/http"
	"sort"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/handlers/render"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/security/permissions"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/security/roles"
	"go.uber.org/zap"
)

// GetRolePermissions godoc
// @Summary Get all permissions associated with a role
// @Description Get all permissions associated with a specified role id
// @Tags Roles
// @Produce json
// @Param id path string true "role ID"
// @Security Bearer
// @Success 200 {array} permissions.Permission "permission"
// @Failure 400 {string} string "Bad Request"
// @Failure 404 {string} string "Not Found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /admin/security/roles/{id}/permissions [get]
func GetRolePermissions(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	roleID, err := uuid.Parse(id)
	if err != nil {
		zap.L().Warn("Parse role id", zap.Error(err))
		render.Error(w, r, render.ErrAPIParsingInteger, err)
		return
	}

	user, _ := GetUserFromContext(r)
	if !user.HasPermission(permissions.New(permissions.TypeRole, roleID.String(), permissions.ActionGet)) {
		render.Error(w, r, render.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	role, found, err := roles.R().Get(roleID)
	if err != nil {
		zap.L().Error("Cannot get role", zap.String("uuid", roleID.String()), zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("Role not found", zap.String("uuid", roleID.String()))
		render.Error(w, r, render.ErrAPIDBResourceNotFound, err)
		return
	}

	permissionsSlice, err := permissions.R().GetAllForRole(role.ID)
	if err != nil {
		zap.L().Error("GetPermissions", zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}

	sort.SliceStable(permissionsSlice, func(i, j int) bool {
		return permissionsSlice[i].ResourceType < permissionsSlice[j].ResourceType
	})

	render.JSON(w, r, permissionsSlice)
}
