package utils

import "fmt"

func ToMiliseconds(min, sec, milisec int) int64 {
	return int64((min * 60000) + (sec * 1000) + (milisec * 10))
}

func FromMiliseconds(miliseconds int64) (int, int, int) {
	min := int(miliseconds / 60000)
	sec := int((miliseconds % 60000) / 1000)
	milisec := int(((miliseconds % 60000) % 1000))

	return min, sec, (milisec / 10)
}

func FormatMiliseconds(miliseconds int64) string {
	min, sec, milisec := FromMiliseconds(miliseconds)
	return FormatTime(min, sec, milisec)
}

func FormatTime(min, sec, milisec int) string {
	return fmt.Sprintf("%s:%s:%s", fmt.Sprintf("%02d", min), fmt.Sprintf("%02d", sec), fmt.Sprintf("%02d", milisec))
}
