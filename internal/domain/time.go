package domain

import (
	"fmt"
	"time"
)

const (
	TimeLayout    = "15:04:05"
	SecondsInHour = 60 * 60
	SecondsInDay  = 24 * SecondsInHour
)

func ParseClock(value string) (int, error) {
	if len(value) != len("00:00:00") {
		return 0, fmt.Errorf("time must have HH:MM:SS format: %q", value)
	}

	parsed, err := time.Parse(TimeLayout, value)
	if err != nil {
		return 0, fmt.Errorf("time must have HH:MM:SS format: %w", err)
	}

	return parsed.Hour()*SecondsInHour + parsed.Minute()*60 + parsed.Second(), nil
}

func FormatClock(seconds int) string {
	seconds %= SecondsInDay
	if seconds < 0 {
		seconds += SecondsInDay
	}
	return fmt.Sprintf("%02d:%02d:%02d", seconds/SecondsInHour, seconds%SecondsInHour/60, seconds%60)
}

func FormatDuration(seconds int) string {
	if seconds < 0 {
		seconds = 0
	}
	return fmt.Sprintf("%02d:%02d:%02d", seconds/SecondsInHour, seconds%SecondsInHour/60, seconds%60)
}
