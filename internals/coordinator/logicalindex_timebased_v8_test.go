package coordinator

import (
	"reflect"
	"testing"
	"time"
)

func TestTimebasedV8FindIndices(t *testing.T) {
	// t time.Time,
	// depthDays int64

	logical := LogicalIndexTimeBasedV8{
		Name: "myproject",
		LiveIndices: []string{
			"myproject-2023-08",
			"myproject-2023-07",
			"myproject-2023-06",
			"myproject-2023-05",
			"myproject-2023-04",
			"myproject-2023-03",
			"myproject-2023-02",
			"myproject-2023-01",
		},
	}

	type TestCase struct {
		ti       time.Time
		depth    int64
		expected []string
	}

	for _, testCase := range []TestCase{
		{ti: time.Date(2023, time.August, 10, 12, 30, 0, 0, time.UTC), depth: 5,
			expected: []string{"myproject-2023-08"}},
		{ti: time.Date(2023, time.August, 10, 12, 30, 0, 0, time.UTC), depth: 15,
			expected: []string{"myproject-2023-08", "myproject-2023-07"}},
		{ti: time.Date(2023, time.August, 10, 12, 30, 0, 0, time.UTC), depth: 30,
			expected: []string{"myproject-2023-08", "myproject-2023-07"}},
		{ti: time.Date(2023, time.August, 10, 12, 30, 0, 0, time.UTC), depth: 39,
			expected: []string{"myproject-2023-08", "myproject-2023-07"}},
		{ti: time.Date(2023, time.August, 10, 12, 30, 0, 0, time.UTC), depth: 40,
			expected: []string{"myproject-2023-08", "myproject-2023-07", "myproject-2023-06"}},
		{ti: time.Date(2023, time.August, 10, 12, 30, 0, 0, time.UTC), depth: 45,
			expected: []string{"myproject-2023-08", "myproject-2023-07", "myproject-2023-06"}},
		{ti: time.Date(2023, time.August, 10, 12, 30, 0, 0, time.UTC), depth: 90,
			expected: []string{"myproject-2023-08", "myproject-2023-07", "myproject-2023-06", "myproject-2023-05"}},
	} {
		indices, _ := logical.FindIndices(testCase.ti, testCase.depth)

		if !reflect.DeepEqual(testCase.expected, indices) {
			t.Log(testCase.ti.Format("2006-01-02"),
				testCase.ti.Add(time.Duration(testCase.depth)*-1*24*time.Hour).Format("2006-01-02"))
			t.Error("invalid indices expected=", testCase.expected, ", received=", indices)
			t.Fail()
		}
	}
}
