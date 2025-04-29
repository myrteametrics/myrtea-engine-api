package notifier

import (
	"testing"
)

func TestNewNotifier(t *testing.T) {
	notifier := NewNotifier()
	if notifier == nil {
		t.Error("notifier constructor returns nil")
	}
}

func TestReplaceGlobal(t *testing.T) {
	notifier1 := NewNotifier()
	notifier2 := C()
	if notifier1 == notifier2 {
		t.Error("Global notifier is weirdly defined")
	}

	ReplaceGlobals(notifier1)
	notifier2 = C()
	if notifier1 != notifier2 {
		t.Error("Global notifier is not a singleton")
	}
}
