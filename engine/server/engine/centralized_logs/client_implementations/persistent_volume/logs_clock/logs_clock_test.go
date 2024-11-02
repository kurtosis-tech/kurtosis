package logs_clock

import (
	"testing"
)

func TestMockLogsClockPerDay(t *testing.T) {
	tests := []struct {
		year        int
		week        int
		day         int
		description string
	}{
		{2024, 1, 0, "First day of ISO week 1 (Sunday)"},
		{2024, 1, 1, "Second day of ISO week 1 (Monday)"},
		{2024, 1, 2, "Third day of ISO week 1 (Tuesday)"},
		{2024, 1, 3, "Fourth day of ISO week 1 (Wednesday)"},
		{2024, 1, 4, "Fifth day of ISO week 1 (Thursday)"},
		{2024, 1, 5, "Sixth day of ISO week 1 (Friday)"},
		{2024, 1, 6, "Last day of ISO week 1 (Saturday)"},
		{2024, 52, 0, "First day of ISO week 52 (Sunday)"},
		{2024, 52, 1, "First day of ISO week 52 (Monday)"},
		{2024, 52, 5, "Fifth day of ISO week 52 (Friday)"},
		{2024, 52, 6, "Last day of ISO week 52 (Saturday)"},
		{2024, 48, 0, "First day of ISO week 48 (Sunday)"},
		{2024, 48, 4, "Fifth day of ISO week 48 (Thursday)"},
		{2024, 24, 2, "Third day of ISO week 24 (Tuesday)"},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			clock := NewMockLogsClockPerDay(test.year, test.week, test.day)
			result := clock.Now()

			// Get the ISO week and day from the result
			year, week := result.ISOWeek()
			day := int(result.Weekday())

			if year != test.year || week != test.week || day != test.day {
				t.Errorf("Expected (year: %d, week: %d, day: %d) but got (year: %d, week: %d, day: %d)",
					test.year, test.week, test.day, year, week, day)
			}
		})
	}
}

func TestMockLogsClockPerHour(t *testing.T) {
	tests := []struct {
		year        int
		week        int
		day         int
		hour        int
		description string
	}{
		{2024, 1, 0, 0, "First day of ISO week 1, hour 0 (Sunday)"},
		{2024, 1, 0, 12, "First day of ISO week 1, hour 12 (Sunday)"},
		{2024, 1, 1, 0, "Second day of ISO week 1, hour 0 (Monday)"},
		{2024, 1, 1, 6, "Second day of ISO week 1, hour 6 (Monday)"},
		{2024, 1, 2, 0, "Third day of ISO week 1, hour 0 (Tuesday)"},
		{2024, 1, 6, 23, "Last day of ISO week 1, hour 23 (Saturday)"},
		{2024, 52, 0, 0, "First day of ISO week 52, hour 0 (Sunday)"},
		{2024, 52, 1, 0, "First day of ISO week 52, hour 0 (Monday)"},
		{2024, 52, 5, 12, "Fifth day of ISO week 52, hour 12 (Friday)"},
		{2024, 52, 6, 0, "Last day of ISO week 52, hour 0 (Saturday)"},
		{2024, 48, 0, 0, "First day of ISO week 48, hour 0 (Sunday)"},
		{2024, 48, 4, 18, "Fifth day of ISO week 48, hour 18 (Thursday)"},
		{2024, 24, 2, 15, "Third day of ISO week 24, hour 15 (Tuesday)"},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			clock := NewMockLogsClockPerHour(test.year, test.week, test.day, test.hour)
			result := clock.Now()

			// Get the ISO week, day, and hour from the result
			year, week := result.ISOWeek()
			day := int(result.Weekday())
			hour := result.Hour()

			if year != test.year || week != test.week || day != test.day || hour != test.hour {
				t.Errorf("Expected (year: %d, week: %d, day: %d, hour: %d) but got (year: %d, week: %d, day: %d, hour: %d)",
					test.year, test.week, test.day, test.hour, year, week, day, hour)
			}
		})
	}
}
