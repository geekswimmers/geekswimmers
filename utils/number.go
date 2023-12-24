package utils

import (
	"math"
	"strconv"
)

func IsNumeric(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

func Abs(n int64) int64 {
	return int64(math.Abs(float64(n)))
}
