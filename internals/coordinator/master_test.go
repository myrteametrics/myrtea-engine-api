package coordinator

import "testing"

func TestMasterSingleton(t *testing.T) {
	instance := GetInstance()
	instance2 := GetInstance()

	if instance != instance2 {
		t.Errorf("Two call on GetInstance() do not return a singleton")
	}
}
