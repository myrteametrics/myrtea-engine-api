package handler

import (
	"encoding/json"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/utils/httputil"
	"net/http"
	"sort"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/explainer/rootcause"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/model"
	"go.uber.org/zap"
)

// GetRootCauses godoc
//
//	@Summary		Get all rootcause definitions
//	@Description	Get all rootcause definitions
//	@Tags			RootCauses
//	@Produce		json
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{array}	model.RootCause	"list of rootcauses"
//	@Failure		500	"internal server error"
//	@Router			/engine/rootcauses [get]
func GetRootCauses(w http.ResponseWriter, r *http.Request) {
	rootcauses, err := rootcause.R().GetAll()
	if err != nil {
		zap.L().Error("Error getting rootcauses", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}

	rootcausesSlice := make([]model.RootCause, 0)
	for _, rootcause := range rootcauses {
		rootcausesSlice = append(rootcausesSlice, rootcause)
	}

	sort.SliceStable(rootcausesSlice, func(i, j int) bool {
		return rootcausesSlice[i].ID < rootcausesSlice[j].ID
	})

	httputil.JSON(w, r, rootcausesSlice)
}

// GetRootCause godoc
//
//	@Summary		Get a rootcause definition
//	@Description	Get a rootcause definition
//	@Tags			RootCauses
//	@Produce		json
//	@Param			id	path	string	true	"RootCause ID"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	model.RootCause	"rootcause"
//	@Failure		400	"Status Bad Request"
//	@Router			/engine/rootcauses/{id} [get]
func GetRootCause(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idRootCause, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("Error on parsing rootcause id", zap.String("rootcauseID", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	rc, found, err := rootcause.R().Get(idRootCause)
	if err != nil {
		zap.L().Error("Get Rootcause by ID", zap.Int64("rootcauseid", idRootCause), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Error("RootCause does not exists", zap.Int64("rootcauseid", idRootCause))
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFound, err)
		return
	}

	httputil.JSON(w, r, rc)
}

// ValidateRootCause godoc
//
//	@Summary		Validate a new rootcause definition
//	@Description	Validate a new rootcause definition
//	@Tags			RootCauses
//	@Accept			json
//	@Produce		json
//	@Param			rootcause	body	model.RootCause	true	"RootCause definition (json)"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	model.RootCause	"rootcause"
//	@Failure		400	"Status Bad Request"
//	@Failure		500	"Status"	internal	server	error"
//	@Router			/engine/rootcauses/validate [post]
func ValidateRootCause(w http.ResponseWriter, r *http.Request) {
	var newRootCause model.RootCause
	err := json.NewDecoder(r.Body).Decode(&newRootCause)
	if err != nil {
		zap.L().Warn("Rootcause json decoding", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDecodeJSONBody, err)
		return
	}

	if ok, err := newRootCause.IsValid(); !ok {
		zap.L().Warn("Rootcause is not valid", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIResourceInvalid, err)
		return
	}

	httputil.JSON(w, r, newRootCause)
}

// PostRootCause godoc
//
//	@Summary		Create a new rootcause definition
//	@Description	Create a new rootcause definition
//	@Tags			RootCauses
//	@Accept			json
//	@Produce		json
//	@Param			rootcause	body	model.RootCause	true	"RootCause definition (json)"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	model.RootCause	"rootcause"
//	@Failure		400	"Status Bad Request"
//	@Failure		500	"Status"	internal	server	error"
//	@Router			/engine/rootcauses [post]
func PostRootCause(w http.ResponseWriter, r *http.Request) {
	var newRootCause model.RootCause
	err := json.NewDecoder(r.Body).Decode(&newRootCause)
	if err != nil {
		zap.L().Warn("Rootcause json decoding", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDecodeJSONBody, err)
		return
	}

	if ok, err := newRootCause.IsValid(); !ok {
		zap.L().Warn("Rootcause is not valid", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIResourceInvalid, err)
		return
	}

	idRootCause, err := rootcause.R().Create(nil, newRootCause)
	if err != nil {
		zap.L().Error("Error while creating the RootCause", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBInsertFailed, err)
		return
	}

	rc, found, err := rootcause.R().Get(idRootCause)
	if err != nil {
		zap.L().Error("Get Rootcause by ID", zap.Int64("rootcauseid", idRootCause), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Error("RootCause does not exists", zap.Int64("rootcauseid", idRootCause))
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFoundAfterInsert, err)
		return
	}

	httputil.JSON(w, r, rc)
}

// PutRootCause godoc
//
//	@Summary		Create or remplace a rootcause definition
//	@Description	Create or remplace a rootcause definition
//	@Tags			RootCauses
//	@Accept			json
//	@Produce		json
//	@Param			id			path	string			true	"RootCause ID"
//	@Param			rootcause	body	model.RootCause	true	"RootCause definition (json)"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	model.RootCause	"rootcause"
//	@Failure		400	"Status Bad Request"
//	@Failure		500	"Status"	internal	server	error"
//	@Router			/engine/rootcauses/{id} [put]
func PutRootCause(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idRootCause, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("Error on parsing rootcause id", zap.String("rootcauseID", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	var newRootCause model.RootCause
	err = json.NewDecoder(r.Body).Decode(&newRootCause)
	if err != nil {
		zap.L().Warn("Rootcause json decoding", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDecodeJSONBody, err)
		return
	}
	newRootCause.ID = idRootCause

	if ok, err := newRootCause.IsValid(); !ok {
		zap.L().Warn("Rootcause is not valid", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIResourceInvalid, err)
		return
	}

	err = rootcause.R().Update(nil, idRootCause, newRootCause)
	if err != nil {
		zap.L().Error("Error while updating the RootCause", zap.Int64("idRootCause", idRootCause), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBUpdateFailed, err)
		return
	}

	rc, found, err := rootcause.R().Get(idRootCause)
	if err != nil {
		zap.L().Error("Get Rootcause by ID", zap.Int64("rootcauseid", idRootCause), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Error("RootCause does not exists", zap.Int64("rootcauseid", idRootCause))
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFoundAfterInsert, err)
		return
	}

	httputil.JSON(w, r, rc)
}

// DeleteRootCause godoc
//
//	@Summary		Delete a rootcause definition
//	@Description	Delete a rootcause definition
//	@Tags			RootCauses
//	@Produce		json
//	@Param			id	path	string	true	"RootCause ID"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	"Status OK"
//	@Failure		400	"Status Bad Request"
//	@Router			/engine/rootcauses/{id} [delete]
func DeleteRootCause(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idRootCause, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("Error on parsing rootcause id", zap.String("rootcauseID", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	err = rootcause.R().Delete(nil, idRootCause)
	if err != nil {
		zap.L().Error("Error while deleting the RootCause", zap.Int64("RootCause ID", idRootCause), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBDeleteFailed, err)
		return
	}

	httputil.OK(w, r)
}
