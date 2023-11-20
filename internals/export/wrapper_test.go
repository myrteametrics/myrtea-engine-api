package export

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/security/users"
	"github.com/myrteametrics/myrtea-sdk/v4/expression"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewExportWrapper(t *testing.T) {
	wrapper := NewExportWrapper("/tmp", 1, 1)
	expression.AssertEqual(t, wrapper.BasePath, "/tmp")
	expression.AssertEqual(t, wrapper.QueueMaxSize, 1)
	expression.AssertEqual(t, wrapper.DiskRetentionDays, 1)
}

func TestNewExportWrapperItem(t *testing.T) {
	item := NewExportWrapperItem(1, CSVParameters{}, users.User{ID: uuid.New()})
	expression.AssertNotEqual(t, item.Data.Id, "")
	expression.AssertEqual(t, item.Data.FactID, int64(1))
	expression.AssertEqual(t, item.Data.Params.Equals(CSVParameters{}), true)
	expression.AssertEqual(t, item.Data.Status, StatusPending)
}

func TestExportWrapperItem_SetError(t *testing.T) {
	item := NewExportWrapperItem(1, CSVParameters{}, users.User{ID: uuid.New()})
	expression.AssertEqual(t, item.Data.Status, StatusPending)
	item.SetError(fmt.Errorf("error"))
	expression.AssertEqual(t, item.Data.Status, StatusError)
	expression.AssertNotEqual(t, item.Data.Error, nil)
}

func TestExportWrapperItem_SetStatus(t *testing.T) {
	item := NewExportWrapperItem(1, CSVParameters{}, users.User{ID: uuid.New()})
	expression.AssertEqual(t, item.Data.Status, StatusPending)
	item.SetStatus(StatusRunning)
	expression.AssertEqual(t, item.Data.Status, StatusRunning)
}

func TestAddToQueue(t *testing.T) {
	wrapper := NewExportWrapper("/tmp", 1, 1)
	user1 := users.User{ID: uuid.New()}
	user2 := users.User{ID: uuid.New()}
	csvParams := CSVParameters{}
	expression.AssertEqual(t, wrapper.AddToQueue(1, csvParams, user1), CodeAdded, "AddToQueue should return CodeAdded")
	expression.AssertEqual(t, wrapper.AddToQueue(1, csvParams, user1), CodeUserExists, "AddToQueue should return CodeUserExists")
	expression.AssertEqual(t, wrapper.AddToQueue(1, csvParams, user2), CodeUserAdded, "AddToQueue should return CodeUserAdded")
	expression.AssertEqual(t, wrapper.AddToQueue(2, csvParams, user2), CodeQueueFull, "AddToQueue should return CodeQueueFull")
}

func TestFindAvailableWorker(t *testing.T) {
	wrapper := NewExportWrapper("/tmp", 1, 1)
	// since wrapper.Init() starts the dispatcher worker that we don't want to run in this test, we initialize the workers manually
	for i := 0; i < 2; i++ {
		wrapper.Workers = append(wrapper.Workers, NewExportWorker("/tmp"))
	}
	w1 := wrapper.FindAvailableWorker()
	expression.AssertNotEqual(t, w1, nil)
	w2 := wrapper.FindAvailableWorker()
	expression.AssertNotEqual(t, w2, nil)
	w3 := wrapper.FindAvailableWorker()
	expression.AssertEqual(t, w3, (*ExportWorker)(nil))
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

	wrapper := NewExportWrapper(dname, 1, 1)
	wrapper.Init(1)
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
	expression.AssertEqual(t, wrapper.AddToQueue(1, CSVParameters{FileName: fileName}, users.User{ID: uuid.New()}), CodeAdded, "AddToQueue should return CodeAdded")
	wrapper.QueueMutex.Lock()
	item := wrapper.Queue[0]
	expression.AssertEqual(t, len(wrapper.Queue), 1)
	wrapper.QueueMutex.Unlock()

	// sleep another 5 seconds to let the goroutine handle the task
	fmt.Println("Sleeping 5 seconds to let the goroutine handle the task")
	time.Sleep(5 * time.Second)

	wrapper.QueueMutex.Lock()
	expression.AssertEqual(t, len(wrapper.Queue), 0)
	wrapper.QueueMutex.Unlock()

	wrapper.DoneMutex.Lock()
	expression.AssertEqual(t, len(wrapper.Done), 1)
	foundItem := wrapper.Done[0]
	wrapper.DoneMutex.Unlock()

	expression.AssertEqual(t, item.Data.Id, foundItem.Data.Id)

	fmt.Println("Sleeping 1 second to wait for status")
	time.Sleep(2 * time.Second)

	// could not create file
	foundItem.Mutex.Lock()
	expression.AssertEqual(t, foundItem.Data.Status, StatusError)
	foundItem.Mutex.Unlock()
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

	wrapper := NewExportWrapper(dname, 1, 1)
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
	wrapper.Done = append(wrapper.Done, &WrapperItem{Data: ExportWrapperItemData{Date: time.Now().AddDate(0, 0, -2)}})
	wrapper.Done = append(wrapper.Done, &WrapperItem{Data: ExportWrapperItemData{Date: goodDate}})
	expression.AssertEqual(t, len(wrapper.Done), 2)
	err = wrapper.CheckForExpiredFiles()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	expression.AssertEqual(t, len(wrapper.Done), 1)
	expression.AssertEqual(t, wrapper.Done[0].Data.Date, goodDate)
}
