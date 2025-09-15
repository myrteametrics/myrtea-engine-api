package handler

import (
	"encoding/json"
	"errors"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/security/permissions"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/utils/httputil"
	"net/http"
	"sort"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/rule"
	"go.uber.org/zap"
)

// GetRules godoc
//
//	@Id				GetRules
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
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
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
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}

	rulesSlice := make([]rule.Rule, 0)
	for _, rule := range rulesMap {
		rulesSlice = append(rulesSlice, rule)
	}

	sort.SliceStable(rulesSlice, func(i, j int) bool {
		return rulesSlice[i].ID < rulesSlice[j].ID
	})

	httputil.JSON(w, r, rulesSlice)
}

// GetRule godoc
//
//	@Id				GetRule
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
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeRule, strconv.FormatInt(idRule, 10), permissions.ActionGet)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	rule, found, err := rule.R().Get(idRule)
	if err != nil {
		zap.L().Error("Get rule from repository", zap.Int64("id", idRule), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("Rule does not exists", zap.String("ruleid", id))
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFound, err)
		return
	}

	httputil.JSON(w, r, rule)
}

// GetRuleByVersion godoc
//
//	@Id				GetRuleByVersion
//
//	@Summary		Get a rule by version
//	@Description	Get a specific rule by its ID and version ID
//	@Tags			Rules
//	@Produce		json
//	@Param			id			path	int	true	"Rule ID"
//	@Param			versionId	path	int	true	"Rule Version ID"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	rule.Rule	"rule"
//	@Failure		400	"bad request - invalid parameters"
//	@Failure		403	"forbidden - insufficient permissions"
//	@Failure		404	"not found - rule does not exist"
//	@Failure		500	"internal server error"
//	@Router			/engine/rules/{id}/versions/{versionId} [get]
func GetRuleByVersion(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idRule, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("Parsing rule id", zap.String("RuleID", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeRule, strconv.FormatInt(idRule, 10), permissions.ActionGet)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	versionID := chi.URLParam(r, "versionId")
	idVersion, err := strconv.ParseInt(versionID, 10, 64)
	if err != nil {
		zap.L().Warn("Parsing rule id", zap.String("RuleID", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	rule, found, err := rule.R().GetByVersion(idRule, idVersion)
	if err != nil {
		zap.L().Error("Get rule from repository", zap.Int64("id", idRule), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("Rule does not exists", zap.String("ruleid", id))
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFound, err)
		return
	}

	httputil.JSON(w, r, rule)
}

// ValidateRule godoc
//
//	@Id				ValidateRule
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
		httputil.Error(w, r, httputil.ErrAPIDecodeJSONBody, err)
		return
	}

	if ok, err := newRule.IsValid(); !ok {
		zap.L().Warn("Rule is not valid", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIResourceInvalid, err)
		return
	}

	httputil.JSON(w, r, newRule)
}

// PostRule godoc
//
//	@Id				PostRule
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
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	var newRule rule.Rule
	err := json.NewDecoder(r.Body).Decode(&newRule)
	if err != nil {
		zap.L().Warn("Decode rule json", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDecodeJSONBody, err)
		return
	}

	if ok, err := newRule.IsValid(); !ok {
		zap.L().Warn("Rule is not valid", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIResourceInvalid, err)
		return
	}

	exists, err := rule.R().CheckByName(newRule.Name)
	if err != nil {
		zap.L().Error("Cannot retrieve rule", zap.String("Rule.Name", newRule.Name), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}
	if exists {
		zap.L().Info("Rule name already exists", zap.String("Rule.Name", newRule.Name))
		httputil.Error(w, r, httputil.ErrAPIResourceDuplicate, err)
		return
	}

	idRule, err := rule.R().Create(newRule)
	if err != nil {
		zap.L().Error("Error while creating Rule", zap.String("Rule.Name", newRule.Name), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBInsertFailed, err)
		return
	}

	rule, found, err := rule.R().Get(idRule)
	if err != nil {
		zap.L().Error("Get rule from repository", zap.Int64("id", idRule), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("Rule does not exists after creation", zap.Int64("ruleid", idRule))
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFoundAfterInsert, err)
		return
	}

	httputil.JSON(w, r, rule)
}

// PutRule godoc
//
//	@Id				PutRule
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
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeRule, strconv.FormatInt(idRule, 10), permissions.ActionUpdate)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	var newRule rule.Rule
	err = json.NewDecoder(r.Body).Decode(&newRule)
	if err != nil {
		zap.L().Warn("Decode rule json", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDecodeJSONBody, err)
		return
	}
	newRule.ID = idRule

	if ok, err := newRule.IsValid(); !ok {
		zap.L().Warn("Rule is not valid", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIResourceInvalid, err)
		return
	}

	err = rule.R().Update(newRule)
	if err != nil {
		zap.L().Error("Error while updating Rule", zap.String("Name", newRule.Name), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBUpdateFailed, err)
		return
	}

	rule, found, err := rule.R().Get(idRule)
	if err != nil {
		zap.L().Error("Get rule from repository", zap.Int64("id", idRule), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("Rule does not exists after update", zap.Int64("ruleid", idRule))
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFound, err)
		return
	}

	httputil.JSON(w, r, rule)
}

// DeleteRule godoc
//
//	@Id				DeleteRule
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
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeRule, strconv.FormatInt(idRule, 10), permissions.ActionDelete)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	err = rule.R().Delete(idRule)
	if err != nil {
		zap.L().Error("Delete rule", zap.String("ID", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBDeleteFailed, err)
		return
	}

	httputil.OK(w, r)
}
