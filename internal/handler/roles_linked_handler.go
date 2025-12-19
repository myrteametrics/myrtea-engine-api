package handler

import (
	"errors"
	"net/http"
	"sort"

	"github.com/go-chi/chi/v5"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/security/permissions"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/security/roles"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/utils/httputil"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// GetRolePermissions godoc
//
//	@Id				GetRolePermissions
//
//	@Summary		Get all permissions associated with a role
//	@Description	Get all permissions associated with a specified role id
//	@Tags			Roles
//	@Produce		json
//	@Param			id	path	string	true	"role ID"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{array}		permissions.Permission	"permission"
//	@Failure		400	{string}	string					"Bad Request"
//	@Failure		404	{object}	httputil.APIError		"Status Not Found"
//	@Failure		500	{object}	httputil.APIError		"Internal Server Error"
//	@Router			/admin/security/roles/{id}/permissions [get]
func GetRolePermissions(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	roleID, err := uuid.Parse(id)
	if err != nil {
		zap.L().Warn("Parse role id", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeRole, roleID.String(), permissions.ActionGet)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	role, found, err := roles.R().Get(roleID)
	if err != nil {
		zap.L().Error("Cannot get role", zap.String("uuid", roleID.String()), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("Role not found", zap.String("uuid", roleID.String()))
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFound, err)
		return
	}

	permissionsSlice, err := permissions.R().GetAllForRole(role.ID)
	if err != nil {
		zap.L().Error("GetPermissions", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}

	sort.SliceStable(permissionsSlice, func(i, j int) bool {
		return permissionsSlice[i].ResourceType < permissionsSlice[j].ResourceType
	})

	httputil.JSON(w, r, permissionsSlice)
}
