package persistent_volume_helpers

import (
	"testing"
	"time"
)

func TestConvertWeeksToDuration(t *testing.T) {
	tests := []struct {
		name     string
		weeks    int
		expected time.Duration
	}{
		{
			name:     "Zero weeks",
			weeks:    0,
			expected: 0,
		},
		{
			name:     "One week",
			weeks:    1,
			expected: 7 * 24 * time.Hour,
		},
		{
			name:     "Two weeks",
			weeks:    2,
			expected: 2 * 7 * 24 * time.Hour,
		},
		{
			name:     "Negative weeks",
			weeks:    -1,
			expected: -7 * 24 * time.Hour,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertWeeksToDuration(tt.weeks)
			if result != tt.expected {
				t.Errorf("ConvertWeeksToDuration(%d) = %v; want %v", tt.weeks, result, tt.expected)
			}
		})
	}
}
