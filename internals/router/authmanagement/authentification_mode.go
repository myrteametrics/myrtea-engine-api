package authmanagement

import (
	"errors"
	"sync"

	"github.com/spf13/viper"
)

type AuthenticationMode struct {
	Mode string `json:"mode"`
}

type AuthenticationModeQuerier struct{}

var (
	_globalQuerierMu sync.RWMutex
	_globalQuerier   *AuthenticationModeQuerier
)

func New() *AuthenticationModeQuerier {
	querier := &AuthenticationModeQuerier{}
	return querier
}

func S() *AuthenticationModeQuerier {
	_globalQuerierMu.RLock()
	defer _globalQuerierMu.RUnlock()

	return _globalQuerier
}

func ReplaceGlobals(querier *AuthenticationModeQuerier) func() {
	_globalQuerierMu.Lock()
	defer _globalQuerierMu.Unlock()

	prev := _globalQuerier
	_globalQuerier = querier

	return func() {
		ReplaceGlobals(prev)
	}
}

func (querier AuthenticationModeQuerier) GetMode() (AuthenticationMode, error) {

	mode := viper.GetString("AUTHENTICATION_MODE")
	if mode == "" {
		return AuthenticationMode{}, errors.New("failed to retrieve AUTHENTICATION_MODE from configuration")
	}
	return AuthenticationMode{Mode: mode}, nil
}
