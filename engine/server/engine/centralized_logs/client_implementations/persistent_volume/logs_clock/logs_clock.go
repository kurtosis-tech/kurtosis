package logs_clock

import (
	"time"
)

const (
	daysInWeek = 7
)

// This interface is for enabling unit testing for log operations that rely on time
// specifically log retention features
type LogsClock interface {
	Now() time.Time
}

type RealLogsClock struct {
}

func NewRealClock() *RealLogsClock {
	return &RealLogsClock{}
}

func (clock *RealLogsClock) Now() time.Time {
	return time.Now()
}

// week 00-52
// day 0-7
// hour 0-23
type MockLogsClock struct {
	year, week, day, hour int
}

func NewMockLogsClockPerDay(year, week, day int) *MockLogsClock {
	return &MockLogsClock{
		year: year,
		week: week,
		day:  day,
		hour: 0,
	}
}

func NewMockLogsClockPerHour(year, week, day, hour int) *MockLogsClock {
	return &MockLogsClock{
		year: year,
		week: week,
		day:  day,
		hour: hour,
	}
}

func (clock *MockLogsClock) Now() time.Time {
	// Create a time object for January 4th of the given year (ISO week 1 always includes January 4th).
	startOfYear := time.Date(clock.year, time.January, 4, clock.hour, 0, 0, 0, time.UTC)

	// Get the Monday of the first ISO week of the year
	isoYearStart := startOfYear.AddDate(0, 0, int(time.Monday-startOfYear.Weekday()))

	// Adjust for Sunday as day 0 in the tests (Go uses Sunday as the first day of the week, but ISO uses Monday).
	var dayToAdd int
	if clock.day == 0 {
		// If the test input day is 0 (Sunday), we need to handle it as the 7th day of the week.
		dayToAdd = 6
	} else {
		// Otherwise, shift the day back by 1 to align with ISO (Monday as 1, etc.).
		dayToAdd = clock.day - 1
	}

	// Calculate the number of days to add based on the week and adjusted day.
	daysToAdd := (clock.week-1)*daysInWeek + dayToAdd

	// Add the calculated days to the ISO week start and return the result.
	mockTime := isoYearStart.AddDate(0, 0, daysToAdd)

	return mockTime
}
