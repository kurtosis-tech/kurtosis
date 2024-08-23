package logs_clock

import (
	"fmt"
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

type MockLogsClock struct {
	year, week, day, hour int
}

// week 00-52
// day 0-7
// hour 0-23
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
	// Start with the first day of the year
	firstDay := time.Date(clock.year, time.January, 1, 0, 0, 0, 0, time.UTC)

	// Adjust to the first ISO week of the year
	isoYear, isoWeek := firstDay.ISOWeek()
	if isoYear != clock.year {
		// The first day of the year might be in the last week of the previous year
		firstDay = firstDay.AddDate(0, 0, 8)
	}

	// Find the difference to the desired week
	days := (clock.week - isoWeek) * 7

	// Move to the desired day and hour
	targetDay := firstDay.AddDate(0, 0, days+(clock.day-1))
	targetTime := targetDay.Add(time.Duration(clock.hour) * time.Hour)

	// Adjust if needed to ensure the correct ISO week and day
	resultYear, resultWeek := targetTime.ISOWeek()
	if resultYear != clock.year || resultWeek != clock.week || int(targetTime.Weekday()) != clock.day || targetTime.Hour() != clock.hour {
		fmt.Printf("invalid year, week, day, hour combination")
	}

	return targetTime
}
