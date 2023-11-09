package handlers

import (
	"context"
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

type ExportHandler struct {
	exportWrapper *export.ExportWrapper
}

func NewExportHandler(exportWrapper *export.ExportWrapper) *ExportHandler {
	return &ExportHandler{
		exportWrapper: exportWrapper,
	}
}

// ExportFactStreamed godoc
// @Summary CSV streamed export facts in chunks
// @Description CSV Streamed export for facts in chunks
// @Tags ExportFactStreamed
// @Produce octet-stream
// @Security Bearer
// @Success 200 {file} Returns data to be saved into a file
// @Failure 500 "internal server error"
// @Router /engine/export/facts/{id} [get]
func ExportFactStreamed(w http.ResponseWriter, r *http.Request) {
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

	var filename = r.URL.Query().Get("fileName")
	if filename == "" {
		filename = f.Name + "_export_" + time.Now().Format("02_01_2006_15-04") + ".csv"
	}

	// suppose that type is csv
	params := GetCSVParameters(r)

	var combineFacts []engine.Fact
	combineFacts = append(combineFacts, f)

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
			combineFacts = append(combineFacts, combineFact)
		}
	}

	err = HandleStreamedExport(r.Context(), w, combineFacts, filename, params)
	if err != nil {
		render.Error(w, r, render.ErrAPIProcessError, err)
	}
	return

}

func GetCSVParameters(r *http.Request) export.CSVParameters {
	result := export.CSVParameters{Separator: ','}

	limit, err := QueryParamToOptionalInt64(r, "limit", -1)
	if err != nil {
		result.Limit = -1
	} else {
		result.Limit = limit
	}

	result.Columns = QueryParamToOptionalStringArray(r, "columns", ",", []string{})
	result.ColumnsLabel = QueryParamToOptionalStringArray(r, "columnsLabel", ",", []string{})

	formatColumnsData := QueryParamToOptionalStringArray(r, "formateColumns", ",", []string{})
	result.FormatColumnsData = make(map[string]string)
	for _, formatData := range formatColumnsData {
		parts := strings.Split(formatData, ";")
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		result.FormatColumnsData[key] = parts[1]
	}
	separator := r.URL.Query().Get("separator")
	if separator != "" {
		sep, size := utf8.DecodeRuneInString(separator)
		if size != 1 {
			result.Separator = ','
		} else {
			result.Separator = sep
		}
	}

	return result
}

// HandleStreamedExport actually only handles CSV
func HandleStreamedExport(requestContext context.Context, w http.ResponseWriter, facts []engine.Fact, fileName string, params export.CSVParameters) error {
	w.Header().Set("Connection", "Keep-Alive")
	w.Header().Set("Transfer-Encoding", "chunked")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Disposition", "attachment; filename="+strconv.Quote(fileName))
	w.Header().Set("Content-Type", "application/octet-stream")
	streamedExport := export.NewStreamedExport()
	var wg sync.WaitGroup

	flusher, ok := w.(http.Flusher)
	if !ok {
		return errors.New("expected http.ResponseWriter to be an http.Flusher")
	}

	// Increment the WaitGroup counter
	wg.Add(2) // 2 goroutines

	var err error = nil
	var writerErr error = nil
	ctx, cancel := context.WithCancel(context.Background())

	/**
	 * How streamed export works:
	 * 1. Browser opens connection
	 * 2. Two goroutines are started:
	 *    - Export goroutine: each fact is processed one by one
	 *      Each bulk of data is sent through a channel to the receiver goroutine
	 *    - The receiver handles the incoming channel data and converts them to the CSV format
	 *      After the conversion, the data is written and sent to the browser
	 */

	go func() {
		defer wg.Done()
		defer close(streamedExport.Data)

		for _, f := range facts {
			writerErr = streamedExport.StreamedExportFactHitsFull(ctx, f, params.Limit)
			if writerErr != nil {
				zap.L().Error("Error during export (StreamedExportFactHitsFullV8)", zap.Error(err))
				break // break here when error occurs?
			}
		}

	}()

	// Chunk handler goroutine
	go func() {
		defer wg.Done()
		first := true
		labels := params.ColumnsLabel

		for {
			select {
			case hits, ok := <-streamedExport.Data:
				if !ok { // channel closed
					return
				}

				data, err := export.ConvertHitsToCSV(hits, params.Columns, labels, params.FormatColumnsData, params.Separator)

				if err != nil {
					zap.L().Error("ConvertHitsToCSV error during export (StreamedExportFactHitsFullV8)", zap.Error(err))
					cancel()
					return
				}

				// Write data
				_, err = w.Write(data)
				if err != nil {
					zap.L().Error("Write error during export (StreamedExportFactHitsFullV8)", zap.Error(err))
					cancel()
					return
				}
				// Flush data to be sent directly to browser
				flusher.Flush()

				if first {
					first = false
					labels = []string{}
				}

			case <-requestContext.Done():
				// Browser unexpectedly closed connection
				writerErr = errors.New("browser unexpectedly closed connection")
				cancel()
				return
			}
		}
	}()

	wg.Wait()

	// Writer could have some errors
	if writerErr != nil {
		return writerErr
	}

	return err
}

// GetFacts godoc
// @Summary Get all user exports
// @Description Get all user exports
// @Tags Exports
// @Produce json
// @Security Bearer
// @Success 200 "Status OK"
// @Failure 500 "internal server error"
// @Router /service/exports [post]
func (e *ExportHandler) GetExports(w http.ResponseWriter, r *http.Request) {

}

// GetFacts godoc
// @Summary Get all user exports
// @Description Get all user exports
// @Tags Exports
// @Produce json
// @Security Bearer
// @Success 200 "Status OK"
// @Failure 500 "internal server error"
// @Router /service/exports/{id} [post]
func (e *ExportHandler) GetExport(w http.ResponseWriter, r *http.Request) {

}

// GetFacts godoc
// @Summary Get all user exports
// @Description Get all user exports
// @Tags Exports
// @Produce json
// @Security Bearer
// @Success 200 "Status OK"
// @Failure 500 "internal server error"
// @Router /service/exports/{id} [post]
func (e *ExportHandler) DeleteExport(w http.ResponseWriter, r *http.Request) {

}

// GetFacts godoc
// @Summary Get all user exports
// @Description Get all user exports
// @Tags Exports
// @Produce json
// @Security Bearer
// @Success 200 "Status OK"
// @Failure 500 "internal server error"
// @Router /service/exports/fact/{id} [post]
func (e *ExportHandler) ExportFact(w http.ResponseWriter, r *http.Request) {

}
