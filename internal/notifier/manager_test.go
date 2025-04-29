package notifier

import (
	"testing"
)

func TestNewClientManager(t *testing.T) {
	manager := NewClientManager()
	if manager == nil {
		t.Error("manager constructor returns nil")
	}
}
