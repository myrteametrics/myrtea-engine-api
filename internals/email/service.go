package email

import (
	"os"
	"sync"

	"github.com/spf13/viper"
)

var (
	_globalMu    sync.RWMutex
	_globalSender *Sender
)

// S is used to access the global Email Sender 
func S() *Sender {
	_globalMu.RLock()
	defer _globalMu.RUnlock()

	c_sender := _globalSender
	return  c_sender
}

// Init initialize the global Email Sender singleton
func Init() {
	_globalMu.Lock()
	defer _globalMu.Unlock()

	username := viper.GetString("AUTHENTICATION_SMTP_USERNAME")
	password := viper.GetString("AUTHENTICATION_SMTP_PASSWORD")
	host := viper.GetString("AUTHENTICATION_SMTP_HOST")
	port := viper.GetString("AUTHENTICATION_SMTP_PORT")
	_globalSender = NewSender(username, password, host, port)
}

// InitUnitTest initialize the global Email sender singleton in unite test mode
func InitUnitTestSenderSend() {
	_globalMu.Lock()
	defer _globalMu.Unlock()

	host := os.Getenv("TEST_SMTP_HOST")
	port := os.Getenv("TEST_SMTP_PORT")
	username := os.Getenv("TEST_SMTP_USERNAME")
	password := os.Getenv("TEST_SMTP_PASSWORD")

	_globalSender = NewSender(username, password, host, port)
}

// InitUnitTest initialize the global Email sender for gmail  singleton in unite test mode
func InitUnitTestSenderGmail(){
     
	_globalMu.Lock()
	defer _globalMu.Unlock()

	username := os.Getenv("TEST_SMTP_USERNAME")
	password := os.Getenv("TEST_SMTP_PASSWORD")
	host := "smtp.gmail.com"
	port := "587"

	_globalSender = NewSender(username, password, host, port)
}

