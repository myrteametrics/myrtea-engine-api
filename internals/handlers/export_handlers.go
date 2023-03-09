package handlers

import (
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/export"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/fact"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/handlers/render"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/security/permissions"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"time"
)

// ExportFact godoc
// @Summary Export facts
// @Description Get all action definitions
// @Tags ExportFact
// @Produce file
// @Security Bearer
// @Success 200 {file} Returns data to be saved into a file
// @Failure 500 "internal server error"
// @Router /engine/export/facts/{id}?type=csv [get]
func ExportFact(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	idFact, err := strconv.ParseInt(id, 10, 64)

	if err != nil {
		zap.L().Warn("Error on parsing fact id", zap.String("idFact", id), zap.Error(err))
		render.Error(w, r, render.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r) // TODO: set the right permission
	if !userCtx.HasPermission(permissions.New(permissions.TypeFact, strconv.FormatInt(idFact, 10), permissions.ActionGet)) {
		render.Error(w, r, render.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	f, found, err := fact.R().Get(idFact)
	if err != nil {
		zap.L().Error("Cannot retrieve fact", zap.Int64("factID", idFact), zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("fact does not exists", zap.Int64("factID", idFact))
		render.Error(w, r, render.ErrAPIDBResourceNotFound, err)
		return
	}

	var file []byte

	// type checker, to handle other files types like xml or json (defaults to csv)
	switch r.URL.Query().Get("type") {
	default:
		// Process needed parameters
		columns := []string{"test"}
		columnsLabel := []string{"test"}
		separator := ','

		fullHits, err := export.ExportFactHitsFull(idFact)

		if err != nil {
			zap.L().Error("Error getting actions", zap.Error(err))
			render.Error(w, r, render.ErrAPIProcessError, err)
			return
		}

		file, err = export.ConvertHitsToCSV(fullHits, columns, columnsLabel, separator)
		break
	}

	filename := f.Name + "_export_" + time.Now().Format("02_01_2006_15_04") + ".csv"

	render.File(w, filename, file)
}
