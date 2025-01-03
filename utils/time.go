package utils

import (
	"fmt"
	"time"
)

func ToMiliseconds(min, sec, milisec int) int64 {
	return int64((min * 60000) + (sec * 1000) + (milisec * 10))
}

func FromMiliseconds(milliseconds int64) (int, int, int) {
	min := int(milliseconds / 60000)
	sec := int((milliseconds % 60000) / 1000)
	milisec := int(((milliseconds % 60000) % 1000))

	return min, sec, (milisec / 10)
}

func FormatMiliseconds(milliseconds int64) string {
	min, sec, milisec := FromMiliseconds(milliseconds)
	return FormatTime(min, sec, milisec)
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
