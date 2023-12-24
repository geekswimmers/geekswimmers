package utils

import "fmt"

func ToMiliseconds(min int, sec int, milisec int) int64 {
	if milisec < 100 {
		milisec = milisec * 10
	}

	return int64((min * 60000) + (sec * 1000) + milisec)
}

func DecomposeMiliseconds(miliseconds int64) (int, int, int) {
	min := int(miliseconds / 60000)
	sec := int((miliseconds % 60000) / 1000)
	milisec := int(((miliseconds % 60000) % 1000))

	return min, sec, milisec
}

func FormatMiliseconds(miliseconds int64) string {
	min, sec, milisec := DecomposeMiliseconds(miliseconds)

	if milisec >= 100 {
		milisec = milisec / 10
	}

	return fmt.Sprintf("%s:%s:%s", fmt.Sprintf("%02d", min), fmt.Sprintf("%02d", sec), fmt.Sprintf("%02d", milisec))
}
