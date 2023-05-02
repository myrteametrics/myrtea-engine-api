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
