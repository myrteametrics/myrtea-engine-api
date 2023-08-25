package handlers

import (
	"errors"
	"github.com/myrteametrics/myrtea-sdk/v4/engine"
	"net/http"
	"strconv"
	"strings"
	"sync"
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
	streamed          bool
	chunkSize         int64
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
		if filename == "" {
			filename = f.Name + "_export_" + time.Now().Format("02_01_2006_15-04") + ".csv"
		}

		if params.streamed {
			err = HandleStreamedExport(w, f, filename, params)
			if err != nil {
				render.Error(w, r, render.ErrAPIProcessError, err)
			}
		} else {
			file, err = export.ConvertHitsToCSV(fullHits, params.columns, params.columnsLabel, params.formatColumnsData, params.separator)
			render.File(w, filename, file)
		}

		break
	}

}

func GetCSVParameters(r *http.Request) CSVParameters {
	result := CSVParameters{separator: ','}

	streamed, err := QueryParamToOptionalBool(r, "streamed", false)
	if err != nil {
		result.streamed = false
	} else {
		result.streamed = streamed
	}

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

// HandleStreamedExport actually only handles CSV
func HandleStreamedExport(w http.ResponseWriter, f engine.Fact, fileName string, params CSVParameters) error {
	w.Header().Set("Connection", "Keep-Alive")
	w.Header().Set("Transfer-Encoding", "chunked")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Disposition", "attachment; filename="+strconv.Quote(fileName))
	w.Header().Set("Content-Type", "application/octet-stream")
	streamedExport := export.NewStreamedExport(params.chunkSize)
	var wg sync.WaitGroup

	flusher, ok := w.(http.Flusher)
	if !ok {
		return errors.New("expected http.ResponseWriter to be an http.Flusher")
	}

	// TODO: HANDLE COMBINED FACTS

	// Increment the WaitGroup counter
	wg.Add(2)

	var err error = nil

	go func() {
		defer wg.Done()
		err = streamedExport.StreamedExportFactHitsFullV8(f)
		if err != nil {
			wg.Done() // End WaitGroup since an error happened
			return
		}
	}()

	// Chunk handler goroutine
	go func() {
		defer wg.Done()
		first := true
		labels := params.columnsLabel

		for hits := range streamedExport.Data {
			data, err := export.ConvertHitsToCSV(hits, params.columns, labels, params.formatColumnsData, params.separator)

			if err != nil {
				// TODO: how to handle error? stop all or continue and ignore
				continue
			}

			// Write data
			w.Write(data)
			// Flush data to be sent directly to browser
			flusher.Flush()

			if first {
				first = false
				labels = []string{}
			}
		}

	}()

	wg.Wait()
	return err
}
