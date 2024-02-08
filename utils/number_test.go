package utils

import (
	"math"
	"testing"
)

func TestAbs(t *testing.T) {

	tests := []struct {
		input int64
		want  int64
	}{
		{10, 10},
		{-10, 10},
		{0, 0},
		{math.MaxInt64, math.MaxInt64},
		{math.MinInt64 + 1, math.MaxInt64},
	}

	for _, tt := range tests {
		got := Abs(tt.input)
		if got != tt.want {
			t.Errorf("Abs(%d) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestIsNumeric(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"123", true},
		{"123.45", true},
		{"-123.45", true},
		{"abc", false},
		{"", false},
	}

	for _, tt := range tests {
		got := IsNumeric(tt.input)
		if got != tt.want {
			t.Errorf("IsNumeric(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}
