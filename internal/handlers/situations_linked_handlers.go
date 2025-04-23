package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/fact"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/handlers/render"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/rule"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/situation"
	"github.com/myrteametrics/myrtea-sdk/v5/engine"
	"go.uber.org/zap"
)

// GetSituationEvaluation godoc
// @Summary Get the last evaluation of a situation
// @Description Get the last evaluation of a situation
// @Tags Situations
// @Produce json
// @Param id path string true "Situation ID"
// @Param instanceid path string true "Situation Template Instance ID"
// @Security Bearer
// @Security ApiKeyAuth
// @Success 200 "Status OK"
// @Failure 400 "Status Bad Request"
// @Router /engine/situations/{id}/evaluation/{instanceid} [get]
func GetSituationEvaluation(w http.ResponseWriter, r *http.Request) {

	// TODO: Fixme or remove handler to get one specific situation evaluation (is it even usefull ?)
	render.NotImplemented(w, r)

	// id := chi.URLParam(r, "id")
	// idSituation, err := strconv.ParseInt(id, 10, 64)
	// if err != nil {
	// 	zap.L().Warn("Error on parsing situation id", zap.String("situationID", id), zap.Error(err))
	// 	render.Error(w, r, render.ErrAPIParsingInteger, err)
	// 	return
	// }

	// id = chi.URLParam(r, "instanceid")
	// instanceID, err := strconv.ParseInt(id, 10, 64)
	// if err != nil {
	// 	zap.L().Warn("Error on parsing situation template instance id, using default value (0)", zap.String("instanceID", id), zap.Error(err))
	// 	instanceID = 0
	// }

	// // FIXME: security check !

	// metaDatas, err := situation.GetLastHistoryMetadata(idSituation, instanceID)
	// if err != nil {
	// 	zap.L().Error("Error on getting situation last evaluation id", zap.String("situationID", id), zap.Error(err))
	// 	render.Error(w, r, render.ErrAPIDBSelectFailed, err)
	// 	return
	// }

	// render.JSON(w, r, metaDatas)
}

// GetSituationFacts godoc
// @Summary Get the list of facts for the evaluation of a situation
// @Description Get the list of facts for the evaluation of a situation
// @Tags Situations
// @Produce json
// @Param id path string true "Situation ID"
// @Security Bearer
// @Security ApiKeyAuth
// @Success 200 "Status OK"
// @Failure 400 "Status Bad Request"
// @Failure 401 "Status Unauthorized"
// @Router /engine/situations/{id}/facts [get]
func GetSituationFacts(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idSituation, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("Error on parsing situation id", zap.String("situationID", id), zap.Error(err))
		render.Error(w, r, render.ErrAPIParsingInteger, err)
		return
	}

	// groups := GetUserGroupsFromContext(r)
	// inGroups, err := situation.R().IsInGroups(idSituation, groups)
	// if err != nil {
	// 	zap.L().Error("Error while validating authorization", zap.Int64("idSituation", idSituation), zap.String("groups", fmt.Sprint(groups)), zap.Error(err))
	// 	render.Error(w, r, render.ErrAPIDBSelectFailed, err)
	// 	return
	// }
	// if !inGroups {
	// 	zap.L().Warn("Error while validating authorization", zap.Int64("idSituation", idSituation), zap.String("groups", fmt.Sprint(groups)), zap.Error(err))
	// 	render.Error(w, r, render.ErrAPISecurityNoPermissions, err)
	// 	return
	// }
	// FIXME: security check !

	factIDs, err := situation.R().GetFacts(idSituation)
	if err != nil {
		zap.L().Error("Error on getting situation rules", zap.Int64("idSituation", idSituation), zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}

	facts := make([]engine.Fact, 0)
	for _, factID := range factIDs {
		f, found, err := fact.R().Get(factID)
		if err != nil {
			zap.L().Error("Cannot fetch situation facts", zap.Int64("idSituation", idSituation), zap.Int64("factID", factID), zap.Error(err))
			render.Error(w, r, render.ErrAPIDBSelectFailed, err)
			return
		}
		if !found {
			zap.L().Warn("Situation is linked to a non-existing fact", zap.Int64("idSituation", idSituation), zap.Int64("factID", factID), zap.Error(err))
		} else {
			facts = append(facts, f)
		}
	}
	sort.SliceStable(facts, func(i, j int) bool {
		return facts[i].ID < facts[j].ID
	})

	render.JSON(w, r, facts)
}

// GetSituationRules godoc
// @Summary Get the list of rules for the evaluation of a situation
// @Description Get the list of rules for the evaluation of a situation
// @Tags Situations
// @Produce json
// @Param id path string true "Situation ID"
// @Security Bearer
// @Security ApiKeyAuth
// @Success 200 {array} rule.Rule
// @Failure 400 "Status Bad Request"
// @Failure 401 "Status Unauthorized"
// @Router /engine/situations/{id}/rules [get]
func GetSituationRules(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idSituation, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("Error on parsing situation id", zap.String("situationID", id), zap.Error(err))
		render.Error(w, r, render.ErrAPIParsingInteger, err)
		return
	}

	// groups := GetUserGroupsFromContext(r)
	// inGroups, err := situation.R().IsInGroups(idSituation, groups)
	// if err != nil {
	// 	zap.L().Error("Error while validating authorization", zap.String("Situation ID", id), zap.String("groups", fmt.Sprint(groups)), zap.Error(err))
	// 	render.Error(w, r, render.ErrAPIDBSelectFailed, err)
	// 	return
	// }
	// if !inGroups {
	// 	zap.L().Warn("Error while validating authorization", zap.String("Situation ID", id), zap.String("groups", fmt.Sprint(groups)), zap.Error(err))
	// 	render.Error(w, r, render.ErrAPISecurityNoPermissions, err)
	// 	return
	// }

	// FIXME: security check !

	ruleIDs, err := situation.R().GetRules(idSituation)
	if err != nil {
		zap.L().Error("Error on getting situation rules", zap.String("situationID", id), zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}

	ruleList := make([]rule.Rule, 0)
	for _, ruleID := range ruleIDs {
		rule, found, err := rule.R().Get(ruleID)
		if err != nil {
			zap.L().Error("Cannot fetch situation rules", zap.Int64("idSituation", idSituation), zap.Int64("ruleID", ruleID), zap.Error(err))
			render.Error(w, r, render.ErrAPIDBSelectFailed, err)
			return
		}
		if !found {
			zap.L().Warn("Situation is linked to a non-existing rule", zap.Int64("idSituation", idSituation), zap.Int64("ruleID", ruleID), zap.Error(err))
		} else {
			ruleList = append(ruleList, rule)
		}
	}
	sort.SliceStable(ruleList, func(i, j int) bool {
		return ruleList[i].ID < ruleList[j].ID
	})

	render.JSON(w, r, ruleList)
}

// SetSituationRules godoc
// @Summary Set the list of rules for the evaluation of a situation
// @Description Set the list of rules for the evaluation of a situation
// @Tags Situations
// @Param id path string true "Situation ID"
// @Param ruleIds body []int64 true "Situation Rules"
// @Security Bearer
// @Security ApiKeyAuth
// @Success 200 "Status OK"
// @Failure 400 "Status Bad Request"
// @Failure 401 "Status Unauthorized"
// @Router /engine/situations/{id}/rules [put]
func SetSituationRules(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idSituation, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("Error on parsing situation id", zap.String("situationID", id), zap.Error(err))
		render.Error(w, r, render.ErrAPIParsingInteger, err)
		return
	}

	// groups := GetUserGroupsFromContext(r)
	// inGroups, err := situation.R().IsInGroups(idSituation, groups)
	// if err != nil {
	// 	zap.L().Error("Error while validating authorization", zap.String("Situation ID", id), zap.String("groups", fmt.Sprint(groups)), zap.Error(err))
	// 	render.Error(w, r, render.ErrAPIDBSelectFailed, err)
	// 	return
	// }
	// if !inGroups {
	// 	zap.L().Warn("Error while validating authorization", zap.String("Situation ID", id), zap.String("groups", fmt.Sprint(groups)), zap.Error(err))
	// 	render.Error(w, r, render.ErrAPISecurityNoPermissions, err)
	// 	return
	// }

	// FIXME: security check !

	var ruleIDs []int64
	err = json.NewDecoder(r.Body).Decode(&ruleIDs)
	if err != nil {
		zap.L().Warn("RuleIds json decode", zap.Error(err))
		render.Error(w, r, render.ErrAPIDecodeJSONBody, err)
		return
	}

	err = situation.R().SetRules(idSituation, ruleIDs)
	if err != nil {
		zap.L().Info("Error while setting the situation rules", zap.String("Situation ID", id), zap.Error(err))
		render.Error(w, r, render.ErrAPIDBUpdateFailed, err)
		return
	}

	render.OK(w, r)
}

// PostSituationTemplateInstance godoc
// @Summary Creates a situation template instance
// @Description Creates a situation template instance
// @Tags Situations
// @Accept json
// @Produce json
// @Param id path string true "Situation ID"
// @Param templateInstance body situation.TemplateInstance true "Situation template instance (json)"
// @Security Bearer
// @Security ApiKeyAuth
// @Success 200 {object} situation.TemplateInstance "situation template instance"
// @Failure 400 "Status Bad Request"
// @Failure 500 "Status" internal server error"
// @Router /engine/situations/{id}/instances [post]
func PostSituationTemplateInstance(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idSituation, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("Error on parsing situation id", zap.String("situationID", id), zap.Error(err))
		render.Error(w, r, render.ErrAPIParsingInteger, err)
		return
	}

	// groups := GetUserGroupsFromContext(r)
	// inGroups, err := situation.R().IsInGroups(idSituation, groups)
	// if err != nil {
	// 	zap.L().Error("Error while validating authorization", zap.String("Situation ID", id), zap.String("groups", fmt.Sprint(groups)), zap.Error(err))
	// 	render.Error(w, r, render.ErrAPIDBSelectFailed, err)
	// 	return
	// }
	// if !inGroups {
	// 	zap.L().Warn("Error while validating authorization", zap.String("Situation ID", id), zap.String("groups", fmt.Sprint(groups)), zap.Error(err))
	// 	render.Error(w, r, render.ErrAPISecurityNoPermissions, err)
	// 	return
	// }

	// FIXME: security check !

	var newInstance situation.TemplateInstance
	err = json.NewDecoder(r.Body).Decode(&newInstance)
	if err != nil {
		zap.L().Warn("TemplateInstance json decode", zap.Error(err))
		render.Error(w, r, render.ErrAPIDecodeJSONBody, err)
		return
	}
	newInstance.SituationID = idSituation

	if valid, err := newInstance.IsValid(); !valid {
		if err != nil {
			err = fmt.Errorf("instance is invalid: %s", err.Error())
		} else {
			err = fmt.Errorf("instance is invalid")
		}
		render.Error(w, r, render.ErrAPIResourceInvalid, err)
		return
	}

	instanceID, err := situation.R().CreateTemplateInstance(idSituation, newInstance)
	if err != nil {
		zap.L().Info("Error while creating the situation template instance ", zap.String("Situation ID", id), zap.Error(err))
		render.Error(w, r, render.ErrAPIDBInsertFailed, err)
		return
	}

	instance, found, err := situation.R().GetTemplateInstance(instanceID, gvalParsingEnabled(r.URL.Query()))
	if err != nil {
		zap.L().Error("Cannot retrieve situation template instance", zap.Int64("situationID", idSituation), zap.Int64("instanceID", instanceID), zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("Situation template instance does not exists after update", zap.Int64("situationID", idSituation), zap.Int64("instanceID", instanceID))
		render.Error(w, r, render.ErrAPIDBResourceNotFoundAfterInsert, err)
		return
	}

	render.JSON(w, r, instance)
}

// ValidateSituationTemplateInstance godoc
// @Summary Validate a new situation template instance definition
// @Description Validate a new  situation template instance definition
// @Tags Situations
// @Accept json
// @Produce json
// @Param templateInstance body situation.TemplateInstance true "Situation template instance (json)"
// @Security Bearer
// @Security ApiKeyAuth
// @Success 200 {object} situation.TemplateInstance "Situation template instance"
// @Failure 400 "Status Bad Request"
// @Failure 500 "Status" internal server error"
// @Router /engine/situations/{id}/instances/validate [post]
func ValidateSituationTemplateInstance(w http.ResponseWriter, r *http.Request) {
	var newInstance situation.TemplateInstance
	err := json.NewDecoder(r.Body).Decode(&newInstance)
	if err != nil {
		zap.L().Warn("Template instance json decode", zap.Error(err))
		render.Error(w, r, render.ErrAPIDecodeJSONBody, err)
		return
	}

	if ok, err := newInstance.IsValid(); !ok {
		zap.L().Warn("Template instance is invalid", zap.Error(err))
		render.Error(w, r, render.ErrAPIResourceInvalid, err)
		return
	}

	render.JSON(w, r, newInstance)
}

// PutSituationTemplateInstance godoc
// @Summary replace a situation template instance
// @Description replace a situation template instance
// @Tags Situations
// @Accept json
// @Produce json
// @Param id path string true "Situation ID"
// @Param instanceid path string true "Situation Template Instance ID"
// @Param templateInstance body situation.TemplateInstance true "Situation template instance (json)"
// @Security Bearer
// @Security ApiKeyAuth
// @Success 200 {object} situation.TemplateInstance "situation template instance"
// @Failure 400 "Status Bad Request"
// @Failure 500 "Status" internal server error"
// @Router /engine/situations/{id}/instances/{instanceid} [put]
func PutSituationTemplateInstance(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idSituation, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("Error on parsing situation id", zap.String("situationID", id), zap.Error(err))
		render.Error(w, r, render.ErrAPIParsingInteger, err)
		return
	}

	// groups := GetUserGroupsFromContext(r)
	// inGroups, err := situation.R().IsInGroups(idSituation, groups)
	// if err != nil {
	// 	zap.L().Error("Error while validating authorization", zap.String("Situation ID", id), zap.String("groups", fmt.Sprint(groups)), zap.Error(err))
	// 	render.Error(w, r, render.ErrAPIDBSelectFailed, err)
	// 	return
	// }
	// if !inGroups {
	// 	zap.L().Warn("Error while validating authorization", zap.String("Situation ID", id), zap.String("groups", fmt.Sprint(groups)), zap.Error(err))
	// 	render.Error(w, r, render.ErrAPISecurityNoPermissions, err)
	// 	return
	// }

	// FIXME: security check !

	id = chi.URLParam(r, "instanceid")
	instanceID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("Error on parsing situation template instance id", zap.String("instanceID", id), zap.Error(err))
		render.Error(w, r, render.ErrAPIParsingInteger, err)
		return
	}

	var newInstance situation.TemplateInstance
	err = json.NewDecoder(r.Body).Decode(&newInstance)
	if err != nil {
		zap.L().Warn("TemplateInstance json decode", zap.Error(err))
		render.Error(w, r, render.ErrAPIDecodeJSONBody, err)
		return
	}
	newInstance.SituationID = idSituation

	if valid, err := newInstance.IsValid(); !valid {
		if err != nil {
			err = fmt.Errorf("instance is invalid: %s", err.Error())
		} else {
			err = fmt.Errorf("instance is invalid")
		}
		render.Error(w, r, render.ErrAPIResourceInvalid, err)
		return
	}

	err = situation.R().UpdateTemplateInstance(instanceID, newInstance)
	if err != nil {
		zap.L().Info("Error while updating the situation template instance ", zap.Int64("SituationID", idSituation), zap.String("instanceID", id), zap.Error(err))
		render.Error(w, r, render.ErrAPIDBUpdateFailed, err)
		return
	}

	instance, found, err := situation.R().GetTemplateInstance(instanceID, gvalParsingEnabled(r.URL.Query()))
	if err != nil {
		zap.L().Error("Cannot retrieve situation template instance", zap.Int64("situationID", idSituation), zap.Int64("instanceID", instanceID), zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("Situation template instance does not exists after update", zap.Int64("situationID", idSituation), zap.Int64("instanceID", instanceID))
		render.Error(w, r, render.ErrAPIDBResourceNotFoundAfterInsert, err)
		return
	}

	render.JSON(w, r, instance)
}

// PutSituationTemplateInstances godoc
// @Summary set the template instances of the situation
// @Description set the template instances of the situation
// @Tags Situations
// @Accept json
// @Produce json
// @Param id path string true "Situation ID"
// @Param templateInstances body []situation.TemplateInstance true "Situation template instance list (json array)"
// @Security Bearer
// @Security ApiKeyAuth
// @Success 200 "Status OK"
// @Failure 400 "Status Bad Request"
// @Failure 500 "Status" internal server error"
// @Router /engine/situations/{id}/instances [put]
func PutSituationTemplateInstances(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idSituation, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("Error on parsing situation id", zap.String("situationID", id), zap.Error(err))
		render.Error(w, r, render.ErrAPIParsingInteger, err)
		return
	}

	// groups := GetUserGroupsFromContext(r)
	// inGroups, err := situation.R().IsInGroups(idSituation, groups)
	// if err != nil {
	// 	zap.L().Error("Error while validating authorization", zap.String("Situation ID", id), zap.String("groups", fmt.Sprint(groups)), zap.Error(err))
	// 	render.Error(w, r, render.ErrAPIDBSelectFailed, err)
	// 	return
	// }
	// if !inGroups {
	// 	zap.L().Warn("Error while validating authorization", zap.String("Situation ID", id), zap.String("groups", fmt.Sprint(groups)), zap.Error(err))
	// 	render.Error(w, r, render.ErrAPISecurityNoPermissions, err)
	// 	return
	// }

	// FIXME: security check !

	var newInstances []situation.TemplateInstance
	err = json.NewDecoder(r.Body).Decode(&newInstances)
	if err != nil {
		zap.L().Warn("TemplateInstances json decode", zap.Error(err))
		render.Error(w, r, render.ErrAPIDecodeJSONBody, err)
		return
	}

	var resolvedNewInstances []situation.TemplateInstance
	for _, instance := range newInstances {
		instance.SituationID = idSituation

		if valid, err := instance.IsValid(); !valid {
			if err != nil {
				err = fmt.Errorf("instance '%s' is invalid: %s", instance.Name, err.Error())
			} else {
				err = fmt.Errorf("instance '%s' is invalid", instance.Name)
			}
			render.Error(w, r, render.ErrAPIResourceInvalid, err)
			return
		}

		resolvedNewInstances = append(resolvedNewInstances, instance)
	}

	oldInstances, err := situation.R().GetAllTemplateInstances(idSituation, gvalParsingEnabled(r.URL.Query()))
	if err != nil {
		zap.L().Error("Error while getting existing situation template instances", zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}

	for _, instance := range resolvedNewInstances {

		if instance.ID == 0 {
			_, err = situation.R().CreateTemplateInstance(idSituation, instance)
			if err != nil {
				zap.L().Error("Error while creating template instance", zap.Error(err))
				render.Error(w, r, render.ErrAPIDBInsertFailed, err)
				return
			}
			continue
		}

		if _, found := oldInstances[instance.ID]; found {
			err = situation.R().UpdateTemplateInstance(instance.ID, instance)
			if err != nil {
				zap.L().Error("Error while updating template instance", zap.Error(err))
				render.Error(w, r, render.ErrAPIDBUpdateFailed, err)
				return
			}
			delete(oldInstances, instance.ID)
		} else {
			zap.L().Warn("Error: unknown template instance ID", zap.Error(err))
			render.Error(w, r, render.ErrAPIResourceInvalid, err)
			return
		}
	}

	for instanceID := range oldInstances {
		err = situation.R().DeleteTemplateInstance(instanceID)
		if err != nil {
			zap.L().Error("Error while deleting template instance", zap.Error(err))
			render.Error(w, r, render.ErrAPIDBDeleteFailed, err)
			return
		}
	}

	render.OK(w, r)
}

// DeleteSituationTemplateInstance godoc
// @Summary Delete a situation template instance
// @Description Delete a situation template instance
// @Tags Situations
// @Param id path string true "Situation ID"
// @Param instanceid path string true "Situation Template Instance ID"
// @Security Bearer
// @Security ApiKeyAuth
// @Success 200 "Status OK"
// @Failure 400 "Status Bad Request"
// @Router /engine/situations/{id}/instances/{instanceid} [delete]
func DeleteSituationTemplateInstance(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idSituation, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("Error on parsing situation id", zap.String("situationID", id), zap.Error(err))
		render.Error(w, r, render.ErrAPIParsingInteger, err)
		return
	}
	_ = idSituation

	// groups := GetUserGroupsFromContext(r)
	// inGroups, err := situation.R().IsInGroups(idSituation, groups)
	// if err != nil {
	// 	zap.L().Error("Error while validating authorization", zap.String("Situation ID", id), zap.String("groups", fmt.Sprint(groups)), zap.Error(err))
	// 	render.Error(w, r, render.ErrAPIDBSelectFailed, err)
	// 	return
	// }
	// if !inGroups {
	// 	zap.L().Warn("Error while validating authorization", zap.String("Situation ID", id), zap.String("groups", fmt.Sprint(groups)), zap.Error(err))
	// 	render.Error(w, r, render.ErrAPISecurityNoPermissions, err)
	// 	return
	// }

	// FIXME: security check !

	id = chi.URLParam(r, "instanceid")
	instanceID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("Error on parsing situation template instance id", zap.String("instanceID", id), zap.Error(err))
		render.Error(w, r, render.ErrAPIParsingInteger, err)
		return
	}

	err = situation.R().DeleteTemplateInstance(instanceID)
	if err != nil {
		zap.L().Error("Error while deleting the situation template instance", zap.String("Situation ID", id), zap.String("instanceID", id), zap.Error(err))
		render.Error(w, r, render.ErrAPIDBDeleteFailed, err)
		return
	}

	render.OK(w, r)
}

// GetSituationTemplateInstances godoc
// @Summary Get the list of situation template instances
// @Description Get the list of situation template instances
// @Tags Situations
// @Produce json
// @Param id path string true "Situation ID"
// @Security Bearer
// @Security ApiKeyAuth
// @Success 200 {array} situation.TemplateInstance
// @Failure 400 "Status Bad Request"
// @Failure 401 "Status Unauthorized"
// @Router /engine/situations/{id}/instances [get]
func GetSituationTemplateInstances(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idSituation, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("Error on parsing situation id", zap.String("situationID", id), zap.Error(err))
		render.Error(w, r, render.ErrAPIParsingInteger, err)
		return
	}

	// groups := GetUserGroupsFromContext(r)
	// inGroups, err := situation.R().IsInGroups(idSituation, groups)
	// if err != nil {
	// 	zap.L().Error("Error while validating authorization", zap.String("Situation ID", id), zap.String("groups", fmt.Sprint(groups)), zap.Error(err))
	// 	render.Error(w, r, render.ErrAPIDBSelectFailed, err)
	// 	return
	// }
	// if !inGroups {
	// 	zap.L().Warn("Error while validating authorization", zap.String("Situation ID", id), zap.String("groups", fmt.Sprint(groups)), zap.Error(err))
	// 	render.Error(w, r, render.ErrAPISecurityNoPermissions, err)
	// 	return
	// }

	// FIXME: security check !

	instances, err := situation.R().GetAllTemplateInstances(idSituation, gvalParsingEnabled(r.URL.Query()))
	if err != nil {
		zap.L().Error("Error on getting situation template instances", zap.String("situationID", id), zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}

	instancesSlice := make([]situation.TemplateInstance, 0)
	for _, instance := range instances {
		instancesSlice = append(instancesSlice, instance)
	}

	sort.SliceStable(instancesSlice, func(i, j int) bool {
		return instancesSlice[i].ID < instancesSlice[j].ID
	})

	render.JSON(w, r, instancesSlice)
}

// GetSituationTemplateInstancesUnprotected godoc
// @Summary Get the list of situation template instances
// @Description Get the list of situation template instances
// @Tags Situations
// @Produce json
// @Param id path string true "Situation ID"
// @Security Bearer
// @Security ApiKeyAuth
// @Success 200 {array} situation.TemplateInstance
// @Failure 400 "Status Bad Request"
// @Failure 401 "Status Unauthorized"
// @Router /engine/situations/{id}/instances/unprotected [get]
func GetSituationTemplateInstancesUnprotected(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idSituation, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("Error on parsing situation id", zap.String("situationID", id), zap.Error(err))
		render.Error(w, r, render.ErrAPIParsingInteger, err)
		return
	}

	instances, err := situation.R().GetAllTemplateInstances(idSituation)
	if err != nil {
		zap.L().Error("Error on getting situation template instances", zap.String("situationID", id), zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}

	instancesSlice := make([]situation.TemplateInstance, 0)
	for _, instance := range instances {
		instancesSlice = append(instancesSlice, instance)
	}

	sort.SliceStable(instancesSlice, func(i, j int) bool {
		return instancesSlice[i].ID < instancesSlice[j].ID
	})

	render.JSON(w, r, instancesSlice)
}
