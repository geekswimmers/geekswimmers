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

func TestFormatTime(t *testing.T) {

	// Happy path
	min, sec, milisec := 1, 2, 3
	want := "01:02:03"
	got := FormatTime(min, sec, milisec)
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	// Edge cases
	cases := []struct {
		min, sec, milisec int
		want              string
	}{
		{0, 0, 0, "00:00:00"},
		{59, 59, 99, "59:59:99"},
	}

	for _, c := range cases {
		got := FormatTime(c.min, c.sec, c.milisec)
		if got != c.want {
			t.Errorf("got %q, want %q", got, c.want)
		}
	}
}
func TestMonthName(t *testing.T) {
	tests := []struct {
		month int64
		want  string
	}{
		{1, "Jan"},
		{2, "Feb"},
		{12, "Dec"},
		{0, ""},
		{-1, ""},
		{13, ""},
	}

	for _, tt := range tests {
		got := MonthName(tt.month)
		if got != tt.want {
			t.Errorf("MonthName(%d) = %q, want %q", tt.month, got, tt.want)
		}
	}
}
