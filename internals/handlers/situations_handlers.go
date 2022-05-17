package handlers

import (
	"encoding/json"
	"net/http"
	"sort"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/fact"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/handlers/render"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/situation"
	"go.uber.org/zap"
)

// GetSituations godoc
// @Summary Get all situation definitions
// @Description Get all situation definitions
// @Tags Situations
// @Produce json
// @Security Bearer
// @Success 200 {array} situation.Situation "list of situations"
// @Failure 500 "internal server error"
// @Router /engine/situations [get]
func GetSituations(w http.ResponseWriter, r *http.Request) {
	groups := GetUserGroupsFromContext(r)
	situations, err := situation.R().GetAll(groups)
	if err != nil {
		zap.L().Warn("Cannot retrieve situations", zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}
	situationsSlice := make([]situation.Situation, 0)
	for _, schedule := range situations {
		situationsSlice = append(situationsSlice, schedule)
	}

	sort.SliceStable(situationsSlice, func(i, j int) bool {
		return situationsSlice[i].ID < situationsSlice[j].ID
	})

	render.JSON(w, r, situationsSlice)
}

// GetSituation godoc
// @Summary Get a situation definition
// @Description Get a situation definition
// @Tags Situations
// @Produce json
// @Param id path string true "Situation ID"
// @Security Bearer
// @Success 200 {object} situation.Situation "situation"
// @Failure 400 "Status Bad Request"
// @Router /engine/situations/{id} [get]
func GetSituation(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idSituation, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("Error on parsing situation id", zap.String("situationID", id), zap.Error(err))
		render.Error(w, r, render.ErrAPIParsingInteger, err)
		return
	}

	groups := GetUserGroupsFromContext(r)
	situation, found, err := situation.R().Get(idSituation, groups)
	if err != nil {
		zap.L().Error("Cannot retrieve situation", zap.Int64("situationID", idSituation), zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("Situation does not exists", zap.Int64("situationID", idSituation))
		render.Error(w, r, render.ErrAPIDBResourceNotFound, err)
		return
	}

	render.JSON(w, r, situation)
}

// ValidateSituation godoc
// @Summary Validate a new situation definition
// @Description Validate a new situation definition
// @Tags Situations
// @Accept json
// @Produce json
// @Param situation body situation.Situation true "Situation definition (json)"
// @Security Bearer
// @Success 200 {object} situation.Situation "situation"
// @Failure 400 "Status Bad Request"
// @Failure 500 "Status" internal server error"
// @Router /engine/situations/validate [post]
func ValidateSituation(w http.ResponseWriter, r *http.Request) {
	var newSituation situation.Situation
	err := json.NewDecoder(r.Body).Decode(&newSituation)
	if err != nil {
		zap.L().Warn("Situation json decode", zap.Error(err))
		render.Error(w, r, render.ErrAPIDecodeJSONBody, err)
		return
	}

	if ok, err := newSituation.IsValid(); !ok {
		zap.L().Warn("Situation is invalid", zap.Error(err))
		render.Error(w, r, render.ErrAPIResourceInvalid, err)
		return
	}

	render.JSON(w, r, newSituation)
}

// PostSituation godoc
// @Summary Creates a situation definition
// @Description Creates a situation definition
// @Tags Situations
// @Accept json
// @Produce json
// @Param factsByName query string false "Find fact by it's name"
// @Param situation body situation.Situation true "Situation definition (json)"
// @Security Bearer
// @Success 200 {object} situation.Situation "situation"
// @Failure 400 "Status Bad Request"
// @Failure 500 "Status" internal server error"
// @Router /engine/situations [post]
func PostSituation(w http.ResponseWriter, r *http.Request) {
	factsByName := false
	_factByName := r.URL.Query().Get("factsByName")
	if _factByName == "true" {
		factsByName = true
	}

	var newSituation situation.Situation
	if factsByName {
		type situationWithFactsName struct {
			ID         int64             `json:"id,omitempty"`
			Name       string            `json:"name"`
			Facts      []string          `json:"facts"`
			CalendarID int64             `json:"calendarId"`
			Groups     []int64           `json:"groups"`
			Parameters map[string]string `json:"parameters"`
			IsTemplate bool              `json:"isTemplate"`
			IsObject   bool              `json:"isObject"`
		}
		var s situationWithFactsName
		err := json.NewDecoder(r.Body).Decode(&s)
		if err != nil {
			zap.L().Warn("Situation json decode", zap.Error(err))
			render.Error(w, r, render.ErrAPIDecodeJSONBody, err)
			return
		}
		factIDs := make([]int64, 0)
		for _, name := range s.Facts {
			f, found, err := fact.R().GetByName(name)
			if err != nil {
				zap.L().Error("Get fact by name", zap.String("name", name), zap.Error(err))
				render.Error(w, r, render.ErrAPIDBSelectFailed, err)
				return
			}
			if !found {
				zap.L().Error("fact not found", zap.String("name", name))
				render.Error(w, r, render.ErrAPIDBResourceNotFound, err)
				return
			}
			factIDs = append(factIDs, f.ID)
		}
		newSituation = situation.Situation{
			ID:         s.ID,
			Name:       s.Name,
			Facts:      factIDs,
			CalendarID: s.CalendarID,
			Groups:     s.Groups,
			Parameters: s.Parameters,
			IsTemplate: s.IsTemplate,
			IsObject:   s.IsObject,
		}
	} else {
		err := json.NewDecoder(r.Body).Decode(&newSituation)
		if err != nil {
			zap.L().Warn("Situation json decode", zap.Error(err))
			render.Error(w, r, render.ErrAPIDecodeJSONBody, err)
			return
		}
	}

	if ok, err := newSituation.IsValid(); !ok {
		zap.L().Warn("Situation is invalid", zap.Error(err))
		render.Error(w, r, render.ErrAPIResourceInvalid, err)
		return
	}

	idSituation, err := situation.R().Create(newSituation)
	if err != nil {
		zap.L().Error("Error while creating the situation", zap.Error(err))
		render.Error(w, r, render.ErrAPIDBInsertFailed, err)
		return
	}

	groups := GetUserGroupsFromContext(r)
	situation, found, err := situation.R().Get(idSituation, groups)
	if err != nil {
		zap.L().Error("Cannot retrieve situation", zap.Int64("situationID", idSituation), zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("Situation does not exists after creation", zap.Int64("situationID", idSituation))
		render.Error(w, r, render.ErrAPIDBResourceNotFoundAfterInsert, err)
		return
	}

	render.JSON(w, r, situation)
}

// PutSituation godoc
// @Summary replace a situation definition
// @Description replace a situation definition
// @Tags Situations
// @Accept json
// @Produce json
// @Param id path string true "Situation ID"
// @Param situation body situation.Situation true "Situation definition (json)"
// @Security Bearer
// @Success 200 {object} situation.Situation "situation"
// @Failure 400 "Status Bad Request"
// @Failure 500 "Status" internal server error"
// @Router /engine/situations/{id} [put]
func PutSituation(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idSituation, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("Error on parsing situation id", zap.String("situationID", id), zap.Error(err))
		render.Error(w, r, render.ErrAPIParsingInteger, err)
		return
	}

	var newSituation situation.Situation
	err = json.NewDecoder(r.Body).Decode(&newSituation)
	if err != nil {
		zap.L().Warn("Situation json decode", zap.Error(err))
		render.Error(w, r, render.ErrAPIDecodeJSONBody, err)
		return
	}
	newSituation.ID = idSituation

	if ok, err := newSituation.IsValid(); !ok {
		zap.L().Warn("Situation is invalid", zap.Error(err))
		render.Error(w, r, render.ErrAPIResourceInvalid, err)
		return
	}

	err = situation.R().Update(idSituation, newSituation)
	if err != nil {
		zap.L().Info("Error while updating the situation", zap.String("situation ID", id), zap.Error(err))
		render.Error(w, r, render.ErrAPIDBUpdateFailed, err)
		return
	}

	groups := GetUserGroupsFromContext(r)
	situation, found, err := situation.R().Get(idSituation, groups)
	if err != nil {
		zap.L().Error("Cannot retrieve situation", zap.Int64("situationID", idSituation), zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("Situation does not exists after update", zap.Int64("situationID", idSituation))
		render.Error(w, r, render.ErrAPIDBResourceNotFoundAfterInsert, err)
		return
	}

	render.JSON(w, r, situation)
}

// DeleteSituation godoc
// @Summary Delete a situation definition
// @Description Delete a situation definition
// @Tags Situations
// @Param id path string true "Situation ID"
// @Security Bearer
// @Success 200 "Status OK"
// @Failure 400 "Status Bad Request"
// @Router /engine/situations/{id} [delete]
func DeleteSituation(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idSituation, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("Error on parsing situation id", zap.String("situationID", id), zap.Error(err))
		render.Error(w, r, render.ErrAPIParsingInteger, err)
		return
	}

	err = situation.R().Delete(idSituation)
	if err != nil {
		zap.L().Error("Error while deleting the situation", zap.String("Situation ID", id), zap.Error(err))
		render.Error(w, r, render.ErrAPIDBDeleteFailed, err)
		return
	}

	render.OK(w, r)
}
