package calendar

import (
	"sync"
)

var (
	_globalMu    sync.RWMutex
	_globalCBase *Base
)

// CBase is used to access the global calendar base
func CBase() *Base {
	_globalMu.RLock()
	defer _globalMu.RUnlock()

	cbase := _globalCBase
	return cbase
}

// Init initialize the global calendar base singleton
func Init() {
	_globalMu.Lock()
	defer _globalMu.Unlock()

	_globalCBase = NewCalendarBase()
	_globalCBase.Update()
}

// InitUnitTest initialize the global calendar base singleton in unite test mode
func InitUnitTest() {
	_globalMu.Lock()
	defer _globalMu.Unlock()

	_globalCBase = NewCalendarBase()
}
