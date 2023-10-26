package situation

import "testing"

func TestGetTranslateValue(t *testing.T) {

	if getTranslateValue() != true {
		t.Error("Expected true for empty input, got true")
	}

	if getTranslateValue(false) != false {
		t.Error("Expected false for input of false, got true")
	}

	if getTranslateValue(true) != true {
		t.Error("Expected true for input of true, got false")
	}

	if getTranslateValue(true, false) != true {
		t.Error("Expected true for multiple input values, but function should only consider the first")
	}
}
