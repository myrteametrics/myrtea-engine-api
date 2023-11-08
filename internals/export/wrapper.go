package export

import (
	"context"
	"github.com/google/uuid"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/security/users"
	"github.com/myrteametrics/myrtea-sdk/v4/engine"
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	CodeUserAdded  = 1
	CodeAdded      = 0
	CodeUserExists = -1
	CodeQueueFull  = -2

	// ExportWrapperItem statuses
	StatusPending  = 0
	StatusRunning  = 1
	StatusDone     = 2
	StatusError    = 3
	StatusCanceled = 4
)

type ExportWrapperItem struct {
	Mutex  sync.Mutex
	Error  error
	Status int
	Users  []users.User // handles export ownership
	// non-critical fields
	Id     string // unique id that represents an export demand
	FactID int64
	Params CSVParameters
	Date   time.Time
	Facts  []engine.Fact
}

type ExportWrapper struct {
	QueueMutex sync.Mutex
	DoneMutex  sync.Mutex
	Workers    []*ExportWorker
	Queue      []*ExportWrapperItem
	// Done contains all tasks that have been taken out of the queue, including the ones that are still running
	Done              []*ExportWrapperItem
	DiskRetentionDays int
	BasePath          string
	QueueMaxSize      int
}

// NewExportWrapperItem creates a new export wrapper item
func NewExportWrapperItem(factID int64, params CSVParameters, user users.User) *ExportWrapperItem {
	return &ExportWrapperItem{
		Id:     uuid.New().String(),
		FactID: factID,
		Params: params,
		Users:  append([]users.User{}, user),
		Date:   time.Now(),
		Status: StatusPending,
		Error:  nil,
	}
}

// NewExportWrapper creates a new export wrapper
func NewExportWrapper(basePath string, diskRetentionDays, queueMaxSize int) *ExportWrapper {
	return &ExportWrapper{
		Workers:           make([]*ExportWorker, 0),
		Queue:             make([]*ExportWrapperItem, 0),
		Done:              make([]*ExportWrapperItem, 0),
		QueueMaxSize:      queueMaxSize,
		BasePath:          basePath,
		DiskRetentionDays: diskRetentionDays,
	}
}

// SetStatus sets the status of the export wrapper item
func (ew *ExportWrapperItem) SetStatus(status int) {
	ew.Mutex.Lock()
	defer ew.Mutex.Unlock()
	ew.Status = status
}

// SetError sets the error and status of the export wrapper item
func (ew *ExportWrapperItem) SetError(err error) {
	ew.Mutex.Lock()
	defer ew.Mutex.Unlock()
	ew.Error = err
	ew.Status = StatusError
	zap.L().Error("Error happened during export worker execution", zap.Error(err))
}

// Init initializes the export wrapper
func (ew *ExportWrapper) Init(workers int) {
	// instantiate workers
	for i := 0; i < workers; i++ {
		ew.Workers = append(ew.Workers, NewExportWorker(ew.BasePath))
	}
	go ew.StartDispatcher(context.Background())
}

// AddToQueue Adds a new export to the export worker queue
func (ew *ExportWrapper) AddToQueue(factID int64, params CSVParameters, user users.User) int {
	ew.QueueMutex.Lock()
	defer ew.QueueMutex.Unlock()

	for _, queueItem := range ew.Queue {
		if queueItem.FactID == factID && queueItem.Params.Equals(params) {

			// check if user not already in queue.users
			for _, u := range queueItem.Users {
				if u.ID == user.ID {
					return CodeUserExists
				}
			}

			queueItem.Users = append(queueItem.Users, user)
			return CodeUserAdded
		}
	}

	if len(ew.Queue) >= ew.QueueMaxSize {
		return CodeQueueFull
	}

	ew.Queue = append(ew.Queue, NewExportWrapperItem(factID, params, user))
	return CodeAdded
}

// FindAvailableWorker finds an available worker and sets it to unavailable
func (ew *ExportWrapper) FindAvailableWorker() *ExportWorker {
	for _, worker := range ew.Workers {
		worker.Mutex.Lock()
		if worker.Available {
			worker.Available = false
			worker.Mutex.Unlock()
			return worker
		}
		worker.Mutex.Unlock()
	}

	return nil
}

// StartDispatcher starts the export tasks dispatcher & the expired files checker
func (ew *ExportWrapper) StartDispatcher(context context.Context) {
	zap.L().Info("Starting export tasks dispatcher")
	// every 5 seconds check if there is a new task to process in queue then check if there is an available worker
	// if yes, start the worker with the task
	// if no, continue to check
	ticker := time.NewTicker(5 * time.Second)
	expiredFileTicker := time.NewTicker(24 * time.Hour)
	for {
		select {
		case <-ticker.C:
			ew.QueueMutex.Lock()
			if len(ew.Queue) > 0 {
				for i := 0; i < len(ew.Queue); i++ {
					x := ew.Queue[i]
					w := ew.FindAvailableWorker()

					// if no worker available, stop the loop since no worker will be available for the next tasks
					if w == nil {
						break
					}

					// attach the task to the worker and start the worker
					go w.Start(x)

					// dequeue the task
					ew.Queue = ew.Queue[1:]

					// add the task to the done list
					ew.DoneMutex.Lock()
					ew.Done = append(ew.Done, x)
					ew.DoneMutex.Unlock()

				}
			}
			ew.QueueMutex.Unlock()
		case <-expiredFileTicker.C:
			err := ew.CheckForExpiredFiles()

			if err != nil {
				zap.L().Error("Error during expired files check", zap.Error(err))
			}

		case <-context.Done():
			ticker.Stop()
			return
		}
	}
}

// CheckForExpiredFiles checks for expired files in the export directory and deletes them
// it also deletes the done tasks that are older than diskRetentionDays
func (ew *ExportWrapper) CheckForExpiredFiles() error {
	// Get all files in directory and check the last edit date
	// if last edit date is older than diskRetentionDays, delete the file
	zap.L().Info("Checking for expired files")
	files, err := os.ReadDir(ew.BasePath)
	if err != nil {
		return err
	}

	// delete all done tasks of ew.Done that are older than diskRetentionDays
	ew.DoneMutex.Lock()
	for i := 0; i < len(ew.Done); i++ {
		x := ew.Done[i]
		if time.Since(x.Date).Hours() > float64(ew.DiskRetentionDays*24) {
			ew.Done = append(ew.Done[:i], ew.Done[i+1:]...)
			i--
		}
	}
	ew.DoneMutex.Unlock()

	// count the number of deleted files
	count := 0

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filePath := filepath.Join(ew.BasePath, file.Name())

		fi, err := os.Stat(filePath)
		if err != nil {
			zap.L().Error("Cannot get file info", zap.String("file", filePath), zap.Error(err))
			continue
		}

		// skip if file is not a zip
		//if filepath.Ext(file.Name()) != ".zip" {
		//	continue
		//}

		if time.Since(fi.ModTime()).Hours() > float64(ew.DiskRetentionDays*24) {
			err = os.Remove(filePath)
			if err != nil {
				zap.L().Error("Cannot delete file", zap.String("file", filePath), zap.Error(err))
				continue
			}
			count++
		}
	}

	zap.L().Info("Deleted expired files", zap.Int("count", count))
	return nil
}
