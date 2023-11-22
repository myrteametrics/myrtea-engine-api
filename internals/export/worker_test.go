package export

import (
	"github.com/myrteametrics/myrtea-sdk/v4/expression"
	"testing"
)

func TestNewExportWorker(t *testing.T) {
	worker := NewExportWorker(0, "/tmp", make(chan<- int))
	expression.AssertEqual(t, worker.BasePath, "/tmp")
	expression.AssertEqual(t, worker.Available, true)
	expression.AssertEqual(t, worker.Id, 0)
}

func TestExportWorker_SetError(t *testing.T) {
	worker := NewExportWorker(0, "/tmp", make(chan<- int))
	worker.SetError(nil)
	expression.AssertEqual(t, worker.QueueItem.Status, StatusError)
	expression.AssertEqual(t, worker.QueueItem.Error, nil)
}

func TestExportWorker_SetStatus(t *testing.T) {
	worker := NewExportWorker(0, "/tmp", make(chan<- int))
	worker.SetStatus(StatusPending)
	expression.AssertEqual(t, worker.QueueItem.Status, StatusPending)
}

func TestExportWorker_SwapAvailable(t *testing.T) {
	worker := NewExportWorker(0, "/tmp", make(chan<- int))
	expression.AssertEqual(t, worker.SwapAvailable(false), true)
	expression.AssertEqual(t, worker.Available, false)
	expression.AssertEqual(t, worker.SwapAvailable(true), false)
	expression.AssertEqual(t, worker.Available, true)
}

func TestExportWorker_IsAvailable(t *testing.T) {
	worker := NewExportWorker(0, "/tmp", make(chan<- int))
	expression.AssertEqual(t, worker.IsAvailable(), true)
	worker.SwapAvailable(false)
	expression.AssertEqual(t, worker.IsAvailable(), false)
}

func TestExportWorker_DrainCancelChannel(t *testing.T) {
	worker := NewExportWorker(0, "/tmp", make(chan<- int))
	worker.Cancel <- true
	worker.DrainCancelChannel()
	expression.AssertEqual(t, len(worker.Cancel), 0)
}
