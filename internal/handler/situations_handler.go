package handler

import (
	"encoding/json"
	"errors"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/fact"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/security/permissions"
	situation2 "github.com/myrteametrics/myrtea-engine-api/v5/pkg/situation"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/utils/httputil"
	"github.com/myrteametrics/myrtea-sdk/v5/handlers/render"
	"net/http"
	"sort"
	"strconv"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// GetSituations godoc
//
//	@Id				GetSituations
//
//	@Summary		Get all situation definitions
//	@Description	Get all situation definitions
//	@Tags			Situations
//	@Produce		json
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{array}	situation.Situation	"list of situations"
//	@Failure		500	"internal server error"
//	@Router			/engine/situations [get]
func GetSituations(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeSituation, permissions.All, permissions.ActionList)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	var situations map[int64]situation2.Situation
	var err error
	if userCtx.HasPermission(permissions.New(permissions.TypeSituation, permissions.All, permissions.ActionGet)) {
		situations, err = situation2.R().GetAll(gvalParsingEnabled(r.URL.Query()))
	} else {
		resourceIDs := userCtx.GetMatchingResourceIDsInt64(permissions.New(permissions.TypeSituation, permissions.All, permissions.ActionGet))
		situations, err = situation2.R().GetAllByIDs(resourceIDs, gvalParsingEnabled(r.URL.Query()))
	}
	if err != nil {
		zap.L().Warn("Cannot retrieve situations", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}
	situationsSlice := make([]situation2.Situation, 0)
	for _, schedule := range situations {
		situationsSlice = append(situationsSlice, schedule)
	}

	sort.SliceStable(situationsSlice, func(i, j int) bool {
		return situationsSlice[i].ID < situationsSlice[j].ID
	})

	httputil.JSON(w, r, situationsSlice)
}

// GetSituation godoc
//
//	@Id				GetSituation
//
//	@Summary		Get a situation definition
//	@Description	Get a situation definition
//	@Tags			Situations
//	@Produce		json
//	@Param			id	path	string	true	"Situation ID"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	situation.Situation	"situation"
//	@Failure		400	"Status Bad Request"
//	@Router			/engine/situations/{id} [get]
func GetSituation(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idSituation, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("Error on parsing situation id", zap.String("situationID", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeSituation, strconv.FormatInt(idSituation, 10), permissions.ActionGet)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	situation, found, err := situation2.R().Get(idSituation, gvalParsingEnabled(r.URL.Query()))
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

	httputil.JSON(w, r, situation)
}

// GetSituationOverview godoc
//
//	@Id				GetSituationOverview
//
//	@Summary		Get situation overview
//	@Description	Get situation overview
//	@Tags			Situations
//	@Produce		json
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{array}	situation.SituationOverview	"situation overview"
//	@Failure		500	"internal server error"
//	@Router			/engine/situations/overview [get]
func GetSituationOverview(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeSituation, permissions.All, permissions.ActionList)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	situations, err := situation2.R().GetSituationOverview()
	if err != nil {
		zap.L().Warn("Cannot retrieve situations", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}

	render.JSON(w, r, situations)
}

// ValidateSituation godoc
//
//	@Id				ValidateSituation
//
//	@Summary		Validate a new situation definition
//	@Description	Validate a new situation definition
//	@Tags			Situations
//	@Accept			json
//	@Produce		json
//	@Param			situation	body	situation.Situation	true	"Situation definition (json)"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	situation.Situation	"situation"
//	@Failure		400	"Status Bad Request"
//	@Failure		500	"Status"	internal	server	error"
//	@Router			/engine/situations/validate [post]
func ValidateSituation(w http.ResponseWriter, r *http.Request) {
	var newSituation situation2.Situation
	err := json.NewDecoder(r.Body).Decode(&newSituation)
	if err != nil {
		zap.L().Warn("Situation json decode", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDecodeJSONBody, err)
		return
	}

	if ok, err := newSituation.IsValid(); !ok {
		zap.L().Warn("Situation is invalid", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIResourceInvalid, err)
		return
	}

	httputil.JSON(w, r, newSituation)
}

// PostSituation godoc
//
//	@Id				PostSituation
//
//	@Summary		Creates a situation definition
//	@Description	Creates a situation definition
//	@Tags			Situations
//	@Accept			json
//	@Produce		json
//	@Param			factsByName	query	string				false	"Find fact by it's name"
//	@Param			situation	body	situation.Situation	true	"Situation definition (json)"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	situation.Situation	"situation"
//	@Failure		400	"Status Bad Request"
//	@Failure		500	"Status"	internal	server	error"
//	@Router			/engine/situations [post]
func PostSituation(w http.ResponseWriter, r *http.Request) {

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeSituation, permissions.All, permissions.ActionCreate)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	factsByName := false
	_factByName := r.URL.Query().Get("factsByName")
	if _factByName == "true" {
		factsByName = true
	}

	var newSituation situation2.Situation
	if factsByName {
		type situationWithFactsName struct {
			ID         int64                  `json:"id,omitempty"`
			Name       string                 `json:"name"`
			Facts      []string               `json:"facts"`
			CalendarID int64                  `json:"calendarId"`
			Groups     []int64                `json:"groups"`
			Parameters map[string]interface{} `json:"parameters"`
			IsTemplate bool                   `json:"isTemplate"`
			IsObject   bool                   `json:"isObject"`
		}
		var s situationWithFactsName
		err := json.NewDecoder(r.Body).Decode(&s)
		if err != nil {
			zap.L().Warn("Situation json decode", zap.Error(err))
			httputil.Error(w, r, httputil.ErrAPIDecodeJSONBody, err)
			return
		}
		factIDs := make([]int64, 0)
		for _, name := range s.Facts {
			f, found, err := fact.R().GetByName(name)
			if err != nil {
				zap.L().Error("Get fact by name", zap.String("name", name), zap.Error(err))
				httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
				return
			}
			if !found {
				zap.L().Error("fact not found", zap.String("name", name))
				httputil.Error(w, r, httputil.ErrAPIDBResourceNotFound, err)
				return
			}
			factIDs = append(factIDs, f.ID)
		}
		newSituation = situation2.Situation{
			ID:         s.ID,
			Name:       s.Name,
			Facts:      factIDs,
			CalendarID: s.CalendarID,
			Parameters: s.Parameters,
			IsTemplate: s.IsTemplate,
			IsObject:   s.IsObject,
		}
	} else {
		err := json.NewDecoder(r.Body).Decode(&newSituation)
		if err != nil {
			zap.L().Warn("Situation json decode", zap.Error(err))
			httputil.Error(w, r, httputil.ErrAPIDecodeJSONBody, err)
			return
		}
	}

	if ok, err := newSituation.IsValid(); !ok {
		zap.L().Warn("Situation is invalid", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIResourceInvalid, err)
		return
	}

	idSituation, err := situation2.R().Create(newSituation)
	if err != nil {
		zap.L().Error("Error while creating the situation", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBInsertFailed, err)
		return
	}

	situation, found, err := situation2.R().Get(idSituation, gvalParsingEnabled(r.URL.Query()))
	if err != nil {
		zap.L().Error("Cannot retrieve situation", zap.Int64("situationID", idSituation), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("Situation does not exists after creation", zap.Int64("situationID", idSituation))
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFoundAfterInsert, err)
		return
	}

	httputil.JSON(w, r, situation)
}

// PutSituation godoc
//
//	@Id				PutSituation
//
//	@Summary		replace a situation definition
//	@Description	replace a situation definition
//	@Tags			Situations
//	@Accept			json
//	@Produce		json
//	@Param			id			path	string				true	"Situation ID"
//	@Param			situation	body	situation.Situation	true	"Situation definition (json)"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	situation.Situation	"situation"
//	@Failure		400	"Status Bad Request"
//	@Failure		500	"Status"	internal	server	error"
//	@Router			/engine/situations/{id} [put]
func PutSituation(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idSituation, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("Error on parsing situation id", zap.String("situationID", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeSituation, strconv.FormatInt(idSituation, 10), permissions.ActionUpdate)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	var newSituation situation2.Situation
	err = json.NewDecoder(r.Body).Decode(&newSituation)
	if err != nil {
		zap.L().Warn("Situation json decode", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDecodeJSONBody, err)
		return
	}
	newSituation.ID = idSituation

	if ok, err := newSituation.IsValid(); !ok {
		zap.L().Warn("Situation is invalid", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIResourceInvalid, err)
		return
	}

	err = situation2.R().Update(idSituation, newSituation)
	if err != nil {
		zap.L().Info("Error while updating the situation", zap.String("situation ID", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBUpdateFailed, err)
		return
	}

	situation, found, err := situation2.R().Get(idSituation, gvalParsingEnabled(r.URL.Query()))
	if err != nil {
		zap.L().Error("Cannot retrieve situation", zap.Int64("situationID", idSituation), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("Situation does not exists after update", zap.Int64("situationID", idSituation))
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFoundAfterInsert, err)
		return
	}

	httputil.JSON(w, r, situation)
}

// DeleteSituation godoc
//
//	@Id				DeleteSituation
//
//	@Summary		Delete a situation definition
//	@Description	Delete a situation definition
//	@Tags			Situations
//	@Param			id	path	string	true	"Situation ID"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	"Status OK"
//	@Failure		400	"Status Bad Request"
//	@Router			/engine/situations/{id} [delete]
func DeleteSituation(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idSituation, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("Error on parsing situation id", zap.String("situationID", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeSituation, strconv.FormatInt(idSituation, 10), permissions.ActionDelete)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	err = situation2.R().Delete(idSituation)
	if err != nil {
		zap.L().Error("Error while deleting the situation", zap.String("Situation ID", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBDeleteFailed, err)
		return
	}

	httputil.OK(w, r)
}
