package export

import (
	"github.com/myrteametrics/myrtea-sdk/v4/expression"
	"testing"
)

func TestNewExportWorker(t *testing.T) {
	worker := NewExportWorker("/tmp")
	expression.AssertEqual(t, worker.BasePath, "/tmp")
	expression.AssertEqual(t, worker.Available, true)
	expression.AssertEqual(t, worker.QueueItemId, "")
}
