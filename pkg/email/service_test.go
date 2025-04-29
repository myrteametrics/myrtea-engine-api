package email

import (
	"os"
	"testing"
)

// InitUnitTest initialize the global Email sender singleton in unite test mode
func TestInitSender(t *testing.T) {
	host := os.Getenv("TEST_SMTP_HOST")
	port := os.Getenv("TEST_SMTP_PORT")
	username := os.Getenv("TEST_SMTP_USERNAME")
	password := os.Getenv("TEST_SMTP_PASSWORD")

	InitSender(username, password, host, port)

	sender := S()
	if sender == nil {
		t.Error("Expected a non-nil Sender instancer, but got nil")
	}

	sender2 := S()
	if sender != sender2 {
		t.Error("Expected the same Sender instance, but got different instances")
	}
}
