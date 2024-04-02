package export

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/notifier"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/security"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/security/users"
	"github.com/myrteametrics/myrtea-sdk/v4/engine"
	"go.uber.org/zap"
)

const (
	CodeUserAdded  = 1
	CodeAdded      = 0
	CodeUserExists = -1
	CodeQueueFull  = -2

	// WrapperItem statuses
	StatusPending   = 0
	StatusRunning   = 1
	StatusDone      = 2
	StatusError     = 3
	StatusCanceled  = 4
	StatusCanceling = 5

	// Delete return codes
	DeleteExportNotFound    = 0
	DeleteExportDeleted     = 1
	DeleteExportUserDeleted = 2
	DeleteExportCanceled    = 3

	randCharSet = "abcdefghijklmnopqrstuvwxyz0123456789"
)

// WrapperItem represents an export demand
type WrapperItem struct {
	Id           string            `json:"id"`      // unique id that represents an export demand
	FactIDs      []int64           `json:"factIds"` // list of fact ids that are part of the export (for archive and json)
	Facts        []engine.Fact     `json:"-"`
	Error        string            `json:"error"`
	Status       int               `json:"status"`
	FileName     string            `json:"fileName"`
	Title        string            `json:"title"`
	Date         time.Time         `json:"date"`
	Users        []string          `json:"-"`
	Params       CSVParameters     `json:"-"`
	Placeholders map[string]string `json:"placeholders"`
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
	BasePath          string // public for export_handlers
	queueMaxSize      int
	workerCount       int
}

// NewWrapperItem creates a new export wrapper item
func NewWrapperItem(facts []engine.Fact, title string, params CSVParameters, user users.User, placeholders map[string]string) *WrapperItem {
	var factIDs []int64
	for _, fact := range facts {
		factIDs = append(factIDs, fact.ID)
	}

	// file extension should be gz
	// add random string to avoid multiple files with same name
	fileName := security.RandStringWithCharset(5, randCharSet) + "_" +
		strings.ReplaceAll(title, " ", "_") + ".csv.gz"

	return &WrapperItem{
		Users:        append([]string{}, user.Login),
		Id:           uuid.New().String(),
		Facts:        facts,
		FactIDs:      factIDs,
		Date:         time.Now(),
		Status:       StatusPending,
		Error:        "",
		FileName:     fileName,
		Title:        title,
		Params:       params,
		Placeholders: placeholders,
	}
}

// NewWrapper creates a new export wrapper
func NewWrapper(basePath string, workersCount, diskRetentionDays, queueMaxSize int) *Wrapper {
	wrapper := &Wrapper{
		workers:           make([]*ExportWorker, 0),
		queue:             make([]*WrapperItem, 0),
		success:           make(chan int),
		archive:           sync.Map{},
		queueMaxSize:      queueMaxSize,
		BasePath:          basePath,
		diskRetentionDays: diskRetentionDays,
		workerCount:       workersCount,
	}

	return wrapper
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
		ew.workers = append(ew.workers, NewExportWorker(i, ew.BasePath, ew.success))
	}

	// check if destination folder exists
	_, err := os.Stat(ew.BasePath)
	if err != nil {

		if os.IsNotExist(err) {
			zap.L().Info("The export directory not exists, trying to create...", zap.String("EXPORT_BASE_PATH", ew.BasePath))

			if err := os.MkdirAll(ew.BasePath, os.ModePerm); err != nil {
				zap.L().Error("Couldn't create export directory", zap.String("EXPORT_BASE_PATH", ew.BasePath), zap.Error(err))
			} else {
				zap.L().Info("The export directory has been successfully created.")
			}

		} else {
			zap.L().Error("Couldn't access to export directory", zap.String("EXPORT_BASE_PATH", ew.BasePath), zap.Error(err))
		}

	}

	go ew.startDispatcher(ctx)
}

// factsEquals checks if two slices of facts are equal
func factsEquals(a, b []engine.Fact) bool {
	if len(a) != len(b) {
		return false
	}
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
	return true
}

// AddToQueue Adds a new export to the export worker queue
func (ew *Wrapper) AddToQueue(facts []engine.Fact, title string, params CSVParameters, user users.User, placeholders map[string]string) (*WrapperItem, int) {
	ew.queueMutex.Lock()
	defer ew.queueMutex.Unlock()

	for _, queueItem := range ew.queue {
		if !factsEquals(queueItem.Facts, facts) || !queueItem.Params.Equals(params) || queueItem.Title != title {
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

	item := NewWrapperItem(facts, title, params, user, placeholders)
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

			// archive item when finished
			worker.Mutex.Lock()
			ew.workers[w].Available = true
			item := worker.QueueItem
			worker.QueueItem = WrapperItem{}
			worker.Mutex.Unlock()

			// archive item
			item.Facts = []engine.Fact{} // empty facts to avoid storing them in the archive
			ew.archive.Store(item.Id, item)

			// send notification to user (non-blocking)
			go func(wrapperItem WrapperItem) {
				_ = notifier.C().SendToUserLogins(
					createExportNotification(ExportNotificationArchived, &wrapperItem),
					wrapperItem.Users)
			}(item)
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
	files, err := os.ReadDir(ew.BasePath)
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

			// send notification to user (non-blocking)
			go func(wrapperItem WrapperItem) {
				_ = notifier.C().SendToUserLogins(
					createExportNotification(ExportNotificationDeleted, &wrapperItem),
					wrapperItem.Users)
			}(data)

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
	result := make([]WrapperItem, 0)

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
		if queueItem.Id != item.Id {
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
		if !worker.Available {
			worker.Mutex.Unlock()
			continue
		}
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
// this function is similar to GetUserExports, but it avoids iterating over all exports, thus it is faster
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

// DeleteExport removes an export from the queue / archive, or cancels it if it is running
// returns :
// DeleteExportNotFound (0): if the export was not found
// DeleteExportDeleted (1): if the export was found and deleted
// DeleteExportUserDeleted (2): if the export was found and the user was removed
// DeleteExportCanceled (3): if the export was found and the cancellation request was made
// this function is similar to GetUserExport, but it avoids iterating over all exports, thus it is faster
func (ew *Wrapper) DeleteExport(id string, user users.User) int {
	// start with archived items
	if item, ok := ew.FindArchive(id, user); ok {
		if len(item.Users) == 1 {
			ew.archive.Delete(id)
			return DeleteExportDeleted
		}
		// remove user from item
		for i, u := range item.Users {
			if u == user.Login {
				item.Users = append(item.Users[:i], item.Users[i+1:]...)
				break
			}
		}
		ew.archive.Store(id, item)
		return DeleteExportUserDeleted
	}

	// then check the queue
	ew.queueMutex.Lock()
	for i, item := range ew.queue {
		if item.Id == id && item.ContainsUser(user) {
			// remove user from item
			for j, u := range item.Users {
				if u == user.Login {
					item.Users = append(item.Users[:j], item.Users[j+1:]...)
					break
				}
			}
			if len(item.Users) == 0 {
				ew.queue = append(ew.queue[:i], ew.queue[i+1:]...)
				ew.queueMutex.Unlock()
				return DeleteExportDeleted
			}

			ew.queueMutex.Unlock()
			return DeleteExportUserDeleted
		}
	}
	ew.queueMutex.Unlock()

	// finally check the workers
	for _, worker := range ew.workers {
		worker.Mutex.Lock()
		if worker.Available || worker.QueueItem.Id != id || !worker.QueueItem.ContainsUser(user) {
			worker.Mutex.Unlock()
			continue
		}

		// worker found but already canceling
		if worker.QueueItem.Status == StatusCanceling {
			worker.Mutex.Unlock()
			return DeleteExportNotFound
		}

		// remove user from item
		if len(worker.QueueItem.Users) == 1 {
			// cancel worker by sending a message on the cancel channel
			// the worker will check this channel and stop if it receives a message
			// it can happen that the worker is already stopped, in this case, the message will be ignored
			select { // non-blocking send
			case worker.Cancel <- true:
			default:
			}
			worker.QueueItem.Status = StatusCanceling
			worker.Mutex.Unlock()
			return DeleteExportCanceled
		}

		for i, u := range worker.QueueItem.Users {
			if u == user.Login {
				worker.QueueItem.Users = append(worker.QueueItem.Users[:i], worker.QueueItem.Users[i+1:]...)
				worker.Mutex.Unlock()
				return DeleteExportUserDeleted
			}
		}
		worker.Mutex.Unlock()
		return DeleteExportNotFound
	}

	return DeleteExportNotFound
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
