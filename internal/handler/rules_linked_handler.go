package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"strconv"

	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/security/permissions"
	situation2 "github.com/myrteametrics/myrtea-engine-api/v5/pkg/situation"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/utils/httputil"

	"github.com/go-chi/chi/v5"
	"github.com/myrteametrics/myrtea-sdk/v5/postgres"

	"go.uber.org/zap"
)

// GetRuleSituations godoc
//
//	@Id				GetRuleSituations
//
//	@Summary		Get the list of situations associated with a rule
//	@Description	Get the list of situations associated with a rule
//	@Tags			Rules
//	@Produce		json
//	@Param			id	path	int	true	"Rule ID"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	"list of situations"
//	@Failure		400	{object}	httputil.APIError	"Bad Request"
//	@Failure		401	"Status Unauthorized"
//	@Router			/engine/rules/{id}/situations [get]
func GetRuleSituations(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idRule, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("Error on parsing rule id", zap.String("RuleID", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeRule, permissions.All, permissions.ActionList)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	situationsMap, err := situation2.R().GetAllByRuleID(idRule, gvalParsingEnabled(r.URL.Query()))
	if err != nil {
		zap.L().Error("Error on getting rule situations", zap.String("situationID", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}

	situationSlice := make([]situation2.Situation, 0)
	for _, v := range situationsMap {
		situationSlice = append(situationSlice, v)
	}

	sort.SliceStable(situationSlice, func(i, j int) bool {
		return situationSlice[i].ID < situationSlice[j].ID
	})

	httputil.JSON(w, r, situationSlice)
}

// PostRuleSituations godoc
//
//	@Id				PostRuleSituations
//
//	@Summary		Add the rule at the end of the rules list of each situation
//	@Description	Add the rule at the end of the rules list of each situation
//	@Tags			Rules
//	@Produce		json
//	@Param			id				path	int		true	"Rule ID"
//	@Param			situationIds	body	[]int64	true	"Situation association"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	"Status OK"
//	@Failure		400	{object}	httputil.APIError	"Bad Request"
//	@Failure		401	"Status Unauthorized"
//	@Router			/engine/rules/{id}/situations [post]
func PostRuleSituations(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idRule, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("Error on parsing rule id", zap.String("RuleID", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeRule, permissions.All, permissions.ActionUpdate)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	var situationIDs []int64
	err = json.NewDecoder(r.Body).Decode(&situationIDs)
	if err != nil {
		zap.L().Warn("SituationsIds json decode", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDecodeJSONBody, err)
		return
	}

	situationsMap, err := situation2.R().GetAllByRuleID(idRule, gvalParsingEnabled(r.URL.Query()))
	if err != nil {
		zap.L().Warn("Error getting situations by rulesID", zap.Int64("ruleID", idRule), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}

	tx, err := postgres.DB().Beginx()
	if err != nil {
		zap.L().Warn("Error beginning DB transaction")
		httputil.Error(w, r, httputil.ErrAPIDBTransactionBegin, err)
		return
	}
	defer func() { _ = tx.Rollback() }()

	for _, situationID := range situationIDs {
		if _, ok := situationsMap[situationID]; ok {
			delete(situationsMap, situationID)
		} else {
			err = situation2.R().AddRule(tx, situationID, idRule)
			if err != nil {
				zap.L().Warn("Error adding the rule to the situation", zap.Int64("situationID", situationID), zap.Error(err))
				httputil.Error(w, r, httputil.ErrAPIDBInsertFailed, err)
				return
			}

		}
	}

	for situationID := range situationsMap {
		err = situation2.R().RemoveRule(tx, situationID, idRule)
		if err != nil {
			zap.L().Warn("Error removing the rule from the situation", zap.Int64("situationID", situationID), zap.Error(err))
			httputil.Error(w, r, httputil.ErrAPIDBInsertFailed, err)
			return
		}
	}

	err = tx.Commit()
	if err != nil {
		httputil.Error(w, r, httputil.ErrAPIDBTransactionCommit, err)
		return
	}

	httputil.OK(w, r)
}

// GetRuleSituationInstances godoc
//
//	@Id				GetRuleSituationInstances
//
//	@Summary		Get the list of situation instances associated to a rule
//	@Description	Get the list of situation instances associated to a rule
//	@Tags			Rules
//	@Produce		json
//	@Param			id	path	int	true	"Rule ID"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	"list of situation instances"
//	@Failure		400	{object}	httputil.APIError	"Bad Request"
//	@Failure		401	"Status Unauthorized"
//	@Router			/engine/rules/{id}/situation-instances [get]
func GetRuleSituationInstances(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idRule, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("Error on parsing rule id", zap.String("RuleID", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeRule, permissions.All, permissions.ActionList)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	situationInstances, err := situation2.R().GetAllTemplateInstancesByRuleID(idRule, gvalParsingEnabled(r.URL.Query()))
	if err != nil {
		zap.L().Error("Error on getting rule situation instances", zap.String("situationID", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}

	httputil.JSON(w, r, situationInstances)
}
