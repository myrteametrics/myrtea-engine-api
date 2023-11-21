package export

import (
	"compress/gzip"
	"context"
	"encoding/csv"
	"fmt"
	"github.com/myrteametrics/myrtea-sdk/v4/engine"
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"sync"
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
		Cancel:    make(chan bool),
		Success:   success,
	}
}

// SetError sets the error and the status of the worker
func (e *ExportWorker) SetError(error error) {
	e.Mutex.Lock()
	defer e.Mutex.Unlock()
	e.QueueItem.Status = StatusError
	e.QueueItem.Error = error
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

// finalise sets the worker availability to true and clears the queueItem
func (e *ExportWorker) finalise() {
	e.Mutex.Lock()

	// set status to error if error occurred
	if e.QueueItem.Error != nil {
		e.QueueItem.Status = StatusError
	}
	// set status to done if no error occurred
	if e.QueueItem.Status != StatusError {
		e.QueueItem.Status = StatusDone
	}

	e.Mutex.Unlock()

	// notify to the dispatcher that this worker is now available
	e.Success <- e.Id
}

// Start starts the export task
// It handles one queueItem at a time and when finished it stops the goroutine
func (e *ExportWorker) Start(item WrapperItem, ctx context.Context) {
	defer e.finalise()
	e.Mutex.Lock()
	e.QueueItem = item
	e.Mutex.Unlock()

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

	// opens a gzip writer
	gzipWriter := gzip.NewWriter(file)
	defer gzipWriter.Close()

	csvWriter := csv.NewWriter(gzipWriter)
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
	 *      After the conversion, the data is written and gzipped to a local file
	 */

	go func() {
		defer wg.Done()
		defer close(streamedExport.Data)

		for _, f := range item.FactIDs {
			_ = f // TODO: facts
			writerErr = streamedExport.StreamedExportFactHitsFull(ctx, engine.Fact{}, item.Params.Limit)
			if writerErr != nil {
				break // break here when error occurs?
			}
		}
	}()

	// Chunk handler
	first := true
	labels := item.Params.ColumnsLabel

loop:
	for {
		select {
		case hits, ok := <-streamedExport.Data:
			if !ok { // channel closed
				break loop
			}

			err = WriteConvertHitsToCSV(csvWriter, hits, item.Params.Columns, labels, item.Params.FormatColumnsData, item.Params.Separator)

			if err != nil {
				zap.L().Error("WriteConvertHitsToCSV error during export", zap.Error(err))
				cancel()
				break loop
			}

			// Flush data
			csvWriter.Flush()

			if first {
				first = false
				labels = []string{}
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
		_ = gzipWriter.Close()
		_ = file.Close()

		err = os.Remove(path)
		if err != nil {
			zap.L().Error("Export worker: couldn't delete file", zap.String("filePath", path), zap.Error(err))
		}
	}

}
