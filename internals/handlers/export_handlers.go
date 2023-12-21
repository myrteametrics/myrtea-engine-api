package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/export"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/handlers/render"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/security/permissions"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"sync"
)

type ExportHandler struct {
	exportWrapper *export.Wrapper
}

// NewExportHandler returns a new ExportHandler
func NewExportHandler(exportWrapper *export.Wrapper) *ExportHandler {
	return &ExportHandler{
		exportWrapper: exportWrapper,
	}
}

// ExportRequest represents a request for an export
type ExportRequest struct {
	export.CSVParameters
	FactIDs []int64 `json:"factIDs"`
	Title   string  `json:"title"`
}

// ExportFactStreamed godoc
// @Summary CSV streamed export facts in chunks
// @Description CSV Streamed export for facts in chunks
// @Tags ExportFactStreamed
// @Produce octet-stream
// @Security Bearer
// @Param request body handlers.ExportRequest true "request (json)"
// @Success 200 {file} Returns data to be saved into a file
// @Failure 500 "internal server error"
// @Router /engine/facts/streamedexport [post]
func ExportFactStreamed(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeExport, permissions.All, permissions.ActionGet)) {
		render.Error(w, r, render.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	var request ExportRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		zap.L().Warn("Decode export request json", zap.Error(err))
		render.Error(w, r, render.ErrAPIDecodeJSONBody, err)
		return
	}

	if len(request.FactIDs) == 0 {
		zap.L().Warn("Missing factIDs in export request")
		render.Error(w, r, render.ErrAPIMissingParam, errors.New("missing factIDs"))
		return
	}

	err = HandleStreamedExport(r.Context(), w, request)
	if err != nil {
		render.Error(w, r, render.ErrAPIProcessError, err)
	}
	return

}

// HandleStreamedExport actually only handles CSV
func HandleStreamedExport(requestContext context.Context, w http.ResponseWriter, request ExportRequest) error {
	w.Header().Set("Connection", "Keep-Alive")
	w.Header().Set("Transfer-Encoding", "chunked")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Disposition", "attachment; filename="+strconv.Quote(request.Title+".csv"))
	w.Header().Set("Content-Type", "application/octet-stream")

	facts := findCombineFacts(request.FactIDs)
	if len(facts) == 0 {
		return errors.New("no fact found")
	}

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
			writerErr = streamedExport.StreamedExportFactHitsFull(ctx, f, request.Limit)
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

		for {
			select {
			case hits, ok := <-streamedExport.Data:
				if !ok { // channel closed
					return
				}

				data, err := export.ConvertHitsToCSV(hits, request.CSVParameters, first)

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
// @Success 200 {array} export.WrapperItem "Returns a list of exports"
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
// @Router /engine/exports/{id} [get]
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
// @Router /engine/exports/{id} [delete]
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
// @Param request body handlers.ExportRequest true "request (json)"
// @Success 200 {object} export.WrapperItem "Status OK: user was added to existing export in queue"
// @Success 201 {object} export.WrapperItem "Status Created: new export was added in queue"
// @Failure 400 "Bad Request: missing fact id / fact id is not an integer"
// @Failure 403 "Status Forbidden: missing permission"
// @Failure 409 {object} export.WrapperItem "Status Conflict: user already exists in export queue"
// @Failure 429 "Status Too Many Requests: export queue is full"
// @Failure 500 "internal server error"
// @Router /engine/exports/fact [post]
func (e *ExportHandler) ExportFact(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeExport, permissions.All, permissions.ActionCreate)) {
		render.Error(w, r, render.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	var request ExportRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		zap.L().Warn("Decode export request json", zap.Error(err))
		render.Error(w, r, render.ErrAPIDecodeJSONBody, err)
		return
	}

	if len(request.FactIDs) == 0 {
		zap.L().Warn("Missing factIDs in export request")
		render.Error(w, r, render.ErrAPIMissingParam, errors.New("missing factIDs"))
		return
	}

	facts := findCombineFacts(request.FactIDs)
	if len(facts) == 0 {
		zap.L().Warn("No fact was found in export request")
		render.Error(w, r, render.ErrAPIDBResourceNotFound, errors.New("no fact was found in export request"))
		return
	}

	item, status := e.exportWrapper.AddToQueue(facts, request.Title, request.CSVParameters, userCtx.User)

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
