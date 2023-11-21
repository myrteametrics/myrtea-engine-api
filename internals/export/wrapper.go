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
	Facts    []engine.Fact `json:"-"`
	Error    error         `json:"error"`
	Status   int           `json:"status"`
	FileName string        `json:"fileName"`
	Date     time.Time     `json:"date"`
	Users    []string      `json:"-"`
	Params   CSVParameters `json:"-"`
}

type Wrapper struct {
	// Queue handling
	queueMutex sync.RWMutex
	queue      []*WrapperItem // stores queue to handle duplicates, state

	// contains also current handled items
	// workers is final, its only instanced once and thus does not change size (ExportWorker have there indexes in this slice stored)
	workers []*ExportWorker

	// success is passed to all workers, they write on this channel when they've finished with there export
	success chan int

	// Archived WrapperItem's
	archive sync.Map // map of all exports that have been done, key is the id of the export

	// Non-critical fields
	// Read-only parameters
	diskRetentionDays int
	basePath          string
	queueMaxSize      int
	workerCount       int
}

// NewWrapperItem creates a new export wrapper item
func NewWrapperItem(facts []engine.Fact, fileName string, params CSVParameters, user users.User) *WrapperItem {
	return &WrapperItem{
		Users:    append([]string{}, user.Login),
		Id:       uuid.New().String(),
		Facts:    facts,
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
		workers:           make([]*ExportWorker, 0),
		queue:             make([]*WrapperItem, 0),
		success:           make(chan int),
		archive:           sync.Map{},
		queueMaxSize:      queueMaxSize,
		basePath:          basePath,
		diskRetentionDays: diskRetentionDays,
		workerCount:       workersCount,
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
	for i := 0; i < ew.workerCount; i++ {
		ew.workers = append(ew.workers, NewExportWorker(i, ew.basePath, ew.success))
	}
	go ew.startDispatcher(ctx)
}

// factsEquals checks if two slices of facts are equal
func factsEquals(a, b []engine.Fact) bool {
	for _, fact := range a {
		found := false
		for _, fact2 := range b {
			if fact.ID == fact2.ID {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return false
}

// AddToQueue Adds a new export to the export worker queue
func (ew *Wrapper) AddToQueue(facts []engine.Fact, fileName string, params CSVParameters, user users.User) (*WrapperItem, int) {
	ew.queueMutex.Lock()
	defer ew.queueMutex.Unlock()

	for _, queueItem := range ew.queue {
		if !factsEquals(queueItem.Facts, facts) || !queueItem.Params.Equals(params) {
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

	if len(ew.queue) >= ew.queueMaxSize {
		return nil, CodeQueueFull
	}

	item := NewWrapperItem(facts, fileName, params, user)
	ew.queue = append(ew.queue, item)
	return item, CodeAdded
}

// startDispatcher starts the export tasks dispatcher & the expired files checker
func (ew *Wrapper) startDispatcher(context context.Context) {
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
		case w := <-ew.success:
			worker := ew.workers[w]
			// TODO: send notifications here

			// archive item when finished
			worker.Mutex.Lock()
			ew.workers[w].Available = true
			item := worker.QueueItem
			worker.QueueItem = WrapperItem{}
			worker.Mutex.Unlock()
			// archive item
			ew.archive.Store(item.Id, item)
		case <-ticker.C:
			ew.dispatchExportQueue(context)
		case <-expiredFileTicker.C:
			err := ew.checkForExpiredFiles()

			if err != nil {
				zap.L().Error("Error during expired files check", zap.Error(err))
			}
		case <-context.Done():
			return
		}
	}
}

// checkForExpiredFiles checks for expired files in the export directory and deletes them
// it also deletes the done tasks that are older than diskRetentionDays
func (ew *Wrapper) checkForExpiredFiles() error {
	// Get all files in directory and check the last edit date
	// if last edit date is older than diskRetentionDays, delete the file
	zap.L().Info("Checking for expired files")
	files, err := os.ReadDir(ew.basePath)
	if err != nil {
		return err
	}

	// delete all done archives of ew.archive that are older than diskRetentionDays
	ew.archive.Range(func(key, value any) bool {
		data, ok := value.(WrapperItem)
		if !ok {
			return true
		}
		if time.Since(data.Date).Hours() > float64(ew.diskRetentionDays*24) {
			ew.archive.Delete(key)
		}
		return true
	})

	// count the number of deleted files
	count := 0

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filePath := filepath.Join(ew.basePath, file.Name())

		fi, err := os.Stat(filePath)
		if err != nil {
			zap.L().Error("Cannot get file info", zap.String("file", filePath), zap.Error(err))
			continue
		}

		// skip if file is not a zip
		//if filepath.Ext(file.Name()) != ".zip" {
		//	continue
		//}

		if time.Since(fi.ModTime()).Hours() > float64(ew.diskRetentionDays*24) {
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
	for _, worker := range ew.workers {
		worker.Mutex.Lock()
		if worker.QueueItem.ContainsUser(user) {
			result = append(result, worker.QueueItem)
		}
		worker.Mutex.Unlock()
	}

	// then, gather all exports that are archived
	ew.archive.Range(func(key, value any) bool {
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
	ew.queueMutex.Lock()
	defer ew.queueMutex.Unlock()

	for _, item := range ew.queue {
		if item.ContainsUser(user) {
			result = append(result, *item)
		}
	}

	return result
}

// dequeueWrapperItem Dequeues an item, returns size of queue and true if item was found and dequeued
func (ew *Wrapper) dequeueWrapperItem(item *WrapperItem) (int, bool) {
	ew.queueMutex.Lock()
	defer ew.queueMutex.Unlock()

	for i, queueItem := range ew.queue {
		// comparing pointer should work
		if queueItem != item {
			continue
		}

		ew.queue = append(ew.queue[:i], ew.queue[i+1:]...)
		return len(ew.queue), true
	}

	return len(ew.queue), false
}

// dispatchExportQueue dispatches the export queue to the available workers
func (ew *Wrapper) dispatchExportQueue(ctx context.Context) {
	for _, worker := range ew.workers {
		worker.Mutex.Lock()
		if worker.Available {
			// check if there is an item in the queue
			ew.queueMutex.Lock()

			if len(ew.queue) == 0 {
				ew.queueMutex.Unlock()
				worker.Mutex.Unlock()
				return // Nothing in queue
			}

			item := *ew.queue[0]
			ew.queue = append(ew.queue[:0], ew.queue[1:]...)
			ew.queueMutex.Unlock()

			worker.Available = false
			worker.Mutex.Unlock()
			go worker.Start(item, ctx)
		} else {
			worker.Mutex.Unlock()
		}
	}
}

// FindArchive returns the archive item for the given id and user
func (ew *Wrapper) FindArchive(id string, user users.User) (WrapperItem, bool) {
	item, found := ew.archive.Load(id)
	if found {
		if data, ok := item.(WrapperItem); ok && data.ContainsUser(user) {
			return data, true
		}
	}
	return WrapperItem{}, false
}

// GetUserExport returns the export item for the given id and user
// this function is similar to GetUserExports but it avoids iterating over all exports, thus it is faster
func (ew *Wrapper) GetUserExport(id string, user users.User) (item WrapperItem, ok bool) {
	// start with archived items
	if item, ok = ew.FindArchive(id, user); ok {
		return item, ok
	}

	// then check the workers
	for _, worker := range ew.workers {
		worker.Mutex.Lock()
		if worker.QueueItem.Id == id && worker.QueueItem.ContainsUser(user) {
			item = worker.QueueItem
			ok = true
		}
		worker.Mutex.Unlock()
		if ok {
			return item, ok
		}
	}

	// finally check the queue
	ew.queueMutex.Lock()
	defer ew.queueMutex.Unlock()

	for _, it := range ew.queue {
		ok = it.ContainsUser(user)
		if ok {
			item = *it
			break
		}
	}

	return item, ok
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
//	for i, worker := range ew.workers {
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
