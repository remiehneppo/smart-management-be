package utils

import "time"

func GetWeekTimestampCurrent(currentTimestamp int64) (from int64, end int64) {
	// Get the current date
	currentDate := time.Unix(currentTimestamp, 0)

	// Get the start of the week (Monday)
	startOfWeek := currentDate.AddDate(0, 0, -int(currentDate.Weekday()-1))

	// Get the end of the week (Sunday)
	endOfWeek := startOfWeek.AddDate(0, 0, 6)

	// Convert to timestamps
	from = startOfWeek.Unix()
	end = endOfWeek.Unix()

	return from, end
}
