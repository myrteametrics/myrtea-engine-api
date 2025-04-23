package export

import (
	"context"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8/typedapi/core/search"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/notifier"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/security/users"
	"github.com/myrteametrics/myrtea-sdk/v5/engine"
	"github.com/myrteametrics/myrtea-sdk/v5/expression"
)

func TestNewWrapper(t *testing.T) {
	wrapper := NewWrapper("/tmp", 1, 1, 1)
	expression.AssertEqual(t, wrapper.BasePath, "/tmp")
	expression.AssertEqual(t, wrapper.queueMaxSize, 1)
	expression.AssertEqual(t, wrapper.diskRetentionDays, 1)
	expression.AssertEqual(t, wrapper.queueMaxSize, 1)
}

func TestFactsEquals(t *testing.T) {
	expression.AssertEqual(t, factsEquals([]engine.Fact{{ID: 1}}, []engine.Fact{{ID: 1}}), true)
	expression.AssertEqual(t, factsEquals([]engine.Fact{{ID: 1}}, []engine.Fact{{ID: 2}}), false)
	expression.AssertEqual(t, factsEquals([]engine.Fact{{ID: 1}, {ID: 2}}, []engine.Fact{{ID: 2}, {ID: 1}}), true)
	expression.AssertEqual(t, factsEquals([]engine.Fact{{ID: 1}, {ID: 2}}, []engine.Fact{{ID: 1}, {ID: 3}}), false)
	expression.AssertEqual(t, factsEquals([]engine.Fact{{ID: 1}, {ID: 2}}, []engine.Fact{{ID: 1}, {ID: 2}, {ID: 3}}), false)
	expression.AssertEqual(t, factsEquals([]engine.Fact{{ID: 2}, {ID: 1}, {ID: 3}}, []engine.Fact{{ID: 1}, {ID: 2}}), false)
}

func TestNewWrapperItem(t *testing.T) {
	item := NewWrapperItem([]engine.Fact{{ID: 1}}, "test", CSVParameters{}, users.User{Login: "test"}, make(map[string]string), false)
	expression.AssertNotEqual(t, item.Id, "")
	expression.AssertEqual(t, factsEquals(item.Facts, []engine.Fact{{ID: 1}}), true)
	expression.AssertEqual(t, item.Params.Equals(CSVParameters{}), true)
	expression.AssertEqual(t, item.Status, StatusPending)
	expression.AssertEqual(t, item.FileName, "test.csv.gz")
	expression.AssertNotEqual(t, len(item.Users), 0)
	expression.AssertEqual(t, item.Users[0], "test")

	item = NewWrapperItem([]engine.Fact{{ID: 1}}, "test.csv.gz", CSVParameters{}, users.User{Login: "test"}, make(map[string]string), false)
	expression.AssertEqual(t, item.FileName, "test.csv.gz")

	// test file name formatting
	item = NewWrapperItem([]engine.Fact{{ID: 1}}, "test", CSVParameters{}, users.User{Login: "test"}, make(map[string]string), true)
	expression.AssertEqual(t, strings.HasSuffix(item.FileName, "test.csv.gz"), true, "Expected file name to end with test.csv.gz")
}

func TestWrapperItem_ContainsFact(t *testing.T) {
	item := NewWrapperItem([]engine.Fact{{ID: 1}, {ID: 22}, {ID: 33}}, "test.txt", CSVParameters{}, users.User{Login: "test"}, make(map[string]string), false)
	expression.AssertEqual(t, item.ContainsFact(1), true)
	expression.AssertEqual(t, item.ContainsFact(22), true)
	expression.AssertEqual(t, item.ContainsFact(3), false)
}

func TestWrapper_Init(t *testing.T) {
	wrapper := NewWrapper("/tmp", 1, 1, 1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	wrapper.Init(ctx)
	time.Sleep(500 * time.Millisecond)
	expression.AssertEqual(t, len(wrapper.workers), 1)
	worker := wrapper.workers[0]
	expression.AssertEqual(t, worker.Id, 0)
	worker.Mutex.Lock()
	defer worker.Mutex.Unlock()
	expression.AssertEqual(t, worker.Available, true)
}

func TestAddToQueue(t *testing.T) {
	wrapper := NewWrapper("/tmp", 1, 1, 1)
	user1 := users.User{Login: "bla"}
	user2 := users.User{Login: "blabla"}
	csvParams := CSVParameters{}
	_, result := wrapper.AddToQueue([]engine.Fact{{ID: 1}}, "test.txt", csvParams, user1, make(map[string]string), false)
	expression.AssertEqual(t, result, CodeAdded, "AddToQueue should return CodeAdded")
	_, result = wrapper.AddToQueue([]engine.Fact{{ID: 1}}, "test.txt", csvParams, user1, make(map[string]string), false)
	expression.AssertEqual(t, result, CodeUserExists, "AddToQueue should return CodeUserExists")
	_, result = wrapper.AddToQueue([]engine.Fact{{ID: 1}}, "test.txt", csvParams, user2, make(map[string]string), false)
	expression.AssertEqual(t, result, CodeUserAdded, "AddToQueue should return CodeUserAdded")
	_, result = wrapper.AddToQueue([]engine.Fact{{ID: 2}}, "test.txt", csvParams, user2, make(map[string]string), false)
	expression.AssertEqual(t, result, CodeQueueFull, "AddToQueue should return CodeQueueFull")
}

func TestAddToQueueCustom(t *testing.T) {
	wrapper := NewWrapper("/tmp", 1, 1, 1)
	user1 := users.User{Login: "bla"}
	user2 := users.User{Login: "blabla"}
	esParams := ElasticParams{}
	csvParams := CSVParameters{}
	_, result := wrapper.AddToQueueCustom("test.txt", esParams, []search.Request{}, csvParams, user1, false)
	expression.AssertEqual(t, result, CodeAdded, "AddToQueueCustom should return CodeAdded")
	_, result = wrapper.AddToQueueCustom("test.txt", esParams, []search.Request{}, csvParams, user1, false)
	expression.AssertEqual(t, result, CodeUserExists, "AddToQueueCustom should return CodeUserExists")
	_, result = wrapper.AddToQueueCustom("test.txt", esParams, []search.Request{}, csvParams, user2, false)
	expression.AssertEqual(t, result, CodeUserAdded, "AddToQueueCustom should return CodeUserAdded")
	_, result = wrapper.AddToQueueCustom("test2.txt", esParams, []search.Request{}, csvParams, user2, false)
	expression.AssertEqual(t, result, CodeQueueFull, "AddToQueueCustom should return CodeQueueFull")
}

func TestStartDispatcher(t *testing.T) {
	// we don't want that the worker try to export data, therefore we will create a temporary directory with a temp file
	// so that the worker will not be able to create the file and will return an error
	dname, err := os.MkdirTemp("", "exportdispatcher")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	defer os.RemoveAll(dname)

	file, err := os.Create(filepath.Join(dname, "exportdispatcher.csv.gz"))
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	fileName := filepath.Base(file.Name())
	_ = file.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	notifier.ReplaceGlobals(notifier.NewNotifier()) // for notifications
	wrapper := NewWrapper(dname, 1, 1, 1)
	wrapper.Init(ctx)
	expression.AssertEqual(t, len(wrapper.workers), 1)
	// sleep one second to let the goroutine start
	fmt.Println("Sleeping 1 second to let the goroutine start")
	time.Sleep(1 * time.Second)

	worker := wrapper.workers[0]

	// check if the worker is available
	worker.Mutex.Lock()
	expression.AssertEqual(t, worker.Available, true)
	worker.Mutex.Unlock()

	// add a task to the queue and check if the task was added to queue
	user := users.User{Login: "test"}
	_, result := wrapper.AddToQueue([]engine.Fact{{ID: 1, Intent: &engine.IntentFragment{}}}, fileName, CSVParameters{}, user, make(map[string]string), false)
	expression.AssertEqual(t, result, CodeAdded, "AddToQueue should return CodeAdded")
	wrapper.queueMutex.Lock()
	expression.AssertEqual(t, len(wrapper.queue), 1)
	itemId := wrapper.queue[0].Id
	wrapper.queueMutex.Unlock()

	// sleep another 5 seconds to let the goroutine handle the task
	fmt.Println("Sleeping 5 seconds to let the goroutine handle the task")
	time.Sleep(5 * time.Second)

	wrapper.queueMutex.Lock()
	expression.AssertEqual(t, len(wrapper.queue), 0)
	wrapper.queueMutex.Unlock()

	worker.Mutex.Lock()
	expression.AssertEqual(t, worker.Available, true)
	worker.Mutex.Unlock()

	time.Sleep(50 * time.Millisecond)

	item, ok := wrapper.FindArchive(itemId, user)
	expression.AssertEqual(t, ok, true)
	expression.AssertEqual(t, item.Status, StatusError) // could not create file
}

func TestCheckForExpiredFiles(t *testing.T) {
	// first test : check if files are deleted
	dname, err := os.MkdirTemp("", "export")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	defer os.RemoveAll(dname)

	// create a file that is 2 days old
	file, err := os.CreateTemp(dname, "export")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	file1Name := file.Name()
	_ = file.Close()
	err = os.Chtimes(file1Name, time.Now().AddDate(0, 0, -2), time.Now().AddDate(0, 0, -2))
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	// create a freshly created file
	file2, err := os.CreateTemp(dname, "export")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	file2Name := file2.Name()
	_ = file2.Close()

	wrapper := NewWrapper(dname, 1, 1, 1)
	err = wrapper.checkForExpiredFiles()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	// check that the file has been deleted
	_, err = os.Stat(file1Name)
	if !os.IsNotExist(err) {
		t.Error("File1 should have been deleted")
		t.FailNow()
	}

	_, err = os.Stat(file2Name)
	if os.IsNotExist(err) {
		t.Error("File2 should not have been deleted")
		t.FailNow()
	}

	// second test : check if expired exports are deleted
	goodDate := time.Now()
	id1 := uuid.New()
	id2 := uuid.New()
	wrapper.archive.Store(id1, WrapperItem{Date: time.Now().AddDate(0, 0, -2)})
	wrapper.archive.Store(id2, WrapperItem{Date: goodDate})

	_, found := wrapper.archive.Load(id1)
	expression.AssertEqual(t, found, true)
	_, found = wrapper.archive.Load(id2)
	expression.AssertEqual(t, found, true)

	err = wrapper.checkForExpiredFiles()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	_, found = wrapper.archive.Load(id1)
	expression.AssertEqual(t, found, false)
	_, found = wrapper.archive.Load(id2)
	expression.AssertEqual(t, found, true)
}

func TestWrapper_GetUserExports(t *testing.T) {
	wrapper := NewWrapper("/tmp", 1, 1, 2)
	user1 := users.User{Login: "bla"}
	user2 := users.User{Login: "blabla"}
	item1 := NewWrapperItem([]engine.Fact{{ID: 1}}, "test.txt", CSVParameters{}, user1, make(map[string]string), false)
	item2 := NewWrapperItem([]engine.Fact{{ID: 2}}, "test.txt", CSVParameters{}, user1, make(map[string]string), false)
	item3 := NewWrapperItem([]engine.Fact{{ID: 3}}, "test.txt", CSVParameters{}, user1, make(map[string]string), false)
	item4 := NewWrapperItem([]engine.Fact{{ID: 4}}, "test.txt", CSVParameters{}, user2, make(map[string]string), false)
	wrapper.archive.Store(item1.Id, *item1)
	wrapper.archive.Store(item2.Id, *item2)
	wrapper.archive.Store(item3.Id, *item3)
	wrapper.archive.Store(item4.Id, *item4)
	wrapper.AddToQueue([]engine.Fact{{ID: 5}}, "test.txt", CSVParameters{}, user1, make(map[string]string), false)
	wrapper.AddToQueue([]engine.Fact{{ID: 6}}, "test.txt", CSVParameters{}, user2, make(map[string]string), false)
	exports := wrapper.GetUserExports(user1)
	expression.AssertEqual(t, len(exports), 4)
	exports = wrapper.GetUserExports(user2)
	expression.AssertEqual(t, len(exports), 2)
}

func TestWrapper_DequeueWrapperItem(t *testing.T) {
	wrapper := NewWrapper("/tmp", 1, 1, 2)
	i, ok := wrapper.dequeueWrapperItem(&WrapperItem{})
	expression.AssertEqual(t, ok, false)
	expression.AssertEqual(t, i, 0)
	wrapper.AddToQueue([]engine.Fact{{ID: 5}}, "test.txt", CSVParameters{}, users.User{Login: "bla"}, make(map[string]string), false)
	wrapper.AddToQueue([]engine.Fact{{ID: 6}}, "test.txt", CSVParameters{}, users.User{Login: "blabla"}, make(map[string]string), false)

	expression.AssertEqual(t, len(wrapper.queue), 2)
	item1 := wrapper.queue[0]
	item2 := wrapper.queue[1]

	i, ok = wrapper.dequeueWrapperItem(item1)
	expression.AssertEqual(t, ok, true)
	expression.AssertEqual(t, i, 1)

	i, ok = wrapper.dequeueWrapperItem(item2)
	expression.AssertEqual(t, ok, true)
	expression.AssertEqual(t, i, 0)
}

func TestWrapper_dispatchExportQueue(t *testing.T) {
	// we don't want that the worker try to export data, therefore we will create a temporary directory with a temp file
	// so that the worker will not be able to create the file and will return an error
	dname, err := os.MkdirTemp("", "exportdispatcher")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	defer os.RemoveAll(dname)

	// create a file that is 2 days old
	file, err := os.Create(filepath.Join(dname, "exportdispatcher.csv.gz"))
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	fileName := filepath.Base(file.Name())
	_ = file.Close()

	notifier.ReplaceGlobals(notifier.NewNotifier()) // for notifications
	wrapper := NewWrapper(dname, 1, 1, 2)
	ctx, cancel := context.WithCancel(context.Background())
	wrapper.Init(ctx)
	cancel() // stop dispatcher since we don't want him to interact with the workers or the queue

	// wait until dispatcher stops
	time.Sleep(50 * time.Millisecond)

	expression.AssertEqual(t, len(wrapper.workers), 1)
	worker := wrapper.workers[0]

	// no items in queue -> nothing should happen
	expression.AssertEqual(t, worker.IsAvailable(), true)
	wrapper.dispatchExportQueue(context.Background())
	expression.AssertEqual(t, worker.IsAvailable(), true, "worker should still be available, because no items in queue")

	// we add an item to the queue
	wrapper.AddToQueue([]engine.Fact{{ID: 1}}, fileName, CSVParameters{}, users.User{Login: "test"}, make(map[string]string), false)

	// we test if dispatchExportQueue will not dispatch the item, no worker available
	worker.SwapAvailable(false)

	wrapper.dispatchExportQueue(context.Background())

	// the item should still be in the queue
	wrapper.queueMutex.Lock()
	expression.AssertEqual(t, len(wrapper.queue), 1, "item should still be in the queue, since no worker is available")
	wrapper.queueMutex.Unlock()

	// we test if dispatchExportQueue will dispatch the item, worker is now set to available
	expression.AssertEqual(t, worker.SwapAvailable(true), false)

	wrapper.dispatchExportQueue(context.Background())

	expression.AssertEqual(t, worker.IsAvailable(), false, "worker should not be available, because it is working on an item")
	expression.AssertEqual(t, len(wrapper.queue), 0)

	// wait until worker has finished
	time.Sleep(1 * time.Second)

	worker.Mutex.Lock()
	defer worker.Mutex.Unlock()

	expression.AssertEqual(t, worker.QueueItem.Status, StatusError, fmt.Sprintf("worker processed item should have StatusError(%d) because the file already exists", StatusError)) // could not create file
}

func TestWrapper_FindArchive(t *testing.T) {
	wrapper := NewWrapper("/tmp", 1, 1, 2)
	item := NewWrapperItem([]engine.Fact{{ID: 1}}, "test.txt", CSVParameters{}, users.User{Login: "bla"}, make(map[string]string), false)
	wrapper.archive.Store(item.Id, *item)

	// testing with non-existing item in archive
	_, ok := wrapper.FindArchive("test", users.User{Login: "bla"})
	expression.AssertEqual(t, ok, false)

	// testing with existing item but not good user in archive
	_, ok = wrapper.FindArchive("test", users.User{Login: "blabla"})
	expression.AssertEqual(t, ok, false)

	// testing with existing item in archive
	_, ok = wrapper.FindArchive(item.Id, users.User{Login: "bla"})
	expression.AssertEqual(t, ok, true)
}

func TestWrapper_ContainsUser(t *testing.T) {
	item := NewWrapperItem([]engine.Fact{{ID: 1}}, "test.txt", CSVParameters{}, users.User{Login: "bla"}, make(map[string]string), false)
	expression.AssertEqual(t, item.ContainsUser(users.User{Login: "bla"}), true)
	expression.AssertEqual(t, item.ContainsUser(users.User{Login: "blabla"}), false)
}

func TestWrapper_DeleteExport(t *testing.T) {
	wrapper := NewWrapper("/tmp", 1, 1, 2)
	item := NewWrapperItem([]engine.Fact{{ID: 1}}, "test.txt", CSVParameters{}, users.User{Login: "bla"}, make(map[string]string), false)

	// test archive
	wrapper.archive.Store(item.Id, *item)
	expression.AssertEqual(t, wrapper.DeleteExport(item.Id, users.User{Login: "bla"}), DeleteExportDeleted, "item should have been deleted")
	_, ok := wrapper.archive.Load(item.Id)
	expression.AssertEqual(t, ok, false, "item should not be in archive anymore")

	// test archive multi-user
	item.Users = []string{"bla", "blabla"}
	wrapper.archive.Store(item.Id, *item)
	expression.AssertEqual(t, wrapper.DeleteExport(item.Id, users.User{Login: "bla"}), DeleteExportUserDeleted, "user should have been deleted from existing export")
	_, ok = wrapper.archive.Load(item.Id)
	expression.AssertEqual(t, ok, true, "item should be in archive")
	item.Users = []string{"bla"}

	// test queue
	queueItem, code := wrapper.AddToQueue([]engine.Fact{{ID: 1}}, "test.txt", CSVParameters{}, users.User{Login: "bla"}, make(map[string]string), false)
	expression.AssertEqual(t, code, CodeAdded, "item should have been added to queue")
	wrapper.queueMutex.Lock()
	expression.AssertEqual(t, len(wrapper.queue), 1, "item should be in queue")
	wrapper.queueMutex.Unlock()
	expression.AssertEqual(t, wrapper.DeleteExport(queueItem.Id, users.User{Login: "bla"}), DeleteExportDeleted, "item should have been deleted")
	wrapper.queueMutex.Lock()
	expression.AssertEqual(t, len(wrapper.queue), 0, "item should not be in queue anymore")
	wrapper.queueMutex.Unlock()

	// test queue multi-user
	queueItem, code = wrapper.AddToQueue([]engine.Fact{{ID: 1}}, "test.txt", CSVParameters{}, users.User{Login: "bla"}, make(map[string]string), false)
	expression.AssertEqual(t, code, CodeAdded, "item should have been added to queue")
	_, code = wrapper.AddToQueue([]engine.Fact{{ID: 1}}, "test.txt", CSVParameters{}, users.User{Login: "blabla"}, make(map[string]string), false)
	expression.AssertEqual(t, code, CodeUserAdded, "user should have been added to existing item in queue")
	wrapper.queueMutex.Lock()
	expression.AssertEqual(t, len(wrapper.queue), 1, "item should be in queue")
	wrapper.queueMutex.Unlock()
	expression.AssertEqual(t, wrapper.DeleteExport(queueItem.Id, users.User{Login: "bla"}), DeleteExportUserDeleted, "user should have been deleted from existing export")
	wrapper.queueMutex.Lock()
	expression.AssertEqual(t, len(wrapper.queue), 1, "item should be in queue")
	wrapper.queueMutex.Unlock()

	// test workers
	item.Users = []string{"bla", "blibli"}
	worker := NewExportWorker(0, "/tmp", make(chan<- int))
	wrapper.workers = append(wrapper.workers, worker)
	worker.Mutex.Lock()
	worker.QueueItem = *item
	worker.Available = true
	worker.Mutex.Unlock()
	expression.AssertEqual(t, wrapper.DeleteExport(item.Id, users.User{Login: "bla"}), DeleteExportNotFound, "item should have not been deleted")
	worker.SwapAvailable(false)
	expression.AssertEqual(t, wrapper.DeleteExport(item.Id, users.User{Login: "blibli"}), DeleteExportUserDeleted, "user should have been deleted from export")
	expression.AssertEqual(t, len(worker.Cancel), 0, "worker cancel channel should not have been filled")
	expression.AssertEqual(t, wrapper.DeleteExport(item.Id, users.User{Login: "bla"}), DeleteExportCanceled, "item should have been deleted")
	expression.AssertEqual(t, len(worker.Cancel), 1, "worker cancel channel should have been filled")

	// clean cancel channel (non-blocking)
	worker.DrainCancelChannel()
	worker.Mutex.Lock()
	worker.QueueItem.Users = []string{"bla", "blabla"}
	worker.Mutex.Unlock()
	expression.AssertEqual(t, wrapper.DeleteExport(item.Id, users.User{Login: "bla"}), DeleteExportNotFound, "user should have been deleted from existing export")
	expression.AssertEqual(t, len(worker.Cancel), 0, "worker cancel channel should not have been filled")
}

func TestWrapper_GetUserExport(t *testing.T) {
	wrapper := NewWrapper("/tmp", 1, 1, 2)
	item := NewWrapperItem([]engine.Fact{{ID: 1}}, "test.txt", CSVParameters{}, users.User{Login: "bla"}, make(map[string]string), false)

	// test item in archive
	wrapper.archive.Store(item.Id, *item)
	export, ok := wrapper.GetUserExport(item.Id, users.User{Login: "bla"})
	expression.AssertEqual(t, ok, true)
	expression.AssertEqual(t, export.Id, item.Id)
	export, ok = wrapper.GetUserExport(item.Id, users.User{Login: "blabla"})
	expression.AssertEqual(t, ok, false)
	wrapper.archive.Delete(item.Id)

	// test item in queue queue
	queueItem, code := wrapper.AddToQueue([]engine.Fact{{ID: 1}}, "test.txt", CSVParameters{}, users.User{Login: "bla"}, make(map[string]string), false)
	expression.AssertEqual(t, code, CodeAdded, "item should have been added to queue")
	export, ok = wrapper.GetUserExport(queueItem.Id, users.User{Login: "bla"})
	expression.AssertEqual(t, ok, true)
	expression.AssertEqual(t, export.Id, queueItem.Id)
	_, ok = wrapper.GetUserExport(queueItem.Id, users.User{Login: "blabla"})
	expression.AssertEqual(t, ok, false)
	_, ok = wrapper.dequeueWrapperItem(&export)
	expression.AssertEqual(t, ok, true)

	// test worker
	worker := NewExportWorker(0, "/tmp", make(chan<- int))
	wrapper.workers = append(wrapper.workers, worker)
	worker.Mutex.Lock()
	worker.QueueItem = *item
	worker.Available = false
	worker.Mutex.Unlock()

	_, ok = wrapper.GetUserExport(item.Id, users.User{Login: "blabla"})
	expression.AssertEqual(t, ok, false)
	_, ok = wrapper.GetUserExport(item.Id, users.User{Login: "bla"})
	expression.AssertEqual(t, ok, true)
}
