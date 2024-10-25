package content

import "testing"

func Test_getQuoteSequence(t *testing.T) {
	var seq = []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10,
		11, 12, 13, 14, 15, 16, 17, 18, 19, 20,
		21, 22, 23, 24, 25, 26, 27, 28, 29, 30,
		31, 32, 33}

	for i := 1; i <= 366; i++ {
		got := getQuoteSequence(i, 33)

		if got <= 0 || got > 33 {
			t.Errorf("getQuoteSequence() = %v", got)
		}

		if got != seq[got-1] {
			t.Errorf("getQuoteSequence() = %v - %v", got, seq[got-1])
		}
	}
}
