package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8/typedapi/core/search"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/config/esconfig"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/security/permissions"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/utils/httputil"
	"github.com/spf13/viper"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/export"
	"go.uber.org/zap"
)

type ExportHandler struct {
	exportWrapper       *export.Wrapper
	directDownload      bool
	indirectDownloadUrl string
}

// NewExportHandler returns a new ExportHandler
func NewExportHandler(exportWrapper *export.Wrapper, directDownload bool, indirectDownloadUrl string) *ExportHandler {
	return &ExportHandler{
		exportWrapper:       exportWrapper,
		directDownload:      directDownload,
		indirectDownloadUrl: indirectDownloadUrl,
	}
}

// ExportRequest represents a request for an export
type ExportRequest struct {
	export.CSVParameters
	FactIDs []int64 `json:"factIDs"`
	Title   string  `json:"title"`
}

// CustomExportRequest represents a request for an custom export
type CustomExportRequest struct {
	export.CSVParameters
	Title                    string           `json:"title"`
	Indices                  string           `json:"indices"`
	SearchRequests           []search.Request `json:"searchRequests"`
	ElasticName              string           `json:"elasticName"`
	IgnoreUnavailableIndices bool             `json:"ignoreUnavailableIndices"`
	AllowNoIndices           string           `json:"allowNoIndices"`
}

// ExportFactStreamed godoc
//
//	@Id				ExportFactStreamed
//
//	@Summary		CSV streamed export facts in chunks
//	@Description	CSV Streamed export for facts in chunks
//	@Tags			ExportFactStreamed
//	@Produce		octet-stream
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Param			request	body		handler.ExportRequest	true	"request (json)"
//	@Success		200		{file}		result					file
//	@Failure		500		{object}	render.APIError			"internal server error"
//	@Router			/engine/facts/streamedexport [post]
func ExportFactStreamed(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeExport, permissions.All, permissions.ActionGet)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	var request ExportRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		zap.L().Warn("Decode export request json", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDecodeJSONBody, err)
		return
	}

	if len(request.FactIDs) == 0 {
		zap.L().Warn("Missing factIDs in export request")
		httputil.Error(w, r, httputil.ErrAPIMissingParam, errors.New("missing factIDs"))
		return
	}

	if request.Separator == request.ListSeparator {
		zap.L().Warn("CSV file separator column and list separator in a column cannot be the same")
		httputil.Error(w, r, httputil.ErrAPIExportSeparatorConflict, errors.New("CSV file separator column and list separator in a column cannot be the same"))
		return
	}

	err = handleStreamedExport(r.Context(), w, request)
	if err != nil {
		httputil.Error(w, r, httputil.ErrAPIProcessError, err)
	}
	return

}

// handleStreamedExport actually only handles CSV
func handleStreamedExport(requestContext context.Context, w http.ResponseWriter, request ExportRequest) error {
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
			writerErr = streamedExport.StreamedExportFactHitsFull(ctx, f, request.Limit, make(map[string]interface{}))
			if writerErr != nil {
				zap.L().Error("Error during export (StreamedExportFactHitsFull)", zap.Error(err))
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
					zap.L().Error("ConvertHitsToCSV error during export (StreamedExportFactHitsFull)", zap.Error(err))
					cancel()
					return
				}

				// Write data
				_, err = w.Write(data)
				if err != nil {
					zap.L().Error("Write error during export (StreamedExportFactHitsFull)", zap.Error(err))
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
//
//	@Id				GetExports
//
//	@Summary		Get user exports
//	@Description	Get in memory user exports
//	@Produce		json
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{array}		export.WrapperItem	"Returns a list of exports"
//	@Failure		403	{object}	render.APIError		"Status Forbidden: missing permission"
//	@Failure		500	{object}	render.APIError		"internal server error"
//	@Router			/engine/exports [get]
func (e *ExportHandler) GetExports(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeExport, permissions.All, permissions.ActionList)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}
	httputil.JSON(w, r, e.exportWrapper.GetUserExports(userCtx.User))
}

// GetExport godoc
//
//	@Id				GetExport
//
//	@Summary		Get single export from user
//	@Description	Get single export from user
//	@Tags			Exports
//	@Produce		json
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Param			id	path		string				true	"Export ID"
//	@Success		200	{object}	export.WrapperItem	"Status OK"
//	@Failure		400	{object}	render.APIError		"Bad Request: missing export id"
//	@Failure		403	{object}	render.APIError		"Status Forbidden: missing permission"
//	@Failure		404	{object}	render.APIError		"Status Not Found: export not found"
//	@Failure		500	{object}	render.APIError		"internal server error"
//	@Router			/engine/exports/{id} [get]
func (e *ExportHandler) GetExport(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		httputil.Error(w, r, httputil.ErrAPIMissingParam, errors.New("missing id"))
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeExport, permissions.All, permissions.ActionGet)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	item, ok := e.exportWrapper.GetUserExport(id, userCtx.User)
	if !ok {
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFound, errors.New("export not found"))
		return
	}

	httputil.JSON(w, r, item)
}

// DeleteExport godoc
//
//	@Id				DeleteExport
//
//	@Summary		Deletes a single export
//	@Description	Deletes a single export, when running it is canceled
//	@Tags			Exports
//	@Produce		json
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Param			id	path		string			true	"Export ID"
//	@Success		202	{string}	string			"Status Accepted: export found, cancellation request has been taken into account and will be processed"
//	@Success		204	{string}	string			"Status OK: export was found and deleted"
//	@Failure		400	{object}	render.APIError	"Bad Request: missing export id"
//	@Failure		403	{object}	render.APIError	"Status Forbidden: missing permission"
//	@Failure		404	{object}	render.APIError	"Status Not Found: export not found"
//	@Failure		500	{object}	render.APIError	"internal server error"
//	@Router			/engine/exports/{id} [delete]
func (e *ExportHandler) DeleteExport(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		httputil.Error(w, r, httputil.ErrAPIMissingParam, errors.New("missing id"))
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeExport, permissions.All, permissions.ActionDelete)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	status := e.exportWrapper.DeleteExport(id, userCtx.User)

	switch status {
	case export.DeleteExportDeleted:
		fallthrough
	case export.DeleteExportUserDeleted:
		w.WriteHeader(http.StatusNoContent)
	case export.DeleteExportCanceled:
		w.WriteHeader(http.StatusAccepted)
	default:
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFound, errors.New("export not found"))
	}

}

// ExportFact godoc
//
//	@Id				ExportFact
//
//	@Summary		Creates a new export request for a fact (or multiple facts)
//	@Description	Creates a new export request for a fact (or multiple facts)
//	@Tags			Exports
//	@Produce		json
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Param			request	body		handler.ExportRequest	true	"request (json)"
//	@Success		200		{object}	export.WrapperItem		"Status OK: user was added to existing export in queue"
//	@Success		201		{object}	export.WrapperItem		"Status Created: new export was added in queue"
//	@Failure		400		{object}	render.APIError			"Bad Request: missing fact id / fact id is not an integer"
//	@Failure		403		{object}	render.APIError			"Status Forbidden: missing permission"
//	@Failure		409		{object}	export.WrapperItem		"Status Conflict: user already exists in export queue"
//	@Failure		429		{object}	render.APIError			"Status Too Many Requests: export queue is full"
//	@Failure		500		{object}	render.APIError			"internal server error"
//	@Router			/engine/exports/fact [post]
func (e *ExportHandler) ExportFact(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeExport, permissions.All, permissions.ActionCreate)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	var request ExportRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		zap.L().Warn("Decode export request json", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDecodeJSONBody, err)
		return
	}

	if len(request.FactIDs) == 0 {
		zap.L().Warn("Missing factIDs in export request")
		httputil.Error(w, r, httputil.ErrAPIMissingParam, errors.New("missing factIDs"))
		return
	}

	if len(request.Title) == 0 {
		zap.L().Warn("Missing title (len is 0) in export request")
		httputil.Error(w, r, httputil.ErrAPIMissingParam, errors.New("missing title (len is 0)"))
		return
	}

	if request.Separator == request.ListSeparator {
		zap.L().Warn("CSV file separator column and list separator in a column cannot be the same")
		httputil.Error(w, r, httputil.ErrAPIExportSeparatorConflict, errors.New("CSV file separator column and list separator in a column cannot be the same"))
		return
	}

	factParameters, err := ParseFactParameters(r.URL.Query().Get("factParameters"))
	if err != nil {
		zap.L().Error("Parse input Fact Parametres", zap.Error(err), zap.String("raw offset", r.URL.Query().Get("factParameters")))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	facts := findCombineFacts(request.FactIDs)
	if len(facts) == 0 {
		zap.L().Warn("No fact was found in export request")
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFound, errors.New("no fact was found in export request"))
		return
	}

	item, status := e.exportWrapper.AddToQueue(facts, request.Title, request.CSVParameters, userCtx.User, factParameters, true)

	e.handleAddToQueueResponse(w, r, status, item)
}

func (e *ExportHandler) handleAddToQueueResponse(w http.ResponseWriter, r *http.Request, status int, item *export.WrapperItem) {
	switch status {
	case export.CodeAdded:
		w.WriteHeader(http.StatusCreated)
	case export.CodeUserAdded:
		w.WriteHeader(http.StatusOK)
	case export.CodeUserExists:
		w.WriteHeader(http.StatusConflict)
	case export.CodeQueueFull:
		httputil.Error(w, r, httputil.ErrAPIQueueFull, fmt.Errorf("export queue is full"))
		return
	default:
		httputil.Error(w, r, httputil.ErrAPIProcessError, fmt.Errorf("unknown status code (%d)", status))
		return
	}

	httputil.JSON(w, r, item)
}

// ExportCustom godoc
//
//	@Id				ExportCustom
//
//	@Summary		Creates a new custom export request
//	@Description	Creates a new custom export request
//	@Tags			Exports
//	@Produce		json
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Param			request	body		handler.ExportRequest	true	"request (json)"
//	@Success		200		{object}	export.WrapperItem		"Status OK: user was added to existing export in queue"
//	@Success		201		{object}	export.WrapperItem		"Status Created: new export was added in queue"
//	@Failure		400		{object}	render.APIError			"Bad Request: missing fact id / fact id is not an integer"
//	@Failure		403		{object}	render.APIError			"Status Forbidden: missing permission"
//	@Failure		409		{object}	export.WrapperItem		"Status Conflict: user already exists in export queue"
//	@Failure		429		{object}	render.APIError			"Status Too Many Requests: export queue is full"
//	@Failure		500		{object}	render.APIError			"internal server error"
//	@Router			/engine/exports/custom [post]
func (e *ExportHandler) ExportCustom(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeExport, permissions.All, permissions.ActionCreate)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	var request CustomExportRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		zap.L().Warn("Decode export custom request json", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDecodeJSONBody, err)
		return
	}

	if request.Indices == "" {
		zap.L().Warn("Missing indices in export request")
		httputil.Error(w, r, httputil.ErrAPIMissingParam, errors.New("missing indices"))
		return
	}

	if len(request.Title) == 0 {
		zap.L().Warn("Missing title (len is 0) in export request")
		httputil.Error(w, r, httputil.ErrAPIMissingParam, errors.New("missing title (len is 0)"))
		return
	}

	if request.Separator == request.ListSeparator {
		zap.L().Warn("CSV file separator column and list separator in a column cannot be the same")
		httputil.Error(w, r, httputil.ErrAPIExportSeparatorConflict, errors.New("CSV file separator column and list separator in a column cannot be the same"))
		return
	}

	if len(request.SearchRequests) == 0 {
		zap.L().Warn("Missing searchRequests (len is 0) in export request")
		httputil.Error(w, r, httputil.ErrAPIMissingParam, errors.New("missing searchRequests (len is 0)"))
		return
	}

	if len(request.SearchRequests) > viper.GetInt("EXPORT_MAX_CUSTOM_SEARCH_REQUESTS") {
		zap.L().Warn("Maximum single custom export search requests reached", zap.Int("count", len(request.SearchRequests)))
		httputil.Error(w, r, httputil.ErrAPITooManyRequests, errors.New("maximum single custom export search requests reached"))
		return
	}

	elastic, found, err := esconfig.R().GetByName(request.ElasticName)
	if err != nil {
		zap.L().Warn("Cannot get esconfig config", zap.String("name", request.ElasticName), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFound, err)
		return
	}
	if !found {
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFound, errors.New("export not found"))
		return
	}
	if !elastic.ExportActivated {
		httputil.Error(w, r, httputil.ErrAPIElasticExportDisabled, nil)
		return
	}

	params := export.ElasticParams{
		Indices:           request.Indices,
		Limit:             request.Limit,
		Client:            elastic.Name,
		IgnoreUnavailable: request.IgnoreUnavailableIndices,
		AllowNoIndices:    request.AllowNoIndices,
	}
	item, status := e.exportWrapper.AddToQueueCustom(request.Title, params, request.SearchRequests, request.CSVParameters, userCtx.User, true)

	e.handleAddToQueueResponse(w, r, status, item)
}

// DownloadExport godoc
//
//	@Id				DownloadExport
//
//	@Summary		Download export
//	@Description	Download export
//	@Tags			Exports
//	@Produce		octet-stream
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Param			id	path		string			true	"Export ID"
//	@Success		200	{file}		result			file
//	@Success		308	{object}	render.APIError	"Redirects	to		the		export	file	location
//	@Failure		400	{object}	render.APIError	"Bad Request: missing export id"
//	@Failure		403	{object}	render.APIError	"Status Forbidden: missing permission"
//	@Failure		404	{object}	render.APIError	"Status Not Found: export not found"
//	@Failure		500	{object}	render.APIError	"internal server error"
//	@Router			/engine/exports/{id}/download [get]
func (e *ExportHandler) DownloadExport(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		httputil.Error(w, r, httputil.ErrAPIMissingParam, errors.New("missing id"))
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeExport, permissions.All, permissions.ActionGet)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	item, ok := e.exportWrapper.GetUserExport(id, userCtx.User)
	if !ok {
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFound, errors.New("export not found"))
		return
	}

	if e.directDownload {
		path := filepath.Join(e.exportWrapper.BasePath, item.FileName)
		httputil.StreamFile(path, item.FileName, w, r)
		return
	}

	path, err := url.JoinPath(e.indirectDownloadUrl, item.FileName)
	if err != nil {
		httputil.Error(w, r, httputil.ErrAPIProcessError, err)
		return
	}

	http.Redirect(w, r, path, http.StatusPermanentRedirect)
}
