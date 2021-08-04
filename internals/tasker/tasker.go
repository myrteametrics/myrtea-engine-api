package tasker

import (
	"sync"

	"go.uber.org/zap"
)

var (
	_globalTaskerMu sync.RWMutex
	_globalTasker   *Tasker
)

// T is used to access the global tasker singleton
func T() *Tasker {
	_globalTaskerMu.RLock()
	defer _globalTaskerMu.RUnlock()

	tasker := _globalTasker
	return tasker
}

// ReplaceGlobals affect a new tasker to the global manager singleton
func ReplaceGlobals(tasker *Tasker) func() {
	_globalTaskerMu.Lock()
	defer _globalTaskerMu.Unlock()

	prev := _globalTasker
	_globalTasker = tasker
	return func() { ReplaceGlobals(prev) }
}

//Tasker represents the actions router, it process the BRMS results and triggers the actions.
type Tasker struct {
	BatchReceiver chan []TaskBatch
	Close         chan struct{}
}

//NewTasker renders a new Tasker
func NewTasker() *Tasker {
	return &Tasker{
		BatchReceiver: make(chan []TaskBatch),
		Close:         make(chan struct{}),
	}
}

//GetBatch retrieve the current tasker batch
func (t *Tasker) GetBatch() []TaskBatch {
	return <-t.BatchReceiver
}

//StartBatchProcessor starts the go routines that will listen to all the incoming batchs
func (t *Tasker) StartBatchProcessor() {
	go func() {
		for {
			select {
			case batchs := <-t.BatchReceiver:
				if len(batchs) > 0 {
					zap.L().Info("Tasker started the batch processing")
					ApplyBatchs(batchs)
					zap.L().Info("Tasker Batch done")
				}
			case <-t.Close:
				return
			}
		}
	}()

}

//StopBatchProcessor stops the batchprocessor
func (t *Tasker) StopBatchProcessor() {
	zap.L().Info("Stopping batchProcessor...")
	zap.L().Info("Stopping batchProcessor...Done")

	t.Close <- struct{}{}
}
