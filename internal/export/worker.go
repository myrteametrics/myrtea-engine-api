package export

import (
	"compress/gzip"
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/myrteametrics/myrtea-engine-api/v5/internal/notifier"
	"go.uber.org/zap"
)

type ExportWorker struct {
	Mutex    sync.Mutex
	Id       int
	Success  chan<- int
	Cancel   chan bool // channel to cancel the worker
	BasePath string    // base path where the file will be saved
	// critical fields
	Available bool
	QueueItem WrapperItem
}

func NewExportWorker(id int, basePath string, success chan<- int) *ExportWorker {
	return &ExportWorker{
		Id:        id,
		Available: true,
		BasePath:  basePath,
		Cancel:    make(chan bool, 3), // buffered channel to avoid blocking
		Success:   success,
	}
}

// SetError sets the error and the status of the worker
func (e *ExportWorker) SetError(error error) {
	e.Mutex.Lock()
	defer e.Mutex.Unlock()
	e.QueueItem.Status = StatusError
	if error == nil {
		e.QueueItem.Error = ""
	} else {
		e.QueueItem.Error = error.Error()
	}
}

// SetStatus sets the status of the worker
func (e *ExportWorker) SetStatus(status int) {
	e.Mutex.Lock()
	defer e.Mutex.Unlock()
	e.QueueItem.Status = status
}

// SwapAvailable swaps the availability of the worker
func (e *ExportWorker) SwapAvailable(available bool) (old bool) {
	e.Mutex.Lock()
	defer e.Mutex.Unlock()
	old = e.Available
	e.Available = available
	return old
}

// IsAvailable returns the availability of the worker
func (e *ExportWorker) IsAvailable() bool {
	e.Mutex.Lock()
	defer e.Mutex.Unlock()
	return e.Available
}

// DrainCancelChannel drains the cancel channel
func (e *ExportWorker) DrainCancelChannel() {
	for {
		select {
		case <-e.Cancel:
		default:
			return
		}
	}
}

// finalize sets the worker availability to true and clears the queueItem
func (e *ExportWorker) finalize() {
	e.Mutex.Lock()

	// set status to error if error occurred
	if e.QueueItem.Error != "" {
		e.QueueItem.Status = StatusError
	}
	// set status to done if no error occurred
	if e.QueueItem.Status != StatusError && e.QueueItem.Status != StatusCanceled {
		e.QueueItem.Status = StatusDone
	}
	e.Mutex.Unlock()

	// clear Cancel channel, to avoid blocking
	e.DrainCancelChannel()

	// notify to the dispatcher that this worker is now available
	e.Success <- e.Id
}

// Start starts the export task
// It handles one queueItem at a time and when finished it stops the goroutine
func (e *ExportWorker) Start(item WrapperItem, ctx context.Context) {
	defer e.finalize()
	item.Status = StatusRunning

	e.Mutex.Lock()
	e.QueueItem = item
	e.Mutex.Unlock()

	// send notification to user (non-blocking)
	go func(wrapperItem WrapperItem) {
		_ = notifier.C().SendToUserLogins(
			createExportNotification(ExportNotificationStarted, &item),
			wrapperItem.Users)
	}(item)

	// create file
	path := filepath.Join(e.BasePath, item.FileName)
	// check if file not already exists
	if _, err := os.Stat(path); err == nil {
		e.SetError(fmt.Errorf("file with same name already exists"))
		return
	}

	file, err := os.Create(path)
	if err != nil {
		e.SetError(err)
		return
	}
	defer file.Close()

	var csvWriter *csv.Writer
	var gzipWriter *gzip.Writer

	// Conditionally apply gzip compression based on the UncompressedOutput parameter
	// Gzip is enabled by default unless UncompressedOutput is true
	if !item.Params.UncompressedOutput {
		// opens a gzip writer
		gzipWriter = gzip.NewWriter(file)
		defer gzipWriter.Close()
		csvWriter = csv.NewWriter(gzipWriter)
	} else {
		// write directly to file without compression
		csvWriter = csv.NewWriter(file)
	}

	streamedExport := NewStreamedExport()
	var wg sync.WaitGroup
	var writerErr error

	// local context handling
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Increment the WaitGroup counter
	wg.Add(1)

	/**
	 * How streamed export works:
	 *    - Export goroutine: each fact is processed one by one
	 *      Each bulk of data is sent through a channel to the receiver
	 *    - The receiver handles the incoming channel data and converts them to the CSV format
	 *      After the conversion, the data is written to a local file (optionally gzipped)
	 */

	go func() {
		defer wg.Done()
		defer close(streamedExport.Data)

		if item.Custom {
			params := ElasticParams{
				Indices:           item.Indices,
				Limit:             item.Params.Limit,
				Client:            item.ElasticName,
				IgnoreUnavailable: item.IgnoreUnavailable,
				AllowNoIndices:    item.AllowNoIndices,
			}

			for _, searchRequest := range item.SearchRequests {
				writerErr = streamedExport.ProcessStreamedExport(ctx, &searchRequest, params)
				if writerErr != nil {
					break
				}
			}

		} else {
			for _, f := range item.Facts {
				writerErr = streamedExport.StreamedExportFactHitsFull(ctx, f, item.Params.Limit, item.FactParameters)
				if writerErr != nil {
					break // break here when error occurs?
				}
			}
		}

	}()

	// Chunk handler
	first := true

loop:
	for {
		select {
		case hits, ok := <-streamedExport.Data:
			if !ok { // channel closed
				break loop
			}

			err = WriteConvertHitsToCSV(csvWriter, hits, item.Params, first)

			if err != nil {
				zap.L().Error("WriteConvertHitsToCSV error during export", zap.Error(err))
				cancel()
				break loop
			}

			// Flush data
			csvWriter.Flush()

			if first {
				first = false
			}
		case <-ctx.Done():
			break loop
		case <-e.Cancel:
			cancel()
			break loop
		}
	}

	wg.Wait()

	// error occurred, close file and delete
	if writerErr != nil || err != nil {
		if ctx.Err() != nil {
			e.SetStatus(StatusCanceled)
			zap.L().Warn("Export worker: canceled, deleting file...", zap.String("filePath", path))
		} else {
			if err != nil { // priority to err
				e.SetError(err)
			} else {
				e.SetError(writerErr)
			}
			zap.L().Error("Export worker: error, deleting file...", zap.String("filePath", path),
				zap.NamedError("err", err), zap.NamedError("writerErr", writerErr))
		}

		// close writer and file access before trying to delete file
		if gzipWriter != nil {
			_ = gzipWriter.Close()
		}
		_ = file.Close()

		err = os.Remove(path)
		if err != nil {
			zap.L().Error("Export worker: couldn't delete file", zap.String("filePath", path), zap.Error(err))
		}
	}

}
