package persistent_volume_helpers

import "time"

const (
	hoursInWeek = 7 * 24
)

func ConvertWeeksToDuration(weeks int) time.Duration {
	return time.Duration(weeks*hoursInWeek) * time.Hour
}
