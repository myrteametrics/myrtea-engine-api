package email

import (
	"os"
	"testing"
)

// InitUnitTest initialize the global Email sender singleton in unite test mode
func TestInitSender_and_S(t *testing.T) {

    // Initializing the values ​​for the test 
	host := os.Getenv("TEST_SMTP_HOST")
	port := os.Getenv("TEST_SMTP_PORT")
	username := os.Getenv("TEST_SMTP_USERNAME")
	password := os.Getenv("TEST_SMTP_PASSWORD")

    // Call INitSender function
	InitSender(username,password,host,port)

	// Sender Instance Recovery
	sender := S()

	// Checking if the Sender instance is non-zero
	if sender == nil{
		t.Error("Expected a non-nil Sender instancer, but got nil")
	}

	// Recovery of a second instance of the Sender
	sender2 := S()

	// Checking if the two instances are the same (sngleton )
	if sender != sender2 {
		t.Errorf("Expected the same Sender instance, but got different instances")
	}
    
}

