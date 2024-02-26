package routerauth

import (
	"errors"

	"github.com/spf13/viper"
)

type AuthenticationMode struct {
	Mode string `json:"mode"`
}

func GetMode() (AuthenticationMode, error) {
	mode := viper.GetString("AUTHENTICATION_MODE")
	if mode == "" {
		return AuthenticationMode{}, errors.New("failed to retrieve AUTHENTICATION_MODE from configuration")
	}
	return AuthenticationMode{Mode: mode}, nil
}
