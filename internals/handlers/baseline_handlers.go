package handlers

// import (
// 	"errors"
// 	"net/http"
// 	"sort"

// 	"github.com/myrteametrics/myrtea-engine-api/v5/internals/handlers/render"
// 	"github.com/myrteametrics/myrtea-engine-api/v5/internals/security/permissions"
// 	"github.com/myrteametrics/myrtea-engine-api/v5/plugins/baseline"
// 	"go.uber.org/zap"
// )

// GetBaselines godoc
// @Summary Get all baseline definitions
// @Description Get all baseline definitions
// @Tags Baselines
// @Produce json
// @Security Bearer
// @Success 200 "Status OK"
// @Failure 500 "internal server error"
// @Router /engine/baselines [get]
// func GetBaselines(w http.ResponseWriter, r *http.Request) {
// 	userCtx, _ := GetUserFromContext(r)
// 	if !userCtx.HasPermission(permissions.New(permissions.TypeFact, permissions.All, permissions.ActionList)) {
// 		render.Error(w, r, render.ErrAPISecurityNoPermissions, errors.New("missing permission"))
// 		return
// 	}

// 	var baselines map[int64]baseline.Definition
// 	var err error
// 	if userCtx.HasPermission(permissions.New(permissions.TypeFact, permissions.All, permissions.ActionGet)) {
// 		p, err := baseline.P()
// 		if err != nil {
// 			baselines, _ = p.BaselineService.GetAll()
// 		}
// 	} else {
// 		// resourceIDs := userCtx.GetMatchingResourceIDsInt64(permissions.New(permissions.TypeFact, permissions.All, permissions.ActionGet))
// 		p, err := baseline.P()
// 		if err != nil {
// 			baselines, _ = p.BaselineService.GetAll()
// 		}
// 	}
// 	if err != nil {
// 		zap.L().Error("Error getting baseline definitions", zap.Error(err))
// 		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
// 		return
// 	}

// 	baselinesSlice := make([]baseline.Definition, 0)
// 	for _, baseline := range baselines {
// 		baselinesSlice = append(baselinesSlice, baseline)
// 	}

// 	sort.SliceStable(baselinesSlice, func(i, j int) bool {
// 		return baselinesSlice[i].ID < baselinesSlice[j].ID
// 	})

// 	render.JSON(w, r, baselinesSlice)
// }
