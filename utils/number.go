package utils

import (
	"strconv"
)

func IsNumeric(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

func Abs(n int64) int64 {
	if n < 0 {
		return -n
	}
	return n
}
