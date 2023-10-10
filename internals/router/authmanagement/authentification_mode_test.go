package authmanagement

import (
	"sync"
	"testing"

	"github.com/spf13/viper"
)

func TestSetModeIntegration(t *testing.T) {

	querier := New()

	viper.Set("AUTHENTICATION_MODE", "BASIC")

	mode, err := querier.GetMode()
	if err != nil {
		t.Errorf("error was not expected while selecting mode: %s", err)
	}

	if mode.Mode != "BASIC" {
		t.Errorf("Authentification mod BASIC expected but get %s", mode.Mode)
	}

}

func TestConcurrentAccess(t *testing.T) {

	querier := New()
	restore := ReplaceGlobals(querier)
	defer restore()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			q := S()
			q.GetMode()

		}()
	}
	wg.Wait()
}
