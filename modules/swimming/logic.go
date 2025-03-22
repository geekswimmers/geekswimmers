package swimming

import "math"

func CalculateSwimmingPoints(baseTime int64, time int64) int {
	if time == 0 {
		return 0
	}

	if time < baseTime {
		return 0
	}

	division := float64(baseTime) / float64(time)
	power := math.Pow(division, float64(3))
	points := int(power * 1000)
	return points
}
