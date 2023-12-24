package utils

import "testing"

func TestFormatMiliseconds(t *testing.T) {
	var tests = []struct {
		name        string
		miliseconds int64
		want        string
	}{
		{"miliseconds should be 96", 100960, "01:40:96"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatMiliseconds(tt.miliseconds)
			if got != tt.want {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}
