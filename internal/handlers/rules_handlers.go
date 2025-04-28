package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/handlers/render"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/rule"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/security/permissions"

	"go.uber.org/zap"
)

// GetRules godoc
//
//	@Summary		Get all rules
//	@Description	Get all rules from rules repository
//	@Tags			Rules
//	@Produce		json
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{array}	rule.Rule	"list of rules"
//	@Failure		500	"internal server error"
//	@Router			/engine/rules [get]
func GetRules(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeRule, permissions.All, permissions.ActionList)) {
		render.Error(w, r, render.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	// FIXME : Suport rule advanced security by refactoring rule repository (merge all getter in a single getter with options + getAllByIDs)

	// var rulesMap map[int64]rule.Rule
	// var err error
	// if userCtx.HasPermission(permissions.New(permissions.TypeRule, permissions.All, permissions.ActionGet)) {
	// 	rulesMap, err = rule.R().GetAll()
	// } else {
	// 	resourceIDs := userCtx.GetMatchingResourceIDsInt64(permissions.New(permissions.TypeRule, permissions.All, permissions.ActionGet))
	// 	rulesMap, err = rule.R().GetAllByIDs(resourceIDs)
	// }

	rulesMap, err := rule.R().GetAll()
	if err != nil {
		zap.L().Error("Get rules", zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}

	rulesSlice := make([]rule.Rule, 0)
	for _, rule := range rulesMap {
		rulesSlice = append(rulesSlice, rule)
	}

	sort.SliceStable(rulesSlice, func(i, j int) bool {
		return rulesSlice[i].ID < rulesSlice[j].ID
	})

	render.JSON(w, r, rulesSlice)
}

// GetRule godoc
//
//	@Summary		Get a rule
//	@Description	Get a specific rule by it's ID
//	@Tags			Rules
//	@Produce		json
//	@Param			id	path	string	true	"Rule ID"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	rule.Rule	"rule"
//	@Failure		500	"internal server error"
//	@Router			/engine/rules/{id} [get]
func GetRule(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idRule, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("Parsing rule id", zap.String("RuleID", id), zap.Error(err))
		render.Error(w, r, render.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeRule, strconv.FormatInt(idRule, 10), permissions.ActionGet)) {
		render.Error(w, r, render.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	rule, found, err := rule.R().Get(idRule)
	if err != nil {
		zap.L().Error("Get rule from repository", zap.Int64("id", idRule), zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("Rule does not exists", zap.String("ruleid", id))
		render.Error(w, r, render.ErrAPIDBResourceNotFound, err)
		return
	}

	render.JSON(w, r, rule)
}

// GetRuleByVersion godoc
//
//	@Summary		Get a rule
//	@Description	Get a specific rule by it's ID
//	@Tags			Rules
//	@Produce		json
//	@Param			id	path	string	true	"Rule ID"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	rule.Rule	"rule"
//	@Failure		500	"internal server error"
//	@Router			/engine/rules/{id}/versions/{versionid} [get]
func GetRuleByVersion(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idRule, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("Parsing rule id", zap.String("RuleID", id), zap.Error(err))
		render.Error(w, r, render.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeRule, strconv.FormatInt(idRule, 10), permissions.ActionGet)) {
		render.Error(w, r, render.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	versionID := chi.URLParam(r, "versionid")
	idVersion, err := strconv.ParseInt(versionID, 10, 64)
	if err != nil {
		zap.L().Warn("Parsing rule id", zap.String("RuleID", id), zap.Error(err))
		render.Error(w, r, render.ErrAPIParsingInteger, err)
		return
	}

	rule, found, err := rule.R().GetByVersion(idRule, idVersion)
	if err != nil {
		zap.L().Error("Get rule from repository", zap.Int64("id", idRule), zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("Rule does not exists", zap.String("ruleid", id))
		render.Error(w, r, render.ErrAPIDBResourceNotFound, err)
		return
	}

	render.JSON(w, r, rule)
}

// ValidateRule godoc
//
//	@Summary		validate a new rule definition
//	@Description	validate a new rule definition
//	@Tags			Rules
//	@Accept			json
//	@Produce		json
//	@Param			rule	body	rule.Rule	true	"Rule definition (json)"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	rule.Rule	"rule"
//	@Failure		400	"Status Bad Request"
//	@Failure		500	"Status Internal Server Error"
//	@Router			/engine/rules/validate [post]
func ValidateRule(w http.ResponseWriter, r *http.Request) {
	var newRule rule.Rule
	err := json.NewDecoder(r.Body).Decode(&newRule)
	if err != nil {
		zap.L().Warn("Decode rule json", zap.Error(err))
		render.Error(w, r, render.ErrAPIDecodeJSONBody, err)
		return
	}

	if ok, err := newRule.IsValid(); !ok {
		zap.L().Warn("Rule is not valid", zap.Error(err))
		render.Error(w, r, render.ErrAPIResourceInvalid, err)
		return
	}

	render.JSON(w, r, newRule)
}

// PostRule godoc
//
//	@Summary		create rule
//	@Description	creates new rule
//	@Tags			Rules
//	@Accept			json
//	@Produce		json
//	@Param			rule	body	rule.Rule	true	"Rule definition (json)"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	rule.Rule	"rule"
//	@Failure		400	"Status Bad Request"
//	@Failure		500	"Status Internal Server Error"
//	@Router			/engine/rules [post]
func PostRule(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeRule, permissions.All, permissions.ActionCreate)) {
		render.Error(w, r, render.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	var newRule rule.Rule
	err := json.NewDecoder(r.Body).Decode(&newRule)
	if err != nil {
		zap.L().Warn("Decode rule json", zap.Error(err))
		render.Error(w, r, render.ErrAPIDecodeJSONBody, err)
		return
	}

	if ok, err := newRule.IsValid(); !ok {
		zap.L().Warn("Rule is not valid", zap.Error(err))
		render.Error(w, r, render.ErrAPIResourceInvalid, err)
		return
	}

	exists, err := rule.R().CheckByName(newRule.Name)
	if err != nil {
		zap.L().Error("Cannot retrieve rule", zap.String("Rule.Name", newRule.Name), zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}
	if exists {
		zap.L().Info("Rule name already exists", zap.String("Rule.Name", newRule.Name))
		render.Error(w, r, render.ErrAPIResourceDuplicate, err)
		return
	}

	idRule, err := rule.R().Create(newRule)
	if err != nil {
		zap.L().Error("Error while creating Rule", zap.String("Rule.Name", newRule.Name), zap.Error(err))
		render.Error(w, r, render.ErrAPIDBInsertFailed, err)
		return
	}

	rule, found, err := rule.R().Get(idRule)
	if err != nil {
		zap.L().Error("Get rule from repository", zap.Int64("id", idRule), zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("Rule does not exists after creation", zap.Int64("ruleid", idRule))
		render.Error(w, r, render.ErrAPIDBResourceNotFoundAfterInsert, err)
		return
	}

	render.JSON(w, r, rule)
}

// PutRule godoc
//
//	@Summary		Create or remplace a rule definition
//	@Description	Create or remplace a rule definition
//	@Tags			Rules
//	@Accept			json
//	@Produce		json
//	@Param			id		path	string		true	"Rule ID"
//	@Param			rule	body	rule.Rule	true	"Rule definition (json)"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	rule.Rule	"rule"
//	@Failure		400	"Status Bad Request"
//	@Failure		500	"Status"	internal	server	error"
//	@Router			/engine/rules/{id} [put]
func PutRule(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idRule, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("Error on parsing rule id", zap.String("RuleID", id), zap.Error(err))
		render.Error(w, r, render.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeRule, strconv.FormatInt(idRule, 10), permissions.ActionUpdate)) {
		render.Error(w, r, render.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	var newRule rule.Rule
	err = json.NewDecoder(r.Body).Decode(&newRule)
	if err != nil {
		zap.L().Warn("Decode rule json", zap.Error(err))
		render.Error(w, r, render.ErrAPIDecodeJSONBody, err)
		return
	}
	newRule.ID = idRule

	if ok, err := newRule.IsValid(); !ok {
		zap.L().Warn("Rule is not valid", zap.Error(err))
		render.Error(w, r, render.ErrAPIResourceInvalid, err)
		return
	}

	err = rule.R().Update(newRule)
	if err != nil {
		zap.L().Error("Error while updating Rule", zap.String("Name", newRule.Name), zap.Error(err))
		render.Error(w, r, render.ErrAPIDBUpdateFailed, err)
		return
	}

	rule, found, err := rule.R().Get(idRule)
	if err != nil {
		zap.L().Error("Get rule from repository", zap.Int64("id", idRule), zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("Rule does not exists after update", zap.Int64("ruleid", idRule))
		render.Error(w, r, render.ErrAPIDBResourceNotFound, err)
		return
	}

	render.JSON(w, r, rule)
}

// DeleteRule godoc
//
//	@Summary		delete rule
//	@Description	delete rule
//	@Tags			Rules
//	@Produce		json
//	@Param			id	path	string	true	"Rule ID"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	"Status OK"
//	@Failure		400	"Status Bad Request"
//	@Failure		500	"Status Internal Server Error"
//	@Router			/engine/rules/{id} [delete]
func DeleteRule(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idRule, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("Error on parsing rule id", zap.String("RuleID", id), zap.Error(err))
		render.Error(w, r, render.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeRule, strconv.FormatInt(idRule, 10), permissions.ActionDelete)) {
		render.Error(w, r, render.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	err = rule.R().Delete(idRule)
	if err != nil {
		zap.L().Error("Delete rule", zap.String("ID", id), zap.Error(err))
		render.Error(w, r, render.ErrAPIDBDeleteFailed, err)
		return
	}

	render.OK(w, r)
}
