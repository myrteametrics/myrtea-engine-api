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
	cbase := _globalCBase
	_globalMu.RUnlock()
	return cbase
}

// Init initialize the global calendar base singleton
func Init() {
	_globalMu.Lock()
	_globalCBase = NewCalendarBase()
	_globalCBase.Update()
	_globalMu.Unlock()
}

// InitUnitTest initialize the global calendar base singleton in unite test mode
func InitUnitTest() {
	_globalMu.Lock()
	_globalCBase = NewCalendarBase()
	_globalMu.Unlock()
}
