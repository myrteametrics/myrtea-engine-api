package tests

import (
	"flag"
	"testing"

	"go.uber.org/zap"
)

var debug = flag.Bool("debug", false, "Enable debug mode for tests (using Zap)")

// CheckDebugLogs checks for the "debug" flags while starting a test and enable zap logging
func CheckDebugLogs(t *testing.T) {
	if debug != nil && *debug {
		logger, err := zap.NewDevelopment(zap.AddStacktrace(zap.ErrorLevel))
		if err != nil {
			t.Fatal(err)
		}
		defer logger.Sync()
		zap.ReplaceGlobals(logger)
	}
}
