package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/go-chi/chi/v5"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/export"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/fact"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/handlers/render"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/security/permissions"
	"go.uber.org/zap"
)

type CSVParameters struct {
	columns           []string
	columnsLabel      []string
	formatColumnsData map[string]string
	separator         rune
}

// ExportFact godoc
// @Summary Export facts
// @Description Get all action definitions
// @Tags ExportFact
// @Produce octet-stream
// @Security Bearer
// @Success 200 {file} Returns data to be saved into a file
// @Failure 500 "internal server error"
// @Router /engine/export/facts/{id} [get]
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
		zap.L().Warn("fact does not exist", zap.Int64("factID", idFact))
		render.Error(w, r, render.ErrAPIDBResourceNotFound, err)
		return
	}

	fullHits, err := export.ExportFactHitsFull(f)

	if err != nil {
		zap.L().Error("Error getting fact hits", zap.Error(err))
		render.Error(w, r, render.ErrAPIProcessError, err)
		return
	}

	// export multiple facts into one file
	combineFactIds, err := QueryParamToOptionalInt64Array(r, "combineFactIds", ",", false, []int64{})
	if err != nil {
		zap.L().Warn("Could not parse parameter combineFactIds", zap.Error(err))
	} else {

		for _, factId := range combineFactIds {
			// no duplicates
			if factId == idFact {
				continue
			}

			combineFact, found, err := fact.R().Get(factId)
			if err != nil {
				zap.L().Error("Export combineFact cannot retrieve fact", zap.Int64("factID", factId), zap.Error(err))
				continue
			}
			if !found {
				zap.L().Warn("Export combineFact fact does not exist", zap.Int64("factID", factId))
				continue
			}

			combineFullHits, err := export.ExportFactHitsFull(combineFact)

			if err != nil {
				zap.L().Error("Export combineFact getting fact hits", zap.Int64("factID", factId), zap.Error(err))
			} else {
				fullHits = append(fullHits, combineFullHits...)
			}

		}

	}

	var file []byte
	var filename = r.URL.Query().Get("fileName")

	// type checker, to handle other files types like xml or json (defaults to csv)
	switch r.URL.Query().Get("type") {
	default:
		// Process needed parameters
		params := GetCSVParameters(r)
		file, err = export.ConvertHitsToCSV(fullHits, params.columns, params.columnsLabel, params.formatColumnsData, params.separator)
		if filename == "" {
			filename = f.Name + "_export_" + time.Now().Format("02_01_2006_15-04") + ".csv"
		}
		break
	}

	render.File(w, filename, file)
}

func GetCSVParameters(r *http.Request) CSVParameters {
	result := CSVParameters{separator: ','}

	result.columns = QueryParamToOptionalStringArray(r, "columns", ",", []string{})
	result.columnsLabel = QueryParamToOptionalStringArray(r, "columnsLabel", ",", []string{})

	formatColumnsData := QueryParamToOptionalStringArray(r, "formateColumns", ",", []string{})
	result.formatColumnsData = make(map[string]string)
	for _, formatData := range formatColumnsData {
		parts := strings.Split(formatData, ";")
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		result.formatColumnsData[key] = parts[1]
	}
	separator := r.URL.Query().Get("separator")
	if separator != "" {
		sep, size := utf8.DecodeRuneInString(separator)
		if size != 1 {
			result.separator = ','
		} else {
			result.separator = sep
		}
	}

	return result
}
