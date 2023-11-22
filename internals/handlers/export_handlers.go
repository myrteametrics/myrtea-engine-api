package handlers

import (
	"context"
	"errors"
	"fmt"
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
	exportWrapper *export.Wrapper
}

func NewExportHandler(exportWrapper *export.Wrapper) *ExportHandler {
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

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeExport, strconv.FormatInt(idFact, 10), permissions.ActionGet)) {
		render.Error(w, r, render.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	filename, params, combineFacts, done := handleExportArgs(w, r, err, idFact)
	if done {
		return
	}

	err = HandleStreamedExport(r.Context(), w, combineFacts, filename, params)
	if err != nil {
		render.Error(w, r, render.ErrAPIProcessError, err)
	}
	return

}

// handleExportArgs handles the export arguments and returns the filename, the parameters and the facts to export
// done is true if an error occurred and the response has already been written
func handleExportArgs(w http.ResponseWriter, r *http.Request, err error, idFact int64) (filename string, params export.CSVParameters, combineFacts []engine.Fact, done bool) {
	f, found, err := fact.R().Get(idFact)
	if err != nil {
		zap.L().Error("Cannot retrieve fact", zap.Int64("factID", idFact), zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return "", export.CSVParameters{}, nil, true
	}
	if !found {
		zap.L().Warn("fact does not exist", zap.Int64("factID", idFact))
		render.Error(w, r, render.ErrAPIDBResourceNotFound, err)
		return "", export.CSVParameters{}, nil, true
	}

	filename = r.URL.Query().Get("fileName")
	if filename == "" {
		filename = fmt.Sprintf("%s_export_%s.csv", f.Name, time.Now().Format("02_01_2006"))
	} else {
		filename = fmt.Sprintf("%s_%s.csv", time.Now().Format("02_01_2006"), filename)
	}

	// suppose that type is csv
	params = GetCSVParameters(r)

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
	return filename, params, combineFacts, false
}

// GetCSVParameters returns the parameters for the CSV export
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

// GetExports godoc
// @Summary Get user exports
// @Description Get in memory user exports
// @Produce json
// @Security Bearer
// @Success 200 {array} export.WrapperItem Returns a list of exports
// @Failure 403 "Status Forbidden: missing permission"
// @Failure 500 "internal server error"
// @Router /engine/exports [get]
func (e *ExportHandler) GetExports(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeExport, permissions.All, permissions.ActionList)) {
		render.Error(w, r, render.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}
	render.JSON(w, r, e.exportWrapper.GetUserExports(userCtx.User))
}

// GetExport godoc
// @Summary Get single export from user
// @Description Get single export from user
// @Tags Exports
// @Produce json
// @Security Bearer
// @Success 200 {object} export.WrapperItem "Status OK"
// @Failure 400 "Bad Request: missing export id / id is not an integer"
// @Failure 403 "Status Forbidden: missing permission"
// @Failure 404 "Status Not Found: export not found"
// @Failure 500 "internal server error"
// @Router /service/exports/{id} [get]
func (e *ExportHandler) GetExport(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		render.Error(w, r, render.ErrAPIMissingParam, errors.New("missing id"))
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeExport, permissions.All, permissions.ActionGet)) {
		render.Error(w, r, render.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	item, ok := e.exportWrapper.GetUserExport(id, userCtx.User)
	if !ok {
		render.Error(w, r, render.ErrAPIDBResourceNotFound, errors.New("export not found"))
		return
	}

	render.JSON(w, r, item)
}

// DeleteExport godoc
// @Summary Deletes a single export
// @Description Deletes a single export, when running it is canceled
// @Tags Exports
// @Produce json
// @Security Bearer
// @Success 204 "Status OK"
// @Failure 400 "Bad Request: missing export id / id is not an integer"
// @Failure 403 "Status Forbidden: missing permission"
// @Failure 404 "Status Not Found: export not found"
// @Failure 500 "internal server error"
// @Router /service/exports/{id} [delete]
func (e *ExportHandler) DeleteExport(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		render.Error(w, r, render.ErrAPIMissingParam, errors.New("missing id"))
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeExport, permissions.All, permissions.ActionDelete)) {
		render.Error(w, r, render.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	ok := e.exportWrapper.DeleteExport(id, userCtx.User)
	if !ok {
		render.Error(w, r, render.ErrAPIDBResourceNotFound, errors.New("export not found"))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ExportFact godoc
// @Summary Creates a new export request for a fact (or multiple facts)
// @Description Creates a new export request for a fact (or multiple facts)
// @Tags Exports
// @Produce json
// @Security Bearer
// @Param id path string true "Fact ID"
// @Param fileName query string false "File name"
// @Param limit query int false "Limit"
// @Param columns query string false "Columns"
// @Param columnsLabel query string false "Columns label"
// @Param formateColumns query string false "Formate columns"
// @Param separator query string false "Separator"
// @Success 200 {object} export.WrapperItem "Status OK: user was added to existing export in queue"
// @Success 201 {object} export.WrapperItem "Status Created: new export was added in queue"
// @Failure 400 "Bad Request: missing fact id / fact id is not an integer"
// @Failure 403 "Status Forbidden: missing permission"
// @Failure 409 {object} export.WrapperItem "Status Conflict: user already exists in export queue"
// @Failure 429 "Status Too Many Requests: export queue is full"
// @Failure 500 "internal server error"
// @Router /service/exports/fact/{id} [post]
func (e *ExportHandler) ExportFact(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idFact, err := strconv.ParseInt(id, 10, 64)

	if err != nil {
		zap.L().Warn("Error on parsing fact id", zap.String("idFact", id), zap.Error(err))
		render.Error(w, r, render.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeExport, permissions.All, permissions.ActionCreate)) {
		render.Error(w, r, render.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	filename, params, combinedFacts, done := handleExportArgs(w, r, err, idFact)
	if done {
		return
	}

	item, status := e.exportWrapper.AddToQueue(combinedFacts, filename, params, userCtx.User)

	switch status {
	case export.CodeAdded:
		w.WriteHeader(http.StatusCreated)
	case export.CodeUserAdded:
		w.WriteHeader(http.StatusOK)
	case export.CodeUserExists:
		w.WriteHeader(http.StatusConflict)
	case export.CodeQueueFull:
		render.Error(w, r, render.ErrAPIQueueFull, fmt.Errorf("export queue is full"))
		return
	default:
		render.Error(w, r, render.ErrAPIProcessError, fmt.Errorf("unknown status code (%d)", status))
		return
	}

	render.JSON(w, r, item)
}
