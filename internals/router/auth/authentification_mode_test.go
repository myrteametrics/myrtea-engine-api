package routerauth

import (
	"testing"

	"github.com/spf13/viper"
)

func TestSetModeIntegration(t *testing.T) {

	viper.Set("AUTHENTICATION_MODE", "BASIC")

	mode, err := GetMode()
	if err != nil {
		t.Errorf("error was not expected while selecting mode: %s", err)
	}

	if mode.Mode != "BASIC" {
		t.Errorf("Authentification mod BASIC expected but get %s", mode.Mode)
	}

}
