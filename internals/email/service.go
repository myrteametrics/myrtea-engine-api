package email

import (
	"sync"
)

var (
	_globalMu     sync.RWMutex
	_globalSender *Sender
)

// S is used to access the global Email Sender
func S() *Sender {
	_globalMu.RLock()
	defer _globalMu.RUnlock()
	c_sender := _globalSender
	return c_sender
}

// Init initialize the global Email Sender singleton
func InitSender(username string, password string, host string, port string) {
	_globalMu.Lock()
	defer _globalMu.Unlock()
	_globalSender = NewSender(username, password, host, port)
}
