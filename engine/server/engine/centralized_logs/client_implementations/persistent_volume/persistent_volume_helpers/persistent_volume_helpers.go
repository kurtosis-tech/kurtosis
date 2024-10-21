package persistent_volume_helpers

import "time"

func ConvertWeeksToDuration(weeks int) time.Duration {
	hoursInWeek := 7 * 24
	return time.Duration(weeks*hoursInWeek) * time.Hour
}
