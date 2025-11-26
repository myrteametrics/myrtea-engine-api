package coordinator

import (
	"testing"
	"time"
)

func TestGenerateNextIndexName(t *testing.T) {
	tests := []struct {
		name              string
		logicalIndexName  string
		existingIndices   []string
		now               time.Time
		expectedIndexName string
	}{
		{
			name:              "First index of the month",
			logicalIndexName:  "myrtea-myindex",
			existingIndices:   []string{},
			now:               time.Date(2025, 11, 26, 0, 0, 0, 0, time.UTC),
			expectedIndexName: "myrtea-myindex-2025-11-0001",
		},
		{
			name:             "Second index of the month",
			logicalIndexName: "myrtea-myindex",
			existingIndices: []string{
				"myrtea-myindex-2025-11-0001",
			},
			now:               time.Date(2025, 11, 26, 0, 0, 0, 0, time.UTC),
			expectedIndexName: "myrtea-myindex-2025-11-0002",
		},
		{
			name:             "Multiple indices - should increment from highest",
			logicalIndexName: "myrtea-myindex",
			existingIndices: []string{
				"myrtea-myindex-2025-11-0001",
				"myrtea-myindex-2025-11-0002",
				"myrtea-myindex-2025-11-0003",
			},
			now:               time.Date(2025, 11, 26, 0, 0, 0, 0, time.UTC),
			expectedIndexName: "myrtea-myindex-2025-11-0004",
		},
		{
			name:             "Mixed months - should only count current month",
			logicalIndexName: "myrtea-myindex",
			existingIndices: []string{
				"myrtea-myindex-2025-10-0001",
				"myrtea-myindex-2025-10-0002",
				"myrtea-myindex-2025-11-0001",
			},
			now:               time.Date(2025, 11, 26, 0, 0, 0, 0, time.UTC),
			expectedIndexName: "myrtea-myindex-2025-11-0002",
		},
		{
			name:             "New month - should reset to 0001",
			logicalIndexName: "myrtea-myindex",
			existingIndices: []string{
				"myrtea-myindex-2025-10-0001",
				"myrtea-myindex-2025-10-0002",
			},
			now:               time.Date(2025, 11, 26, 0, 0, 0, 0, time.UTC),
			expectedIndexName: "myrtea-myindex-2025-11-0001",
		},
		{
			name:             "Handles high sequence numbers",
			logicalIndexName: "myrtea-myindex",
			existingIndices: []string{
				"myrtea-myindex-2025-11-0099",
			},
			now:               time.Date(2025, 11, 26, 0, 0, 0, 0, time.UTC),
			expectedIndexName: "myrtea-myindex-2025-11-0100",
		},
		{
			name:             "Ignores old format indices",
			logicalIndexName: "myrtea-myindex",
			existingIndices: []string{
				"myrtea-myindex-active-000001",
				"myrtea-myindex-active-000002",
				"myrtea-myindex-2025-11-0001",
			},
			now:               time.Date(2025, 11, 26, 0, 0, 0, 0, time.UTC),
			expectedIndexName: "myrtea-myindex-2025-11-0002",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateNextIndexName(tt.logicalIndexName, tt.existingIndices, tt.now)
			if result != tt.expectedIndexName {
				t.Errorf("generateNextIndexName() = %v, want %v", result, tt.expectedIndexName)
			}
		})
	}
}
