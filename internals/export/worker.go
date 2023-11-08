package export

import (
	"compress/gzip"
	"context"
	"encoding/csv"
	"fmt"
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"sync"
)

type ExportWorker struct {
	Mutex     sync.Mutex
	Available bool
	Cancel    chan bool // channel to cancel the worker
	//
	QueueItemId string // id of the queueItem currently handled by the worker
	BasePath    string // base path where the file will be saved
}

func NewExportWorker(basePath string) *ExportWorker {
	return &ExportWorker{
		Available:   true,
		QueueItemId: "",
		BasePath:    basePath,
		Cancel:      make(chan bool),
	}
}

// SetAvailable sets the worker availability to true and clears the queueItem
func (e *ExportWorker) SetAvailable(item *ExportWrapperItem) {
	e.Mutex.Lock()
	defer e.Mutex.Unlock()
	e.Available = true

	// set queueItem status to done
	item.Mutex.Lock()
	// set status to error if error occurred
	if item.Error != nil {
		item.Status = StatusError
	}
	// set status to done if no error occurred
	if item.Status != StatusError {
		item.Status = StatusDone
	}
	item.Mutex.Unlock()

	e.QueueItemId = ""
}

// Start starts the export task
// It handles one queueItem at a time and when finished it stops the goroutine
func (e *ExportWorker) Start(item *ExportWrapperItem) {
	defer e.SetAvailable(item)
	e.Mutex.Lock()
	e.QueueItemId = item.Id
	e.Mutex.Unlock()

	item.SetStatus(StatusRunning)

	// create file
	path := filepath.Join(e.BasePath, item.Params.FileName)
	// check if file not already exists
	if _, err := os.Stat(path); err == nil {
		item.SetError(fmt.Errorf("file with same name already exists"))
		return
	}

	file, err := os.Create(path)
	if err != nil {
		item.SetError(err)
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
	ctx, cancel := context.WithCancel(context.Background())
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

		for _, f := range item.Facts {
			writerErr = streamedExport.StreamedExportFactHitsFull(ctx, f, item.Params.Limit)
			if writerErr != nil {
				break // break here when error occurs?
			}
		}
	}()

	// Chunk handler
	first := true
	labels := item.Params.ColumnsLabel
	loop := true

	for loop {
		select {
		case hits, ok := <-streamedExport.Data:
			if !ok { // channel closed
				loop = false
				break
			}

			err := WriteConvertHitsToCSV(csvWriter, hits, item.Params.Columns, labels, item.Params.FormatColumnsData, item.Params.Separator)

			if err != nil {
				zap.L().Error("WriteConvertHitsToCSV error during export", zap.Error(err))
				cancel()
				loop = false
				break
			}

			// Flush data
			csvWriter.Flush()

			if first {
				first = false
				labels = []string{}
			}

		case <-e.Cancel:
			cancel()
			loop = false
		}
	}

	wg.Wait()

	// error occurred, close file and delete
	if writerErr != nil || err != nil {
		if ctx.Err() != nil {
			item.SetStatus(StatusCanceled)
			zap.L().Warn("Export worker: canceled, deleting file...", zap.String("filePath", path))
		} else {
			if err != nil { // priority to err
				item.SetError(err)
			} else {
				item.SetError(writerErr)
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
