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

func NewMockLogsClock(year, week, day, hour int) *MockLogsClock {
	return &MockLogsClock{
		year: year,
		week: week,
		day:  day,
		hour: hour,
	}
}

// The mocked Now() function returns a time object representing the start of date specified by the year, week, and day
func (clock *MockLogsClock) Now() time.Time {
	// Create a time.Time object for January 1st of the given year
	startOfYear := time.Date(clock.year, time.January, 1, 0, 0, 0, 0, time.UTC)

	// Calculate the number of days to add to reach the start of the desired week.
	daysToAdd := time.Duration(clock.week * daysInWeek)

	// Calculate the start of the desired week by adding days to the start of the year.
	startOfWeek := startOfYear.Add(daysToAdd * 24 * time.Hour)

	// Adjust the start of the week to the beginning of the week (usually Sunday or Monday).
	startOfWeek = startOfWeek.Add(time.Duration(clock.day) * 24 * time.Hour)
	return startOfWeek
}
