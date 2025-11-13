package utils

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func ToMilliseconds(min, sec, milisec int) int64 {
	return int64((min * 60000) + (sec * 1000) + (milisec * 10))
}

func FromMilliseconds(milliseconds int64) (int, int, int) {
	minutes := int(milliseconds / 60000)
	seconds := int((milliseconds % 60000) / 1000)
	milisec := int((milliseconds % 60000) % 1000)

	return minutes, seconds, milisec / 10
}

func MillisecondsFromText(value string) (int64, error) {
	value = strings.TrimSpace(value)
	minutes := 0
	if strings.Contains(value, ":") {
		parts := strings.Split(value, ":")

		if len(parts) != 2 {
			return 0, fmt.Errorf("invalid time format, expected M:SS.CS")
		}

		// Parse minutes
		var err error
		minutes, err = strconv.Atoi(parts[0])
		if err != nil {
			return 0, fmt.Errorf("invalid minutes: %v", err)
		}

		value = parts[1]
	}

	// Split seconds and centiseconds by decimal point
	secParts := strings.Split(value, ".")
	if len(secParts) != 2 {
		return 0, fmt.Errorf("invalid seconds format, expected SS.CS")
	}

	// Parse seconds
	seconds, err := strconv.Atoi(secParts[0])
	if err != nil {
		return 0, fmt.Errorf("invalid seconds: %v", err)
	}

	// Parse centiseconds
	centiseconds, err := strconv.Atoi(secParts[1])
	if err != nil {
		return 0, fmt.Errorf("invalid centiseconds: %v", err)
	}

	return ToMilliseconds(minutes, seconds, centiseconds), nil
}

func FormatMilliseconds(milliseconds int64) string {
	ms := milliseconds
	signal := ""
	if milliseconds < 0 {
		ms = -milliseconds
		signal = "-"
	}
	minute, sec, millisecond := FromMilliseconds(ms)
	return fmt.Sprintf("%s%s", signal, FormatTime(minute, sec, millisecond))
}

func FormatTime(min, sec, milisec int) string {
	return fmt.Sprintf("%s:%s.%s", fmt.Sprintf("%02d", min), fmt.Sprintf("%02d", sec), fmt.Sprintf("%02d", milisec))
}

func MonthName(month int64) string {
	if month < 1 || month > 12 {
		return ""
	}

	months := [...]string{
		"Jan", "Feb", "Mar", "Apr", "May", "Jun",
		"Jul", "Aug", "Sep", "Oct", "Nov", "Dec",
	}

	return months[month-1]
}

func DayOfTheYear() int {
	t := time.Now()
	return t.YearDay()
}
