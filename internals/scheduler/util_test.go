package scheduler

import (
	"testing"
	"time"
)

func TestParseDuration(t *testing.T) {
	// Test cases
	testCases := []struct {
		input     string
		want      time.Duration
		shouldErr bool
	}{
		{"7d 3h 35m 3s", time.Hour*(7*24+3) + time.Minute*35 + time.Second*3, false},
		{"2d 12h 30m 45s", time.Hour*(2*24+12) + time.Minute*30 + time.Second*45, false},
		{"72h", time.Hour * 72, false},
		{"15d 5h", time.Hour*365 + time.Minute*0 + time.Second*0, false},
		{"invalid_string", 0, true},
	}

	// Run tests
	for _, tc := range testCases {
		got, err := parseDuration(tc.input)

		if tc.shouldErr && err == nil {
			t.Errorf("parseDuration(%q) should return an error", tc.input)
		} else if !tc.shouldErr && err != nil {
			t.Errorf("parseDuration(%q) returned an unexpected error: %v", tc.input, err)
		} else if !tc.shouldErr && got != tc.want {
			t.Errorf("parseDuration(%q) = %v, want %v", tc.input, got, tc.want)
		}
	}
}

func TestGenerateKeyAndValues(t *testing.T) {

	situation1 := map[string]string{
		IDSituationDependsOn: "123",
		IDInstanceDependsOn:  "456",
	}
	key, id1, id2, err := generateKeyAndValues(situation1)
	if err != nil {
		t.Fatalf("Expected no error, but got: %v", err)
	}
	if key != "123-456" {
		t.Errorf("Expected key to be 123-456, but got: %s", key)
	}
	if id1 != 123 || id2 != 456 {
		t.Errorf("Expected ids to be 123 and 456, but got: %d and %d", id1, id2)
	}

	situation2 := map[string]string{}
	_, _, _, err = generateKeyAndValues(situation2)
	if err == nil {
		t.Error("Expected an error due to missing keys, but got none")
	}

	situation3 := map[string]string{
		IDSituationDependsOn: "abc",
		IDInstanceDependsOn:  "456",
	}
	_, _, _, err = generateKeyAndValues(situation3)
	if err == nil {
		t.Error("Expected an error due to non-int value, but got none")
	}
}
