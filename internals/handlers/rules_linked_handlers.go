package handlers

import (
	"encoding/json"
	"net/http"
	"sort"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/handlers/render"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/situation"
	"github.com/myrteametrics/myrtea-sdk/v4/postgres"

	"go.uber.org/zap"
)

// GetRuleSituations godoc
// @Summary Get the list of situatons associated to a rule
// @Description Get the list of situatons associated to a rule
// @Tags Rules
// @Produce json
// @Param id path string true "Rule ID"
// @Security Bearer
// @Success 200 {array} situation.Situation "list of situations"
// @Failure 400 "Status Bad Request"
// @Failure 401 "Status Unauthorized"
// @Router /engine/rules/{id}/situations [get]
func GetRuleSituations(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idRule, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("Error on parsing rule id", zap.String("RuleID", id), zap.Error(err))
		render.Error(w, r, render.ErrAPIParsingInteger, err)
		return
	}

	situationsMap, err := situation.R().GetAllByRuleID(idRule)
	if err != nil {
		zap.L().Error("Error on getting rule situations", zap.String("situationID", id), zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}

	situationSlice := make([]situation.Situation, 0)
	for _, v := range situationsMap {
		situationSlice = append(situationSlice, v)
	}

	sort.SliceStable(situationSlice, func(i, j int) bool {
		return situationSlice[i].ID < situationSlice[j].ID
	})

	render.JSON(w, r, situationSlice)
}

// PostRuleSituations godoc
// @Summary Add the rule at the end of the rules list of each situation
// @Description Add the rule at the end of the rules list of each situation
// @Tags Rules
// @Produce json
// @Param id path string true "Rule ID"
// @Param situationIds body {array} true "Situation association"
// @Security Bearer
// @Success 200 "Status OK"
// @Failure 400 "Status Bad Request"
// @Failure 401 "Status Unauthorized"
// @Router /engine/rules/{id}/situations [post]
func PostRuleSituations(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idRule, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("Error on parsing rule id", zap.String("RuleID", id), zap.Error(err))
		render.Error(w, r, render.ErrAPIParsingInteger, err)
		return
	}

	var situationIDs []int64
	err = json.NewDecoder(r.Body).Decode(&situationIDs)
	if err != nil {
		zap.L().Warn("SituationsIds json decode", zap.Error(err))
		render.Error(w, r, render.ErrAPIDecodeJSONBody, err)
		return
	}

	situationsMap, err := situation.R().GetAllByRuleID(idRule)
	if err != nil {
		zap.L().Warn("Error getting situations by rulesID", zap.Int64("ruleID", idRule), zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}

	tx, err := postgres.DB().Beginx()
	if err != nil {
		zap.L().Warn("Error beginning DB transaction")
		render.Error(w, r, render.ErrAPIDBTransactionBegin, err)
		return
	}

	for _, situationID := range situationIDs {
		if _, ok := situationsMap[situationID]; ok {
			delete(situationsMap, situationID)
		} else {
			err = situation.R().AddRule(tx, situationID, idRule)
			if err != nil {
				tx.Rollback()
				zap.L().Warn("Error adding the rule to the situation", zap.Int64("situationID", situationID), zap.Error(err))
				render.Error(w, r, render.ErrAPIDBInsertFailed, err)
				return
			}

		}
	}

	for situationID := range situationsMap {
		err = situation.R().RemoveRule(tx, situationID, idRule)
		if err != nil {
			tx.Rollback()
			zap.L().Warn("Error removing the rule from the situation", zap.Int64("situationID", situationID), zap.Error(err))
			render.Error(w, r, render.ErrAPIDBInsertFailed, err)
			return
		}
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		render.Error(w, r, render.ErrAPIDBTransactionCommit, err)
		return
	}

	render.OK(w, r)
}
