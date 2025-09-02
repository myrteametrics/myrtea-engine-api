package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/model"
	fact2 "github.com/myrteametrics/myrtea-engine-api/v5/pkg/fact"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/reader"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/security/permissions"
	situation2 "github.com/myrteametrics/myrtea-engine-api/v5/pkg/situation"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/utils/httputil"
	"github.com/myrteametrics/myrtea-sdk/v5/elasticsearch"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/plugins/baseline"
	"github.com/myrteametrics/myrtea-sdk/v5/engine"
	"go.uber.org/zap"
)

// GetFacts godoc
//
//	@Id				GetFacts
//
//	@Summary		Get all fact definitions
//	@Description	Get all fact definitions
//	@Tags			Facts
//	@Produce		json
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	"Status OK"
//	@Failure		500	"internal server error"
//	@Router			/engine/facts [get]
func GetFacts(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeFact, permissions.All, permissions.ActionList)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	var facts map[int64]engine.Fact
	var err error
	if userCtx.HasPermission(permissions.New(permissions.TypeFact, permissions.All, permissions.ActionGet)) {
		facts, err = fact2.R().GetAll()
	} else {
		resourceIDs := userCtx.GetMatchingResourceIDsInt64(permissions.New(permissions.TypeFact, permissions.All, permissions.ActionGet))
		facts, err = fact2.R().GetAllByIDs(resourceIDs)
	}
	if err != nil {
		zap.L().Error("Error getting facts", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}

	factsSlice := make([]engine.Fact, 0)
	for _, fact := range facts {
		factsSlice = append(factsSlice, fact)
	}

	sort.SliceStable(factsSlice, func(i, j int) bool {
		return factsSlice[i].ID < factsSlice[j].ID
	})

	httputil.JSON(w, r, factsSlice)
}

// GetFact godoc
//
//	@Id				GetFact
//
//	@Summary		Get a fact definition
//	@Description	Get a fact definition
//	@Tags			Facts
//	@Produce		json
//	@Param			id	path	string	true	"Fact ID"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	"Status OK"
//	@Failure		400	"Status Bad Request"
//	@Router			/engine/facts/{id} [get]
func GetFact(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idFact, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("Error on parsing fact id", zap.String("factID", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeFact, strconv.FormatInt(idFact, 10), permissions.ActionGet)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	f, found, err := fact2.R().Get(idFact)
	if err != nil {
		zap.L().Error("Cannot retrieve fact", zap.Int64("factID", idFact), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("fact does not exists", zap.Int64("factID", idFact))
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFound, err)
		return
	}

	httputil.JSON(w, r, f)
}

// ValidateFact godoc
//
//	@Id				ValidateFact
//
//	@Summary		Validate a new fact definition
//	@Description	Validate a new fact definition
//	@Tags			Facts
//	@Accept			json
//	@Produce		json
//	@Param			fact	body	interface{}	true	"Fact definition (json)"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	"Status OK"
//	@Failure		400	"Status Bad Request"
//	@Failure		500	"Status"	internal	server	error"
//	@Router			/engine/facts/validate [post]
func ValidateFact(w http.ResponseWriter, r *http.Request) {

	var newFact engine.Fact
	err := json.NewDecoder(r.Body).Decode(&newFact)
	if err != nil {
		zap.L().Warn("Fact definition json decode", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDecodeJSONBody, err)
		return
	}

	if ok, err := newFact.IsValid(); !ok {
		zap.L().Warn("Fact definition json is invalid", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIResourceInvalid, err)
		return
	}

	httputil.JSON(w, r, newFact)
}

// PostFact godoc
//
//	@Id				PostFact
//
//	@Summary		Create a new fact definition
//	@Description	Create a new fact definition
//	@Tags			Facts
//	@Accept			json
//	@Produce		json
//	@Param			fact	body	interface{}	true	"Fact definition (json)"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	"Status OK"
//	@Failure		400	"Status Bad Request"
//	@Failure		500	"Status"	internal	server	error"
//	@Router			/engine/facts [post]
func PostFact(w http.ResponseWriter, r *http.Request) {

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeFact, permissions.All, permissions.ActionCreate)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	var newFact engine.Fact
	err := json.NewDecoder(r.Body).Decode(&newFact)
	if err != nil {
		zap.L().Warn("Fact definition json decode", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDecodeJSONBody, err)
		return
	}

	if ok, err := newFact.IsValid(); !ok {
		zap.L().Warn("Fact definition json is invalid", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIResourceInvalid, err)
		return
	}

	newFactID, err := fact2.R().Create(newFact)
	if err != nil {
		zap.L().Error("Error while creating the Fact", zap.Any("fact", newFact), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBInsertFailed, err)
		return
	}

	f, found, err := fact2.R().Get(newFactID)
	if err != nil {
		zap.L().Error("Error while fetch the created fact", zap.Any("newfactID", newFactID), zap.Any("newfact", newFact), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Error("Fact cannot be found after creation", zap.Any("newfactID", newFactID), zap.Any("newfact", newFact), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFoundAfterInsert, fmt.Errorf("Resouce with id %d not found after creation", newFactID))
		return
	}

	httputil.JSON(w, r, f)
}

// PutFact godoc
//
//	@Id				PutFact
//
//	@Summary		Create or remplace a fact definition
//	@Description	Create or remplace a fact definition
//	@Tags			Facts
//	@Accept			json
//	@Produce		json
//	@Param			id		path	string		true	"Fact ID"
//	@Param			fact	body	interface{}	true	"Fact definition (json)"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	"Status OK"
//	@Failure		400	"Status Bad Request"
//	@Failure		500	"Status"	internal	server	error"
//	@Router			/engine/facts/{id} [put]
func PutFact(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idFact, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("Error on parsing fact id", zap.String("factID", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeFact, strconv.FormatInt(idFact, 10), permissions.ActionUpdate)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	var newFact engine.Fact
	err = json.NewDecoder(r.Body).Decode(&newFact)
	if err != nil {
		zap.L().Warn("Fact definition json decode", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDecodeJSONBody, err)
		return
	}
	newFact.ID = idFact

	if ok, err := newFact.IsValid(); !ok {
		zap.L().Warn("Fact definition json is invalid", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIResourceInvalid, err)
		return
	}

	err = fact2.R().Update(idFact, newFact)
	if err != nil {
		zap.L().Error("Error while updating the Fact", zap.Int64("idFact", idFact), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBUpdateFailed, err)
		return
	}

	f, found, err := fact2.R().Get(idFact)
	if err != nil {
		zap.L().Error("Error while fetch the created fact", zap.Any("factID", idFact), zap.Any("newfact", newFact), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Error("Error while creating the Fact", zap.Any("factID", idFact), zap.Any("newfact", newFact), zap.Error(errors.New("fact not properly created")))
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFoundAfterInsert, fmt.Errorf("Resouce with id %d not found after update", idFact))
		return
	}

	httputil.JSON(w, r, f)
}

// DeleteFact godoc
//
//	@Id				DeleteFact
//
//	@Summary		Delete a fact definition
//	@Description	Delete a fact definition
//	@Tags			Facts
//	@Produce		json
//	@Param			id	path	string	true	"Fact ID"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	"Status OK"
//	@Failure		400	"Status Bad Request"
//	@Router			/engine/facts/{id} [delete]
func DeleteFact(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idFact, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("Error on parsing fact id", zap.String("factID", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeFact, strconv.FormatInt(idFact, 10), permissions.ActionDelete)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	err = fact2.R().Delete(idFact)
	if err != nil {
		zap.L().Error("Error while deleting the Fact", zap.Int64("factID", idFact), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBDeleteFailed, err)
		return
	}

	httputil.OK(w, r)
}

// ExecuteFact godoc
//
//	@Id				ExecuteFact
//
//	@Summary		Execute a fact with a given timestamp
//	@Description	Execute a fact with a given timestamp (This route is deprecated. Please use POST /engine/facts/execute instead.)
//	@Tags			Facts
//	@Produce		json
//	@Param			id				path	string	true	"Fact ID"
//	@Param			byName			query	string	false	"Find fact by it's name"
//	@Param			time			query	string	false	"Timestamp used for the fact execution"
//	@Param			nhit			query	int		false	"Hit per page"
//	@Param			offset			query	int		false	"Offset number"
//	@Param			placeholders	query	string	false	"Placeholders (format: key1:value1,key2:value2)"
//	@Param			debug			query	string	false	"Debug true/false"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	"Status OK"
//	@Failure		400	"Status Bad Request"
//	@Deprecated		true
//	@Router			/engine/facts/{id}/execute [get]
//
// Deprecated: Use ExecuteFactOrGetHits instead.
func ExecuteFact(w http.ResponseWriter, r *http.Request) {

	t, err := QueryParamToOptionalTime(r, "time", time.Now())
	if err != nil {
		zap.L().Warn("Parse input time", zap.Error(err), zap.String("rawTime", r.URL.Query().Get("time")))
		httputil.Error(w, r, httputil.ErrAPIParsingDateTime, err)
		return
	}

	nhit, err := QueryParamToOptionalInt(r, "nhit", 0)
	if err != nil {
		zap.L().Warn("Parse input nhit", zap.Error(err), zap.String("rawNhit", r.URL.Query().Get("nhit")))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	offset, err := QueryParamToOptionalInt(r, "offset", 0)
	if err != nil {
		zap.L().Warn("Parse input offset", zap.Error(err), zap.String("raw offset", r.URL.Query().Get("offset")))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	placeholders, err := QueryParamToOptionalKeyValues(r, "placeholders", make(map[string]interface{}))
	if err != nil {
		zap.L().Warn("Parse input placeholders", zap.Error(err), zap.String("raw placeholders", r.URL.Query().Get("placeholders")))
		httputil.Error(w, r, httputil.ErrAPIParsingKeyValue, err)
		return
	}

	byName := false
	_byName := r.URL.Query().Get("byName")
	if _byName == "true" {
		byName = true
	}

	id := chi.URLParam(r, "id")
	f, apiError, err := lookupFact(byName, id)
	if err != nil {
		httputil.Error(w, r, apiError, err)
		return
	}

	// Might be a security Issue (because we lookup for the fact ID / Name before any control)
	// Should be better to just remove the "lookup by name" feature (which is not used anymore, and has no sense in this API)
	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeFact, strconv.FormatInt(f.ID, 10), permissions.ActionGet)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	data, err := fact2.ExecuteFact(t, f, 0, 0, placeholders, nhit, offset, false)
	if err != nil {
		zap.L().Error("Cannot execute fact", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIElasticSelectFailed, err)
		return
	}

	if data.Aggregates != nil {
		pluginBaseline, err := baseline.P()
		if err == nil {
			// value, err := pluginBaseline.Baseline.GetBaselineValue(0, f.ID, situationID, situationInstanceID, t)
			values, err := pluginBaseline.BaselineService.GetBaselineValues(-1, f.ID, 0, 0, t)
			if err != nil {
				zap.L().Error("Cannot fetch fact baselines", zap.Int64("id", f.ID), zap.Error(err))
			}
			data.Aggregates.Baselines = values
		}
	}

	httputil.JSON(w, r, data)
}

// ExecuteFactOrGetHits godoc
//
//	@Id				ExecuteFactOrGetHits
//
//	@Summary		Execute a fact and restitue the hits
//	@Description	Execute a fact and restitue the hits
//	@Tags			Facts
//	@Accept			json
//	@Produce		json
//	@Param			request	body	model.FactHitsRequest	true	"Request parameters"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	"Status OK"
//	@Failure		400	"Status Bad Request"
//	@Router			/engine/facts/execute [POST]
func ExecuteFactOrGetHits(w http.ResponseWriter, r *http.Request) {
	var request model.FactHitsRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		zap.L().Error("Failed to decode request body", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIUnexpectedParamValue, err)
		return
	}

	if err := request.ValidateParseParam(); err != nil {
		zap.L().Error("Failed to validate request body", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIUnexpectedParamValue, err)
		return
	}

	var f engine.Fact
	var found bool

	f, found, err := fact2.R().Get(request.FactId)
	if err != nil {
		zap.L().Error("Error while fetching fact", zap.Int64("factid", request.FactId), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("Fact does not exists", zap.Int64("factid", request.FactId))
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFound, err)
		return
	}

	if f.IsObject {
		zap.L().Warn("Fact is an object fact", zap.Int64("factid", request.FactId))
		httputil.Error(w, r, httputil.ErrAPIResourceInvalid, fmt.Errorf("service not supported on fact object"))
		return
	}

	if f.Intent.Operator == engine.Delete {
		zap.L().Warn("Fact is a delete fact", zap.Int64("factid", request.FactId))
		httputil.Error(w, r, httputil.ErrAPIResourceInvalid, fmt.Errorf("service not supported on delete fact"))
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeFact, strconv.FormatInt(request.FactId, 10), permissions.ActionGet)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	var data *reader.WidgetData
	placeholders := make(map[string]interface{})

	if f.IsTemplate {
		if request.SituationId <= 0 {
			zap.L().Warn("Missing situationId for template fact")
			httputil.Error(w, r, httputil.ErrAPIParsingInteger, errors.New("missing situationId for template fact"))
			return
		}

		situationn, found, err := situation2.R().Get(request.SituationId, true)
		if err != nil {
			zap.L().Error("Cannot retrieve situation", zap.Int64("situationID", request.SituationId), zap.Error(err))
			httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
			return
		}
		if !found {
			zap.L().Warn("Situation does not exists", zap.Int64("situationID", request.SituationId))
			httputil.Error(w, r, httputil.ErrAPIDBResourceNotFound, err)
			return
		}

		var situationInstance situation2.TemplateInstance
		if request.SituationInstanceId <= 0 {
			zap.L().Warn("Missing situationInstanceId for template fact")
			httputil.Error(w, r, httputil.ErrAPIParsingInteger, errors.New("missing situationInstanceId for template fact"))
			return
		}
		situationInstance, found, err = situation2.R().GetTemplateInstance(request.SituationInstanceId, true)
		if err != nil {
			zap.L().Error("Cannot retrieve situation Instance", zap.Int64("situationInstanceID", request.SituationInstanceId), zap.Error(err))
			httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
			return
		}
		if !found {
			zap.L().Warn("Situation Instance does not exists", zap.Int64("situationInstanceID", request.SituationInstanceId))
			httputil.Error(w, r, httputil.ErrAPIDBResourceNotFound, err)
			return
		}

		for key, param := range situationn.Parameters {
			placeholders[key] = param
		}

		for key, param := range situationInstance.Parameters {
			placeholders[key] = param
		}
	}

	if request.FactParameters != nil {
		for key, param := range request.FactParameters {
			placeholders[key] = param
		}
	}

	if request.HitsOnly {
		f.Intent.Operator = engine.Select
	}

	t := time.Now().Truncate(1 * time.Second).UTC()

	data, err = fact2.ExecuteFact(t, f, 0, 0, placeholders, request.Nhit, request.Offset, !*request.Debug)
	if err != nil {
		zap.L().Error("Cannot execute fact", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIElasticSelectFailed, err)
		return
	}

	if !request.HitsOnly && data.Aggregates != nil {
		pluginBaseline, err := baseline.P()
		if err == nil {
			values, err := pluginBaseline.BaselineService.GetBaselineValues(-1, f.ID, 0, 0, t)
			if err != nil {
				zap.L().Error("Cannot fetch fact baselines", zap.Int64("id", f.ID), zap.Error(err))
			}
			data.Aggregates.Baselines = values
		}
	}

	httputil.JSON(w, r, data)
}

// GetFactHits godoc
//
//	@Id				GetFactHits
//
//	@Summary		Execute a fact and restitue the hits
//	@Description	Execute a fact and restitue the hits (This route is deprecated. Please use POST /engine/facts/execute instead.)
//	@Tags			Facts
//	@Produce		json
//	@Param			id					path	string	true	"Fact ID"
//	@Param			time				query	string	true	"Timestamp used for the fact execution"
//	@Param			nhit				query	int		false	"Hit per page"
//	@Param			offset				query	int		false	"Offset number"
//	@Param			situationId			query	string	false	"Situation Id, necessary if the fact is template"
//	@Param			situationInstanceId	query	string	false	"Situation instance Id if applicable"
//	@Param			debug				query	string	false	"Debug true/false"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	"Status OK"
//	@Failure		400	"Status Bad Request"
//	@Deprecated		true
//	@Router			/engine/facts/{id}/hits [get]
//
// Deprecated: Use ExecuteFactOrGetHits instead.
func GetFactHits(w http.ResponseWriter, r *http.Request) {

	t, err := ParseTime(r.URL.Query().Get("time"))
	if err != nil {
		zap.L().Error("Parse input time", zap.Error(err), zap.String("rawTime", r.URL.Query().Get("time")))
		httputil.Error(w, r, httputil.ErrAPIParsingDateTime, err)
		return
	}

	nhit, err := ParseInt(r.URL.Query().Get("nhit"))
	if err != nil {
		zap.L().Error("Parse input nhit", zap.Error(err), zap.String("rawNhit", r.URL.Query().Get("nhit")))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	offset, err := ParseInt(r.URL.Query().Get("offset"))
	if err != nil {
		zap.L().Error("Parse input offset", zap.Error(err), zap.String("raw offset", r.URL.Query().Get("offset")))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	factParameters, err := ParseFactParameters(r.URL.Query().Get("factParameters"))
	if err != nil {
		zap.L().Error("Parse input FactParameters", zap.Error(err), zap.String("raw offset", r.URL.Query().Get("factParameters")))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	id := chi.URLParam(r, "id")
	var f engine.Fact
	var found bool

	idFact, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("Error on parsing fact id", zap.String("factID", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	f, found, err = fact2.R().Get(idFact)
	if err != nil {
		zap.L().Error("Error while fetching fact", zap.String("factid", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("Fact does not exists", zap.String("factid", id))
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFound, err)
		return
	}

	if f.Dimensions != nil {
		zap.L().Warn("Fact does have dimensions", zap.String("factid", id))
		httputil.Error(w, r, httputil.ErrAPIResourceInvalid, fmt.Errorf("service not supported on fact with dimensions"))
		return
	}

	if f.IsObject {
		zap.L().Warn("Fact is an object fact", zap.String("factid", id))
		httputil.Error(w, r, httputil.ErrAPIResourceInvalid, fmt.Errorf("service not supported on fact object"))
		return
	}

	// Might be a security Issue (because we lookup for the fact ID / Name before any control)
	// Should be better to just remove the "lookup by name" feature (which is not used anymore, and has no sense in this API)
	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeFact, strconv.FormatInt(idFact, 10), permissions.ActionGet)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	var data *reader.WidgetData
	placeholders := make(map[string]interface{})

	if f.IsTemplate {
		idSituationStr := r.URL.Query().Get("situationId")
		idSituation, err := strconv.ParseInt(idSituationStr, 10, 64)
		if err != nil {
			zap.L().Warn("Parse input situationId", zap.Error(err), zap.String("rawsituationId", r.URL.Query().Get("situationId")))
			httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
			return
		}

		situationn, found, err := situation2.R().Get(idSituation, gvalParsingEnabled(r.URL.Query()))
		if err != nil {
			zap.L().Error("Cannot retrieve situation", zap.Int64("situationID", idSituation), zap.Error(err))
			httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
			return
		}
		if !found {
			zap.L().Warn("Situation does not exists", zap.Int64("situationID", idSituation))
			httputil.Error(w, r, httputil.ErrAPIDBResourceNotFound, err)
			return
		}

		var situationInstance situation2.TemplateInstance
		situationInstanceIDStr := r.URL.Query().Get("situationInstanceId")
		if situationInstanceIDStr != "" {
			situationInstanceID, err := strconv.ParseInt(situationInstanceIDStr, 10, 64)
			if err != nil {
				zap.L().Warn("Parse input situationInstanceId", zap.Error(err), zap.String("rawSituationInstanceId", r.URL.Query().Get("situationInstanceId")))
				httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
				return
			}

			situationInstance, found, err = situation2.R().GetTemplateInstance(situationInstanceID)
			if err != nil {
				zap.L().Error("Cannot retrieve situation Instance", zap.Int64("situationInstanceID", situationInstanceID), zap.Error(err))
				httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
				return
			}
			if !found {
				zap.L().Warn("Situation Instance does not exists", zap.Int64("situationInstanceID", situationInstanceID))
				httputil.Error(w, r, httputil.ErrAPIDBResourceNotFound, err)
				return
			}
		}

		for key, param := range situationn.Parameters {
			placeholders[key] = param
		}

		for key, param := range situationInstance.Parameters {
			placeholders[key] = param
		}

	}

	// parameters entered from the front-ends
	for key, param := range factParameters {
		placeholders[key] = param
	}

	// Change the behaviour of the Fact
	f.Intent.Operator = engine.Select

	data, err = fact2.ExecuteFact(t, f, 0, 0, placeholders, nhit, offset, false)
	if err != nil {
		zap.L().Error("Cannot execute fact", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIElasticSelectFailed, err)
		return
	}

	httputil.JSON(w, r, data)
}

// FactToESQuery godoc
//
//	@Id				FactToESQuery
//
//	@Summary		Execute a fact with a given timestamp
//	@Description	Execute a fact with a given timestamp
//	@Tags			Facts
//	@Produce		json
//	@Param			id				path	string	true	"Fact ID"
//	@Param			byName			query	string	false	"Find fact by it's name"
//	@Param			situationid		query	string	false	"Optional SituationID"
//	@Param			instanceid		query	string	false	"Optional InstanceID"
//	@Param			time			query	string	true	"Timestamp used for the fact execution"
//	@Param			nhit			query	int		false	"Hit per page"
//	@Param			offset			query	int		false	"Offset number"
//	@Param			placeholders	query	string	false	"Placeholders (format: key1:value1,key2:value2)"
//	@Param			debug			query	string	false	"Debug true/false"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	"Status OK"
//	@Failure		400	"Status Bad Request"
//	@Router			/engine/facts/{id}/es [get]
func FactToESQuery(w http.ResponseWriter, r *http.Request) {

	debug := false
	_debug := r.URL.Query().Get("debug")
	if _debug == "true" {
		debug = true
	}

	t, err := ParseTime(r.URL.Query().Get("time"))
	if err != nil {
		zap.L().Error("Parse input time", zap.Error(err), zap.String("rawTime", r.URL.Query().Get("time")))
		httputil.Error(w, r, httputil.ErrAPIParsingDateTime, err)
		return
	}

	nhit, err := ParseInt(r.URL.Query().Get("nhit"))
	if err != nil {
		zap.L().Error("Parse input nhit", zap.Error(err), zap.String("rawNhit", r.URL.Query().Get("nhit")))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	offset, err := ParseInt(r.URL.Query().Get("offset"))
	if err != nil {
		zap.L().Error("Parse input offset", zap.Error(err), zap.String("raw offset", r.URL.Query().Get("offset")))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	placeholders, err := QueryParamToOptionalKeyValues(r, "placeholders", make(map[string]interface{}))
	if err != nil {
		zap.L().Warn("Parse input placeholders", zap.Error(err), zap.String("raw placeholders", r.URL.Query().Get("placeholders")))
		httputil.Error(w, r, httputil.ErrAPIParsingKeyValue, err)
		return
	}

	situationid, err := ParseInt(r.URL.Query().Get("situationid"))
	if err != nil {
		zap.L().Error("Parse input situationid", zap.Error(err), zap.String("rawsituationid", r.URL.Query().Get("situationid")))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	instanceid, err := ParseInt(r.URL.Query().Get("instanceid"))
	if err != nil {
		zap.L().Error("Parse input instanceid", zap.Error(err), zap.String("rawinstanceid", r.URL.Query().Get("instanceid")))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	byName := false
	_byName := r.URL.Query().Get("byName")
	if _byName == "true" {
		byName = true
	}

	id := chi.URLParam(r, "id")
	f, apiError, err := lookupFact(byName, id)
	if err != nil {
		httputil.Error(w, r, apiError, err)
		return
	}

	parameters := make(map[string]interface{})
	if situationid != 0 {
		s, found, err := situation2.R().Get(int64(situationid))
		if err != nil {
			httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
			return
		}
		if !found {
			httputil.Error(w, r, httputil.ErrAPIDBResourceNotFound, nil)
			return
		}
		for k, v := range s.Parameters {
			parameters[k] = v
		}

		if s.IsTemplate && instanceid != 0 {
			template, found, err := situation2.R().GetTemplateInstance(int64(instanceid))
			if err != nil {
				httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
				return
			}
			if !found {
				httputil.Error(w, r, httputil.ErrAPIDBResourceNotFound, nil)
				return
			}
			for k, v := range template.Parameters {
				parameters[k] = v
			}
		}
	}

	for k, v := range placeholders {
		parameters[k] = v
	}

	zap.L().Debug("Use elasticsearch to resolve query")
	if debug {
		zap.L().Debug("Debugging fact", zap.Any("f", f))
	}

	// Add context to fact, replace params and evaluate queries
	f.ContextualizeDimensions(t)
	err = f.ContextualizeCondition(t, parameters)
	if err != nil {
		httputil.Error(w, r, apiError, err)
		return
	}

	source, err := elasticsearch.ConvertFactToSearchRequestV8(f, t, parameters)
	if err != nil {
		zap.L().Error("Cannot convert fact to search request", zap.Error(err), zap.Any("fact", f))
		httputil.Error(w, r, httputil.ErrAPIResourceInvalid, err)
		return
	}

	source.Size = &nhit
	source.From = &offset

	zap.L().Info("Debugging final elastic query", zap.Any("query", source))

	httputil.JSON(w, r, source)
}

func lookupFact(byName bool, id string) (engine.Fact, httputil.APIError, error) {
	var f engine.Fact
	var err error
	var found bool
	if byName {
		f, found, err = fact2.R().GetByName(id)
		if err != nil {
			zap.L().Error("Error while fetching fact", zap.String("factid", id), zap.Error(err))
			return engine.Fact{}, httputil.ErrAPIDBSelectFailed, err
		}
		if !found {
			zap.L().Warn("Fact does not exists", zap.String("factid", id))
			return engine.Fact{}, httputil.ErrAPIDBResourceNotFound, fmt.Errorf("fact not found with name %s", id)
		}
	} else {
		idFact, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			zap.L().Warn("Error on parsing fact id", zap.String("factID", id), zap.Error(err))
			return engine.Fact{}, httputil.ErrAPIParsingInteger, err
		}
		f, found, err = fact2.R().Get(idFact)
		if err != nil {
			zap.L().Error("Error while fetching fact", zap.Int64("factid", idFact), zap.Error(err))
			return engine.Fact{}, httputil.ErrAPIDBSelectFailed, err
		}
		if !found {
			zap.L().Warn("Fact does not exists", zap.String("factid", id))
			return engine.Fact{}, httputil.ErrAPIDBResourceNotFound, fmt.Errorf("fact not found with id %d", idFact)
		}
	}
	return f, httputil.APIError{}, nil
}
