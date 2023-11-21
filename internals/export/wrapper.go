package export

import (
	"context"
	"github.com/google/uuid"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/security/users"
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

const (
	CodeUserAdded  = 1
	CodeAdded      = 0
	CodeUserExists = -1
	CodeQueueFull  = -2

	// WrapperItem statuses
	StatusPending  = 0
	StatusRunning  = 1
	StatusDone     = 2
	StatusError    = 3
	StatusCanceled = 4
)

type WrapperItem struct {
	Id       string        `json:"id"` // unique id that represents an export demand
	FactIDs  []int64       `json:"factIds"`
	Error    error         `json:"error"`
	Status   int           `json:"status"`
	FileName string        `json:"fileName"`
	Date     time.Time     `json:"date"`
	Users    []string      `json:"-"`
	Params   CSVParameters `json:"-"`
}

type Wrapper struct {
	// Queue handling
	QueueItemsMutex sync.RWMutex
	QueueItems      []*WrapperItem // stores queue to handle duplicates, state
	//Queue           chan *WrapperItem

	// contains also current handled items
	// Workers is final, its only instanced once and thus does not change size (ExportWorker have there indexes in this slice stored)
	Workers []*ExportWorker

	// Success is passed to all workers, they write on this channel when they've finished with there export
	Success chan int

	// Archived WrapperItem's
	Archive sync.Map // map of all exports that have been done, key is the id of the export

	// Non-critical fields
	// Read-only parameters
	DiskRetentionDays int
	BasePath          string
	QueueMaxSize      int
	WorkerCount       int
}

// NewWrapperItem creates a new export wrapper item
func NewWrapperItem(factIDs []int64, fileName string, params CSVParameters, user users.User) *WrapperItem {
	// sort slices (for easy comparison)
	sort.Slice(factIDs, func(i, j int) bool { return factIDs[i] < factIDs[j] })
	return &WrapperItem{
		Users:    append([]string{}, user.Login),
		Id:       uuid.New().String(),
		FactIDs:  factIDs,
		Date:     time.Now(),
		Status:   StatusPending,
		Error:    nil,
		FileName: fileName,
		Params:   params,
	}
}

// NewWrapper creates a new export wrapper
func NewWrapper(basePath string, workersCount, diskRetentionDays, queueMaxSize int) *Wrapper {
	return &Wrapper{
		Workers:           make([]*ExportWorker, 0),
		QueueItems:        make([]*WrapperItem, 0),
		Success:           make(chan int),
		Archive:           sync.Map{},
		QueueMaxSize:      queueMaxSize,
		BasePath:          basePath,
		DiskRetentionDays: diskRetentionDays,
		WorkerCount:       workersCount,
	}
}

// ContainsFact checks if fact is part of the WrapperItem data
func (it *WrapperItem) ContainsFact(factID int64) bool {
	for _, d := range it.FactIDs {
		if d == factID {
			return true
		}
	}
	return false
}

// Init initializes the export wrapper
func (ew *Wrapper) Init(ctx context.Context) {
	// instantiate workers
	for i := 0; i < ew.WorkerCount; i++ {
		ew.Workers = append(ew.Workers, NewExportWorker(i, ew.BasePath, ew.Success))
	}
	go ew.StartDispatcher(ctx)
}

// AddToQueue Adds a new export to the export worker queue
func (ew *Wrapper) AddToQueue(factIDs []int64, fileName string, params CSVParameters, user users.User) (*WrapperItem, int) {
	ew.QueueItemsMutex.Lock()
	defer ew.QueueItemsMutex.Unlock()

	for _, queueItem := range ew.QueueItems {
		if !Int64Equals(queueItem.FactIDs, factIDs) || !queueItem.Params.Equals(params) {
			continue
		}

		// check if user not already in queue.users
		for _, u := range queueItem.Users {
			if u == user.Login {
				return nil, CodeUserExists
			}
		}

		queueItem.Users = append(queueItem.Users, user.Login)
		return nil, CodeUserAdded
	}

	if len(ew.QueueItems) >= ew.QueueMaxSize {
		return nil, CodeQueueFull
	}

	item := NewWrapperItem(factIDs, fileName, params, user)
	ew.QueueItems = append(ew.QueueItems, item)
	return item, CodeAdded
}

// StartDispatcher starts the export tasks dispatcher & the expired files checker
func (ew *Wrapper) StartDispatcher(context context.Context) {
	zap.L().Info("Starting export tasks dispatcher")
	// every 5 seconds check if there is a new task to process in queue then check if there is an available worker
	// if yes, start the worker with the task
	// if no, continue to check
	ticker := time.NewTicker(5 * time.Second)
	expiredFileTicker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()
	defer expiredFileTicker.Stop()

	for {
		select {
		case w := <-ew.Success:
			worker := ew.Workers[w]
			// TODO: send notifications here

			// archive item when finished
			worker.Mutex.Lock()
			ew.Workers[w].Available = true
			item := worker.QueueItem
			worker.QueueItem = WrapperItem{}
			worker.Mutex.Unlock()
			// archive item
			ew.Archive.Store(item.Id, item)
		case <-ticker.C:
			ew.dispatchExportQueue(context)
		case <-expiredFileTicker.C:
			err := ew.CheckForExpiredFiles()

			if err != nil {
				zap.L().Error("Error during expired files check", zap.Error(err))
			}
		case <-context.Done():
			return
		}
	}
}

// CheckForExpiredFiles checks for expired files in the export directory and deletes them
// it also deletes the done tasks that are older than diskRetentionDays
func (ew *Wrapper) CheckForExpiredFiles() error {
	// Get all files in directory and check the last edit date
	// if last edit date is older than diskRetentionDays, delete the file
	zap.L().Info("Checking for expired files")
	files, err := os.ReadDir(ew.BasePath)
	if err != nil {
		return err
	}

	// delete all done archives of ew.Archive that are older than diskRetentionDays
	ew.Archive.Range(func(key, value any) bool {
		data, ok := value.(WrapperItem)
		if !ok {
			return true
		}
		if time.Since(data.Date).Hours() > float64(ew.DiskRetentionDays*24) {
			ew.Archive.Delete(key)
		}
		return true
	})

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

func (ew *Wrapper) GetUserExports(user users.User) []WrapperItem {
	var result []WrapperItem

	// first, gather all exports that are in the workers if there are any
	for _, worker := range ew.Workers {
		worker.Mutex.Lock()
		if worker.QueueItem.ContainsUser(user) {
			result = append(result, worker.QueueItem)
		}
		worker.Mutex.Unlock()
	}

	// then, gather all exports that are archived
	ew.Archive.Range(func(key, value any) bool {
		data, ok := value.(WrapperItem)
		if !ok {
			return true
		}
		if data.ContainsUser(user) {
			result = append(result, data)
		}
		return true
	})

	// finally, gather all exports that are in the queue
	ew.QueueItemsMutex.Lock()
	defer ew.QueueItemsMutex.Unlock()

	for _, item := range ew.QueueItems {
		if item.ContainsUser(user) {
			result = append(result, *item)
		}
	}

	return result
}

// DequeueWrapperItem Dequeues an item, returns size of queue and true if item was found and dequeued
func (ew *Wrapper) DequeueWrapperItem(item *WrapperItem) (int, bool) {
	ew.QueueItemsMutex.Lock()
	defer ew.QueueItemsMutex.Unlock()

	for i, queueItem := range ew.QueueItems {
		// comparing pointer should work
		if queueItem != item {
			continue
		}

		ew.QueueItems = append(ew.QueueItems[:i], ew.QueueItems[i+1:]...)
		return len(ew.QueueItems), true
	}

	return len(ew.QueueItems), false
}

// dispatchExportQueue dispatches the export queue to the available workers
func (ew *Wrapper) dispatchExportQueue(ctx context.Context) {
	for _, worker := range ew.Workers {
		worker.Mutex.Lock()
		if worker.Available {
			// check if there is an item in the queue
			ew.QueueItemsMutex.Lock()

			if len(ew.QueueItems) == 0 {
				ew.QueueItemsMutex.Unlock()
				worker.Mutex.Unlock()
				return // Nothing in queue
			}

			item := *ew.QueueItems[0]
			ew.QueueItems = append(ew.QueueItems[:0], ew.QueueItems[1:]...)
			ew.QueueItemsMutex.Unlock()

			worker.Available = false
			worker.Mutex.Unlock()
			go worker.Start(item, ctx)
		} else {
			worker.Mutex.Unlock()
		}
	}
}

func (ew *Wrapper) FindArchive(id string, user users.User) (WrapperItem, bool) {
	item, found := ew.Archive.Load(id)
	if found {
		if data, ok := item.(WrapperItem); ok && data.ContainsUser(user) {
			return data, true
		}
	}
	return WrapperItem{}, false
}

// ContainsUser checks if user is in item
func (it *WrapperItem) ContainsUser(user users.User) bool {
	for _, u := range it.Users {
		if u == user.Login {
			return true
		}
	}
	return false
}

//func (ew *Wrapper) CancelExport(id string, user users.User) error {
//	// first check if the export is in the queue
//	// if it is, we check if the user is the only one in the queueItem.users
//	// if yes, we remove the queueItem from the queue
//	// if no, we remove the user from the queueItem.users
//
//	for i, worker := range ew.Workers {
//
//		worker.Mutex.Lock()
//		if worker.QueueItem == nil || worker.QueueItem.Id != id {
//			worker.Mutex.Unlock()
//			continue
//		}
//		worker.Mutex.Lock()
//
//		if userIdx == -1 {
//			worker.Mutex.Unlock()
//			ew.QueueMutex.Unlock()
//			return fmt.Errorf("user not found")
//		}
//
//		if len(worker.Users) == 1 {
//			ew.Queue = append(ew.Queue[:userIdx], ew.Queue[userIdx+1:]...)
//			worker.Mutex.Unlock()
//			ew.QueueMutex.Unlock()
//			return nil
//		}
//
//		worker.Users = append(worker.Users[:i], worker.Users[i+1:]...)
//		worker.Mutex.Unlock()
//		ew.QueueMutex.Unlock()
//		return nil
//	}
//
//	ew.QueueMutex.Unlock()
//
//	return nil
//}
