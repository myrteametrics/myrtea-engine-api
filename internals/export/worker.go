package export

import (
	"compress/gzip"
	"context"
	"encoding/csv"
	"errors"
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"sync"
)

type ExportWorker struct {
	Mutex     sync.Mutex
	Available bool
	QueueItem *ExportWrapperItem
	Context   context.Context
	BasePath  string
}

func NewExportWorker(basePath string) *ExportWorker {
	return &ExportWorker{
		Available: true,
		QueueItem: nil,
		BasePath:  basePath,
	}
}

// SetAvailable sets the worker availability to true and clears the queueItem
func (e *ExportWorker) SetAvailable() {
	e.Mutex.Lock()
	defer e.Mutex.Unlock()
	e.Available = true

	// set queueItem status to done
	e.QueueItem.Mutex.Lock()
	if e.QueueItem.Error == nil {
		e.QueueItem.Status = StatusDone
	}
	e.QueueItem.Mutex.Unlock()

	e.QueueItem = nil
}

// Start starts the export task
// It handles one queueItem at a time and when finished it stops the goroutine
func (e *ExportWorker) Start(item *ExportWrapperItem) {
	defer e.SetAvailable()
	e.Mutex.Lock()
	e.QueueItem = item
	e.QueueItem.SetStatus(StatusRunning)
	e.Mutex.Unlock()

	// create file
	path := filepath.Join(e.BasePath, item.Params.FileName)
	// check if file exists
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		e.Mutex.Lock()
		e.QueueItem.SetError(err)
		e.Mutex.Unlock()
		return
	}

	file, err := os.Create("data.csv.gz")
	if err != nil {
		e.Mutex.Lock()
		e.QueueItem.SetError(err)
		e.Mutex.Unlock()
		return
	}
	defer file.Close()

	// Create a gzip writer
	gzipWriter := gzip.NewWriter(file)
	defer gzipWriter.Close()

	csvWriter := csv.NewWriter(gzipWriter)

	// start streamed export
	streamedExport := NewStreamedExport()
	var wg sync.WaitGroup

	// Increment the WaitGroup counter
	wg.Add(2) // 2 goroutines

	var writerErr error = nil

	/**
	 * How streamed export works:
	 * 1. Browser opens connection
	 * 2. Two goroutines are started:
	 *    - Export goroutine: each fact is processed one by one
	 *      Each bulk of data is sent through a channel to the receiver goroutine
	 *    - The receiver handles the incoming channel data and converts them to the CSV format
	 *      After the conversion, the data is written and gzipped to a local file
	 */

	go func() {
		defer wg.Done()
		defer close(streamedExport.Data)

		for _, f := range item.Facts {
			writerErr = streamedExport.StreamedExportFactHitsFull(e.Context, f, item.Params.Limit)
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
		labels := item.Params.ColumnsLabel

		for {
			select {
			case hits, ok := <-streamedExport.Data:
				if !ok { // channel closed
					return
				}

				data, err := ConvertHitsToCSV(hits, item.Params.Columns, labels, item.Params.FormatColumnsData, item.Params.Separator)

				if err != nil {
					zap.L().Error("ConvertHitsToCSV error during export (StreamedExportFactHitsFullV8)", zap.Error(err))
					cancel()
					return
				}

				// Write data
				_, err = csvWriter.Write(data)
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

}
