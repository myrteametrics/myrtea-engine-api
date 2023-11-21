package export

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/security/users"
	"github.com/myrteametrics/myrtea-sdk/v4/expression"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewWrapper(t *testing.T) {
	wrapper := NewWrapper("/tmp", 1, 1, 1)
	expression.AssertEqual(t, wrapper.BasePath, "/tmp")
	expression.AssertEqual(t, wrapper.QueueMaxSize, 1)
	expression.AssertEqual(t, wrapper.DiskRetentionDays, 1)
	expression.AssertEqual(t, wrapper.QueueMaxSize, 1)
}

func TestNewWrapperItem(t *testing.T) {
	item := NewWrapperItem([]int64{1}, "test.txt", CSVParameters{}, users.User{Login: "test"})
	expression.AssertNotEqual(t, item.Id, "")
	expression.AssertEqual(t, Int64Equals(item.FactIDs, []int64{1}), true)
	expression.AssertEqual(t, item.Params.Equals(CSVParameters{}), true)
	expression.AssertEqual(t, item.Status, StatusPending)
	expression.AssertEqual(t, item.FileName, "test.txt")
	expression.AssertNotEqual(t, len(item.Users), 0)
	expression.AssertEqual(t, item.Users[0], "test")
}

func TestWrapperItem_ContainsFact(t *testing.T) {
	item := NewWrapperItem([]int64{1, 22, 33}, "test.txt", CSVParameters{}, users.User{Login: "test"})
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
	expression.AssertEqual(t, len(wrapper.Workers), 1)
	worker := wrapper.Workers[0]
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
	_, result := wrapper.AddToQueue([]int64{1}, "test.txt", csvParams, user1)
	expression.AssertEqual(t, result, CodeAdded, "AddToQueue should return CodeAdded")
	_, result = wrapper.AddToQueue([]int64{1}, "test.txt", csvParams, user1)
	expression.AssertEqual(t, result, CodeUserExists, "AddToQueue should return CodeUserExists")
	_, result = wrapper.AddToQueue([]int64{1}, "test.txt", csvParams, user2)
	expression.AssertEqual(t, result, CodeUserAdded, "AddToQueue should return CodeUserAdded")
	_, result = wrapper.AddToQueue([]int64{2}, "test.txt", csvParams, user2)
	expression.AssertEqual(t, result, CodeQueueFull, "AddToQueue should return CodeQueueFull")
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

	// create a file that is 2 days old
	file, err := os.CreateTemp(dname, "exportdispatcher")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	fileName := filepath.Base(file.Name())
	_ = file.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wrapper := NewWrapper(dname, 1, 1, 1)
	wrapper.Init(ctx)
	expression.AssertEqual(t, len(wrapper.Workers), 1)
	// sleep one second to let the goroutine start
	fmt.Println("Sleeping 1 second to let the goroutine start")
	time.Sleep(1 * time.Second)

	worker := wrapper.Workers[0]

	// check if the worker is available
	worker.Mutex.Lock()
	expression.AssertEqual(t, worker.Available, true)
	worker.Mutex.Unlock()

	// add a task to the queue and check if the task was added to queue
	user := users.User{Login: "test"}
	_, result := wrapper.AddToQueue([]int64{1}, fileName, CSVParameters{}, user)
	expression.AssertEqual(t, result, CodeAdded, "AddToQueue should return CodeAdded")
	wrapper.QueueItemsMutex.Lock()
	expression.AssertEqual(t, len(wrapper.QueueItems), 1)
	itemId := wrapper.QueueItems[0].Id
	wrapper.QueueItemsMutex.Unlock()

	// sleep another 5 seconds to let the goroutine handle the task
	fmt.Println("Sleeping 5 seconds to let the goroutine handle the task")
	time.Sleep(5 * time.Second)

	wrapper.QueueItemsMutex.Lock()
	expression.AssertEqual(t, len(wrapper.QueueItems), 0)
	wrapper.QueueItemsMutex.Unlock()

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
	err = wrapper.CheckForExpiredFiles()
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
	wrapper.Archive.Store(id1, WrapperItem{Date: time.Now().AddDate(0, 0, -2)})
	wrapper.Archive.Store(id2, WrapperItem{Date: goodDate})

	_, found := wrapper.Archive.Load(id1)
	expression.AssertEqual(t, found, true)
	_, found = wrapper.Archive.Load(id2)
	expression.AssertEqual(t, found, true)

	err = wrapper.CheckForExpiredFiles()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	_, found = wrapper.Archive.Load(id1)
	expression.AssertEqual(t, found, false)
	_, found = wrapper.Archive.Load(id2)
	expression.AssertEqual(t, found, true)
}

func TestWrapper_GetUserExports(t *testing.T) {
	wrapper := NewWrapper("/tmp", 1, 1, 2)
	user1 := users.User{Login: "bla"}
	user2 := users.User{Login: "blabla"}
	item1 := NewWrapperItem([]int64{1}, "test.txt", CSVParameters{}, user1)
	item2 := NewWrapperItem([]int64{2}, "test.txt", CSVParameters{}, user1)
	item3 := NewWrapperItem([]int64{3}, "test.txt", CSVParameters{}, user1)
	item4 := NewWrapperItem([]int64{4}, "test.txt", CSVParameters{}, user2)
	wrapper.Archive.Store(item1.Id, *item1)
	wrapper.Archive.Store(item2.Id, *item2)
	wrapper.Archive.Store(item3.Id, *item3)
	wrapper.Archive.Store(item4.Id, *item4)
	wrapper.AddToQueue([]int64{5}, "test.txt", CSVParameters{}, user1)
	wrapper.AddToQueue([]int64{6}, "test.txt", CSVParameters{}, user2)
	exports := wrapper.GetUserExports(user1)
	expression.AssertEqual(t, len(exports), 4)
	exports = wrapper.GetUserExports(user2)
	expression.AssertEqual(t, len(exports), 2)
}

func TestWrapper_DequeueWrapperItem(t *testing.T) {
	wrapper := NewWrapper("/tmp", 1, 1, 2)
	i, ok := wrapper.DequeueWrapperItem(&WrapperItem{})
	expression.AssertEqual(t, ok, false)
	expression.AssertEqual(t, i, 0)
	wrapper.AddToQueue([]int64{5}, "test.txt", CSVParameters{}, users.User{Login: "bla"})
	wrapper.AddToQueue([]int64{6}, "test.txt", CSVParameters{}, users.User{Login: "blabla"})

	expression.AssertEqual(t, len(wrapper.QueueItems), 2)
	item1 := wrapper.QueueItems[0]
	item2 := wrapper.QueueItems[1]

	i, ok = wrapper.DequeueWrapperItem(item1)
	expression.AssertEqual(t, ok, true)
	expression.AssertEqual(t, i, 1)

	i, ok = wrapper.DequeueWrapperItem(item2)
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
	file, err := os.CreateTemp(dname, "exportdispatcher")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	fileName := filepath.Base(file.Name())
	_ = file.Close()

	wrapper := NewWrapper(dname, 1, 1, 2)
	ctx, cancel := context.WithCancel(context.Background())
	wrapper.Init(ctx)
	cancel() // stop dispatcher since we don't want him to interact with the workers or the queue

	// wait until dispatcher stops
	time.Sleep(50 * time.Millisecond)

	expression.AssertEqual(t, len(wrapper.Workers), 1)
	worker := wrapper.Workers[0]

	// no items in queue -> nothing should happen
	expression.AssertEqual(t, worker.IsAvailable(), true)
	wrapper.dispatchExportQueue(context.Background())
	expression.AssertEqual(t, worker.IsAvailable(), true, "worker should still be available, because no items in queue")

	// we add an item to the queue
	wrapper.AddToQueue([]int64{1}, fileName, CSVParameters{}, users.User{Login: "test"})

	// we test if dispatchExportQueue will not dispatch the item, no worker available
	worker.SwapAvailable(false)

	wrapper.dispatchExportQueue(context.Background())

	// the item should still be in the queue
	wrapper.QueueItemsMutex.Lock()
	expression.AssertEqual(t, len(wrapper.QueueItems), 1, "item should still be in the queue, since no worker is available")
	wrapper.QueueItemsMutex.Unlock()

	// we test if dispatchExportQueue will dispatch the item, worker is now set to available
	expression.AssertEqual(t, worker.SwapAvailable(true), false)

	wrapper.dispatchExportQueue(context.Background())

	expression.AssertEqual(t, worker.IsAvailable(), false, "worker should not be available, because it is working on an item")
	expression.AssertEqual(t, len(wrapper.QueueItems), 0)

	// wait until worker has finished
	time.Sleep(1 * time.Second)

	worker.Mutex.Lock()
	defer worker.Mutex.Unlock()

	expression.AssertEqual(t, worker.QueueItem.Status, StatusError, fmt.Sprintf("worker processed item should have StatusError(%d) because the file already exists", StatusError)) // could not create file
}

func TestWrapper_FindArchive(t *testing.T) {
	wrapper := NewWrapper("/tmp", 1, 1, 2)
	item := NewWrapperItem([]int64{1}, "test.txt", CSVParameters{}, users.User{Login: "bla"})
	wrapper.Archive.Store(item.Id, *item)

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
	item := NewWrapperItem([]int64{1}, "test.txt", CSVParameters{}, users.User{Login: "bla"})
	expression.AssertEqual(t, item.ContainsUser(users.User{Login: "bla"}), true)
	expression.AssertEqual(t, item.ContainsUser(users.User{Login: "blabla"}), false)
}
